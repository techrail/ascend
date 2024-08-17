package models

type DeployRequest struct {
	RepositoryUrl *string `json:"repositoryUrl"`
	BuildCommand  *string `json:"buildCommand"`
	StartCommand  *string `json:"startCommand"`
	Port          *string `json:"port"`
	Branch        string  `json:"branch"`
}
