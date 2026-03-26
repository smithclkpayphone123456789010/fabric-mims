package model

// Account 账户，虚拟管理员和若干业主账号
type Account struct {
	AccountId string  `json:"accountId"` //账号ID
	UserName  string  `json:"userName"`  //账号名
	Balance   float64 `json:"balance"`   //余额
}

// objectType  对象类型，用于创建复合主键
const (
	AccountKey = "account-key"

	AccountV2Key    = "account-v2-key"
	PrescriptionKey = "prescription-key"
	PatientKey      = "patient-key"
	InsuranceKey    = "insurance-key"
	DrugKey         = "drug-key"

	// Authorization 相关键
	AuthorizationKey             = "authorization-key"
	AuthorizationPatientIndexKey = "authorization-patient-index-key"
	AuthorizationDoctorIndexKey  = "authorization-doctor-index-key"

	OutpatientRegistrationKey           = "outpatient-registration-key"
	OutpatientRegistrationPatientIdxKey = "outpatient-registration-patient-index-key"
	OutpatientRegistrationDoctorIdxKey  = "outpatient-registration-doctor-index-key"
	OutpatientSlotKey                   = "outpatient-slot-key"
	OutpatientPaymentKey                = "outpatient-payment-key"
	OutpatientQueueDoctorIdxKey         = "outpatient-queue-doctor-index-key"
)

// --------------------------------------------------------------------

// AccountV2 账号
type AccountV2 struct {
	AccountId       string `json:"account_id"`        // 账号ID
	AccountName     string `json:"account_name"`      // 账号名
	Role            string `json:"role"`              // 角色
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

// Hospital 医院
type Hospital struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Admins  []HospitalAdmin `json:"admins"`
	Doctors []Doctor        `json:"doctors"`
}

// HospitalAdmin 医院管理员
type HospitalAdmin struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Doctor 医生
type Doctor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Prescription 医疗处方
type Prescription struct {
	ID                 string `json:"id"`                  // 医疗处方ID
	Patient            string `json:"patient"`             // 患者ID
	RecordType         string `json:"record_type"`         // 病历类型
	FileHash           string `json:"file_hash"`           // 文件哈希
	FileName           string `json:"file_name"`           // 文件名
	FilePath           string `json:"file_path"`           // 本地加密存储路径
	SymptomDescription string `json:"symptom_description"` // 兼容旧字段-症状描述
	DoctorDiagnosis    string `json:"doctor_diagnosis"`    // 兼容旧字段-医生诊断
	Diagnosis          string `json:"diagnosis"`           // 兼容旧字段
	Drug               []Drug `json:"drug"`                // 药品列表及用量(兼容旧字段)
	Doctor             string `json:"doctor"`              // 开方医师 AccountV2Id
	Hospital           string `json:"hospital"`            // 医院 ID
	Created            string `json:"created"`             // 创建时间
	Comment            string `json:"comment"`             // 备注

	PatientName      string `json:"patient_name"`
	PatientGender    string `json:"patient_gender"`
	PatientAge       string `json:"patient_age"`
	PatientIDCardNo  string `json:"patient_id_card_no"`
	PatientPhone     string `json:"patient_phone"`
	InsuranceCardNo  string `json:"insurance_card_no"`
	HospitalName     string `json:"hospital_name"`
	Department       string `json:"department"`
	VisitDoctorName  string `json:"visit_doctor_name"`
	ChiefComplaint   string `json:"chief_complaint"`
	PresentIllness   string `json:"present_illness"`
	PastHistory      string `json:"past_history"`
	AllergyHistory   string `json:"allergy_history"`
	FamilyHistory    string `json:"family_history"`
	Temperature      string `json:"temperature"`
	Pulse            string `json:"pulse"`
	BloodPressure    string `json:"blood_pressure"`
	Respiration      string `json:"respiration"`
	PhysicalExam     string `json:"physical_exam"`
	LabExam          string `json:"lab_exam"`
	ImagingExam      string `json:"imaging_exam"`
	DiagnosisResult  string `json:"diagnosis_result"`
	TreatmentPlan    string `json:"treatment_plan"`
	MedicationAdvice string `json:"medication_advice"`
	DoctorAdvice     string `json:"doctor_advice"`
}

// Patient 患者
type Patient struct {
	ID     string `json:"id"`     // 患者 AccountV2Id
	Name   string `json:"name"`   // 患者姓名
	Age    int    `json:"age"`    // 患者年龄
	Gender string `json:"gender"` // 患者性别
}

// Drug 药品
type Drug struct {
	//ID      string `json:"id"`
	Name   string `json:"Name"`   // 药品名
	Amount string `json:"amount"` // 药品数量
}

// DrugOrder 药品订单
type DrugOrder struct {
	ID           string `json:"id"`           // 订单ID
	Name         string `json:"Name"`         // 药品名
	Amount       string `json:"amount"`       // 药品数量
	Prescription string `json:"prescription"` // 处方ID
	Patient      string `json:"patient"`      // 患者ID
	DrugStore    string `json:"drug_store"`   // 药店id
	Created      string `json:"created"`      // 创建时间
}

