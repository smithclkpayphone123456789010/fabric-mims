package model

// ----------------------         Account 用户   ----------------------------------

type AccountIdBody struct {
	AccountId string `json:"account_id"`
}

type AccountRequestBody struct {
	Args []AccountIdBody `json:"args"`
}

type CreateAccountBody struct {
	AccountName     string `json:"account_name"`      // 姓名/名称
	Role            string `json:"role"`              // doctor/patient/drugstore/insurance
	Operator        string `json:"operator"`          // 操作人
	HospitalID      string `json:"hospital_id"`       // 所属医院ID
	HospitalName    string `json:"hospital_name"`     // 所属医院名称
	Department      string `json:"department"`        // 所属科室
	Title           string `json:"title"`             // 职位
	Gender          string `json:"gender"`            // 性别
	EmployeeNo      string `json:"employee_no"`       // 工号
	IDCardNo        string `json:"id_card_no"`        // 身份证号
	InsuranceCardNo string `json:"insurance_card_no"` // 医保卡号
	Age             string `json:"age"`               // 年龄
	BirthDate       string `json:"birth_date"`        // 出生年月
	Phone           string `json:"phone"`             // 联系方式
}

// ----------------------         Prescription 病历   ----------------------------------

type PrescriptionRequestBody struct {
	Doctor             string `json:"doctor" form:"doctor"`                           // 医生ID
	Patient            string `json:"patient" form:"patient"`                         // 患者Id
	RecordType         string `json:"record_type" form:"record_type"`                 // 病历类型: EMR/REPORT/PRESCRIPTION
	FileHash           string `json:"file_hash" form:"file_hash"`                     // 病历文件哈希
	SymptomDescription string `json:"symptom_description" form:"symptom_description"` // 兼容旧字段-症状描述
	DoctorDiagnosis    string `json:"doctor_diagnosis" form:"doctor_diagnosis"`       // 兼容旧字段-医生诊断
	Diagnosis          string `json:"diagnosis" form:"diagnosis"`                     // 兼容旧字段
	DrugName           string `json:"drug_name" form:"drug_name"`                     // 兼容旧字段-药品名
	DrugAmount         string `json:"drug_amount" form:"drug_amount"`                 // 兼容旧字段-药品用量
	Hospital           string `json:"hospital" form:"hospital"`                       // 医院 ID
	Comment            string `json:"comment" form:"comment"`                         // 备注
	FileName           string `json:"file_name" form:"file_name"`                     // 文件名
	FilePath           string `json:"file_path" form:"file_path"`                     // 本地加密文件路径

	PatientName      string `json:"patient_name" form:"patient_name"`
	PatientGender    string `json:"patient_gender" form:"patient_gender"`
	PatientAge       string `json:"patient_age" form:"patient_age"`
	PatientIDCardNo  string `json:"patient_id_card_no" form:"patient_id_card_no"`
	PatientPhone     string `json:"patient_phone" form:"patient_phone"`
	InsuranceCardNo  string `json:"insurance_card_no" form:"insurance_card_no"`
	HospitalName     string `json:"hospital_name" form:"hospital_name"`
	Department       string `json:"department" form:"department"`
	VisitDoctorName  string `json:"visit_doctor_name" form:"visit_doctor_name"`
	ChiefComplaint   string `json:"chief_complaint" form:"chief_complaint"`
	PresentIllness   string `json:"present_illness" form:"present_illness"`
	PastHistory      string `json:"past_history" form:"past_history"`
	AllergyHistory   string `json:"allergy_history" form:"allergy_history"`
	FamilyHistory    string `json:"family_history" form:"family_history"`
	Temperature      string `json:"temperature" form:"temperature"`
	Pulse            string `json:"pulse" form:"pulse"`
	BloodPressure    string `json:"blood_pressure" form:"blood_pressure"`
	Respiration      string `json:"respiration" form:"respiration"`
	PhysicalExam     string `json:"physical_exam" form:"physical_exam"`
	LabExam          string `json:"lab_exam" form:"lab_exam"`
	ImagingExam      string `json:"imaging_exam" form:"imaging_exam"`
	DiagnosisResult  string `json:"diagnosis_result" form:"diagnosis_result"`
	TreatmentPlan    string `json:"treatment_plan" form:"treatment_plan"`
	MedicationAdvice string `json:"medication_advice" form:"medication_advice"`
	DoctorAdvice     string `json:"doctor_advice" form:"doctor_advice"`
}

