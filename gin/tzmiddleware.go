package main

import (
	"github.com/gin-gonic/gin"
	"time"
)

var defaultLocation *time.Location

func init() {
	var err error
	defaultLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
}

// TzMiddleware returns a gin middleware that
// finds a best `*time.Location` for current request and adds it to context
func TzMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取客户端指定的时区
		tzName := c.GetHeader("X-Timezone")
		location, err := time.LoadLocation(tzName)
		if tzName == "" || err != nil {
			// 没有指定时区 或者 出错，则使用默认时区
			location = defaultLocation
		}

		// 2. 放入 context 中
		ctx := WithLocation(c.Request.Context(), location)

		// 3. 替换 context 为最新的
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
