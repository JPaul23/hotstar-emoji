package main

import (
	"os"

	"api/controllers"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func logRequestMiddleware() gin.HandlerFunc {
	logger := zap.NewExample()
	defer logger.Sync() // flushes buffer, if any

	return func(ctx *gin.Context) {
		logger.Info("incoming request",
			zap.String("path", ctx.Request.URL.Path),
			zap.String("method", ctx.Request.Method),
		)
		ctx.Next()
	}
}

func main() {
	// Set up the router
	r := setupRouter()

	// background kafka channels
	controllers.HandleKafkaProducerBg()

	// Start the server
	apiPort := os.Getenv("API_PORT")
	r.Run(":" + apiPort)
}

// setupRouter sets up the router and adds the routes.
func setupRouter() *gin.Engine {
	// Create a new router
	router := gin.Default()
	router.ForwardedByClientIP = true
	router.SetTrustedProxies([]string{"127.0.0.1"})

	router.Use(logRequestMiddleware())
	// Add a welcome route
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello from  API")
	})
	// Create a new group for the API
	api := router.Group("/api/v1/hotstar")
	{
		// Create a new group for the account routes
		public := api.Group("/emoji")
		{
			// Add the login route
			public.GET("/", controllers.GetEmojis)
			// Add the signup route
			public.POST("/react", controllers.EmojiReaction)
		}

	}
	// Return the router
	return router
}
