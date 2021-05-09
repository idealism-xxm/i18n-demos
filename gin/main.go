package main

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"gin-i18n/i18n"
)

func main() {
	startGinServer()
}

func startGinServer() {
	r := gin.Default()
	r.Use(i18n.GinMiddleware())

	r.GET("/hello/:username/", func(c *gin.Context) {
		name := c.Param("username")
		currentLanguage := i18n.CurrentLanguage(c.Request.Context(), name)
		myCats := i18n.MyCats(c.Request.Context(), 1)
		myCats += i18n.MyCats(c.Request.Context(), 2)
		c.String(http.StatusOK, currentLanguage+myCats)
	})

	if err := r.Run(); err != nil {
		panic(err)
	}
}
