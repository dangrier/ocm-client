# https://just.systems

set lazy
versionTag := "-X 'github.com/dangrier/ocm-client/cli.Version=$(git describe --tags --always --long --dirty 2>/dev/null)'"
buildDate := "-X 'github.com/dangrier/ocm-client/cli.BuildDate=" + `date -u '+%Y-%m-%dT%H:%M:%SZ'` + "'"
ldFlags := versionTag + " " + buildDate

default: clean build test

clean:
    rm -rf out/

build: clean
    go build -ldflags="{{ldFlags}}" -o out/ocm ./cmd/...

install:
    go install ./cmd/...

test:
    go test -v ./...

release-build: clean
    GOOS=darwin GOARCH=arm64 go build -ldflags="{{ldFlags}}" -o out/release/ocm-darwin-arm64 ./cmd/...
    GOOS=linux GOARCH=arm64 go build -ldflags="{{ldFlags}}" -o out/release/ocm-linux-arm64 ./cmd/...
    GOOS=linux GOARCH=amd64 go build -ldflags="{{ldFlags}}" -o out/release/ocm-linux-amd64 ./cmd/...
    GOOS=windows GOARCH=amd64 go build -ldflags="{{ldFlags}}" -o out/release/ocm-windows-amd64.exe ./cmd/...
