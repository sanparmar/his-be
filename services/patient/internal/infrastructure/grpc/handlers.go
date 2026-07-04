package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	otcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	patientv1 "github.com/deloitte-us-consulting/his-be/api/patient/v1"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/application/command"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/application/query"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/domain"
)

// CreatePatient implements PatientService.CreatePatient RPC
func (ps *PatientServer) CreatePatient(ctx context.Context, req *patientv1.CreatePatientRequest) (*patientv1.CreatePatientResponse, error) {
	ctx, span := tracer.Start(ctx, "PatientService.CreatePatient")
	defer span.End()

	if req.TenantId == "" || req.Name == "" || req.Phone == "" || req.Gender == "" {
		span.SetStatus(otcodes.Error, "missing required fields")
		return nil, status.Error(codes.InvalidArgument, "missing required fields: tenant_id, name, phone, gender")
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		span.SetStatus(otcodes.Error, fmt.Sprintf("invalid tenant_id: %v", err))
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid tenant_id: %v", err))
	}

	dob, err := time.Parse("2006-01-02", req.Dob)
	if err != nil {
		span.SetStatus(otcodes.Error, fmt.Sprintf("invalid dob format: %v", err))
		return nil, status.Error(codes.InvalidArgument, "dob must be YYYY-MM-DD format")
	}

	userID := uuid.New() // TODO: Extract from JWT context

	address := &domain.Address{}
	if req.Address != nil {
		address = &domain.Address{
			Street:     req.Address.Street,
			City:       req.Address.City,
			State:      req.Address.State,
			PostalCode: req.Address.PostalCode,
			Country:    req.Address.Country,
		}
	}

	emergencyContact := &domain.EmergencyContact{}
	if req.EmergencyContact != nil {
		emergencyContact = &domain.EmergencyContact{
			Name:     req.EmergencyContact.Name,
			Relation: req.EmergencyContact.Relation,
			Phone:    req.EmergencyContact.Phone,
		}
	}

	var insurance *domain.Insurance
	if req.Insurance != nil {
		var policyValidTill *time.Time
		if req.Insurance.PolicyValidTill != "" {
			pvt, err := time.Parse("2006-01-02", req.Insurance.PolicyValidTill)
			if err == nil {
				policyValidTill = &pvt
			}
		}
		insurance = &domain.Insurance{
			Provider:        req.Insurance.Provider,
			PolicyNo:        req.Insurance.PolicyNo,
			TPA:             req.Insurance.Tpa,
			MemberID:        req.Insurance.MemberId,
			SumInsured:      req.Insurance.SumInsured,
			PolicyValidTill: policyValidTill,
			CorporateName:   req.Insurance.CorporateName,
		}
	}

	medications := make([]domain.Medication, len(req.HomeMedications))
	for i, med := range req.HomeMedications {
		medications[i] = domain.Medication{
			Drug:      med.Drug,
			Dose:      med.Dose,
			Frequency: med.Frequency,
			Notes:     med.Notes,
		}
	}

	cmd := command.CreatePatientCommand{
		TenantID:         tenantID,
		Name:             req.Name,
		DOB:              dob,
		Gender:           domain.Gender(req.Gender),
		Phone:            req.Phone,
		Email:            req.Email,
		BloodGroup:       domain.BloodGroup(req.BloodGroup),
		Address:          address,
		EmergencyContact: emergencyContact,
		Insurance:        insurance,
		Allergies:        req.Allergies,
		HomeMedications:  medications,
		CreatedBy:        userID,
	}

	patient, err := ps.createPatientHandler.Handle(ctx, &cmd)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create patient: %v", err))
	}

	span.SetAttributes(
		attribute.String("patient.id", patient.ID.String()),
		attribute.String("patient.mrn", patient.MRN),
	)

	resp := &patientv1.CreatePatientResponse{
		Patient: domainPatientToProto(patient),
		Mrn:     patient.MRN,
	}

	return resp, nil
}

