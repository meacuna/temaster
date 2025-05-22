package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/meacuna/temaster/internal/spotify"
)

func openSpotifyLink(url string) error {
	script := fmt.Sprintf(`
		tell application "Spotify"
			play track "%s"
		end tell`, url)

	return exec.Command("osascript", "-e", script).Start()
}

func main() {
	// Check if running on macOS
	if runtime.GOOS != "darwin" {
		fmt.Println("This program only works on macOS")
		os.Exit(1)
	}

	// Get Spotify credentials from environment
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		fmt.Println("Error: SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET environment variables must be set")
		os.Exit(1)
	}

	fmt.Print("Enter Spotify playlist URL: ")
	var playlistURL string
	if _, err := fmt.Scanln(&playlistURL); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	if playlistURL == "" {
		fmt.Println("No playlist URL provided")
		os.Exit(1)
	}

	client := spotify.NewClient(clientID, clientSecret)

	tracks, err := client.GetPlaylistTracks(playlistURL)
	if err != nil {
		fmt.Printf("Error getting playlist tracks: %v\n", err)
		os.Exit(1)
	}

	if len(tracks) == 0 {
		fmt.Println("No tracks found in the playlist")
		os.Exit(1)
	}

	fmt.Printf("This playlist has %d songs\n", len(tracks))

	// Create a map to track played songs
	playedTracks := make(map[string]bool)
	remainingTracks := len(tracks)

	fmt.Println("\nPress Enter to play a random song, or type 'exit' to quit")

	for {
		var input string
		fmt.Scanln(&input)

		if input == "exit" {
			fmt.Println("Goodbye!")
			return
		}

		if remainingTracks == 0 {
			fmt.Println("No more songs to play!")
			return
		}

		// Get a random track that hasn't been played yet
		var track string
		for {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(tracks))))
			if err != nil {
				fmt.Printf("Error generating random number: %v\n", err)
				os.Exit(1)
			}
			randomIndex := int(n.Int64())
			track = tracks[randomIndex]
			if !playedTracks[track] {
				playedTracks[track] = true
				remainingTracks--
				break
			}
		}

		fmt.Printf("---- Song %d of %d ----\n", len(playedTracks), len(tracks))

		// Get track info
		trackInfo, err := client.GetTrackInfo(track)
		if err != nil {
			fmt.Printf("Failed to get track info: %v\n", err)
			continue
		}

		// Convert the HTTPS URL to a spotify: URI before opening
		spotifyURI := spotify.ConvertToSpotifyURI(track)
		if err := openSpotifyLink(spotifyURI); err != nil {
			fmt.Printf("Failed to open Spotify: %v\n", err)
			continue
		}

		fmt.Println("\nPress Enter to show song info, or type 'exit' to quit")
		fmt.Scanln(&input)

		if input == "exit" {
			fmt.Println("Goodbye!")
			return
		}

		fmt.Printf("\nNow playing: %s by %s (%s)\n",
			trackInfo.Name,
			strings.Join(trackInfo.Artists, ", "),
			trackInfo.Year)

		fmt.Println("Press Enter to play another song, or type 'exit' to quit")
	}
}
