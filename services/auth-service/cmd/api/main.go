package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/application"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/jwt"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/postgres"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/transport/http/handlers"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/transport/http/router"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Setup Database
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"
	}

	db, err := postgres.NewDB(ctx, connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 2. Setup Infrastructure
	jwtService := jwt.NewJWTService(
		os.Getenv("JWT_ACCESS_SECRET"),
		os.Getenv("JWT_REFRESH_SECRET"),
	)
	if os.Getenv("JWT_ACCESS_SECRET") == "" {
		jwtService = jwt.NewJWTService("access-secret-key", "refresh-secret-key")
	}

	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)

	// 3. Setup Application (Use Cases)
	loginUseCase := application.NewLoginUseCase(userRepo, sessionRepo, jwtService)
	refreshUseCase := application.NewRefreshUseCase(sessionRepo, jwtService)
	logoutUseCase := application.NewLogoutUseCase(sessionRepo)
	meUseCase := application.NewMeUseCase(jwtService)

	// 4. Setup Transport (Handlers & Router)
	handler := handlers.NewAuthHandler(
		loginUseCase,
		refreshUseCase,
		logoutUseCase,
		meUseCase,
	)
	r := router.NewRouter(handler)

	// 5. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Auth service starting on port %s...\n", port)
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
