

# Compiling
## Linux

## Requirements
    `sudo apt install build-essential`
    `sudo apt-get install pkg-config`
### Build
    `CGO_ENABLED=1 go build .`

# Running
## Dependencies
    1. yt-dlp
    2. ffmpeg
    3. deno (for yt-dlp JS runtime) *optional

*yt-dlp and deno can be downloaded by the app

### FFMPEG
[FFMPEG](https://www.ffmpeg.org/download.html)

### YT-DLP

[yt-dlp installation guid](https://github.com/yt-dlp/yt-dlp?tab=readme-ov-file#installation)

### DENO
macOS | Linux 

`curl -fsSL https://deno.land/install.sh | sh`
    
Windows (Power Shell) 

`irm https://deno.land/install.ps1 | iex`

# CONFIG

`
max_queue_size=500
sample_rate=44100
music_folders=C:/music;C:/musicb
recursive_search=true
`

### MUSIC FOLDERS
Add the folders separated with a comma (;) for music search into the system

Also add `recursive_search` if your folder has levels
`
C:/music
    /typeA
    /typeB
`