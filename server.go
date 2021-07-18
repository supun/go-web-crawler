package main

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/lestoni/html-version"
)

type (
	formdata struct {
		URL   string    `json:"url"`
	}

	htmlcontentresponse struct{
		Status string `json:"status"`
		Version string `json:"version"`

	}
)
func main() { 
	e := echo.New()
	
	// Routes
	e.File("/","public/index.html")
	e.POST("/api/check-url", checkUrl)
	e.Logger.Fatal(e.Start(":1323"))
}

func checkUrl(c echo.Context) error{
	u := new(formdata)
	if err := c.Bind(u); err != nil {
		return err
	}

	if u.URL == ""{
		return c.JSON(http.StatusBadRequest,"Invalid input")
	}
	
	version, err := HTMLVersion.DetectFromURL(u.URL)
	
	if err == nil{
		return c.JSON(http.StatusOK,version)
	}
	return c.JSON(http.StatusOK,u)
}