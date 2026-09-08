package coolify

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// NewClient creates a new Coolify API client with caching
func NewClient(baseURL, token string, httpClient *http.Client, ttl time.Duration) *Client {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		Client:  httpClient,
		cache:   newCache(ttl),
	}
}

type Client struct {
	BaseURL string
	Token   string
	Client  *http.Client
	cache   *cache
}

func (c *Client) ListProjects() ([]Project, error) {
	if cached, found := c.cache.Get("projects"); found {
		return cached.([]Project), nil
	}

	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/projects", nil)
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

	var projects []Project
	err = json.NewDecoder(resp.Body).Decode(&projects)
	if err != nil {
		return nil, err
	}

	c.cache.Set("projects", projects)
	return projects, nil
}

func (c *Client) GetProjectByUUID(uuid string) (*Project, error) {
	cacheKey := fmt.Sprintf("project_%s", uuid)
	if cached, found := c.cache.Get(cacheKey); found {
		p := cached.(Project)
		return &p, nil
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s", c.BaseURL, uuid)
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
		return nil, errors.New("project not found")
	}

	var project Project
	err = json.NewDecoder(resp.Body).Decode(&project)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, project)
	return &project, nil
}

func (c *Client) ListApplications() ([]Application, error) {
	if cached, found := c.cache.Get("applications"); found {
		return cached.([]Application), nil
	}

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

	c.cache.Set("applications", apps)
	return apps, nil
}

func (c *Client) GetApplicationByUUID(uuid string) (*ApplicationDetail, error) {
	cacheKey := fmt.Sprintf("app_%s", uuid)
	if cached, found := c.cache.Get(cacheKey); found {
		app := cached.(ApplicationDetail)
		return &app, nil
	}

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

	c.cache.Set(cacheKey, app)
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

	c.cache.Delete(fmt.Sprintf("app_%s", uuid))
	c.cache.Delete(fmt.Sprintf("app_envs_%s", uuid))
	c.cache.Delete(fmt.Sprintf("app_start_%s", uuid))
	c.cache.Delete(fmt.Sprintf("app_start_%s_true_true", uuid))
	c.cache.Delete(fmt.Sprintf("app_start_%s_true_false", uuid))
	c.cache.Delete(fmt.Sprintf("app_start_%s_false_true", uuid))
	c.cache.Delete(fmt.Sprintf("app_start_%s_false_false", uuid))
	c.cache.Delete(fmt.Sprintf("app_stop_%s", uuid))
	c.cache.Delete(fmt.Sprintf("app_restart_%s", uuid))
	c.cache.Delete("applications")

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
		return "", errors.New("application logs not found")
	}

	var logs ApplicationLogs
	err = json.NewDecoder(resp.Body).Decode(&logs)
	if err != nil {
		return "", err
	}

	return logs.Logs, nil
}

func (c *Client) GetApplicationEnvsByUUID(uuid string) ([]EnvironmentVariable, error) {
	cacheKey := fmt.Sprintf("app_envs_%s", uuid)
	if cached, found := c.cache.Get(cacheKey); found {
		return cached.([]EnvironmentVariable), nil
	}

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
		return nil, errors.New("application environment variables not found")
	}

	var envs []EnvironmentVariable
	err = json.NewDecoder(resp.Body).Decode(&envs)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, envs)
	return envs, nil
}

func (c *Client) StartApplicationDeployment(uuid string, force, instantDeploy bool) (*StartDeploymentResponse, error) {
	cacheKey := fmt.Sprintf("app_start_%s_%v_%v", uuid, force, instantDeploy)
	if cached, found := c.cache.Get(cacheKey); found {
		dep := cached.(StartDeploymentResponse)
		return &dep, nil
	}

	url := fmt.Sprintf("%s/api/v1/applications/%s/start", c.BaseURL, uuid)
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

	var deployment StartDeploymentResponse
	err = json.NewDecoder(resp.Body).Decode(&deployment)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, deployment)
	return &deployment, nil
}

func (c *Client) StopApplicationByUUID(uuid string) (*StopApplicationResponse, error) {
	cacheKey := fmt.Sprintf("app_stop_%s", uuid)
	if cached, found := c.cache.Get(cacheKey); found {
		stop := cached.(StopApplicationResponse)
		return &stop, nil
	}

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

	var stopResponse StopApplicationResponse
	err = json.NewDecoder(resp.Body).Decode(&stopResponse)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, stopResponse)
	return &stopResponse, nil
}

func (c *Client) RestartApplicationByUUID(uuid string) (*StartDeploymentResponse, error) {
	cacheKey := fmt.Sprintf("app_restart_%s", uuid)
	if cached, found := c.cache.Get(cacheKey); found {
		dep := cached.(StartDeploymentResponse)
		return &dep, nil
	}

	url := fmt.Sprintf("%s/api/v1/applications/%s/restart", c.BaseURL, uuid)

	req, err := http.NewRequest("POST", url, nil)
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

	var deployment StartDeploymentResponse
	err = json.NewDecoder(resp.Body).Decode(&deployment)
	if err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, deployment)
	return &deployment, nil
}

func (c *Client) ListServers() ([]Server, error) {
	if cached, found := c.cache.Get("servers"); found {
		return cached.([]Server), nil
	}

	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/servers", nil)
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

	var servers []Server
	err = json.NewDecoder(resp.Body).Decode(&servers)
	if err != nil {
		return nil, err
	}

	c.cache.Set("servers", servers)
	return servers, nil
}

func (c *Client) ListDatabases() ([]Database, error) {
	if cached, found := c.cache.Get("databases"); found {
		return cached.([]Database), nil
	}

	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/databases", nil)
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

	var databases []Database
	err = json.NewDecoder(resp.Body).Decode(&databases)
	if err != nil {
		return nil, err
	}

	c.cache.Set("databases", databases)
	return databases, nil
}

func (c *Client) ListDeployments() ([]Deployment, error) {
	if cached, found := c.cache.Get("deployments"); found {
		return cached.([]Deployment), nil
	}

	req, err := http.NewRequest("GET", c.BaseURL+"/api/v1/deployments", nil)
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

	var deployments []Deployment
	err = json.NewDecoder(resp.Body).Decode(&deployments)
	if err != nil {
		return nil, err
	}

	c.cache.Set("deployments", deployments)
	return deployments, nil
}
