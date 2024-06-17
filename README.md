# Video Creater

This bot will create a quote video every time the main function runs. The bot is not perfect and can fail in the creation of the video.

## Needed for working bot

**Need to get:**
* **Google cloud** json key. This key should be in the root folder and have the name client_secret.json.
* **.env** this file should be in the root folder and have:
    * GOOGLE_API_KEY=""
    * PEXELS_API_KEY=""
    * FAVQS_API_KEY=""

**Need to download**
* ffmpeg - https://ffmpeg.org/download.html or ```sudo apt install ffmpeg```

**Note** The *first* time the bot runs, you will get a link in the terminal. Follow that link and confirm whats needed to make the bot able to upload to youtube. This will create a token.json file.

## How to use the bot

The file *global/variables.go* is a makeshift control panel. Here you can change the variables based on how you want the outputed video. Note that the variables at the bottom of the screen should not be changed.

When you are happy with the *settings* you may run the bot. in your terminal change your directory to the root of the prodject. There you can run the command ```go run main.go``` If you want to build the the prodject in to an executable file. Then run the command ```go build -o X``` where X is the name you want the built prodject to have. However, the code loads some env variables. This will have to be manualy set or you can create a script to load the env variables then run the executable. Example code:
```sh
#!/bin/bash

# Source the environment variables
source /home/heier/videoCreater/.env

# Run the executable
/home/heier/videoCreater/videomaker

```
Remember to: ```chmod +x script.sh```

This script can be anywhere and still work as intended.

### Cross platfor compile

If you want to compile the program on windows, then run it on linux. Then you will have to set sett these env variables before you run the build command.
Powershell:
``` powershell
$env:GOOS="linux"
$env:GOARCH="amd64"
```