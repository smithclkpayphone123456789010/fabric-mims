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