// DrugStore 药店
type DrugStore struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RecordAuthorization 病历授权记录
// 状态说明：active=生效中、expired=已过期(查询计算态)、revoked=已撤销
// 其中 expired 为查询展示态，账本内只持久化 active/revoked
// 复合主键：AuthorizationKey(patient_id, record_id, auth_id)
type RecordAuthorization struct {
	ID           string `json:"id"`            // 授权ID
	RecordID     string `json:"record_id"`     // 病历ID
	PatientID    string `json:"patient_id"`    // 患者ID
	DoctorID     string `json:"doctor_id"`     // 医生ID
	HospitalName string `json:"hospital_name"` // 医院名称
	Department   string `json:"department"`    // 科室
	Scope        string `json:"scope"`         // 权限范围(read)
	Status       string `json:"status"`        // active/revoked/expired(计算态)
	StartTime    string `json:"start_time"`    // 开始时间 2006-01-02 15:04:05
	EndTime      string `json:"end_time"`      // 截止时间 2006-01-02 15:04:05
	CreatedTime  string `json:"created_time"`  // 创建时间
	UpdatedTime  string `json:"updated_time"`  // 更新时间
	Remark       string `json:"remark"`        // 备注
}

// OutpatientRegistration 门诊挂号
// 状态：BOOKED/CANCELLED/VISITED
// 费用状态：UNPAID/PAID
type OutpatientRegistration struct {
	ID             string `json:"id"`
	PatientID      string `json:"patient_id"`
	DoctorID       string `json:"doctor_id"`
	DepartmentID   string `json:"department_id"`
	ScheduleSlotID string `json:"schedule_slot_id"`
	VisitDate      string `json:"visit_date"`
	Status         string `json:"status"`
	FeeAmount      string `json:"fee_amount"`
	FeeStatus      string `json:"fee_status"`
	QueueNo        string `json:"queue_no"`
	CreatedTime    string `json:"created_time"`
	UpdatedTime    string `json:"updated_time"`
	TxID           string `json:"tx_id"`
}

// OutpatientScheduleSlot 门诊号源
// 状态：OPEN/CLOSED
type OutpatientScheduleSlot struct {
	ID           string `json:"id"`
	DoctorID     string `json:"doctor_id"`
	DepartmentID string `json:"department_id"`
	VisitDate    string `json:"visit_date"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
	Capacity     int    `json:"capacity"`
	BookedCount  int    `json:"booked_count"`
	Status       string `json:"status"`
	CreatedTime  string `json:"created_time"`
	UpdatedTime  string `json:"updated_time"`
	TxID         string `json:"tx_id"`
}

// OutpatientPayment 门诊缴费
// 状态：UNPAID/PAID
type OutpatientPayment struct {
	ID             string `json:"id"`
	OrderType      string `json:"order_type"`
	RegistrationID string `json:"registration_id"`
	PatientID      string `json:"patient_id"`
	Amount         string `json:"amount"`
	Status         string `json:"status"`
	PaidTime       string `json:"paid_time"`
	CreatedTime    string `json:"created_time"`
	TxID           string `json:"tx_id"`
}

// OutpatientQueueItem 门诊排队
// 状态：WAITING/IN_PROGRESS/DONE
type OutpatientQueueItem struct {
	ID             string `json:"id"`
	RegistrationID string `json:"registration_id"`
	DoctorID       string `json:"doctor_id"`
	PatientID      string `json:"patient_id"`
	QueueNo        string `json:"queue_no"`
	Status         string `json:"status"`
	CalledTime     string `json:"called_time"`
	FinishedTime   string `json:"finished_time"`
	CreatedTime    string `json:"created_time"`
	TxID           string `json:"tx_id"`
}

// Insurance 保险机构
type Insurance struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// InsuranceCover 保险报销订单
type InsuranceCover struct {
	ID           string `json:"id"`           // 订单ID
	Prescription string `json:"prescription"` // 处方ID
	Patient      string `json:"patient"`      // 患者ID
	Status       string `json:"status"`       // 订单状态
	Created      string `json:"created"`      // 创建时间
}

// InsuranceStatusConstant 保险状态
var InsuranceStatusConstant = func() map[string]string {
	return map[string]string{
		"processing": "处理中", // 患者发起保险报销申请，等待保险公司确认报销
		"cancelled":  "已取消", // 患者在保险公司确认报销之前取消保险报销申请
		"refused":    "已拒绝", // 保险公司拒绝确认报销
		"approved":   "已通过", // 保险公司确认报销，保险报销完成
	}
}

// DrugStatusConstant 药品状态
//var DrugStatusConstant = func() map[string]string {
//	return map[string]string{
//		"processing": "处理中", //
//		"done":       "完成",   //
//	}
//}
