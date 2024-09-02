#!/bin/sh

path=$(pwd)
# shellcheck disable=SC2044
for f in $(find "$path" -name '*.proto')
do
    dir=$(dirname "$f")
    cd "${dir}" || exit
    file=$(basename "$f")
    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative "$file";
    cd "${path}" || exit
  done
