#!/bin/bash

function buildFunctions() {
    for d in ./cmd/functions/* ; do (cd "$d" && go get . && GOOS=linux GOARCH=arm64 go build -o ./bootstrap && chmod +x bootstrap); done
}

while getopts s: flag
do
    case "${flag}" in
        s) stack=${OPTARG};;
    esac
done

if [ -z "$stack" ]; then
    echo 'Missing -s flag' >&2
    exit 1
fi

buildFunctions 

pushd ./cmd/infrastructure &> /dev/null
    go get .
    pulumi up -s $stack -f
popd &> /dev/null
