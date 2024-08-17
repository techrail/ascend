package controllers

import (
	"encoding/json"

	"github.com/techrail/ascend/deploy"
	"github.com/techrail/ascend/models"
	"github.com/valyala/fasthttp"
)

func HandleDeploy(ctx *fasthttp.RequestCtx) {
	body := ctx.Request.Body()
	if len(body) == 0 {
		ctx.Error("Empty request", fasthttp.StatusBadRequest)
		return
	}

	var deployRequest models.DeployRequest
	if err := json.Unmarshal(body, &deployRequest); err != nil {
		ctx.Error("Invalid request body structure", fasthttp.StatusBadRequest)
		return
	}
	go deploy.DockerAPI(deployRequest)
	ctx.Response.SetStatusCode(fasthttp.StatusAccepted)

}
