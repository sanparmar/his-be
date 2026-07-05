package main

import (
	"context"
	"log"
	stdhttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/appointment/internal/application/command"
	"github.com/deloitte-us-consulting/his-be/services/appointment/internal/application/query"
	"github.com/deloitte-us-consulting/his-be/services/appointment/internal/domain"
	appointmenthttp "github.com/deloitte-us-consulting/his-be/services/appointment/internal/infrastructure/http"
	"github.com/deloitte-us-consulting/his-be/services/appointment/internal/infrastructure/memory"
	"github.com/deloitte-us-consulting/his-be/services/appointment/pkg/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	repo := memory.NewRepository()
	service := domain.NewAppointmentService(repo)
	handler := appointmenthttp.NewHandler(
		command.NewCreateAppointmentHandler(service),
		command.NewUpdateAppointmentStatusHandler(service),
		query.NewListAppointmentsHandler(service),
		query.NewSearchAppointmentsHandler(service),
	)

	mux := stdhttp.NewServeMux()
	handler.Register(mux)

	server := &stdhttp.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           mux,
		ReadHeaderTimeout: cfg.ReadTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			log.Fatalf("listen and serve: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
