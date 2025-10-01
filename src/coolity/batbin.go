package coolify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// max chunk size
const maxChunkSize = 500_00

func uploadToBatbin(content string) ([]string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}
	var chunks []string
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
		req, err := http.NewRequest("POST", "https://batbin.me/api/v2/paste", bytes.NewBufferString(chunk))
		if err != nil {
			return nil, fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("Content-Type", "text/plain; charset=UTF-8")
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
			Success bool   `json:"success"`
			Message string `json:"message"`
		}

		if err = json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("error parsing JSON: %v\nResponse body: %s", err, body)
		}

		if !result.Success {
			return nil, fmt.Errorf("batbin upload failed: %s", result.Message)
		}

		urls = append(urls, fmt.Sprintf("https://batbin.me/%s", result.Message))
	}

	return urls, nil
}
