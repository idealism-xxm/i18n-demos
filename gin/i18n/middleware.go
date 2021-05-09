package i18n

import (
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// GinMiddleware returns a gin middleware that
// finds a best localizer for current request and adds it to context
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取客户端接受的语言
		acceptLanguage := c.GetHeader("Accept-Language")

		// 2. 新建一个最适配当前请求的本地化器
		localizer := i18n.NewLocalizer(bundle, acceptLanguage)

		// 3. 放入 context 中
		ctx := c.Request.Context()
		ctx = WithLocalizer(ctx, localizer)

		// 4. 替换 context 为最新的
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
