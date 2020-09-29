#!/usr/bin/env bash

rm -rf ../dist/*
package_name=dfs
platforms=("darwin/386" "darwin/amd64"  "linux/386" "linux/amd64" "linux/arm64" "windows/amd64" "windows/386")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=../dist/$package_name'-'$GOOS'-'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o $output_name ../cmd/dfs
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
    zip $output_name.zip $output_name
    rm $output_name
done