type PrescriptionQueryRequestBody struct {
	Patient  string `json:"patient"`   // 患者AccountId
	DoctorID string `json:"doctor_id"` // 医生ID（按授权过滤）
}

// ----------------------         DrugOrder 药品订单   ----------------------------------

type DrugOrderRequestBody struct {
	//Drug      []Drug `json:"drug"`      // 药品列表及用量
	DrugName     string `json:"drug_name"`    // 药品名
	DrugAmount   string `json:"drug_amount"`  // 药品用量
	Prescription string `json:"prescription"` // 处方ID
	Patient      string `json:"patient"`      // 患者Id
	DrugStore    string `json:"drug_store"`   // 药店Id
}

type DrugOrderQueryRequestBody struct {
	Patient   string `json:"patient"` // 患者AccountId
	DrugStore string `json:"drug_store"`
}

// ----------------------         InsuranceCover 保险报销   ----------------------------------

type InsuranceCoverRequestBody struct {
	Prescription string `json:"prescription"` // 处方ID
	Patient      string `json:"patient"`      // 患者Id
	Status       string `json:"status"`       // 订单状态
}

type InsuranceCoverQueryRequestBody struct {
	Patient        string `json:"patient"`         // 患者Id
	InsuranceCover string `json:"insurance_cover"` // 报销订单ID
}

type UpdateInsuranceCoverRequestBody struct {
	InsuranceCover string `json:"insurance_cover"` // 报销订单ID
	Patient        string `json:"patient"`         // 病人ID
	InsuranceID    string `json:"insurance_id"`    // 保险机构ID
	Status         string `json:"status"`          // 订单状态
}

// ----------------------         Authorization 病历授权   ----------------------------------

type GrantRecordAuthorizationRequestBody struct {
	PatientID    string `json:"patient_id"`
	RecordID     string `json:"record_id"`
	DoctorID     string `json:"doctor_id"`
	HospitalName string `json:"hospital_name"`
	Department   string `json:"department"`
	EndTime      string `json:"end_time"`
	Remark       string `json:"remark"`
}

type QueryMyAuthorizationsRequestBody struct {
	PatientID string `json:"patient_id"`
}

type RevokeRecordAuthorizationRequestBody struct {
	PatientID string `json:"patient_id"`
	AuthID    string `json:"auth_id"`
}

type CheckRecordAccessRequestBody struct {
	DoctorID string `json:"doctor_id" form:"doctor_id"`
	RecordID string `json:"record_id" form:"record_id"`
}

type RenewRecordAuthorizationRequestBody struct {
	PatientID string `json:"patient_id"`
	AuthID    string `json:"auth_id"`
	EndTime   string `json:"end_time"`
}

type QueryAccessibleRecordsByDoctorRequestBody struct {
	DoctorID           string `json:"doctor_id"`
	PatientNameKeyword string `json:"patient_name_keyword"`
	IdCardKeyword      string `json:"id_card_keyword"`
	RecordTypeKeyword  string `json:"record_type_keyword"`
	CreatedStart       string `json:"created_start"`
	CreatedEnd         string `json:"created_end"`
}

// ----------------------         Outpatient 门诊管理   ----------------------------------

type CreateOutpatientRegistrationRequestBody struct {
	PatientID    string `json:"patient_id"`
	DoctorID     string `json:"doctor_id"`
	DepartmentID string `json:"department_id"`
	SlotID       string `json:"slot_id"`
	VisitDate    string `json:"visit_date"`
}

