package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/transport/http/handlers"
)

func NewRouter(h *handlers.PatientHandler) *mux.Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/patients", h.CreatePatient).Methods(http.MethodPost)
	api.HandleFunc("/patients", h.ListPatients).Methods(http.MethodGet)
	api.HandleFunc("/patients/{mrn}", h.GetPatientByMRN).Methods(http.MethodGet)
	api.HandleFunc("/patients/{mrn}/demographics", h.UpdateDemographics).Methods(http.MethodPut)
	api.HandleFunc("/patients/{mrn}/insurance", h.UpdateInsurance).Methods(http.MethodPut)
	api.HandleFunc("/patients/{mrn}/allergies", h.RecordAllergies).Methods(http.MethodPut)
	api.HandleFunc("/patients/{mrn}/medications", h.RecordMedications).Methods(http.MethodPut)
	api.HandleFunc("/patients/{mrn}/vitals", h.RecordVitalSigns).Methods(http.MethodPut)

	return r
}
