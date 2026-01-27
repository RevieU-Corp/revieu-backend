package handler

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(r *gin.Engine) {
	// Initialize handlers
	testHandler := NewTestHandler()

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Test routes
		test := v1.Group("/test")
		{
			test.GET("", testHandler.GetTest)
			test.POST("", testHandler.PostTest)
		}

		// Example: User routes
		// users := v1.Group("/users")
		// {
		//     users.GET("", userHandler.List)
		//     users.GET("/:id", userHandler.Get)
		//     users.POST("", userHandler.Create)
		//     users.PUT("/:id", userHandler.Update)
		//     users.DELETE("/:id", userHandler.Delete)
		// }

		// Add your route groups here
	}
}
