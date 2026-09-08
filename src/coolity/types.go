package coolify

type Application struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	FQDN        string `json:"fqdn"`
	Status      string `json:"status"`
	ProjectUUID string `json:"project_uuid"`
}

type ApplicationDetail struct {
	ID                      int64  `json:"id"`
	UUID                    string `json:"uuid"`
	Name                    string `json:"name"`
	FQDN                    string `json:"fqdn"`
	Status                  string `json:"status"`
	Description             string `json:"description"`
	GitRepository           string `json:"git_repository"`
	GitBranch               string `json:"git_branch"`
	DockerRegistryImageName string `json:"docker_registry_image_name"`
	Dockerfile              string `json:"dockerfile"`
	BuildPack               string `json:"build_pack"`
	ProjectUUID             string `json:"project_uuid"`
	CreatedAt               string `json:"created_at"`
	UpdatedAt               string `json:"updated_at"`
}

type Project struct {
	ID           int64         `json:"id"`
	UUID         string        `json:"uuid"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Environments []Environment `json:"environments,omitempty"`
}

type Environment struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ProjectID   int64  `json:"project_id"`
	Description string `json:"description"`
}

type ApplicationLogs struct {
	Logs string `json:"logs"`
}

type EnvironmentVariable struct {
	ID               int64  `json:"id"`
	UUID             string `json:"uuid"`
	ResourceableType string `json:"resourceable_type"`
	ResourceableID   int64  `json:"resourceable_id"`
	IsBuildTime      bool   `json:"is_build_time"`
	IsLiteral        bool   `json:"is_literal"`
	IsMultiline      bool   `json:"is_multiline"`
	IsPreview        bool   `json:"is_preview"`
	IsShared         bool   `json:"is_shared"`
	IsShownOnce      bool   `json:"is_shown_once"`
	Key              string `json:"key"`
	Value            string `json:"value"`
	RealValue        string `json:"real_value"`
	Version          string `json:"version"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type StartDeploymentResponse struct {
	Message        string `json:"message"`
	DeploymentUUID string `json:"deployment_uuid"`
}

type StopApplicationResponse struct {
	Message string `json:"message"`
}

type Server struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IP          string `json:"ip"`
	User        string `json:"user"`
	Port        int    `json:"port"`
}

type Database struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type Deployment struct {
	ID             int64  `json:"id"`
	ApplicationID  string `json:"application_id"`
	DeploymentUUID string `json:"deployment_uuid"`
	Status         string `json:"status"`
	Commit         string `json:"commit"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}
