package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/deloitte-us-consulting/his-be/services/patient/internal/domain"
)

// PatientRepository implements the domain.PatientRepository interface using PostgreSQL.
type PatientRepository struct {
	db *sqlx.DB
}

// NewPatientRepository creates a new PostgreSQL patient repository.
func NewPatientRepository(db *sqlx.DB) *PatientRepository {
	return &PatientRepository{db: db}
}

// dbPatient is the database model mapping to patients table (47 columns)
type dbPatient struct {
	ID                       uuid.UUID       `db:"id"`
	TenantID                 uuid.UUID       `db:"tenant_id"`
	MRN                      string          `db:"mrn"`
	Name                     string          `db:"name"`
	DOB                      time.Time       `db:"dob"`
	Gender                   string          `db:"gender"`
	Phone                    string          `db:"phone"`
	Email                    string          `db:"email"`
	BloodGroup               string          `db:"blood_group"`
	ABHA                     sql.NullString  `db:"abha"`
	AadhaarLastFour          sql.NullString  `db:"aadhaar_last_four"`
	AddressStreet            sql.NullString  `db:"address_street"`
	AddressCity              sql.NullString  `db:"address_city"`
	AddressState             sql.NullString  `db:"address_state"`
	AddressPostalCode        sql.NullString  `db:"address_postal_code"`
	AddressCountry           sql.NullString  `db:"address_country"`
	EmergencyContactName     sql.NullString  `db:"emergency_contact_name"`
	EmergencyContactRelation sql.NullString  `db:"emergency_contact_relation"`
	EmergencyContactPhone    sql.NullString  `db:"emergency_contact_phone"`
	InsuranceProvider        sql.NullString  `db:"insurance_provider"`
	InsurancePolicyNo        sql.NullString  `db:"insurance_policy_no"`
	InsuranceTPA             sql.NullString  `db:"insurance_tpa"`
	InsuranceMemberID        sql.NullString  `db:"insurance_member_id"`
	InsuranceSumInsured      sql.NullInt64   `db:"insurance_sum_insured"`
	InsurancePolicyValidTill sql.NullTime    `db:"insurance_policy_valid_till"`
	InsuranceCorporateName   sql.NullString  `db:"insurance_corporate_name"`
	HomeMedications          pq.StringArray  `db:"home_medications"`
	Allergies                pq.StringArray  `db:"allergies"`
	VitalSignsBP             sql.NullString  `db:"vital_signs_bp"`
	VitalSignsHR             sql.NullInt64   `db:"vital_signs_hr"`
	VitalSignsTemp           sql.NullFloat64 `db:"vital_signs_temp"`
	VitalSignsSPO2           sql.NullInt64   `db:"vital_signs_spo2"`
	VitalSignsRR             sql.NullInt64   `db:"vital_signs_rr"`
	VitalSignsWeight         sql.NullFloat64 `db:"vital_signs_weight"`
	VitalSignsHeight         sql.NullInt64   `db:"vital_signs_height"`
	VitalSignsBMI            sql.NullFloat64 `db:"vital_signs_bmi"`
	VitalSignsRecordedAt     sql.NullTime    `db:"vital_signs_recorded_at"`
	Status                   string          `db:"status"`
	CreatedAt                time.Time       `db:"created_at"`
	UpdatedAt                time.Time       `db:"updated_at"`
	DeletedAt                sql.NullTime    `db:"deleted_at"`
	CreatedBy                uuid.UUID       `db:"created_by"`
	UpdatedBy                uuid.UUID       `db:"updated_by"`
}

