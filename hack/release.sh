#!/usr/bin/env bash

set -e
[ -z "$DEBUG" ] || set -x;

usage() {
  echo "$0 <repo> <tag> [<release name>] [-- <asset>...]" >&2;
}

if [ "$1" = "-h" -o "$1" = "--help" ]; then
  usage
  cat >&2 <<EOS
Pass the following arguments:
	* \`<repo>\`: ":user/:name" of the repository. For example, "foca/mpp".
	* \`<tag>\`: Name of the tag for this release. For example, "v1.0.0".
	* \`<release name>\`: Optional suffix for the release name. (disabled arg at present)
You can pass a list of files to upload as release assets by giving them after a
\`--\` argument.
If you supply text on \`STDIN\` it will be used as the release notes.
EXAMPLES:
	$ $0 foca/mpp v1.0.0 -- pkg/*.tar.gz
	Creates a release named "mpp v1.0.0" and adds any tar.gz file in
	\`./pkg\` as an asset.
	$ $0 foca/mpp v1.0.1 "Bugfixes" -- pkg/*.tar.gz < notes.md
	Creates a release named "mpp v1.0.1: Bugfixes", adds any tar.gz
	file in \`./pkg\` as an asset, and uses the contents of \`notes.md\`
	as the release notes.
NOTE:
This uses your \`.netrc\` file to authenticate with GitHub. In order to run the
script, make sure you have **both** \`api.github.com\` and \`upload.github.com\` in
this file. For example:
machine api.github.com
  login foca
  password <an access token>
machine uploads.github.com
  login foca
  password <an access token>
Generate this access token at https://github.com/settings/tokens and make sure
it has access to the \`"repo"\` scope.
EOS
  exit 1;
fi

[ -n "$2" ] || (usage; exit 1);

REPO="$1"
shift

TAG="$1"
shift

NAME="${TAG}"
# Note: release name is disabled to have only version name
# if [ -n "$1" -a "$1" != "--" ]; then
#   NAME="${NAME}: $1";
#   shift
# fi

BODY=""
[ -t 0 ] || BODY=`cat`;

if [ "$1" = "--" -a "$#" -ge "2" ]; then
  shift
  ASSETS="$@"
fi

payload=$(
  jq --null-input \
     --arg tag "$TAG" \
     --arg name "$NAME" \
     --arg body "$BODY" \
     '{ tag_name: $tag, name: $name, body: $body, draft: true }'
)

echo "creating new tag $TAG"
git tag $TAG

echo "building release assets"
make all-container
cp -f bin/darwin_amd64/sqsmv bin/sqsmv_darwin_amd64
cp -f bin/linux_amd64/sqsmv bin/sqsmv_linux_amd64

echo "creating draft release"
response=$(
  curl --fail \
       --netrc \
       --silent \
       --location \
       --data "$payload" \
       "https://api.github.com/repos/${REPO}/releases"
)

upload_url="$(echo "$response" | jq -r .upload_url | sed -e "s/{?name,label}//")"

for file in $ASSETS; do
  echo "uploading asset $file"
  curl --netrc \
       --header "Content-Type:application/gzip" \
       --data-binary "@$file" \
       "$upload_url?name=$(basename "$file")"
done
