package i18n

import (
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// GinMiddleware returns a gin middleware that
// finds a best localizer for current request and adds it to context
func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取客户端接受的语言，并选择最适合的一个语言（方法和 go-i18n 自带第一步一致）
		acceptLanguage := c.GetHeader("Accept-Language")
		languageTags, _, _ := language.ParseAcceptLanguage(acceptLanguage)
		supportedLanguages := bundle.LanguageTags()
		_, index, _ := bundleMatcher.Match(languageTags...)
		languageTag := supportedLanguages[index]

		// 2. 新建一个最适配当前请求的本地化器
		localizer := i18n.NewLocalizer(bundle, languageTag.String())

		// 3. 放入 context 中
		ctx := c.Request.Context()
		ctx = WithLanguageTag(ctx, languageTag)
		ctx = WithLocalizer(ctx, localizer)

		// 4. 替换 context 为最新的
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
