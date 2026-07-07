package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/deloitte-us-consulting/his-be/services/patient/internal/application/command"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/application/query"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/domain"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type PatientHandler struct {
	createPatientHandler      *command.CreatePatientHandler
	updateDemographicsHandler *command.UpdateDemographicsHandler
	updateInsuranceHandler    *command.UpdateInsuranceHandler
	recordAllergiesHandler    *command.RecordAllergiesHandler
	recordMedicationsHandler  *command.RecordMedicationsHandler
	recordVitalSignsHandler   *command.RecordVitalSignsHandler
	getPatientByMRNHandler    *query.GetPatientByMRNHandler
	searchPatientsHandler     *query.SearchPatientsHandler
	listPatientsHandler       *query.ListPatientsHandler
}

func NewPatientHandler(
	createPatientHandler *command.CreatePatientHandler,
	updateDemographicsHandler *command.UpdateDemographicsHandler,
	updateInsuranceHandler *command.UpdateInsuranceHandler,
	recordAllergiesHandler *command.RecordAllergiesHandler,
	recordMedicationsHandler *command.RecordMedicationsHandler,
	recordVitalSignsHandler *command.RecordVitalSignsHandler,
	getPatientByMRNHandler *query.GetPatientByMRNHandler,
	searchPatientsHandler *query.SearchPatientsHandler,
	listPatientsHandler *query.ListPatientsHandler,
) *PatientHandler {
	return &PatientHandler{
		createPatientHandler:      createPatientHandler,
		updateDemographicsHandler: updateDemographicsHandler,
		updateInsuranceHandler:    updateInsuranceHandler,
		recordAllergiesHandler:    recordAllergiesHandler,
		recordMedicationsHandler:  recordMedicationsHandler,
		recordVitalSignsHandler:   recordVitalSignsHandler,
		getPatientByMRNHandler:    getPatientByMRNHandler,
		searchPatientsHandler:     searchPatientsHandler,
		listPatientsHandler:       listPatientsHandler,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, title, detail string) {
	writeJSON(w, status, map[string]string{
		"title":  title,
		"detail": detail,
	})
}

type createPatientRequest struct {
	TenantID   string `json:"tenant_id"`
	Name       string `json:"name"`
	DOB        string `json:"dob"`
	Gender     string `json:"gender"`
	Phone      string `json:"phone"`
	Email      string `json:"email,omitempty"`
	BloodGroup string `json:"blood_group"`
	CreatedBy  string `json:"created_by"`
}

func (h *PatientHandler) CreatePatient(w http.ResponseWriter, r *http.Request) {
	var req createPatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant_id", err.Error())
		return
	}

	dob, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid dob (expects YYYY-MM-DD)", err.Error())
		return
	}

	gender, err := domain.ParseGender(req.Gender)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid gender", err.Error())
		return
	}

	bloodGroup, err := domain.ParseBloodGroup(req.BloodGroup)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid blood_group", err.Error())
		return
	}

	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid created_by", err.Error())
		return
	}

	cmd := &command.CreatePatientCommand{
		TenantID:   tenantID,
		Name:       req.Name,
		DOB:        dob,
		Gender:     gender,
		Phone:      req.Phone,
		Email:      req.Email,
		BloodGroup: bloodGroup,
		CreatedBy:  createdBy,
	}

	patient, err := h.createPatientHandler.Handle(r.Context(), cmd)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Failed to create patient", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, patientToResponse(patient))
}

