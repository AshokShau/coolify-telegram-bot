package coolify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func uploadToSpacebin(content string) ([]string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	const maxChunkSize = 400_000 // 400 KB
	var chunks []string
	log.Println("Uploading to spaceb.in...", len(content))
	for start := 0; start < len(content); start += maxChunkSize {
		end := start + maxChunkSize
		if end > len(content) {
			end = len(content)
		}
		chunks = append(chunks, content[start:end])
	}

	client := &http.Client{Timeout: 15 * time.Second}
	var urls []string

	for _, chunk := range chunks {
		data := map[string]string{"content": chunk}
		jsonBody, _ := json.Marshal(data)

		req, err := http.NewRequest("POST", "https://spaceb.in/api/", bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "Go-http-client/1.1")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("HTTP request error: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("bad status: %s\nResponse: %s", resp.Status, body)
		}

		var result struct {
			Error   string `json:"error"`
			Payload struct {
				ID string `json:"id"`
			} `json:"payload"`
		}

		if err = json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("error parsing JSON: %v\nResponse body: %s", err, body)
		}

		if result.Error != "" {
			return nil, fmt.Errorf("spaceb.in upload failed: %s", result.Error)
		}

		urls = append(urls, fmt.Sprintf("https://spaceb.in/%s", result.Payload.ID))
		time.Sleep(500 * time.Millisecond)
	}

	return urls, nil
}
