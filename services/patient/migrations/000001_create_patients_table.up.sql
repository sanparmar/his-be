-- Create patients table for patient bounded context
-- Supports FHIR R4 Patient resource alignment
-- PHI fields are encrypted at application layer (names, DOB, contact info, etc.)
CREATE TABLE IF NOT EXISTS patients (
    -- Unique identifiers
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    mrn VARCHAR(50) NOT NULL,

    -- Demographics (PHI - encrypted at application layer)
    name VARCHAR(500) NOT NULL,
    dob DATE NOT NULL, -- Date of birth
    gender VARCHAR(20) NOT NULL, -- MALE, FEMALE, OTHER, UNKNOWN
    blood_group VARCHAR(5), -- A+, A-, B+, B-, AB+, AB-, O+, O-, UNKNOWN
    
    -- Contact information (PHI - encrypted)
    phone VARCHAR(20),
    email VARCHAR(255),
    
    -- Address (PHI - encrypted)
    address_street VARCHAR(500),
    address_city VARCHAR(100),
    address_state VARCHAR(100),
    address_postal_code VARCHAR(20),
    address_country VARCHAR(100),
    
    -- Ayushman Bharat Health Account & Aadhaar (PHI - encrypted)
    abha VARCHAR(100),
    aadhaar_last_four VARCHAR(4),
    
    -- Emergency Contact
    emergency_contact_name VARCHAR(500),
    emergency_contact_relation VARCHAR(100),
    emergency_contact_phone VARCHAR(20),
    
    -- Insurance (PHI - encrypted)
    insurance_provider VARCHAR(255),
    insurance_policy_no VARCHAR(100),
    insurance_tpa VARCHAR(255),
    insurance_member_id VARCHAR(100),
    insurance_sum_insured BIGINT,
    insurance_policy_valid_till DATE,
    insurance_corporate_name VARCHAR(255),
    
    -- Home medications and allergies (JSON for flexibility)
    home_medications JSONB, -- [{drug, dose, frequency, notes}]
    allergies TEXT[], -- Array of allergy strings
    
    -- Vital signs (latest recorded)
    vital_signs_bp VARCHAR(20), -- Format: "SYS/DIA", e.g., "120/80"
    vital_signs_hr INT,
    vital_signs_temp NUMERIC(4,1),
    vital_signs_spo2 INT,
    vital_signs_rr INT,
    vital_signs_weight NUMERIC(5,2),
    vital_signs_height INT,
    vital_signs_bmi NUMERIC(4,1),
    vital_signs_recorded_at TIMESTAMP WITH TIME ZONE,
    
    -- Status and audit
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, INACTIVE, ARCHIVED
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    updated_by VARCHAR(255),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT unique_mrn_per_tenant UNIQUE (tenant_id, mrn),
    CONSTRAINT valid_blood_group CHECK (blood_group IN ('A+', 'A-', 'B+', 'B-', 'AB+', 'AB-', 'O+', 'O-', 'UNKNOWN', NULL)),
    CONSTRAINT valid_gender CHECK (gender IN ('MALE', 'FEMALE', 'OTHER', 'UNKNOWN')),
    CONSTRAINT valid_status CHECK (status IN ('ACTIVE', 'INACTIVE', 'ARCHIVED')),
    CONSTRAINT tenant_not_null CHECK (tenant_id IS NOT NULL)
);

-- Create indexes for common queries
CREATE INDEX idx_patients_tenant_id ON patients(tenant_id);
CREATE INDEX idx_patients_tenant_mrn ON patients(tenant_id, mrn);
CREATE INDEX idx_patients_status ON patients(status);
CREATE INDEX idx_patients_created_at ON patients(created_at DESC);
CREATE INDEX idx_patients_updated_at ON patients(updated_at DESC);
CREATE INDEX idx_patients_created_by ON patients(created_by);
CREATE INDEX idx_patients_deleted_at ON patients(deleted_at);

-- JSONB index for medications search
CREATE INDEX idx_patients_home_medications ON patients USING GIN(home_medications);

-- GIN index for allergies array search
CREATE INDEX idx_patients_allergies ON patients USING GIN(allergies);

-- Create patient audit log table (for mutations tracking per HIPAA requirements)
CREATE TABLE IF NOT EXISTS patient_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL, -- CREATE, UPDATE, DELETE, ACTIVATE, ARCHIVE
    changed_fields JSONB, -- Which fields were changed
    old_values JSONB, -- Previous values (without PHI)
    new_values JSONB, -- New values (without PHI)
    changed_by VARCHAR(255) NOT NULL, -- User ID or service name
    correlation_id VARCHAR(255), -- For distributed tracing
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_patient_audit_logs_patient_id ON patient_audit_logs(patient_id);
CREATE INDEX idx_patient_audit_logs_tenant_id ON patient_audit_logs(tenant_id);
CREATE INDEX idx_patient_audit_logs_changed_at ON patient_audit_logs(changed_at DESC);
CREATE INDEX idx_patient_audit_logs_changed_by ON patient_audit_logs(changed_by);
