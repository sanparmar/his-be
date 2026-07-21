INSERT INTO roles (id, name, display_name, category, parent_role_id, is_system, description) VALUES
-- Executive Roles
('11111111-1111-1111-1111-111111111111', 'super_admin', 'System Administrator', 'executive', NULL, TRUE, 'Full system access across all tenants'),
('22222222-2222-2222-2222-222222222222', 'tenant_admin', 'Tenant Administrator', 'executive', '11111111-1111-1111-1111-111111111111', TRUE, 'Full access within tenant'),
('33333333-3333-3333-3333-333333333333', 'cmo', 'Chief Medical Officer', 'executive', '22222222-2222-2222-2222-222222222222', TRUE, 'Clinical oversight, policy'),
('44444444-4444-4444-4444-444444444444', 'cno', 'Chief Nursing Officer', 'executive', '22222222-2222-2222-2222-222222222222', TRUE, 'Nursing oversight, policy'),

-- Administrative Roles
('55555555-5555-5555-5555-555555555555', 'hospital_admin', 'Hospital Administrator', 'administrative', '22222222-2222-2222-2222-222222222222', TRUE, 'Hospital operations'),
('66666666-6666-6666-6666-666666666666', 'dept_head', 'Department Head', 'administrative', '55555555-5555-5555-5555-555555555555', TRUE, 'Department management'),
('77777777-7777-7777-7777-777777777777', 'unit_manager', 'Unit Manager', 'administrative', '66666666-6666-6666-6666-666666666666', TRUE, 'Unit/ward management'),

-- Clinical Roles
('88888888-8888-8888-8888-888888888888', 'doctor', 'Doctor', 'clinical', NULL, TRUE, 'General physician privileges'),
('99999999-9999-9999-9999-999999999999', 'senior_doctor', 'Senior Doctor', 'clinical', '88888888-8888-8888-8888-888888888888', TRUE, 'Additional approval privileges'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'specialist', 'Specialist', 'clinical', '99999999-9999-9999-9999-999999999999', TRUE, 'Specialty-specific privileges'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'surgeon', 'Surgeon', 'clinical', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', TRUE, 'Surgical privileges'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'anesthetist', 'Anesthetist', 'clinical', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', TRUE, 'Anesthesia privileges'),

-- Nursing Roles
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'nurse', 'Nurse', 'nursing', NULL, TRUE, 'General nursing privileges'),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'senior_nurse', 'Senior Nurse', 'nursing', 'dddddddd-dddd-dddd-dddd-dddddddddddd', TRUE, 'Charge nurse privileges'),
('ffffffff-ffff-ffff-ffff-ffffffffffff', 'nursing_super', 'Nursing Superintendent', 'nursing', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', TRUE, 'Nursing admin privileges'),

-- Allied Health
('11111111-2222-3333-4444-555555555555', 'pharmacist', 'Pharmacist', 'allied_health', NULL, TRUE, 'Pharmacy privileges'),
('22222222-3333-4444-5555-666666666666', 'lab_tech', 'Lab Technician', 'allied_health', NULL, TRUE, 'Laboratory privileges'),
('33333333-4444-5555-6666-777777777777', 'rad_tech', 'Radiology Technician', 'allied_health', NULL, TRUE, 'Radiology privileges'),
('44444444-5555-6666-7777-888888888888', 'physio', 'Physiotherapist', 'allied_health', NULL, TRUE, 'Physiotherapy privileges'),
('55555555-6666-7777-8888-999999999999', 'dietitian', 'Dietitian', 'allied_health', NULL, TRUE, 'Dietetics privileges'),

-- Support Staff
('66666666-7777-8888-9999-000000000000', 'front_desk', 'Front Desk Executive', 'support', NULL, TRUE, 'Registration, appointments'),
('77777777-8888-9999-0000-111111111111', 'billing_officer', 'Billing Officer', 'support', NULL, TRUE, 'Billing, invoicing'),
('88888888-9999-0000-1111-222222222222', 'insurance_officer', 'Insurance Officer', 'support', '77777777-8888-9999-0000-111111111111', TRUE, 'Insurance, claims'),
('99999999-0000-1111-2222-333333333333', 'bed_manager', 'Bed Manager', 'support', NULL, TRUE, 'Bed assignment'),
('aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee', 'medical_records', 'Medical Records Officer', 'support', NULL, TRUE, 'Records management'),

-- Patient-Facing
('bbbbbbbb-cccc-dddd-eeee-ffffffffffff', 'patient', 'Patient', 'patient_facing', NULL, TRUE, 'Patient portal access'),
('cccccccc-dddd-eeee-ffff-000000000000', 'caregiver', 'Family Caregiver', 'patient_facing', 'bbbbbbbb-cccc-dddd-eeee-ffffffffffff', TRUE, 'Limited patient access');