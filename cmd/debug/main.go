package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	gqlEndpoint = "https://gql.twitch.tv/gql"
	clientID    = "kimne78kx3ncx6brgo4mv6wki5h1ko"

	// Хэш нужно брать актуальный.
	// Если этот не сработает, возьми новый из DevTools.
	sha256Hash = "ed230aa1e33e07eebb8928504583da78a5173989fadfb1ac94be06a04f3cdbe9"
)

type PlaybackTokenResponse struct {
	Data struct {
		VideoPlaybackAccessToken *struct {
			Value     string `json:"value"`
			Signature string `json:"signature"`
		} `json:"videoPlaybackAccessToken"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func checkVOD(vodID string) (string, error) {
	payload := map[string]interface{}{
		"operationName": "PlaybackAccessToken",
		"variables": map[string]interface{}{
			"isLive":     false,
			"login":      "",
			"isVod":      true,
			"vodID":      vodID,
			"playerType": "site",
			"platform":   "web",
		},
		"extensions": map[string]interface{}{
			"persistedQuery": map[string]interface{}{
				"version":    1,
				"sha256Hash": sha256Hash,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", gqlEndpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Client-ID", clientID)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp PlaybackTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("bad json: %w", err)
	}

	if len(tokenResp.Errors) > 0 {
		return "", fmt.Errorf("graphql error: %s", tokenResp.Errors[0].Message)
	}

	token := tokenResp.Data.VideoPlaybackAccessToken
	if token == nil || token.Value == "" || token.Signature == "" {
		return "", fmt.Errorf("vod недоступен или удалён")
	}

	m3u8URL := fmt.Sprintf(
		"https://usher.ttvnw.net/vod/%s?allow_source=true&allow_audio_only=true&nauthsig=%s&nauth=%s",
		vodID,
		token.Signature,
		url.QueryEscape(token.Value),
	)

	return m3u8URL, nil
}

func main() {

	vodID := "2012345678"

	m3u8URL, err := checkVOD(vodID)
	if err != nil {
		fmt.Printf("VOD %s: %v\n", vodID, err)
		os.Exit(1)
	}

	fmt.Println("m3u8 найден:")
	fmt.Println(m3u8URL)
}
