package i18n

import (
	"github.com/gin-gonic/gin"
)

// GinMiddleware returns a gin middleware that
// finds a best localizer for current request and adds it to context
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取客户端接受的语言，并选择最适合的一个语言（方法和 go-i18n 自带第一步一致）
		acceptLanguage := c.GetHeader("Accept-Language")
		ctx := WithLanguageAndTag(c.Request.Context(), acceptLanguage)

		// 2. 替换 context 为最新的
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
