package coolify

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

func (c *Client) ListApplications() ([]Application, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/applications", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("unauthenticated: invalid or missing token (401)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, errors.New("invalid token (400)")
	}

	var apps []Application
	err = json.NewDecoder(resp.Body).Decode(&apps)
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (c *Client) GetApplicationByUUID(uuid string) (*ApplicationDetail, error) {
	url := fmt.Sprintf("%s/api/v1/applications/%s", c.BaseURL, uuid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("unauthenticated: invalid or missing token (401)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, errors.New("invalid token (400)")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("application not found")
	}

	var app ApplicationDetail
	err = json.NewDecoder(resp.Body).Decode(&app)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (c *Client) DeleteApplicationByUUID(uuid string) error {
	url := fmt.Sprintf("%s/api/v1/applications/%s", c.BaseURL, uuid)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.New("unauthenticated: invalid or missing token (401)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return errors.New("invalid token (400)")
	}
	if resp.StatusCode == http.StatusNotFound {
		return errors.New("application not found")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected response: %s", resp.Status)
	}

	return nil
}

func (c *Client) GetApplicationLogsByUUID(uuid string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/applications/%s/logs?lines=-1", c.BaseURL, uuid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("unauthenticated: invalid or missing token (401)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return "", errors.New("invalid token (400)")
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New("application not found")
	}

	var result ApplicationLogs
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return uploadToBatbin(result.Logs)
}

func (c *Client) GetApplicationEnvsByUUID(uuid string) ([]EnvironmentVariable, error) {
	url := fmt.Sprintf("%s/api/v1/applications/%s/envs", c.BaseURL, uuid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("unauthenticated: invalid or missing token (401)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, errors.New("invalid token (400)")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("application not found")
	}

	var envs []EnvironmentVariable
	err = json.NewDecoder(resp.Body).Decode(&envs)
	if err != nil {
		return nil, err
	}

	return envs, nil
}

func (c *Client) StartApplicationDeployment(uuid string, force, instantDeploy bool) (*StartDeploymentResponse, error) {
	url := fmt.Sprintf("%s/api/v1/applications/%s/start", c.BaseURL, uuid)

	// Build query parameters
	query := url + "?"
	if force {
		query += "force=true&"
	}
	if instantDeploy {
		query += "instant_deploy=true"
	}

	req, err := http.NewRequest("GET", query, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("unauthenticated: invalid or missing token (401)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, errors.New("invalid token (400)")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("application not found")
	}

	var result StartDeploymentResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) StopApplicationByUUID(uuid string) (*StopApplicationResponse, error) {
	url := fmt.Sprintf("%s/api/v1/applications/%s/stop", c.BaseURL, uuid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("unauthenticated: invalid or missing token (401)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, errors.New("invalid token (400)")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("application not found")
	}

	var result StopApplicationResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) RestartApplicationByUUID(uuid string) (*StartDeploymentResponse, error) {
	url := fmt.Sprintf("%s/api/v1/applications/%s/restart", c.BaseURL, uuid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errors.New("unauthenticated: invalid or missing token (401)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, errors.New("invalid token (400)")
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("application not found")
	}

	var result StartDeploymentResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
