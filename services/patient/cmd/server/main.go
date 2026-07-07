package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	kafkaclient "github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	patientv1 "github.com/deloitte-us-consulting/his-be/api/patient/v1"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/application/command"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/application/query"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/domain"
	grpcserver "github.com/deloitte-us-consulting/his-be/services/patient/internal/infrastructure/grpc"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/infrastructure/kafka"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/infrastructure/postgres"
	httphandlers "github.com/deloitte-us-consulting/his-be/services/patient/internal/transport/http/handlers"
	httproutes "github.com/deloitte-us-consulting/his-be/services/patient/internal/transport/http/router"
	"github.com/deloitte-us-consulting/his-be/services/patient/pkg/config"
	"github.com/deloitte-us-consulting/his-be/services/patient/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Logging.Level)
	logger.Info("starting patient service", logger.String("version", "1.0.0"))

	// Connect to database
	db, err := sqlx.Connect("postgres", cfg.Database.DSN())
	if err != nil {
		logger.Error("failed to connect to database", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.Database.MaxConns)
	db.SetMaxIdleConns(cfg.Database.MaxConns / 2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := db.PingContext(ctx); err != nil {
		logger.Error("database ping failed", err)
		cancel()
		os.Exit(1)
	}
	cancel()

	logger.Info("connected to database", logger.String("host", cfg.Database.Host))

	// Initialize Kafka writer
	kafkaWriter := &kafkaclient.Writer{
		Addr:     kafkaclient.TCP(cfg.Kafka.Brokers...),
		Topic:    "his.patient.patient.events",
		Balancer: &kafkaclient.LeastBytes{},
	}
	defer kafkaWriter.Close()

	logger.Info("connected to kafka", logger.String("brokers", cfg.Kafka.Brokers[0]))

	// Wire dependencies
	patientRepo := postgres.NewPatientRepository(db)
	patientSvc := domain.NewPatientService(patientRepo)
	eventPublisher := kafka.NewEventPublisher(kafkaWriter)

	// CQRS handlers - Command handlers
	createPatientHandler := command.NewCreatePatientHandler(patientSvc, eventPublisher)
	updateDemographicsHandler := command.NewUpdateDemographicsHandler(patientSvc, eventPublisher)
	updateInsuranceHandler := command.NewUpdateInsuranceHandler(patientSvc, eventPublisher)
	recordAllergiesHandler := command.NewRecordAllergiesHandler(patientSvc, eventPublisher)
	recordMedicationsHandler := command.NewRecordMedicationsHandler(patientSvc, eventPublisher)
	recordVitalSignsHandler := command.NewRecordVitalSignsHandler(patientSvc, eventPublisher)

	// CQRS handlers - Query handlers
	getPatientByMRNHandler := query.NewGetPatientByMRNHandler(patientSvc)
	searchPatientsHandler := query.NewSearchPatientsHandler(patientSvc)
	listPatientsHandler := query.NewListPatientsHandler(patientSvc)

	// Create HTTP handler and router (REST API)
	patientHTTPHandler := httphandlers.NewPatientHandler(
		createPatientHandler,
		updateDemographicsHandler,
		updateInsuranceHandler,
		recordAllergiesHandler,
		recordMedicationsHandler,
		recordVitalSignsHandler,
		getPatientByMRNHandler,
		searchPatientsHandler,
		listPatientsHandler,
	)
	httpRouter := httproutes.NewRouter(patientHTTPHandler)

	// Create gRPC server
	grpcSrv := grpc.NewServer()

	// Create patient server
	patientServer := grpcserver.NewPatientServer(
		patientSvc,
		createPatientHandler,
		updateDemographicsHandler,
		updateInsuranceHandler,
		recordAllergiesHandler,
		recordMedicationsHandler,
		recordVitalSignsHandler,
		getPatientByMRNHandler,
		searchPatientsHandler,
		listPatientsHandler,
	)

	// Register patient service with gRPC
	patientv1.RegisterPatientServiceServer(grpcSrv, patientServer)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthServer)
	healthServer.SetServingStatus("grpc.health.v1.Health", grpc_health_v1.HealthCheckResponse_SERVING)

	// Listen on gRPC port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		logger.Error("failed to create listener", err)
		os.Exit(1)
	}

	logger.Info("gRPC server listening", logger.String("addr", fmt.Sprintf(":%d", cfg.Server.Port)))

	// Start gRPC server in goroutine
	go func() {
		if err := grpcSrv.Serve(listener); err != nil && err.Error() != "http: Server closed" {
			logger.Error("gRPC server error", err)
			os.Exit(1)
		}
	}()

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler: httpRouter,
	}

	go func() {
		logger.Info("HTTP server listening", logger.String("addr", fmt.Sprintf(":%d", cfg.Server.HTTPPort)))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", err)
			os.Exit(1)
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("shutting down patient service")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	httpServer.Shutdown(shutdownCtx)
	grpcSrv.GracefulStop()

	logger.Info("patient service shutdown complete")
}
