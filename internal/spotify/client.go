package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	authURL       = "https://accounts.spotify.com/api/token"
	apiURL        = "https://api.spotify.com/v1"
	playlistURL   = "https://api.spotify.com/v1/playlists/%s/tracks?limit=%d&offset=%d"
	spotifyWebURL = "https://open.spotify.com/"
	trackURL      = "https://api.spotify.com/v1/tracks/%s"
	// Timeout for HTTP requests
	httpTimeout = 45 * time.Second
)

// Client is a Spotify client
type Client struct {
	clientID     string
	clientSecret string
	accessToken  string
	httpClient   *http.Client
}

// NewClient creates a new Spotify client
func NewClient(clientID, clientSecret string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// TrackInfo represents the information we want to get about a track
type TrackInfo struct {
	Name    string
	Artists []string
	Year    string
}

// GetAccessToken retrieves an access token using Client Credentials Flow
func (c *Client) refreshAccessToken() error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", authURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("error parsing token response: %w", err)
	}

	token, ok := result["access_token"].(string)
	if !ok {
		return fmt.Errorf("failed to retrieve access token")
	}
	c.accessToken = token
	return nil
}

// GetPlaylistTracks fetches all track links from a given Spotify playlist ID
func (c *Client) GetPlaylistTracks(rawURL string) ([]string, error) {
	if !strings.HasPrefix(rawURL, spotifyWebURL) {
		return nil, fmt.Errorf("invalid Spotify URL format")
	}

	if c.accessToken == "" {
		if err := c.refreshAccessToken(); err != nil {
			return nil, fmt.Errorf("failed to refresh access token: %w", err)
		}
	}
	playlistID := c.GetPlaylistID(rawURL)
	if playlistID == "" {
		return nil, fmt.Errorf("invalid playlist ID")
	}

	trackLinks := []string{}
	offset := 0
	limit := 100 // Spotify's maximum limit per request

	for {
		apiURL := fmt.Sprintf(playlistURL, playlistID, limit, offset)
		req, err := http.NewRequest(http.MethodGet, apiURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.accessToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get playlist: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("error parsing playlist response: %w", err)
		}

		total, _ := result["total"].(float64)

		if items, exists := result["items"].([]interface{}); exists {
			for _, item := range items {
				if track, ok := item.(map[string]interface{})["track"].(map[string]interface{}); ok {
					if url, exists := track["external_urls"].(map[string]interface{})["spotify"].(string); exists {
						trackLinks = append(trackLinks, url)
					}
				}
			}
		}

		if float64(len(trackLinks)) >= total {
			break
		}

		offset += limit
	}

	return trackLinks, nil
}

func (c *Client) GetPlaylistID(playlistURL string) string {
	parts := strings.Split(playlistURL, "/")
	playlistID := parts[len(parts)-1]
	playlistID = strings.Split(playlistID, "?")[0] // Remove query params if any
	return playlistID
}

// ConvertToSpotifyURI converts a Spotify HTTP URL to a Spotify URI
// eg: https://open.spotify.com/track/xxx to spotify:track:xxx:play
func ConvertToSpotifyURI(url string) string {
	if strings.HasPrefix(url, spotifyWebURL) {
		path := strings.TrimPrefix(url, spotifyWebURL)
		parts := strings.Split(path, "/")
		if len(parts) >= 2 {
			trackID := strings.Split(parts[1], "?")[0]
			return fmt.Sprintf("spotify:%s:%s:play", parts[0], trackID)
		}
	}
	return url
}

// GetTrackInfo fetches detailed information about a track
func (c *Client) GetTrackInfo(rawURL string) (*TrackInfo, error) {
	if c.accessToken == "" {
		if err := c.refreshAccessToken(); err != nil {
			log.Fatalf("Failed to refresh access token: %v", err)
			return nil, err
		}
	}

	// Convert URL to track ID
	trackID := c.GetTrackID(rawURL)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(trackURL, trackID), nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get track: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Error parsing track response: %v", err)
		return nil, err
	}

	// Extract track name
	name, _ := result["name"].(string)

	// Extract artists
	var artists []string
	if artistsData, exists := result["artists"].([]interface{}); exists {
		for _, artist := range artistsData {
			if artistMap, ok := artist.(map[string]interface{}); ok {
				if artistName, exists := artistMap["name"].(string); exists {
					artists = append(artists, artistName)
				}
			}
		}
	}

	// Extract release date
	var year string
	if album, exists := result["album"].(map[string]interface{}); exists {
		if releaseDate, exists := album["release_date"].(string); exists {
			// Spotify returns dates in YYYY-MM-DD format, we'll just take the year
			year = strings.Split(releaseDate, "-")[0]
		}
	}

	return &TrackInfo{
		Name:    name,
		Artists: artists,
		Year:    year,
	}, nil
}

// GetTrackID extracts the track ID from a Spotify URL
func (c *Client) GetTrackID(trackURL string) string {
	parts := strings.Split(trackURL, "/")
	trackID := parts[len(parts)-1]
	trackID = strings.Split(trackID, "?")[0] // Remove query params if any
	return trackID
}
