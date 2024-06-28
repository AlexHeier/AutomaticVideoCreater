#!/bin/bash

GIT_REPO_DIR="/home/heier/videoCreater"

cd "$GIT_REPO_DIR" || exit 1


git fetch origin main


git checkout FETCH_HEAD -- .


if ! git diff --quiet; then
    go build -o videomaker main.go
    chmod +x videomaker
fi
