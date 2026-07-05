package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/deloitte-us-consulting/his-be/services/appointment/internal/domain"
)

type Repository struct {
	mu    sync.RWMutex
	items map[uuid.UUID]map[string]*domain.Appointment
}

func NewRepository() *Repository {
	return &Repository{items: make(map[uuid.UUID]map[string]*domain.Appointment)}
}

func (r *Repository) Create(_ context.Context, appointment *domain.Appointment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	bucket := r.ensureTenantBucket(appointment.TenantID)
	if _, exists := bucket[appointment.ID]; exists {
		return domain.ErrAppointmentExists
	}
	bucket[appointment.ID] = clone(appointment)
	return nil
}

func (r *Repository) GetByID(_ context.Context, tenantID uuid.UUID, appointmentID string) (*domain.Appointment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bucket := r.items[tenantID]
	if bucket == nil {
		return nil, domain.ErrAppointmentNotFound
	}
	appointment, ok := bucket[appointmentID]
	if !ok {
		return nil, domain.ErrAppointmentNotFound
	}
	return clone(appointment), nil
}

func (r *Repository) Update(_ context.Context, appointment *domain.Appointment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	bucket := r.items[appointment.TenantID]
	if bucket == nil {
		return domain.ErrAppointmentNotFound
	}
	if _, ok := bucket[appointment.ID]; !ok {
		return domain.ErrAppointmentNotFound
	}
	bucket[appointment.ID] = clone(appointment)
	return nil
}

func (r *Repository) List(_ context.Context, tenantID uuid.UUID, filter domain.ListFilter) ([]*domain.Appointment, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := r.filteredAppointments(tenantID, func(appointment *domain.Appointment) bool {
		if filter.Department != "" && !strings.EqualFold(appointment.Department, filter.Department) {
			return false
		}
		if filter.DoctorID != "" && appointment.Doctor.ID != filter.DoctorID {
			return false
		}
		if filter.Status != nil && appointment.Status != *filter.Status {
			return false
		}
		if filter.Date != nil && !sameDay(appointment.AppointmentDate, *filter.Date) {
			return false
		}
		return true
	})

	return paginate(items, filter.Offset, filter.Limit), int64(len(items)), nil
}

func (r *Repository) Search(_ context.Context, tenantID uuid.UUID, filter domain.SearchFilter) ([]*domain.Appointment, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	needle := strings.ToLower(strings.TrimSpace(filter.Text))
	items := r.filteredAppointments(tenantID, func(appointment *domain.Appointment) bool {
		if filter.Status != nil && appointment.Status != *filter.Status {
			return false
		}
		if filter.Date != nil && !sameDay(appointment.AppointmentDate, *filter.Date) {
			return false
		}
		if needle == "" {
			return true
		}
		haystack := strings.ToLower(strings.Join([]string{
			appointment.ID,
			appointment.Patient.MRN,
			appointment.Patient.Name,
			appointment.Doctor.Name,
			appointment.Department,
			appointment.Notes,
		}, " "))
		return strings.Contains(haystack, needle)
	})

	return paginate(items, filter.Offset, filter.Limit), int64(len(items)), nil
}

func (r *Repository) ensureTenantBucket(tenantID uuid.UUID) map[string]*domain.Appointment {
	if bucket, ok := r.items[tenantID]; ok {
		return bucket
	}
	bucket := make(map[string]*domain.Appointment)
	r.items[tenantID] = bucket
	return bucket
}

func (r *Repository) filteredAppointments(tenantID uuid.UUID, keep func(*domain.Appointment) bool) []*domain.Appointment {
	bucket := r.items[tenantID]
	result := make([]*domain.Appointment, 0, len(bucket))
	for _, appointment := range bucket {
		if keep(appointment) {
			result = append(result, clone(appointment))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if sameDay(result[i].AppointmentDate, result[j].AppointmentDate) {
			return result[i].SlotTime < result[j].SlotTime
		}
		return result[i].AppointmentDate.Before(result[j].AppointmentDate)
	})
	return result
}

func paginate(items []*domain.Appointment, offset, limit int) []*domain.Appointment {
	if limit <= 0 {
		limit = len(items)
	}
	if offset >= len(items) {
		return []*domain.Appointment{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func sameDay(a, b time.Time) bool {
	a = a.UTC()
	b = b.UTC()
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func clone(appointment *domain.Appointment) *domain.Appointment {
	copyValue := *appointment
	return &copyValue
}
