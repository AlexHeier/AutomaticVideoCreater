# Video Creator

This bot will create a video and if wanted, post it to TikTok or YouTube.

## How to call the code

The program takes in one to two calling arguments. Based on the video you want to make. The syntax is:
```go run main.go <Video type> <Specification>```

```<Video type>``` is quote or reddit. Based on if you want a reddit or quote video.
```<Specification>``` is the subreddit for the video.

Examples:

Quote video: ```go run main.go quote```
    Creates a random quote video.

Reddit video: ```go run main.go reddit aitah```
    Creates a video about the last post posted to r/AITAH

## How to use the bot

The file *global/variables.go* is a makeshift control panel. Here you can change the variables based on how you want the output video. Note that the variables at the bottom of the file should not be changed.

When you are happy with the *settings* you may run the bot. in your terminal change your directory to the root of the project. There you can run the command ```go run main.go <Video type> <Specification>``` If you want to build the project in to an executable file. Then run the command ```go build -o X``` where X is the name you want the built project to have. However, the code needs some env variables. This will have to be manually set, or you can create a script to load the env variables then run the executable. 

Example code:
```sh

#!/bin/bash

export PEXELS_API_KEY='VALUE'
export YOUTUBE_API_KEY='VALUE'

export FAVQS_API_KEY='VALUE'
export GOOGLE_API_KEY='VALUE'
export UNREAL_SPEECH_API_KEY='VALUE'

export REDDIT_USER_AGENT='VALUE'
export TIKTOK_API_KEY='VALUE'

# Run the executable
/home/heier/videoCreater/videomaker

```
Remember to ```chmod +x script.sh``` if on Linux.

This script can be anywhere and still works as intended.

## Needed for working bot

**Need to get:**
* **Google cloud** json key. This key should be in the root folder and have the name client_secret.json.

* **.env** this file should be in the root folder and have:
    * PEXELS_API_KEY
    * YOUTUBE_API_KEY
    * FAVQS_API_KEY
    * GOOGLE_API_KEY
    * UNREAL_SPEECH_API_KEY
    * REDDIT_USER_AGENT
    * TIKTOK_API_KEY


**Need to download**
* ffmpeg - https://ffmpeg.org/download.html or ```sudo apt install ffmpeg```

**Note** The *first* time the bot runs, you will get a link in the terminal. Follow that link and confirm what is needed to make the bot able to upload to YouTube. This will create a token.json file.

### Cross platform compile

If you want to compile the program on windows, then run it on Linux. Then you will have to set these env variables before you run the build command.
Powershell:
``` powershell
$env:GOOS="linux"
$env:GOARCH="amd64"
```