// toDomainPatient converts database model to domain entity
func (dbp *dbPatient) toDomainPatient() *domain.Patient {
	p := &domain.Patient{
		ID:         dbp.ID,
		TenantID:   dbp.TenantID,
		MRN:        dbp.MRN,
		Name:       dbp.Name,
		DOB:        dbp.DOB,
		Gender:     domain.Gender(dbp.Gender),
		Phone:      dbp.Phone,
		Email:      dbp.Email,
		BloodGroup: domain.BloodGroup(dbp.BloodGroup),
		Status:     domain.PatientStatus(dbp.Status),
		CreatedAt:  dbp.CreatedAt,
		UpdatedAt:  dbp.UpdatedAt,
		CreatedBy:  dbp.CreatedBy,
		UpdatedBy:  dbp.UpdatedBy,
	}

	if dbp.DeletedAt.Valid {
		p.DeletedAt = &dbp.DeletedAt.Time
	}

	// Health IDs
	if dbp.ABHA.Valid {
		p.ABHA = dbp.ABHA.String
	}
	if dbp.AadhaarLastFour.Valid {
		p.AadhaarLast4 = dbp.AadhaarLastFour.String
	}

	// Address
	if dbp.AddressStreet.Valid {
		p.Address = &domain.Address{
			Street:     dbp.AddressStreet.String,
			City:       dbp.AddressCity.String,
			State:      dbp.AddressState.String,
			PostalCode: dbp.AddressPostalCode.String,
			Country:    dbp.AddressCountry.String,
		}
	}

	// Emergency Contact
	if dbp.EmergencyContactName.Valid {
		p.EmergencyContact = &domain.EmergencyContact{
			Name:     dbp.EmergencyContactName.String,
			Relation: dbp.EmergencyContactRelation.String,
			Phone:    dbp.EmergencyContactPhone.String,
		}
	}

	// Insurance
	if dbp.InsuranceProvider.Valid {
		p.Insurance = &domain.Insurance{
			Provider:      dbp.InsuranceProvider.String,
			PolicyNo:      dbp.InsurancePolicyNo.String,
			TPA:           dbp.InsuranceTPA.String,
			MemberID:      dbp.InsuranceMemberID.String,
			CorporateName: dbp.InsuranceCorporateName.String,
		}
		if dbp.InsuranceSumInsured.Valid {
			p.Insurance.SumInsured = dbp.InsuranceSumInsured.Int64
		}
		if dbp.InsurancePolicyValidTill.Valid {
			p.Insurance.PolicyValidTill = &dbp.InsurancePolicyValidTill.Time
		}
	}

	// Allergies
	if dbp.Allergies != nil && len(dbp.Allergies) > 0 {
		p.Allergies = dbp.Allergies
	}

	// Home Medications
	if dbp.HomeMedications != nil && len(dbp.HomeMedications) > 0 {
		for _, medJSON := range dbp.HomeMedications {
			var med domain.Medication
			if err := json.Unmarshal([]byte(medJSON), &med); err == nil {
				p.HomeMedications = append(p.HomeMedications, med)
			}
		}
	}

	// Vital Signs
	if dbp.VitalSignsBP.Valid {
		hr := int(dbp.VitalSignsHR.Int64)
		temp := float32(dbp.VitalSignsTemp.Float64)
		spo2 := int(dbp.VitalSignsSPO2.Int64)
		rr := int(dbp.VitalSignsRR.Int64)
		weight := float32(dbp.VitalSignsWeight.Float64)
		height := int(dbp.VitalSignsHeight.Int64)
		bmi := float32(dbp.VitalSignsBMI.Float64)

		p.VitalSigns = &domain.VitalSigns{
			BP:     dbp.VitalSignsBP.String,
			HR:     &hr,
			Temp:   &temp,
			SpO2:   &spo2,
			RR:     &rr,
			Weight: &weight,
			Height: &height,
			BMI:    &bmi,
		}
		if dbp.VitalSignsRecordedAt.Valid {
			p.VitalSigns.RecordedAt = &dbp.VitalSignsRecordedAt.Time
		}
	}

	return p
}

