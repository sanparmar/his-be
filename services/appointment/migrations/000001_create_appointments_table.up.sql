CREATE TABLE IF NOT EXISTS appointments (
    tenant_id UUID NOT NULL,
    id VARCHAR(64) NOT NULL,
    patient_mrn VARCHAR(64) NOT NULL,
    patient_name VARCHAR(255) NOT NULL,
    doctor_id VARCHAR(64) NOT NULL,
    doctor_name VARCHAR(255) NOT NULL,
    department VARCHAR(120) NOT NULL,
    appointment_date DATE NOT NULL,
    slot_time VARCHAR(16) NOT NULL,
    visit_type VARCHAR(40) NOT NULL,
    status VARCHAR(40) NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    PRIMARY KEY (tenant_id, id)
);

CREATE INDEX IF NOT EXISTS idx_appointments_tenant_date ON appointments (tenant_id, appointment_date);
CREATE INDEX IF NOT EXISTS idx_appointments_tenant_mrn ON appointments (tenant_id, patient_mrn);
CREATE INDEX IF NOT EXISTS idx_appointments_tenant_status_date ON appointments (tenant_id, status, appointment_date);
CREATE INDEX IF NOT EXISTS idx_appointments_tenant_doctor_date ON appointments (tenant_id, doctor_id, appointment_date);