type updateDemographicsRequest struct {
	TenantID  string `json:"tenant_id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Email     string `json:"email,omitempty"`
	UpdatedBy string `json:"updated_by"`
}

func (h *PatientHandler) UpdateDemographics(w http.ResponseWriter, r *http.Request) {
	mrn := mux.Vars(r)["mrn"]
	if mrn == "" {
		writeError(w, http.StatusBadRequest, "Missing MRN", "MRN is required in the path")
		return
	}

	var req updateDemographicsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant_id", err.Error())
		return
	}

	updatedBy, err := uuid.Parse(req.UpdatedBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid updated_by", err.Error())
		return
	}

	cmd := &command.UpdateDemographicsCommand{
		TenantID:  tenantID,
		MRN:       mrn,
		Name:      req.Name,
		Phone:     req.Phone,
		Email:     req.Email,
		UpdatedBy: updatedBy,
	}

	patient, err := h.updateDemographicsHandler.Handle(r.Context(), cmd)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Failed to update demographics", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, patientToResponse(patient))
}

type updateInsuranceRequest struct {
	TenantID      string `json:"tenant_id"`
	Provider      string `json:"provider"`
	PolicyNo      string `json:"policy_no"`
	TPA           string `json:"tpa,omitempty"`
	MemberID      string `json:"member_id,omitempty"`
	SumInsured    int64  `json:"sum_insured,omitempty"`
	CorporateName string `json:"corporate_name,omitempty"`
	UpdatedBy     string `json:"updated_by"`
}

func (h *PatientHandler) UpdateInsurance(w http.ResponseWriter, r *http.Request) {
	mrn := mux.Vars(r)["mrn"]
	if mrn == "" {
		writeError(w, http.StatusBadRequest, "Missing MRN", "MRN is required in the path")
		return
	}

	var req updateInsuranceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant_id", err.Error())
		return
	}

	updatedBy, err := uuid.Parse(req.UpdatedBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid updated_by", err.Error())
		return
	}

	cmd := &command.UpdateInsuranceCommand{
		TenantID: tenantID,
		MRN:      mrn,
		Insurance: &domain.Insurance{
			Provider:      req.Provider,
			PolicyNo:      req.PolicyNo,
			TPA:           req.TPA,
			MemberID:      req.MemberID,
			SumInsured:    req.SumInsured,
			CorporateName: req.CorporateName,
		},
		UpdatedBy: updatedBy,
	}

	patient, err := h.updateInsuranceHandler.Handle(r.Context(), cmd)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Failed to update insurance", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, patientToResponse(patient))
}

type recordAllergiesRequest struct {
	TenantID  string   `json:"tenant_id"`
	Allergies []string `json:"allergies"`
	UpdatedBy string   `json:"updated_by"`
}

func (h *PatientHandler) RecordAllergies(w http.ResponseWriter, r *http.Request) {
	mrn := mux.Vars(r)["mrn"]
	if mrn == "" {
		writeError(w, http.StatusBadRequest, "Missing MRN", "MRN is required in the path")
		return
	}

	var req recordAllergiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant_id", err.Error())
		return
	}

	updatedBy, err := uuid.Parse(req.UpdatedBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid updated_by", err.Error())
		return
	}

	cmd := &command.RecordAllergiesCommand{
		TenantID:  tenantID,
		MRN:       mrn,
		Allergies: req.Allergies,
		UpdatedBy: updatedBy,
	}

	patient, err := h.recordAllergiesHandler.Handle(r.Context(), cmd)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Failed to record allergies", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, patientToResponse(patient))
}

type medicationItem struct {
	Drug      string `json:"drug"`
	Dose      string `json:"dose"`
	Frequency string `json:"frequency"`
	Notes     string `json:"notes,omitempty"`
}

type recordMedicationsRequest struct {
	TenantID    string           `json:"tenant_id"`
	Medications []medicationItem `json:"medications"`
	UpdatedBy   string           `json:"updated_by"`
}

func (h *PatientHandler) RecordMedications(w http.ResponseWriter, r *http.Request) {
	mrn := mux.Vars(r)["mrn"]
	if mrn == "" {
		writeError(w, http.StatusBadRequest, "Missing MRN", "MRN is required in the path")
		return
	}

	var req recordMedicationsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant_id", err.Error())
		return
	}

	updatedBy, err := uuid.Parse(req.UpdatedBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid updated_by", err.Error())
		return
	}

	medications := make([]domain.Medication, 0, len(req.Medications))
	for _, m := range req.Medications {
		medications = append(medications, domain.Medication{
			Drug:      m.Drug,
			Dose:      m.Dose,
			Frequency: m.Frequency,
			Notes:     m.Notes,
		})
	}

	cmd := &command.RecordMedicationsCommand{
		TenantID:    tenantID,
		MRN:         mrn,
		Medications: medications,
		UpdatedBy:   updatedBy,
	}

	patient, err := h.recordMedicationsHandler.Handle(r.Context(), cmd)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Failed to record medications", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, patientToResponse(patient))
}

type recordVitalSignsRequest struct {
	TenantID    string   `json:"tenant_id"`
	BP          string   `json:"bp"`
	HR          *int     `json:"hr"`
	Temp        *float32 `json:"temp"`
	SpO2        *int     `json:"spo2"`
	RR          *int     `json:"rr"`
	Weight      *float32 `json:"weight"`
	Height      *int     `json:"height"`
	RecordedAt  *string  `json:"recorded_at"`
	UpdatedBy   string   `json:"updated_by"`
}

func (h *PatientHandler) RecordVitalSigns(w http.ResponseWriter, r *http.Request) {
	mrn := mux.Vars(r)["mrn"]
	if mrn == "" {
		writeError(w, http.StatusBadRequest, "Missing MRN", "MRN is required in the path")
		return
	}

	var req recordVitalSignsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant_id", err.Error())
		return
	}

	updatedBy, err := uuid.Parse(req.UpdatedBy)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid updated_by", err.Error())
		return
	}

	vitals := &domain.VitalSigns{
		BP:   req.BP,
		HR:   req.HR,
		Temp: req.Temp,
		SpO2: req.SpO2,
		RR:   req.RR,
	}

	if req.RecordedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.RecordedAt)
		if err == nil {
			vitals.RecordedAt = &t
		}
	}

	cmd := &command.RecordVitalSignsCommand{
		TenantID:   tenantID,
		MRN:        mrn,
		VitalSigns: vitals,
		UpdatedBy:  updatedBy,
	}

	patient, err := h.recordVitalSignsHandler.Handle(r.Context(), cmd)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "Failed to record vital signs", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, patientToResponse(patient))
}

func (h *PatientHandler) GetPatientByMRN(w http.ResponseWriter, r *http.Request) {
	mrn := mux.Vars(r)["mrn"]
	if mrn == "" {
		writeError(w, http.StatusBadRequest, "Missing MRN", "MRN is required in the path")
		return
	}

	tenantIDStr := r.URL.Query().Get("tenant_id")
	if tenantIDStr == "" {
		writeError(w, http.StatusBadRequest, "Missing tenant_id", "tenant_id query parameter is required")
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant_id", err.Error())
		return
	}

	q := &query.GetPatientByMRNQuery{TenantID: tenantID, MRN: mrn}
	patient, err := h.getPatientByMRNHandler.Handle(r.Context(), q)
	if err != nil {
		writeError(w, http.StatusNotFound, "Patient not found", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, patientToResponse(patient))
}

func (h *PatientHandler) ListPatients(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.URL.Query().Get("tenant_id")
	if tenantIDStr == "" {
		writeError(w, http.StatusBadRequest, "Missing tenant_id", "tenant_id query parameter is required")
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid tenant_id", err.Error())
		return
	}

	queryStr := r.URL.Query().Get("q")
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}

	var patients []*domain.Patient
	var total int64

	if queryStr != "" {
		q := &query.SearchPatientsQuery{TenantID: tenantID, Query: queryStr, Offset: offset, Limit: limit}
		patients, total, err = h.searchPatientsHandler.Handle(r.Context(), q)
	} else {
		q := &query.ListPatientsQuery{TenantID: tenantID, Offset: offset, Limit: limit}
		patients, total, err = h.listPatientsHandler.Handle(r.Context(), q)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list patients", err.Error())
		return
	}

	items := make([]any, 0, len(patients))
	for _, p := range patients {
		items = append(items, patientToSummary(p))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":  items,
		"total":  total,
		"offset": offset,
		"limit":  limit,
		"query":  queryStr,
	})
}

func patientToResponse(p *domain.Patient) map[string]any {
	if p == nil {
		return nil
	}

	result := map[string]any{
		"id":         p.ID.String(),
		"tenant_id":  p.TenantID.String(),
		"mrn":        p.MRN,
		"name":       p.Name,
		"dob":        p.DOB.Format("2006-01-02"),
		"gender":     string(p.Gender),
		"phone":      p.Phone,
		"email":      p.Email,
		"blood_group": string(p.BloodGroup),
		"status":     string(p.Status),
		"allergies":  p.Allergies,
		"created_at": p.CreatedAt.Format(time.RFC3339),
		"updated_at": p.UpdatedAt.Format(time.RFC3339),
	}

	if p.Address != nil {
		result["address"] = map[string]any{
			"street":      p.Address.Street,
			"city":        p.Address.City,
			"state":       p.Address.State,
			"postal_code": p.Address.PostalCode,
			"country":     p.Address.Country,
		}
	}

	if p.EmergencyContact != nil {
		result["emergency_contact"] = map[string]any{
			"name":     p.EmergencyContact.Name,
			"relation": p.EmergencyContact.Relation,
			"phone":    p.EmergencyContact.Phone,
		}
	}

	if p.Insurance != nil {
		result["insurance"] = map[string]any{
			"provider":       p.Insurance.Provider,
			"policy_no":      p.Insurance.PolicyNo,
			"tpa":            p.Insurance.TPA,
			"member_id":      p.Insurance.MemberID,
			"sum_insured":    p.Insurance.SumInsured,
			"corporate_name": p.Insurance.CorporateName,
		}
	}

	return result
}

func patientToSummary(p *domain.Patient) map[string]any {
	return map[string]any{
		"id":     p.ID.String(),
		"mrn":    p.MRN,
		"name":   p.Name,
		"status": string(p.Status),
		"gender": string(p.Gender),
		"phone":  p.Phone,
		"created_at": p.CreatedAt.Format(time.RFC3339),
	}
}
