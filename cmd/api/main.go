package main

import (
	"SpaceBookProject/internal/worker"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer database.Close()

	jwtManager := auth.NewJWTManager(cfg.JWT.SecretKey)

	userRepo := repository.NewUserRepository(database)
	bookingRepo := repository.NewBookingRepository(database)
	spaceRepo := repository.NewSpaceRepository(database)
	eventsChan := make(chan domain.BookingEvent, 100)

	authService := services.NewAuthService(userRepo, jwtManager)
	bookingService := services.NewBookingService(bookingRepo, spaceRepo, eventsChan)
	spaceService := services.NewSpaceService(spaceRepo)

	authHandler := handlers.NewAuthHandler(authService)
	bookingHandler := handlers.NewBookingHandler(bookingService)
	spaceHandler := handlers.NewSpaceHandler(spaceService)

	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORSMiddleware())

	api := r.Group(cfg.API.Prefix + "/" + cfg.API.Version)

	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.GET("/me", middleware.AuthMiddleware(jwtManager), authHandler.GetMe)
	}

	spacesGroup := api.Group("/spaces")
	{
		spacesGroup.GET("", spaceHandler.ListSpaces)
	}
	ownerSpaces := api.Group("/spaces", middleware.AuthMiddleware(jwtManager), middleware.OwnerOnlyMiddleware())
	{
		ownerSpaces.POST("", spaceHandler.CreateSpace)
	}

	bookingsGroup := api.Group("/bookings", middleware.AuthMiddleware(jwtManager))
	{
		bookingsGroup.POST("", middleware.RoleMiddleware(domain.RoleTenant), bookingHandler.CreateBooking)
		bookingsGroup.GET("/my", middleware.RoleMiddleware(domain.RoleTenant), bookingHandler.MyBookings)
		bookingsGroup.PATCH("/:id/cancel", middleware.RoleMiddleware(domain.RoleTenant), bookingHandler.CancelBooking)
	}

	ownerBookings := api.Group("/owner/bookings",
		middleware.AuthMiddleware(jwtManager),
		middleware.OwnerOnlyMiddleware(),
	)
	{
		ownerBookings.GET("", bookingHandler.OwnerBookings)
		ownerBookings.PATCH("/:id/approve", bookingHandler.ApproveBooking)
		ownerBookings.PATCH("/:id/reject", bookingHandler.RejectBooking)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer stop()

	bookingWorker := worker.NewBookingEventWorker(eventsChan)
	go bookingWorker.Run(ctx)

	go func() {
		log.Printf("server listening on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited gracefully")
}