type CancelOutpatientRegistrationRequestBody struct {
	RegistrationID string `json:"registration_id"`
	OperatorID     string `json:"operator_id"`
}

type QueryOutpatientRegistrationRequestBody struct {
	PatientID string `json:"patient_id" form:"patient_id"`
	DoctorID  string `json:"doctor_id" form:"doctor_id"`
	Status    string `json:"status" form:"status"`
}

type CreateScheduleSlotRequestBody struct {
	DoctorID     string `json:"doctor_id"`
	DepartmentID string `json:"department_id"`
	VisitDate    string `json:"visit_date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Capacity     int    `json:"capacity"`
}

type QueryScheduleSlotRequestBody struct {
	DepartmentID string `json:"department_id" form:"department_id"`
	DoctorID     string `json:"doctor_id" form:"doctor_id"`
	VisitDate    string `json:"visit_date" form:"visit_date"`
}

type QueryOutpatientPaymentRequestBody struct {
	PatientID string `json:"patient_id" form:"patient_id"`
	Status    string `json:"status" form:"status"`
}

type PayOutpatientOrderRequestBody struct {
	PaymentID  string `json:"payment_id"`
	PatientID  string `json:"patient_id"`
	PayChannel string `json:"pay_channel"`
}

type QueryOutpatientQueueRequestBody struct {
	DoctorID string `json:"doctor_id" form:"doctor_id"`
}

type StartVisitRequestBody struct {
	RegistrationID string `json:"registration_id"`
	DoctorID       string `json:"doctor_id"`
}

type FinishVisitRequestBody struct {
	RegistrationID string `json:"registration_id"`
	DoctorID       string `json:"doctor_id"`
}

type QueryOutpatientRecordRequestBody struct {
	PatientID string `json:"patient_id" form:"patient_id"`
	DoctorID  string `json:"doctor_id" form:"doctor_id"`
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
}

// ----------------------         Audit 审计监控   ----------------------------------

type AuditEventRequestBody struct {
	EventType       string `json:"event_type" form:"event_type"`
	EventLevel      string `json:"event_level" form:"event_level"`
	ActionResult    string `json:"action_result" form:"action_result"`
	TargetPatientID string `json:"target_patient_id" form:"target_patient_id"`
	TargetRecordID  string `json:"target_record_id" form:"target_record_id"`
	DetailJSON      string `json:"detail_json" form:"detail_json"`
	FailReason      string `json:"fail_reason" form:"fail_reason"`
}

type AuditEventListRequestBody struct {
	StartTime       string `form:"start_time" json:"start_time"`
	EndTime         string `form:"end_time" json:"end_time"`
	EventType       string `form:"event_type" json:"event_type"`
	EventLevel      string `form:"event_level" json:"event_level"`
	ActionResult    string `form:"action_result" json:"action_result"`
	TargetPatientID string `form:"target_patient_id" json:"target_patient_id"`
	TargetRecordID  string `form:"target_record_id" json:"target_record_id"`
	TxID            string `form:"tx_id" json:"tx_id"`
	Keyword         string `form:"keyword" json:"keyword"`
	Page            int    `form:"page" json:"page"`
	Size            int    `form:"size" json:"size"`
}

type AuditEventItem struct {
	ID              string `json:"id"`
	EventType       string `json:"event_type"`
	EventLevel      string `json:"event_level"`
	EventTime       string `json:"event_time"`
	ActorID         string `json:"actor_id"`
	TargetPatientID string `json:"target_patient_id"`
	TargetRecordID  string `json:"target_record_id"`
	ActionResult    string `json:"action_result"`
	TxID            string `json:"tx_id"`
	RequestPath     string `json:"request_path"`
	Message         string `json:"message"`
}

type AuditEventDetailResponse struct {
	ID              string `json:"id"`
	EventType       string `json:"event_type"`
	EventLevel      string `json:"event_level"`
	EventTime       string `json:"event_time"`
	ActorID         string `json:"actor_id"`
	OrgID           string `json:"org_id"`
	TargetPatientID string `json:"target_patient_id"`
	TargetRecordID  string `json:"target_record_id"`
	TxID            string `json:"tx_id"`
	ChaincodeFunc   string `json:"chaincode_func"`
	RequestPath     string `json:"request_path"`
	RequestMethod   string `json:"request_method"`
	RequestID       string `json:"request_id"`
	TraceID         string `json:"trace_id"`
	ClientIP        string `json:"client_ip"`
	UserAgent       string `json:"user_agent"`
	ActionResult    string `json:"action_result"`
	FailReason      string `json:"fail_reason"`
	DetailJSON      string `json:"detail_json"`
	HashPrev        string `json:"hash_prev"`
	HashCurrent     string `json:"hash_current"`
}

type AuditEventStatsResponse struct {
	TotalCount   int `json:"total_count"`
	SuccessCount int `json:"success_count"`
	FailCount    int `json:"fail_count"`
	L2Count      int `json:"l2_count"`
	L3Count      int `json:"l3_count"`
}

type AuditAlertListRequestBody struct {
	Level     string `form:"level" json:"level"`
	Status    string `form:"status" json:"status"`
	StartTime string `form:"start_time" json:"start_time"`
	EndTime   string `form:"end_time" json:"end_time"`
	Page      int    `form:"page" json:"page"`
	Size      int    `form:"size" json:"size"`
}

type AuditAlertItem struct {
	ID              string `json:"id"`
	RuleCode        string `json:"rule_code"`
	Level           string `json:"level"`
	Status          string `json:"status"`
	TriggerTime     string `json:"trigger_time"`
	ActorID         string `json:"actor_id"`
	TargetPatientID string `json:"target_patient_id"`
	TargetRecordID  string `json:"target_record_id"`
	Description     string `json:"description"`
}

type AuditAlertDetailResponse struct {
	ID              string `json:"id"`
	RuleCode        string `json:"rule_code"`
	Level           string `json:"level"`
	Status          string `json:"status"`
	TriggerTime     string `json:"trigger_time"`
	EventID         string `json:"event_id"`
	ActorID         string `json:"actor_id"`
	TargetPatientID string `json:"target_patient_id"`
	TargetRecordID  string `json:"target_record_id"`
	Description     string `json:"description"`
	HandleTime      string `json:"handle_time"`
	HandleNote      string `json:"handle_note"`
}

type AuditAlertStatsResponse struct {
	TodayCount     int `json:"today_count"`
	Unresolved     int `json:"unresolved"`
	HighLevelCount int `json:"high_level_count"`
}

type AuditAlertResolveRequestBody struct {
	HandleNote string `json:"handle_note" form:"handle_note"`
}

type AuditExportCreateRequestBody struct {
	Format        string `json:"format" form:"format"`
	MaskSensitive bool   `json:"mask_sensitive" form:"mask_sensitive"`
	FilterJSON    string `json:"filter_json" form:"filter_json"`
}

type AuditExportTaskItem struct {
	ID         string `json:"id"`
	CreatorID  string `json:"creator_id"`
	CreateTime string `json:"create_time"`
	Status     string `json:"status"`
	Format     string `json:"format"`
	FileName   string `json:"file_name"`
	FinishTime string `json:"finish_time"`
	FailReason string `json:"fail_reason"`
}

type AuditExportTaskDetailResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Format     string `json:"format"`
	FilterJSON string `json:"filter_json"`
	FileName   string `json:"file_name"`
	FileHash   string `json:"file_hash"`
	FailReason string `json:"fail_reason"`
	CreateTime string `json:"create_time"`
	FinishTime string `json:"finish_time"`
}

type AuditCollectorHealthResponse struct {
	Status       string `json:"status"`
	LastEventID  string `json:"last_event_id"`
	HashChainOK  bool   `json:"hash_chain_ok"`
	PendingCount int    `json:"pending_count"`
}