// GetPatient implements PatientService.GetPatient RPC
func (ps *PatientServer) GetPatient(ctx context.Context, req *patientv1.GetPatientRequest) (*patientv1.GetPatientResponse, error) {
	ctx, span := tracer.Start(ctx, "PatientService.GetPatient")
	defer span.End()

	if req.TenantId == "" || req.PatientId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields: patient_id, tenant_id")
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid tenant_id: %v", err))
	}

	q := query.GetPatientByMRNQuery{
		TenantID: tenantID,
		MRN:      req.PatientId,
	}

	patient, err := ps.getPatientByMRNHandler.Handle(ctx, &q)
	if err != nil {
		if err == domain.ErrNotFound {
			span.SetStatus(otcodes.Error, "patient not found")
			return nil, status.Error(codes.NotFound, "patient not found")
		}
		span.RecordError(err)
		span.SetStatus(otcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get patient: %v", err))
	}

	span.SetAttributes(
		attribute.String("patient.id", patient.ID.String()),
		attribute.String("patient.mrn", patient.MRN),
	)

	resp := &patientv1.GetPatientResponse{
		Patient: domainPatientToProto(patient),
	}

	return resp, nil
}

// SearchPatients implements PatientService.SearchPatients RPC
func (ps *PatientServer) SearchPatients(ctx context.Context, req *patientv1.SearchPatientsRequest) (*patientv1.SearchPatientsResponse, error) {
	ctx, span := tracer.Start(ctx, "PatientService.SearchPatients")
	defer span.End()

	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing tenant_id")
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid tenant_id: %v", err))
	}

	searchTerm := ""
	if req.Mrn != "" {
		searchTerm = req.Mrn
	} else if req.FirstName != "" || req.LastName != "" {
		searchTerm = req.FirstName + " " + req.LastName
	} else if req.Phone != "" {
		searchTerm = req.Phone
	}

	offset := 0
	limit := int(req.PageSize)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if req.Page > 1 {
		offset = int((req.Page - 1) * req.PageSize)
	}

	q := query.SearchPatientsQuery{
		TenantID: tenantID,
		Query:    searchTerm,
		Offset:   offset,
		Limit:    limit,
	}

	patients, total, err := ps.searchPatientsHandler.Handle(ctx, &q)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to search patients: %v", err))
	}

	protoPatients := make([]*patientv1.Patient, len(patients))
	for i, p := range patients {
		protoPatients[i] = domainPatientToProto(p)
	}

	span.SetAttributes(attribute.Int64("results.count", total))

	resp := &patientv1.SearchPatientsResponse{
		Patients: protoPatients,
		Total:    total,
		Page:     req.Page,
		PageSize: int32(limit),
	}

	return resp, nil
}

// ListPatients implements PatientService.ListPatients RPC
func (ps *PatientServer) ListPatients(ctx context.Context, req *patientv1.ListPatientsRequest) (*patientv1.ListPatientsResponse, error) {
	ctx, span := tracer.Start(ctx, "PatientService.ListPatients")
	defer span.End()

	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing tenant_id")
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid tenant_id: %v", err))
	}

	limit := int(req.PageSize)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := 0
	if req.Page > 1 {
		offset = int((req.Page - 1) * req.PageSize)
	}

	q := query.ListPatientsQuery{
		TenantID: tenantID,
		Offset:   offset,
		Limit:    limit,
	}

	patients, total, err := ps.listPatientsHandler.Handle(ctx, &q)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list patients: %v", err))
	}

	protoPatients := make([]*patientv1.Patient, len(patients))
	for i, p := range patients {
		protoPatients[i] = domainPatientToProto(p)
	}

	span.SetAttributes(attribute.Int64("results.count", total))

	resp := &patientv1.ListPatientsResponse{
		Patients: protoPatients,
		Total:    total,
		Page:     req.Page,
		PageSize: int32(limit),
	}

	return resp, nil
}

