package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"media-server/controllers"
)

func main() {
	_ = godotenv.Load()

	e := echo.New()
	e.Use(middleware.CORS())
	e.Use(middleware.Logger())
	e.GET("/call", controllers.Call)
	if err := e.Start(":8080"); err != nil {
		fmt.Println(err)
	}
}
