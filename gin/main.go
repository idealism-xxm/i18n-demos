package main

import (
	"fmt"
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
		// 获取路径中的 username
		username := c.Param("username")
		// 从 context 中获取 languageTag
		langaugeTag := i18n.LanguageTagFromContext(c.Request.Context())
		currentLanguage := i18n.CurrentLanguage(c.Request.Context(), langaugeTag.String())
		personCat := i18n.PersonCats(c.Request.Context(), username, 1)
		personCats := i18n.PersonCats(c.Request.Context(), username, 2)
		c.String(http.StatusOK, fmt.Sprintf("%v\n%v\n%v", currentLanguage, personCat, personCats))
	})

	if err := r.Run(); err != nil {
		panic(err)
	}
}
