package main

import (
	"fmt"
	"net/http"

	"github.com/techrail/ascend/controllers"
	"github.com/techrail/ground"
)

func main() {
	fmt.Print("Hello")
	server := ground.GiveMeAWebServer()
	server.Router.Handle(http.MethodPost, "/deploy", controllers.HandleDeploy)
	server.BlockOnStart = true
	server.Start()
}