// fromDomainPatient converts domain entity to database model
func fromDomainPatient(p *domain.Patient) *dbPatient {
	dbp := &dbPatient{
		ID:         p.ID,
		TenantID:   p.TenantID,
		MRN:        p.MRN,
		Name:       p.Name,
		DOB:        p.DOB,
		Gender:     string(p.Gender),
		Phone:      p.Phone,
		Email:      p.Email,
		BloodGroup: string(p.BloodGroup),
		Status:     string(p.Status),
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
		CreatedBy:  p.CreatedBy,
		UpdatedBy:  p.UpdatedBy,
	}

	if p.DeletedAt != nil {
		dbp.DeletedAt = sql.NullTime{Time: *p.DeletedAt, Valid: true}
	}

	if p.ABHA != "" {
		dbp.ABHA = sql.NullString{String: p.ABHA, Valid: true}
	}
	if p.AadhaarLast4 != "" {
		dbp.AadhaarLastFour = sql.NullString{String: p.AadhaarLast4, Valid: true}
	}

	if p.Address != nil {
		dbp.AddressStreet = sql.NullString{String: p.Address.Street, Valid: true}
		dbp.AddressCity = sql.NullString{String: p.Address.City, Valid: true}
		dbp.AddressState = sql.NullString{String: p.Address.State, Valid: true}
		dbp.AddressPostalCode = sql.NullString{String: p.Address.PostalCode, Valid: true}
		dbp.AddressCountry = sql.NullString{String: p.Address.Country, Valid: true}
	}

	if p.EmergencyContact != nil {
		dbp.EmergencyContactName = sql.NullString{String: p.EmergencyContact.Name, Valid: true}
		dbp.EmergencyContactRelation = sql.NullString{String: p.EmergencyContact.Relation, Valid: true}
		dbp.EmergencyContactPhone = sql.NullString{String: p.EmergencyContact.Phone, Valid: true}
	}

	if p.Insurance != nil {
		dbp.InsuranceProvider = sql.NullString{String: p.Insurance.Provider, Valid: true}
		dbp.InsurancePolicyNo = sql.NullString{String: p.Insurance.PolicyNo, Valid: true}
		dbp.InsuranceTPA = sql.NullString{String: p.Insurance.TPA, Valid: true}
		dbp.InsuranceMemberID = sql.NullString{String: p.Insurance.MemberID, Valid: true}
		if p.Insurance.SumInsured > 0 {
			dbp.InsuranceSumInsured = sql.NullInt64{Int64: p.Insurance.SumInsured, Valid: true}
		}
		dbp.InsuranceCorporateName = sql.NullString{String: p.Insurance.CorporateName, Valid: true}
		if p.Insurance.PolicyValidTill != nil {
			dbp.InsurancePolicyValidTill = sql.NullTime{Time: *p.Insurance.PolicyValidTill, Valid: true}
		}
	}

	if len(p.Allergies) > 0 {
		dbp.Allergies = p.Allergies
	}

	if len(p.HomeMedications) > 0 {
		meds := make([]string, len(p.HomeMedications))
		for i, med := range p.HomeMedications {
			if data, err := json.Marshal(med); err == nil {
				meds[i] = string(data)
			}
		}
		dbp.HomeMedications = meds
	}

	if p.VitalSigns != nil {
		dbp.VitalSignsBP = sql.NullString{String: p.VitalSigns.BP, Valid: true}
		if p.VitalSigns.HR != nil {
			dbp.VitalSignsHR = sql.NullInt64{Int64: int64(*p.VitalSigns.HR), Valid: true}
		}
		if p.VitalSigns.Temp != nil {
			dbp.VitalSignsTemp = sql.NullFloat64{Float64: float64(*p.VitalSigns.Temp), Valid: true}
		}
		if p.VitalSigns.SpO2 != nil {
			dbp.VitalSignsSPO2 = sql.NullInt64{Int64: int64(*p.VitalSigns.SpO2), Valid: true}
		}
		if p.VitalSigns.RR != nil {
			dbp.VitalSignsRR = sql.NullInt64{Int64: int64(*p.VitalSigns.RR), Valid: true}
		}
		if p.VitalSigns.Weight != nil {
			dbp.VitalSignsWeight = sql.NullFloat64{Float64: float64(*p.VitalSigns.Weight), Valid: true}
		}
		if p.VitalSigns.Height != nil {
			dbp.VitalSignsHeight = sql.NullInt64{Int64: int64(*p.VitalSigns.Height), Valid: true}
		}
		if p.VitalSigns.BMI != nil {
			dbp.VitalSignsBMI = sql.NullFloat64{Float64: float64(*p.VitalSigns.BMI), Valid: true}
		}
		if p.VitalSigns.RecordedAt != nil {
			dbp.VitalSignsRecordedAt = sql.NullTime{Time: *p.VitalSigns.RecordedAt, Valid: true}
		}
	}

	return dbp
}

