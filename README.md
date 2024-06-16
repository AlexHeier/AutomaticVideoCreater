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