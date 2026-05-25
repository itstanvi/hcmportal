package routes

import (
	"net/http"
	"time"

	"hcm-backend/internal/config"
	"hcm-backend/internal/handlers"
	"hcm-backend/internal/middleware"
	"hcm-backend/internal/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CORSOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	auth := handlers.NewAuthHandler(db, cfg)
	me := handlers.NewMeHandler(db)
	fields := handlers.NewFieldHandler(db)
	emps := handlers.NewEmployeeHandler(db)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		api.POST("/login", auth.Login)

		// Authenticated group
		authd := api.Group("")
		authd.Use(middleware.JWTAuth(cfg.JWTSecret))
		{
			authd.GET("/me", me.Me)

			// Fields: list available to all roles. Mutations only for ADMIN.
			authd.GET("/employee-fields", fields.List)
			authd.POST("/employee-fields", middleware.RequireRole(models.RoleAdmin), fields.Create)
			authd.PUT("/employee-fields/:id", middleware.RequireRole(models.RoleAdmin), fields.Update)

			// Employees: ADMIN and HR can manage.
			emp := authd.Group("/employees")
			emp.Use(middleware.RequireRole(models.RoleAdmin, models.RoleHR))
			{
				emp.GET("", emps.List)
				emp.POST("", emps.Create)
				emp.GET("/:id", emps.Get)
				emp.PUT("/:id", emps.Update)
			}
		}
	}

	return r
}
