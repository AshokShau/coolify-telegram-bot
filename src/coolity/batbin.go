package coolify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func uploadToBatbin(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", fmt.Errorf("content cannot be empty")
	}

	req, err := http.NewRequest("POST", "https://batbin.me/api/v2/paste", bytes.NewBufferString(content))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "text/plain; charset=UTF-8")
	req.Header.Set("User-Agent", "Go-http-client/1.1")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s\nResponse: %s", resp.Status, body)
	}

	var result struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error parsing JSON: %v\nResponse body: %s", err, body)
	}

	return fmt.Sprintf("https://batbin.me/%s", result.Message), nil
}
