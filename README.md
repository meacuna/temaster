# Temaster

A command-line tool that turns any [Spotify](https://open.spotify.com/) playlist into a music guessing game, inspired by [Hitster](https://hitstergame.com/). Test your music knowledge by guessing songs from your favorite playlists!

## Features

- 🎵 Randomly plays songs from any Spotify playlist
- 🎮 Interactive command-line interface
- 🎯 Displays song information after each guess
- 🎧 Directly integrates with Spotify for playback

## Requirements

- macOS operating system
- Spotify desktop app installed
- Spotify API credentials (Client ID and Client Secret)
- [Go](https://go.dev/) installed (version 1.16 or higher) - only needed if building from source

## Installation

### Option 1: Download the executable
1. Download the latest release from the [Releases](https://github.com/yourusername/temaster/releases) page
2. Make the file executable:
   ```bash
   chmod +x temaster
   ```
3. Move it to your PATH (optional):
   ```bash
   mv temaster /usr/local/bin/
   ```

### Option 2: Build from source
1. Clone this repository
2. Build the project:
   ```bash
   go build -o temaster
   ```
3. Move the executable to your PATH (optional):
   ```bash
   mv temaster /usr/local/bin/
   ```

## Setup

1. Set up your Spotify API credentials:
   ```bash
   export SPOTIFY_CLIENT_ID="your_client_id"
   export SPOTIFY_CLIENT_SECRET="your_client_secret"
   ```
2. Run the program:
   ```bash
   temaster
   ```

## How to Play

1. Enter a Spotify playlist URL when prompted
2. Try to guess the song and artist
3. Reveal the song information
4. Repeat!

## License

MIT License
