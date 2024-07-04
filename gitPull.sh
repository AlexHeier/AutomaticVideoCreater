#!/bin/bash

shutdown -r now

GIT_REPO_DIR="/home/heier/videoCreater"

cd "$GIT_REPO_DIR" || exit 1

current_head=$(git rev-parse HEAD)
git fetch origin main
git reset --hard FETCH_HEAD
new_head=$(git rev-parse HEAD)

if [ "$current_head" != "$new_head" ]; then
    # Update the go.mod file to ensure the Go version is in the format 1.x
    sed -i -E 's/^go ([0-9]+\.[0-9]+)\.[0-9]+$/go \1/' go.mod

    go build -o videomaker main.go
    chmod +x videomaker
fi

chmod +x gitPull.sh
