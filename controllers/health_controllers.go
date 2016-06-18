package controllers

import "github.com/labstack/echo"

func GetHealthCheck(c echo.Context) error {
	return c.JSON(200, "Lookin' Good!")
}
