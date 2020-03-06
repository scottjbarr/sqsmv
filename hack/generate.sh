#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

unameOut="$(uname -s)"
case "${unameOut}" in
    Linux*)     machine=Linux;;
    Darwin*)    machine=Mac;;
    CYGWIN*)    machine=Cygwin;;
    MINGW*)     machine=MinGw;;
    *)          machine="UNKNOWN:${unameOut}"
esac

if [ $# -eq 0 ]
  then
    echo "no arguments supplied, exiting"
    exit 1
fi

if [ -z "$1" ]; then
    echo "no template supplied, exiting"
    exit 1
fi

if [ -z "${VERSION}" ]; then
    echo "VERSION not set, exiting"
    exit 1
fi
export SQSMV_VERSION=${VERSION}

echo "sourcing environment variables: ./.envs.sh"
source ./.envs.sh

BIN='SQSMV'
while IFS='=' read -r name value ; do
  if [[ $name == *"${BIN}_"* ]]; then
    if [ "$machine" = "Mac" ]; then
        sed -i '' -e "s#{{ ${name} }}#${!name}#g" $1
    else
        sed -i "s#{{ ${name} }}#${!name}#g" $1
    fi
  fi
done < <(env)

if [ "$machine" = "Mac" ]; then
    sed -i '' -e "s#{{ ${BIN}_TAG }}#${VERSION}#g" $1
else
    sed -i "s#{{ ${BIN}_TAG }}#${VERSION}#g" $1
fi
