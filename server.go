package main

import (
	"log"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/tylerb/graceful"

	"github.com/walkerlangley/basic-go-server/controllers"
)

func main() {
	e := echo.New()

	//---------------------------
	// Health Check
	//---------------------------

	e.Get("/health", controllers.GetHealthCheck)

	//---------------------------
	// Main routes
	//---------------------------

	// e.Get("/", controller)

	port := "4001"
	std := standard.New(":" + port)
	std.SetHandler(e)
	log.Println("Server started on port: " + port)
	graceful.ListenAndServe(std.Server, 5*time.Second)
}