// UpdatePatient implements PatientService.UpdatePatient RPC
func (ps *PatientServer) UpdatePatient(ctx context.Context, req *patientv1.UpdatePatientRequest) (*patientv1.UpdatePatientResponse, error) {
	ctx, span := tracer.Start(ctx, "PatientService.UpdatePatient")
	defer span.End()

	if req.TenantId == "" || req.PatientId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing required fields: patient_id, tenant_id")
	}

	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid tenant_id: %v", err))
	}

	updatedBy := uuid.New() // TODO: Extract from JWT context

	if req.Name != "" || req.Phone != "" || req.Email != "" || req.Address != nil {
		var address *domain.Address
		if req.Address != nil {
			address = &domain.Address{
				Street:     req.Address.Street,
				City:       req.Address.City,
				State:      req.Address.State,
				PostalCode: req.Address.PostalCode,
				Country:    req.Address.Country,
			}
		}

		cmd := command.UpdateDemographicsCommand{
			TenantID:  tenantID,
			MRN:       req.PatientId,
			Name:      req.Name,
			Phone:     req.Phone,
			Email:     req.Email,
			Address:   address,
			UpdatedBy: updatedBy,
		}

		_, err := ps.updateDemographicsHandler.Handle(ctx, &cmd)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otcodes.Error, err.Error())
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update patient demographics: %v", err))
		}
	}

	if req.Insurance != nil {
		var policyValidTill *time.Time
		if req.Insurance.PolicyValidTill != "" {
			pvt, err := time.Parse("2006-01-02", req.Insurance.PolicyValidTill)
			if err == nil {
				policyValidTill = &pvt
			}
		}

		insurance := &domain.Insurance{
			Provider:        req.Insurance.Provider,
			PolicyNo:        req.Insurance.PolicyNo,
			TPA:             req.Insurance.Tpa,
			MemberID:        req.Insurance.MemberId,
			SumInsured:      req.Insurance.SumInsured,
			PolicyValidTill: policyValidTill,
			CorporateName:   req.Insurance.CorporateName,
		}

		cmd := command.UpdateInsuranceCommand{
			TenantID:  tenantID,
			MRN:       req.PatientId,
			Insurance: insurance,
			UpdatedBy: updatedBy,
		}

		_, err := ps.updateInsuranceHandler.Handle(ctx, &cmd)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otcodes.Error, err.Error())
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update patient insurance: %v", err))
		}
	}

	if len(req.Allergies) > 0 {
		cmd := command.RecordAllergiesCommand{
			TenantID:  tenantID,
			MRN:       req.PatientId,
			Allergies: req.Allergies,
			UpdatedBy: updatedBy,
		}

		_, err := ps.recordAllergiesHandler.Handle(ctx, &cmd)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otcodes.Error, err.Error())
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record allergies: %v", err))
		}
	}

	if len(req.HomeMedications) > 0 {
		medications := make([]domain.Medication, len(req.HomeMedications))
		for i, med := range req.HomeMedications {
			medications[i] = domain.Medication{
				Drug:      med.Drug,
				Dose:      med.Dose,
				Frequency: med.Frequency,
				Notes:     med.Notes,
			}
		}

		cmd := command.RecordMedicationsCommand{
			TenantID:    tenantID,
			MRN:         req.PatientId,
			Medications: medications,
			UpdatedBy:   updatedBy,
		}

		_, err := ps.recordMedicationsHandler.Handle(ctx, &cmd)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otcodes.Error, err.Error())
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record medications: %v", err))
		}
	}

	if req.VitalSigns != nil {
		vitals := &domain.VitalSigns{
			BP: req.VitalSigns.Bp,
		}

		if req.VitalSigns.Hr > 0 {
			hr := int(req.VitalSigns.Hr)
			vitals.HR = &hr
		}
		if req.VitalSigns.Temp > 0 {
			vitals.Temp = &req.VitalSigns.Temp
		}
		if req.VitalSigns.Spo2 > 0 {
			spo2 := int(req.VitalSigns.Spo2)
			vitals.SpO2 = &spo2
		}
		if req.VitalSigns.Rr > 0 {
			rr := int(req.VitalSigns.Rr)
			vitals.RR = &rr
		}
		if req.VitalSigns.Weight > 0 {
			vitals.Weight = &req.VitalSigns.Weight
		}
		if req.VitalSigns.Height > 0 {
			height := int(req.VitalSigns.Height)
			vitals.Height = &height
		}
		if req.VitalSigns.Bmi > 0 {
			vitals.BMI = &req.VitalSigns.Bmi
		}
		if req.VitalSigns.RecordedAt != nil {
			recordedAt := *req.VitalSigns.RecordedAt
			vitals.RecordedAt = &recordedAt
		}

		cmd := command.RecordVitalSignsCommand{
			TenantID:   tenantID,
			MRN:        req.PatientId,
			VitalSigns: vitals,
			UpdatedBy:  updatedBy,
		}

		_, err := ps.recordVitalSignsHandler.Handle(ctx, &cmd)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otcodes.Error, err.Error())
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record vital signs: %v", err))
		}
	}

	q := query.GetPatientByMRNQuery{
		TenantID: tenantID,
		MRN:      req.PatientId,
	}

	patient, err := ps.getPatientByMRNHandler.Handle(ctx, &q)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(otcodes.Error, err.Error())
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to fetch updated patient: %v", err))
	}

	span.SetAttributes(attribute.String("patient.id", patient.ID.String()))

	resp := &patientv1.UpdatePatientResponse{
		Patient: domainPatientToProto(patient),
	}

	return resp, nil
}

