# command to build and run on the local OS.
GO_BUILD = go build

# command to compiling the distributable. Specify GOOS and GOARCH for
# the target OS.
GO_DIST = CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO_BUILD) -a -tags netgo -ldflags '-w'

APP = sqsmv
TMP_BUILD = tmp/$(APP)
DIST_DIR = dist

TAG ?= `git describe --tags | sed -e 's/^v//'`

.PHONY: dist

all: clean tools lint goimports vet dist

deps:
	go get -t ./...

prepare:
	mkdir -p build dist

tools:
	go get golang.org/x/tools/cmd/goimports
	go get github.com/golang/lint/golint

lint: golint vet goimports vet

vet:
	go vet

golint:
	ret=0 && test -z "$$(golint . | tee /dev/stderr)" || ret=1 ; exit $$ret

goimports:
	ret=0 && test -z "$$(goimports -l . | tee /dev/stderr)" || ret=1 ; exit $$ret

dist: prepare dist-linux dist-darwin dist-windows

dist-linux:
	GOOS=linux GOARCH=amd64 go build -o $(TMP_BUILD)
	tar -C tmp -zcvf $(DIST_DIR)/$(APP)-$(TAG)-linux.gz $(APP)
	rm $(TMP_BUILD)

dist-darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(TMP_BUILD)
	tar -C tmp -zcvf $(DIST_DIR)/$(APP)-$(TAG)-darwin.gz $(APP)
	rm $(TMP_BUILD)

dist-windows:
	GOOS=windows GOARCH=amd64 go build -o $(TMP_BUILD)
	tar -C tmp -zcvf $(DIST_DIR)/$(APP)-$(TAG)-windows.gz $(APP)
	rm $(TMP_BUILD)

clean:
	rm -rf build dist
