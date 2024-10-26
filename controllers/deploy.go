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

	responseChan := make(chan models.DockerResponse)

	go deploy.DockerAPI(deployRequest, responseChan)

	res := <-responseChan
	ctx.Response.Header.Set("Content-Type", "application/json")
	if res.Error != nil {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		json.NewEncoder(ctx).Encode(res)
	} else {
		ctx.Response.SetStatusCode(fasthttp.StatusAccepted)
		data, _ := json.Marshal(res)
		ctx.SetBody(data)
	}
}
