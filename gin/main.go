package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"gin-i18n/i18n"
)

func main() {
	startGinServer()
}

func startGinServer() {
	r := gin.Default()
	r.Use(i18n.GinMiddleware())
	r.Use(TzMiddleware())

	r.GET("/hello/:username/", func(c *gin.Context) {
		// 获取路径中的 username
		username := c.Param("username")
		// 从 context 中获取 languageTag 和 location
		ctx := c.Request.Context()
		langaugeTag := i18n.LanguageTagFromContext(ctx)
		location := LocationFromContext(ctx)
		currentLanguage := i18n.CurrentLanguage(ctx, langaugeTag.String())
		personCat := i18n.PersonCats(ctx, username, 1)
		personCats := i18n.PersonCats(ctx, username, 2)
		nowStr := time.Now().In(location).Format(time.RFC3339Nano)
		timeStr := fmt.Sprintf("(%s) %s", location.String(), nowStr)
		c.String(http.StatusOK, fmt.Sprintf("%v\n%v\n%v\n%v", currentLanguage, personCat, personCats, timeStr))
	})

	if err := r.Run(); err != nil {
		panic(err)
	}
}