// ActivatePatient implements PatientService.ActivatePatient RPC
func (ps *PatientServer) ActivatePatient(ctx context.Context, req *patientv1.ActivatePatientRequest) (*patientv1.ActivatePatientResponse, error) {
	ctx, span := tracer.Start(ctx, "PatientService.ActivatePatient")
	defer span.End()

	// TODO: Implement activation logic
	span.SetStatus(otcodes.Error, "not implemented")
	return nil, status.Error(codes.Unimplemented, "ActivatePatient not yet implemented")
}

// Health implements PatientService.Health RPC
func (ps *PatientServer) Health(ctx context.Context, _ *emptypb.Empty) (*patientv1.HealthResponse, error) {
	ctx, span := tracer.Start(ctx, "PatientService.Health")
	defer span.End()

	resp := &patientv1.HealthResponse{
		Status: "SERVING",
	}

	return resp, nil
}

// domainPatientToProto converts domain.Patient to proto Patient
func domainPatientToProto(p *domain.Patient) *patientv1.Patient {
	pp := &patientv1.Patient{
		Id:              p.ID.String(),
		TenantId:        p.TenantID.String(),
		Mrn:             p.MRN,
		Name:            p.Name,
		Dob:             p.DOB.Format("2006-01-02"),
		Gender:          string(p.Gender),
		BloodGroup:      string(p.BloodGroup),
		Phone:           p.Phone,
		Email:           p.Email,
		Abha:            p.ABHA,
		AadhaarLastFour: p.AadhaarLast4,
		Status:          string(p.Status),
		CreatedBy:       p.CreatedBy.String(),
		UpdatedBy:       p.UpdatedBy.String(),
	}

	// Convert timestamps
	if !p.CreatedAt.IsZero() {
		pp.CreatedAt = &p.CreatedAt
	}
	if !p.UpdatedAt.IsZero() {
		pp.UpdatedAt = &p.UpdatedAt
	}
	if p.DeletedAt != nil {
		pp.DeletedAt = p.DeletedAt
	}

	if p.Address != nil {
		pp.Address = &patientv1.Address{
			Street:     p.Address.Street,
			City:       p.Address.City,
			State:      p.Address.State,
			PostalCode: p.Address.PostalCode,
			Country:    p.Address.Country,
		}
	}

	if p.EmergencyContact != nil {
		pp.EmergencyContact = &patientv1.EmergencyContact{
			Name:     p.EmergencyContact.Name,
			Relation: p.EmergencyContact.Relation,
			Phone:    p.EmergencyContact.Phone,
		}
	}

	if p.Insurance != nil {
		pp.Insurance = &patientv1.Insurance{
			Provider:      p.Insurance.Provider,
			PolicyNo:      p.Insurance.PolicyNo,
			Tpa:           p.Insurance.TPA,
			MemberId:      p.Insurance.MemberID,
			SumInsured:    p.Insurance.SumInsured,
			CorporateName: p.Insurance.CorporateName,
		}
		if p.Insurance.PolicyValidTill != nil {
			pp.Insurance.PolicyValidTill = p.Insurance.PolicyValidTill.Format("2006-01-02")
		}
	}

	pp.Allergies = p.Allergies

	for _, med := range p.HomeMedications {
		pp.HomeMedications = append(pp.HomeMedications, &patientv1.Medication{
			Drug:      med.Drug,
			Dose:      med.Dose,
			Frequency: med.Frequency,
			Notes:     med.Notes,
		})
	}

	if p.VitalSigns != nil {
		vs := &patientv1.VitalSigns{
			Bp: p.VitalSigns.BP,
		}
		if p.VitalSigns.HR != nil {
			vs.Hr = int32(*p.VitalSigns.HR)
		}
		if p.VitalSigns.Temp != nil {
			vs.Temp = *p.VitalSigns.Temp
		}
		if p.VitalSigns.SpO2 != nil {
			vs.Spo2 = int32(*p.VitalSigns.SpO2)
		}
		if p.VitalSigns.RR != nil {
			vs.Rr = int32(*p.VitalSigns.RR)
		}
		if p.VitalSigns.Weight != nil {
			vs.Weight = *p.VitalSigns.Weight
		}
		if p.VitalSigns.Height != nil {
			vs.Height = int32(*p.VitalSigns.Height)
		}
		if p.VitalSigns.BMI != nil {
			vs.Bmi = *p.VitalSigns.BMI
		}
		if p.VitalSigns.RecordedAt != nil && !p.VitalSigns.RecordedAt.IsZero() {
			vs.RecordedAt = p.VitalSigns.RecordedAt
		}
		pp.VitalSigns = vs
	}

	return pp
}
