#!/bin/bash

GIT_REPO_DIR="/home/heier/videoCreater"

cd "$GIT_REPO_DIR" || exit 1

current_head=$(git rev-parse HEAD)
git fetch origin main
git reset --hard FETCH_HEAD
new_head=$(git rev-parse HEAD)

if [ "$current_head" != "$new_head" ]; then
    go build -o videomaker main.go
    chmod +x videomaker
fi

chmod +x gitPull.sh