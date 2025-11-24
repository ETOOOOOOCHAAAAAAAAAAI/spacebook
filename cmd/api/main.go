package main

import (
	"log"

	"SpaceBookProject/internal/auth"
	"SpaceBookProject/internal/config"
	"SpaceBookProject/internal/db"
	"SpaceBookProject/internal/domain"
	"SpaceBookProject/internal/handlers"
	"SpaceBookProject/internal/repository"
	"SpaceBookProject/internal/services"
	"SpaceBookProject/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.InitDB(&cfg.Database)
	if err != nil {
		log.Fatal(err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWT.SecretKey)

	userRepo := repository.NewUserRepository(database)
	bookingRepo := repository.NewBookingRepository(database)
	spaceRepo := repository.NewSpaceRepository(database)

	authService := services.NewAuthService(userRepo, jwtManager)
	bookingService := services.NewBookingService(bookingRepo)
	spaceService := services.NewSpaceService(spaceRepo)

	authHandler := handlers.NewAuthHandler(authService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	spaceHandler := handlers.NewSpaceHandler(spaceService)

	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	api := r.Group(cfg.API.Prefix + "/" + cfg.API.Version)

	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)

		authGroup.Use(middleware.AuthMiddleware(jwtManager))
		authGroup.GET("/me", authHandler.GetMe)
		authGroup.POST("/logout", authHandler.Logout)
	}

	spacesGroup := api.Group("/spaces")
	{
		spacesGroup.GET("", spaceHandler.ListSpaces)
	}
	spacesGroup.Use(
		middleware.AuthMiddleware(jwtManager),
		middleware.RoleMiddleware(domain.RoleOwner),
	)
	spacesGroup.POST("", spaceHandler.CreateSpace)

	bookings := api.Group("/bookings")
	bookings.Use(middleware.AuthMiddleware(jwtManager))

	bookings.POST("", middleware.RoleMiddleware(domain.RoleTenant), bookingHandler.CreateBooking)
	bookings.GET("/my", middleware.RoleMiddleware(domain.RoleTenant), bookingHandler.ListMyBookings)
	bookings.PATCH("/:id/cancel", middleware.RoleMiddleware(domain.RoleTenant), bookingHandler.CancelBooking)

	owner := api.Group("/owner")
	owner.Use(middleware.AuthMiddleware(jwtManager), middleware.RoleMiddleware(domain.RoleOwner))

	owner.GET("/bookings", bookingHandler.ListOwnerBookings)
	owner.PATCH("/bookings/:id/approve", bookingHandler.ApproveBooking)
	owner.PATCH("/bookings/:id/reject", bookingHandler.RejectBooking)

	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}
