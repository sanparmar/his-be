package query

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/deloitte-us-consulting/his-be/services/appointment/internal/domain"
)

type ListAppointmentsQuery struct {
	TenantID   uuid.UUID
	Department string
	DoctorID   string
	Status     *domain.AppointmentStatus
	Date       *time.Time
	Offset     int
	Limit      int
}

type ListAppointmentsHandler struct {
	service *domain.AppointmentService
}

func NewListAppointmentsHandler(service *domain.AppointmentService) *ListAppointmentsHandler {
	return &ListAppointmentsHandler{service: service}
}

func (h *ListAppointmentsHandler) Handle(ctx context.Context, query ListAppointmentsQuery) ([]*domain.Appointment, int64, error) {
	return h.service.ListAppointments(ctx, query.TenantID, domain.ListFilter{
		Department: query.Department,
		DoctorID:   query.DoctorID,
		Status:     query.Status,
		Date:       query.Date,
		Offset:     query.Offset,
		Limit:      query.Limit,
	})
}

type SearchAppointmentsQuery struct {
	TenantID uuid.UUID
	Text     string
	Date     *time.Time
	Status   *domain.AppointmentStatus
	Offset   int
	Limit    int
}

type SearchAppointmentsHandler struct {
	service *domain.AppointmentService
}

func NewSearchAppointmentsHandler(service *domain.AppointmentService) *SearchAppointmentsHandler {
	return &SearchAppointmentsHandler{service: service}
}

func (h *SearchAppointmentsHandler) Handle(ctx context.Context, query SearchAppointmentsQuery) ([]*domain.Appointment, int64, error) {
	return h.service.SearchAppointments(ctx, query.TenantID, domain.SearchFilter{
		Text:   query.Text,
		Date:   query.Date,
		Status: query.Status,
		Offset: query.Offset,
		Limit:  query.Limit,
	})
}
