package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/application"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/jwt"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/postgres"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/infrastructure/redis"
	grpcsvc "github.com/deloitte-us-consulting/his-be/services/auth-service/internal/transport/grpc"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/transport/grpc/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	authv1 "github.com/deloitte-us-consulting/his-be/api/auth/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Database connection
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://postgres:postgres@localhost:5432/auth_db?sslmode=disable"
	}

	db, err := postgres.NewDB(ctx, connString)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// JWT Service
	jwtService := jwt.NewJWTService(
		getEnv("JWT_ACCESS_SECRET", "access-secret-key"),
		getEnv("JWT_REFRESH_SECRET", "refresh-secret-key"),
	)

	// Redis Client (for token revocation)
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisClient, err := redis.NewClient(redisAddr)
	if err != nil {
		logger.Warn("failed to connect to Redis, token revocation will not work", zap.Error(err))
	}
	var revocationStore *redis.TokenRevocationStore
	if redisClient != nil {
		revocationStore = redis.NewTokenRevocationStore(redisClient.Client)
		defer redisClient.Close()
	}

	// Repositories
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	mfaRepo := postgres.NewMFARepository(db)
	roleRepo := postgres.NewRoleRepository(db)
	userRoleRepo := postgres.NewUserRoleRepository(db)

	// Use Cases
	loginUseCase := application.NewLoginUseCase(userRepo, sessionRepo, jwtService)
	refreshUseCase := application.NewRefreshUseCase(sessionRepo, jwtService)
	logoutUseCase := application.NewLogoutUseCase(sessionRepo)
	meUseCase := application.NewMeUseCase(jwtService)
	provisionIdentityUseCase := application.NewProvisionIdentityUseCase(userRepo, sessionRepo, jwtService)
	updateCredentialsUseCase := application.NewUpdateCredentialsUseCase(userRepo, sessionRepo, mfaRepo, jwtService)
	assignRolesUseCase := application.NewAssignRolesUseCase(userRepo, roleRepo, userRoleRepo)

	// Auth validator for interceptor
	authValidator := grpcsvc.NewAuthValidatorImpl(jwtService, sessionRepo)

	// gRPC Server with interceptors
	grpcPort := getEnv("GRPC_PORT", "9090")
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RecoveryInterceptor(logger),
			interceptors.LoggingInterceptor(logger),
			interceptors.MetricsInterceptor(),
			interceptors.AuthInterceptor(logger, authValidator),
		),
	)

	// Create auth service handler
	authService := grpcsvc.NewAuthService(
		loginUseCase,
		refreshUseCase,
		logoutUseCase,
		meUseCase,
		provisionIdentityUseCase,
		updateCredentialsUseCase,
		assignRolesUseCase,
		userRepo,
		sessionRepo,
		roleRepo,
		jwtService,
		revocationStore,
	)

	// Register services
	authv1.RegisterAuthServiceServer(grpcServer, authService)

	// Health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("auth.v1.AuthService", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("grpc.health.v1.Health", grpc_health_v1.HealthCheckResponse_SERVING)

	// Enable reflection for development
	reflection.Register(grpcServer)

	// Start gRPC server
	go func() {
		logger.Info("gRPC server starting", zap.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	// HTTP Gateway (gRPC-Gateway)
	httpPort := getEnv("HTTP_PORT", "8080")
	go func() {
		logger.Info("HTTP gateway starting", zap.String("port", httpPort))
		mux := runtime.NewServeMux(
			runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
				logger.Error("gRPC-Gateway error", zap.Error(err))
				runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
			}),
		)
		
		// Register gRPC-Gateway
		conn, err := grpc.DialContext(
			context.Background(),
			"localhost:"+grpcPort,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			logger.Fatal("failed to dial gRPC server for gateway", zap.Error(err))
		}
		defer conn.Close()

		if err := authv1.RegisterAuthServiceHandler(context.Background(), mux, conn); err != nil {
			logger.Fatal("failed to register gRPC-Gateway handler", zap.Error(err))
		}

		// Add health and metrics endpoints
		httpMux := http.NewServeMux()
		httpMux.Handle("/", mux)
		httpMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})
		httpMux.Handle("/metrics", promhttp.Handler())

		if err := http.ListenAndServe(":"+httpPort, httpMux); err != nil {
			logger.Fatal("HTTP gateway failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down servers...")

	// Graceful shutdown
	_, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	grpcServer.GracefulStop()
	logger.Info("servers stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