// Create inserts a new patient into the database
func (r *PatientRepository) Create(ctx context.Context, patient *domain.Patient) error {
	dbp := fromDomainPatient(patient)

	query := `
		INSERT INTO patients (
			id, tenant_id, mrn, name, dob, gender, phone, email, blood_group,
			abha, aadhaar_last_four,
			address_street, address_city, address_state, address_postal_code, address_country,
			emergency_contact_name, emergency_contact_relation, emergency_contact_phone,
			insurance_provider, insurance_policy_no, insurance_tpa, insurance_member_id,
			insurance_sum_insured, insurance_policy_valid_till, insurance_corporate_name,
			home_medications, allergies,
			vital_signs_bp, vital_signs_hr, vital_signs_temp, vital_signs_spo2, vital_signs_rr,
			vital_signs_weight, vital_signs_height, vital_signs_bmi, vital_signs_recorded_at,
			status, created_at, updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10, $11,
			$12, $13, $14, $15, $16,
			$17, $18, $19,
			$20, $21, $22, $23,
			$24, $25, $26,
			$27, $28,
			$29, $30, $31, $32, $33,
			$34, $35, $36, $37,
			$38, $39, $40, $41, $42
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		dbp.ID, dbp.TenantID, dbp.MRN, dbp.Name, dbp.DOB, dbp.Gender, dbp.Phone, dbp.Email, dbp.BloodGroup,
		dbp.ABHA, dbp.AadhaarLastFour,
		dbp.AddressStreet, dbp.AddressCity, dbp.AddressState, dbp.AddressPostalCode, dbp.AddressCountry,
		dbp.EmergencyContactName, dbp.EmergencyContactRelation, dbp.EmergencyContactPhone,
		dbp.InsuranceProvider, dbp.InsurancePolicyNo, dbp.InsuranceTPA, dbp.InsuranceMemberID,
		dbp.InsuranceSumInsured, dbp.InsurancePolicyValidTill, dbp.InsuranceCorporateName,
		dbp.HomeMedications, dbp.Allergies,
		dbp.VitalSignsBP, dbp.VitalSignsHR, dbp.VitalSignsTemp, dbp.VitalSignsSPO2, dbp.VitalSignsRR,
		dbp.VitalSignsWeight, dbp.VitalSignsHeight, dbp.VitalSignsBMI, dbp.VitalSignsRecordedAt,
		dbp.Status, dbp.CreatedAt, dbp.UpdatedAt, dbp.CreatedBy, dbp.UpdatedBy,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return domain.ErrAlreadyExists
		}
		return fmt.Errorf("create patient: %w", err)
	}

	return nil
}

// GetByID retrieves a patient by ID with tenant isolation
func (r *PatientRepository) GetByID(ctx context.Context, tenantID, patientID uuid.UUID) (*domain.Patient, error) {
	dbp := &dbPatient{}

	query := `
		SELECT
			id, tenant_id, mrn, name, dob, gender, phone, email, blood_group,
			abha, aadhaar_last_four,
			address_street, address_city, address_state, address_postal_code, address_country,
			emergency_contact_name, emergency_contact_relation, emergency_contact_phone,
			insurance_provider, insurance_policy_no, insurance_tpa, insurance_member_id,
			insurance_sum_insured, insurance_policy_valid_till, insurance_corporate_name,
			home_medications, allergies,
			vital_signs_bp, vital_signs_hr, vital_signs_temp, vital_signs_spo2, vital_signs_rr,
			vital_signs_weight, vital_signs_height, vital_signs_bmi, vital_signs_recorded_at,
			status, created_at, updated_at, deleted_at, created_by, updated_by
		FROM patients
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, dbp, query, patientID, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get patient by id: %w", err)
	}

	return dbp.toDomainPatient(), nil
}

// GetByMRN retrieves a patient by MRN with tenant isolation
func (r *PatientRepository) GetByMRN(ctx context.Context, tenantID uuid.UUID, mrn string) (*domain.Patient, error) {
	dbp := &dbPatient{}

	query := `
		SELECT
			id, tenant_id, mrn, name, dob, gender, phone, email, blood_group,
			abha, aadhaar_last_four,
			address_street, address_city, address_state, address_postal_code, address_country,
			emergency_contact_name, emergency_contact_relation, emergency_contact_phone,
			insurance_provider, insurance_policy_no, insurance_tpa, insurance_member_id,
			insurance_sum_insured, insurance_policy_valid_till, insurance_corporate_name,
			home_medications, allergies,
			vital_signs_bp, vital_signs_hr, vital_signs_temp, vital_signs_spo2, vital_signs_rr,
			vital_signs_weight, vital_signs_height, vital_signs_bmi, vital_signs_recorded_at,
			status, created_at, updated_at, deleted_at, created_by, updated_by
		FROM patients
		WHERE mrn = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`

	err := r.db.GetContext(ctx, dbp, query, mrn, tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get patient by mrn: %w", err)
	}

	return dbp.toDomainPatient(), nil
}

