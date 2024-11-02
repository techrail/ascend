package models

type Mount struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

type DeployRequest struct {
	RepositoryUrl *string  `json:"repositoryUrl"`
	BuildCommand  *string  `json:"buildCommand"`
	StartCommand  *string  `json:"startCommand"`
	Port          *string  `json:"port"`
	Branch        string   `json:"branch"`
	MemoryLimit   *int64   `json:"memoryLimit"`
	Mounts        *[]Mount `json:"mounts"`
	CPUs          *float64 `json:"cpus"`
}

type DockerResponse struct {
	Error *string         `json:"error,omitempty"`
	Info  *DeploymentInfo `json:"info,omitempty"`
}

type DeploymentInfo struct {
	ImageName string `json:"imageName"`
	Port      string `json:"port"`
}
