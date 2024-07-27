package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/techrail/ground"
	"github.com/valyala/fasthttp"
)

func handleDeploy(ctx *fasthttp.RequestCtx) {
	var deployRequest DeployRequest
	if err := json.Unmarshal(ctx.PostBody(), &deployRequest); err != nil {
		fmt.Print(err.Error())
	}
	go DockerAPI(deployRequest)
	if res, err := json.Marshal(&deployRequest); err != nil {
		ctx.Response.SetStatusCode(http.StatusAccepted)
		ctx.Response.SetBody(res)
	}
}

func main() {
	fmt.Print("Hello")
	server := ground.GiveMeAWebServer()

	server.Router.Handle(http.MethodPost, "/deploy", handleDeploy)
	server.BlockOnStart = true
	server.Start()
}
