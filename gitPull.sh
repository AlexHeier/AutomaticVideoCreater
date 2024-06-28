#!/bin/bash

GIT_REPO_DIR="/home/heier/videoCreater"

cd "$GIT_REPO_DIR" || exit 1


UPDATE_STATUS=$(git pull origin main)


if [[ $UPDATE_STATUS != *"Already up to date."* ]]; then
    
    go build -o videomaker main.go

    chmod +x videomaker

fi

# */5 * * * * /home/heier/videoCreater/gitPull.sh <-- Crontab selution
