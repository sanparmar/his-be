package command

import (
	"context"

	"github.com/google/uuid"

	"github.com/deloitte-us-consulting/his-be/services/appointment/internal/domain"
)

type CreateAppointmentCommand struct {
	TenantID        uuid.UUID
	Patient         domain.PatientSummary
	Doctor          domain.DoctorSummary
	Department      string
	AppointmentDate string
	SlotTime        string
	VisitType       domain.VisitType
	Notes           string
	CreatedBy       uuid.UUID
}

type CreateAppointmentHandler struct {
	service *domain.AppointmentService
}

func NewCreateAppointmentHandler(service *domain.AppointmentService) *CreateAppointmentHandler {
	return &CreateAppointmentHandler{service: service}
}

func (h *CreateAppointmentHandler) Handle(ctx context.Context, cmd CreateAppointmentCommand) (*domain.Appointment, error) {
	return h.service.CreateAppointment(ctx, domain.CreateAppointmentInput(cmd))
}

type UpdateAppointmentStatusCommand struct {
	TenantID      uuid.UUID
	AppointmentID string
	Status        domain.AppointmentStatus
	UpdatedBy     uuid.UUID
}

type UpdateAppointmentStatusHandler struct {
	service *domain.AppointmentService
}

func NewUpdateAppointmentStatusHandler(service *domain.AppointmentService) *UpdateAppointmentStatusHandler {
	return &UpdateAppointmentStatusHandler{service: service}
}

func (h *UpdateAppointmentStatusHandler) Handle(ctx context.Context, cmd UpdateAppointmentStatusCommand) (*domain.Appointment, error) {
	return h.service.UpdateStatus(ctx, cmd.TenantID, cmd.AppointmentID, cmd.Status, cmd.UpdatedBy)
}