// Update updates an existing patient
func (r *PatientRepository) Update(ctx context.Context, patient *domain.Patient) error {
	dbp := fromDomainPatient(patient)

	query := `
		UPDATE patients
		SET name = $1, phone = $2, email = $3, blood_group = $4,
			abha = $5, aadhaar_last_four = $6,
			address_street = $7, address_city = $8, address_state = $9, address_postal_code = $10, address_country = $11,
			emergency_contact_name = $12, emergency_contact_relation = $13, emergency_contact_phone = $14,
			insurance_provider = $15, insurance_policy_no = $16, insurance_tpa = $17, insurance_member_id = $18,
			insurance_sum_insured = $19, insurance_policy_valid_till = $20, insurance_corporate_name = $21,
			home_medications = $22, allergies = $23,
			vital_signs_bp = $24, vital_signs_hr = $25, vital_signs_temp = $26, vital_signs_spo2 = $27, vital_signs_rr = $28,
			vital_signs_weight = $29, vital_signs_height = $30, vital_signs_bmi = $31, vital_signs_recorded_at = $32,
			status = $33, updated_at = $34, updated_by = $35
		WHERE id = $36 AND tenant_id = $37
	`

	result, err := r.db.ExecContext(ctx, query,
		dbp.Name, dbp.Phone, dbp.Email, dbp.BloodGroup,
		dbp.ABHA, dbp.AadhaarLastFour,
		dbp.AddressStreet, dbp.AddressCity, dbp.AddressState, dbp.AddressPostalCode, dbp.AddressCountry,
		dbp.EmergencyContactName, dbp.EmergencyContactRelation, dbp.EmergencyContactPhone,
		dbp.InsuranceProvider, dbp.InsurancePolicyNo, dbp.InsuranceTPA, dbp.InsuranceMemberID,
		dbp.InsuranceSumInsured, dbp.InsurancePolicyValidTill, dbp.InsuranceCorporateName,
		dbp.HomeMedications, dbp.Allergies,
		dbp.VitalSignsBP, dbp.VitalSignsHR, dbp.VitalSignsTemp, dbp.VitalSignsSPO2, dbp.VitalSignsRR,
		dbp.VitalSignsWeight, dbp.VitalSignsHeight, dbp.VitalSignsBMI, dbp.VitalSignsRecordedAt,
		dbp.Status, dbp.UpdatedAt, dbp.UpdatedBy,
		dbp.ID, dbp.TenantID,
	)

	if err != nil {
		return fmt.Errorf("update patient: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update patient: %w", err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete performs soft delete (sets deleted_at) on a patient
func (r *PatientRepository) Delete(ctx context.Context, tenantID, patientID uuid.UUID, deletedBy uuid.UUID) error {
	now := time.Now().UTC()

	query := `
		UPDATE patients
		SET deleted_at = $1, status = $2, updated_at = $3, updated_by = $4
		WHERE id = $5 AND tenant_id = $6
	`

	result, err := r.db.ExecContext(ctx, query, now, domain.PatientStatusArchived, now, deletedBy, patientID, tenantID)
	if err != nil {
		return fmt.Errorf("delete patient: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete patient: %w", err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// List retrieves patients with pagination (excludes deleted)
func (r *PatientRepository) List(ctx context.Context, tenantID uuid.UUID, offset, limit int, statusFilter *domain.PatientStatus) ([]*domain.Patient, int64, error) {
	var patients []*domain.Patient
	var count int64

	countQuery := "SELECT COUNT(*) FROM patients WHERE tenant_id = $1 AND deleted_at IS NULL"
	countArgs := []interface{}{tenantID}

	if statusFilter != nil {
		countQuery += " AND status = $2"
		countArgs = append(countArgs, string(*statusFilter))
	}

	if err := r.db.GetContext(ctx, &count, countQuery, countArgs...); err != nil {
		return nil, 0, fmt.Errorf("list patients (count): %w", err)
	}

	fetchQuery := `
		SELECT
			id, tenant_id, mrn, name, dob, gender, phone, email, blood_group,
			abha, aadhaar_last_four,
			address_street, address_city, address_state, address_postal_code, address_country,
			emergency_contact_name, emergency_contact_relation, emergency_contact_phone,
			insurance_provider, insurance_policy_no, insurance_tpa, insurance_member_id,
			insurance_sum_insured, insurance_policy_valid_till, insurance_corporate_name,
			home_medications, allergies,
			vital_signs_bp, vital_signs_hr, vital_signs_temp, vital_signs_spo2, vital_signs_rr,
			vital_signs_weight, vital_signs_height, vital_signs_bmi, vital_signs_recorded_at,
			status, created_at, updated_at, deleted_at, created_by, updated_by
		FROM patients
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`
	fetchArgs := []interface{}{tenantID}
	argNum := 2

	if statusFilter != nil {
		fetchQuery += fmt.Sprintf(" AND status = $%d", argNum)
		fetchArgs = append(fetchArgs, string(*statusFilter))
		argNum++
	}

	fetchQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	fetchArgs = append(fetchArgs, limit, offset)

	var dbPatients []*dbPatient
	if err := r.db.SelectContext(ctx, &dbPatients, fetchQuery, fetchArgs...); err != nil {
		return nil, 0, fmt.Errorf("list patients: %w", err)
	}

	for _, dbp := range dbPatients {
		patients = append(patients, dbp.toDomainPatient())
	}

	return patients, count, nil
}

// Search performs full-text search on patient records (name, MRN, phone)
func (r *PatientRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, offset, limit int) ([]*domain.Patient, int64, error) {
	var patients []*domain.Patient
	var count int64

	searchTerm := "%" + query + "%"

	countQuery := `
		SELECT COUNT(*) FROM patients
		WHERE tenant_id = $1 AND deleted_at IS NULL
		AND (name ILIKE $2 OR mrn ILIKE $2 OR phone ILIKE $2)
	`

	if err := r.db.GetContext(ctx, &count, countQuery, tenantID, searchTerm); err != nil {
		return nil, 0, fmt.Errorf("search patients (count): %w", err)
	}

	fetchQuery := `
		SELECT
			id, tenant_id, mrn, name, dob, gender, phone, email, blood_group,
			abha, aadhaar_last_four,
			address_street, address_city, address_state, address_postal_code, address_country,
			emergency_contact_name, emergency_contact_relation, emergency_contact_phone,
			insurance_provider, insurance_policy_no, insurance_tpa, insurance_member_id,
			insurance_sum_insured, insurance_policy_valid_till, insurance_corporate_name,
			home_medications, allergies,
			vital_signs_bp, vital_signs_hr, vital_signs_temp, vital_signs_spo2, vital_signs_rr,
			vital_signs_weight, vital_signs_height, vital_signs_bmi, vital_signs_recorded_at,
			status, created_at, updated_at, deleted_at, created_by, updated_by
		FROM patients
		WHERE tenant_id = $1 AND deleted_at IS NULL
		AND (name ILIKE $2 OR mrn ILIKE $2 OR phone ILIKE $2)
		ORDER BY created_at DESC LIMIT $3 OFFSET $4
	`

	var dbPatients []*dbPatient
	if err := r.db.SelectContext(ctx, &dbPatients, fetchQuery, tenantID, searchTerm, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("search patients: %w", err)
	}

	for _, dbp := range dbPatients {
		patients = append(patients, dbp.toDomainPatient())
	}

	return patients, count, nil
}

// GetAuditLog retrieves audit trail for a patient
func (r *PatientRepository) GetAuditLog(ctx context.Context, tenantID, patientID uuid.UUID, offset, limit int) ([]*domain.AuditLog, error) {
	var auditLogs []*domain.AuditLog

	query := `
		SELECT id, patient_id, tenant_id, action, changed_fields, old_values, new_values, changed_by, correlation_id, changed_at
		FROM patient_audit_logs
		WHERE tenant_id = $1 AND patient_id = $2
		ORDER BY changed_at DESC LIMIT $3 OFFSET $4
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, patientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get audit log: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var audit domain.AuditLog
		if err := rows.Scan(
			&audit.ID, &audit.PatientID, &audit.TenantID, &audit.Action,
			&audit.ChangedFields, &audit.OldValues, &audit.NewValues,
			&audit.ChangedBy, &audit.CorrelationID, &audit.ChangedAt,
		); err != nil {
			return nil, fmt.Errorf("get audit log (scan): %w", err)
		}
		auditLogs = append(auditLogs, &audit)
	}

	return auditLogs, rows.Err()
}
