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
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/redis"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/middleware"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/transport/http/handlers"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/transport/http/router"
)
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"
	}

	db, err := postgres.NewDB(ctx, connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	jwtAccessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if jwtAccessSecret == "" {
		jwtAccessSecret = "access-secret-key-change-in-production"
	}

	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		jwtRefreshSecret = "refresh-secret-key-change-in-production"
	}

	jwtService := jwt.NewJWTService(jwtAccessSecret, jwtRefreshSecret)

	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	permRepo := postgres.NewPermissionRepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	userRoleRepo := postgres.NewUserRoleRepository(db)

	// Redis Client (for permission cache)
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisClient, err := redis.NewClient(redisAddr)
	if err != nil {
		log.Printf("Failed to connect to Redis, permission cache will not work: %v", err)
	}
	var permCache *redis.PermCache
	if redisClient != nil {
		permCache = redis.NewPermCache(redisClient.Client, 15*time.Minute)
		defer redisClient.Close()
	}

	permResolver := application.NewPermissionResolver(userRoleRepo, roleRepo, permRepo, permCache)

	loginUseCase := application.NewLoginUseCase(userRepo, sessionRepo, jwtService, permResolver, userRoleRepo, roleRepo)
	refreshUseCase := application.NewRefreshUseCase(userRepo, sessionRepo, jwtService, permResolver, userRoleRepo, roleRepo)
	logoutUseCase := application.NewLogoutUseCase(sessionRepo)
	meUseCase := application.NewMeUseCase(jwtService)

	assignRoleUseCase := application.NewAssignRoleUseCase(userRoleRepo, roleRepo, permResolver)
	revokeRoleUseCase := application.NewRevokeRoleUseCase(userRoleRepo, permResolver)
	listUserPermsUseCase := application.NewListUserPermissionsUseCase(userRoleRepo, roleRepo, permResolver)
	getEffectivePermsUseCase := application.NewGetEffectivePermissionsUseCase(permResolver)
	listRolesUseCase := application.NewListRolesUseCase(roleRepo)
	listPermsUseCase := application.NewListPermissionsUseCase(permRepo)

	authHandler := handlers.NewAuthHandler(
		loginUseCase,
		refreshUseCase,
		logoutUseCase,
		meUseCase,
		assignRoleUseCase,
		revokeRoleUseCase,
		listUserPermsUseCase,
		getEffectivePermsUseCase,
		listRolesUseCase,
		listPermsUseCase,
	)

	authMiddleware := middleware.NewAuthorizationMiddleware(jwtService, permResolver)

	r := router.NewRouter(authHandler, authMiddleware)

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}