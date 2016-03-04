#!/bin/bash
set -e
#linux build
pushd ./src/renderservice/renderer/mupdf
  rm -rf ./build/release
  rm -rf ./build/release
  make build=release HAVE_X11=no
popd
go get github.com/Sirupsen/logrus
go get gopkg.in/tylerb/graceful.v1
go get github.com/gorilla/mux
env GOOS=linux GOARCH=amd64 go build -o renderer -x -a -i -v --ldflags '-extldflags "-static"' renderservice
#windows build
pushd ./src/renderservice/renderer/mupdf
  rm -rf ./build/release
  rm -rf ./build/debug
  make OS=w64_amd64-cross-mingw32 build=release HAVE_X11=no
popd
# have to use -Wl,--allow-multiple-definition to work around https://github.com/golang/go/issues/9510
env CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o renderer.exe -buildmode=exe -x -a -i -v --ldflags '-extld=x86_64-w64-mingw32-gcc -extldflags "-static -Wl,--allow-multiple-definition"' renderservice
