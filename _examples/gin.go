package examples

import (
	"github.com/gin-gonic/gin"
)

// Rid of debug output
func init() {
	gin.SetMode(gin.TestMode)
}

// GinHandler Create add /example route to gin engine
func GinHandler(r *gin.Engine) *gin.Engine {
	// Add route to the gin engine
	r.GET("/example", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	// return gin engine with newly added route
	return r
}
