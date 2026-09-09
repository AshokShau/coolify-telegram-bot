package coolify

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	Token   string
	Client  *http.Client
	cache   *cache
}

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

func (c *Client) doRequest(method, urlPath string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.BaseURL+urlPath, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		_ = resp.Body.Close()
		return nil, errors.New("unauthenticated: invalid or missing token (401)")
	}
	if resp.StatusCode == http.StatusBadRequest {
		_ = resp.Body.Close()
		return nil, errors.New("invalid token (400)")
	}

	return resp, nil
}

func requestJSON[T any](c *Client, method, urlPath string, notFoundMsg string) (T, error) {
	var result T
	resp, err := c.doRequest(method, urlPath)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		if notFoundMsg != "" {
			return result, errors.New(notFoundMsg)
		}
		return result, errors.New("not found")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return result, fmt.Errorf("unexpected response: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func (c *Client) ListProjects() ([]Project, error) {
	if cached, found := c.cache.Get("projects"); found {
		return cached.([]Project), nil
	}
	projects, err := requestJSON[[]Project](c, "GET", "/api/v1/projects", "")
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
	project, err := requestJSON[Project](c, "GET", fmt.Sprintf("/api/v1/projects/%s", uuid), "project not found")
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
	apps, err := requestJSON[[]Application](c, "GET", "/api/v1/applications", "")
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
	app, err := requestJSON[ApplicationDetail](c, "GET", fmt.Sprintf("/api/v1/applications/%s", uuid), "application not found")
	if err != nil {
		return nil, err
	}
	c.cache.Set(cacheKey, app)
	return &app, nil
}

func (c *Client) DeleteApplicationByUUID(uuid string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/api/v1/applications/%s", uuid))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
	logs, err := requestJSON[ApplicationLogs](c, "GET", fmt.Sprintf("/api/v1/applications/%s/logs?lines=-1", uuid), "application logs not found")
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
	envs, err := requestJSON[[]EnvironmentVariable](c, "GET", fmt.Sprintf("/api/v1/applications/%s/envs", uuid), "application environment variables not found")
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

	path := fmt.Sprintf("/api/v1/applications/%s/start", uuid)
	var params []string
	if force {
		params = append(params, "force=true")
	}
	if instantDeploy {
		params = append(params, "instant_deploy=true")
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	deployment, err := requestJSON[StartDeploymentResponse](c, "GET", path, "application not found")
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
	stopResponse, err := requestJSON[StopApplicationResponse](c, "GET", fmt.Sprintf("/api/v1/applications/%s/stop", uuid), "application not found")
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
	deployment, err := requestJSON[StartDeploymentResponse](c, "POST", fmt.Sprintf("/api/v1/applications/%s/restart", uuid), "application not found")
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
	servers, err := requestJSON[[]Server](c, "GET", "/api/v1/servers", "")
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
	databases, err := requestJSON[[]Database](c, "GET", "/api/v1/databases", "")
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
	deployments, err := requestJSON[[]Deployment](c, "GET", "/api/v1/deployments", "")
	if err != nil {
		return nil, err
	}
	c.cache.Set("deployments", deployments)
	return deployments, nil
}
