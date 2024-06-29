#!/bin/bash

GIT_REPO_DIR="/home/heier/videoCreater"

cd "$GIT_REPO_DIR" || exit 1

current_head=$(git rev-parse HEAD)
git fetch origin main
git reset --hard FETCH_HEAD
new_head=$(git rev-parse HEAD)

if [ "$current_head" != "$new_head" ]; then

    # Retrieve the Go version installed on the server
    go_version=$(go version | awk '{print $3}')
    
    # Update the go.mod file to use the server's Go version
    sed -i "s/^go [0-9]\+\.[0-9]\+\.[0-9]\+$/go ${go_version#go}/" go.mod

    go build -o videomaker main.go
    chmod +x videomaker
fi

chmod +x gitPull.sh
