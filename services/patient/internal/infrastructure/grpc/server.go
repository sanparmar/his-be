package grpc

import (
	"go.opentelemetry.io/otel"

	patientv1 "github.com/deloitte-us-consulting/his-be/api/patient/v1"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/application/command"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/application/query"
	"github.com/deloitte-us-consulting/his-be/services/patient/internal/domain"
)

var tracer = otel.Tracer("his-patient-service")

// PatientServer implements the gRPC PatientService.
type PatientServer struct {
	patientv1.UnimplementedPatientServiceServer
	patientService            *domain.PatientService
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

// NewPatientServer creates a new gRPC patient server.
func NewPatientServer(
	patientSvc *domain.PatientService,
	createPtHandler *command.CreatePatientHandler,
	updateDemoHandler *command.UpdateDemographicsHandler,
	updateInsHandler *command.UpdateInsuranceHandler,
	recordAllergiesH *command.RecordAllergiesHandler,
	recordMedsH *command.RecordMedicationsHandler,
	recordVitalsH *command.RecordVitalSignsHandler,
	getByMRNHandler *query.GetPatientByMRNHandler,
	searchPtsHandler *query.SearchPatientsHandler,
	listPtsHandler *query.ListPatientsHandler,
) *PatientServer {
	return &PatientServer{
		patientService:            patientSvc,
		createPatientHandler:      createPtHandler,
		updateDemographicsHandler: updateDemoHandler,
		updateInsuranceHandler:    updateInsHandler,
		recordAllergiesHandler:    recordAllergiesH,
		recordMedicationsHandler:  recordMedsH,
		recordVitalSignsHandler:   recordVitalsH,
		getPatientByMRNHandler:    getByMRNHandler,
		searchPatientsHandler:     searchPtsHandler,
		listPatientsHandler:       listPtsHandler,
	}
}
