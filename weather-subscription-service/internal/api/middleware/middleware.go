package middleware

import "github.com/gin-gonic/gin"

func ExtractParam(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		param := c.Param(key)

		if param == "" {
			c.Next()
			return
		}

		c.Set(key, param)

		c.Next()
	}
}

func ExtractQuery(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query(key)

		if query == "" {
			c.Next()
			return
		}

		c.Set(key, query)

		c.Next()
	}
}
