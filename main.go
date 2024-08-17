package main

import (
	"fmt"
	"net/http"

	"github.com/techrail/ascend/controllers"
	"github.com/techrail/ground"
	"github.com/valyala/fasthttp"
)

func main() {
	fmt.Print("Hello")
	server := ground.GiveMeAWebServer()
	server.Router.Handle(http.MethodPost, "/deploy", controllers.HandleDeploy)
	server.Router.Handle(http.MethodGet, "/", index)
	server.BlockOnStart = true
	server.Start()
}

func index(ctx *fasthttp.RequestCtx) {
	ctx.SetBodyString("Hello")
}
