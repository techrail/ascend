package main

type DeployRequest struct {
	Git_URL   string `json:"git_url"`
	Build_CMD string `json:"build_cmd"`
	Start_CMD string `json:"start_cmd"`
	Port      string `json:"port"`
}
