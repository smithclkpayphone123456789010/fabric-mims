package api

import (
	"chaincode/model"
	"chaincode/pkg/utils"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// CreatePrescription 创建处方(医生)
func CreatePrescription(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 参数（38个）:
	// doctor, patient, recordType, fileHash, fileName, filePath, symptomDescription, doctorDiagnosis, diagnosis, drugName, drugAmount, hospital, comment,
	// patientName, patientGender, patientAge, patientIDCardNo, patientPhone, insuranceCardNo,
	// hospitalName, department, visitDoctorName,
	// chiefComplaint, presentIllness, pastHistory, allergyHistory, familyHistory,
	// temperature, pulse, bloodPressure, respiration, physicalExam, labExam, imagingExam,
	// diagnosisResult, treatmentPlan, medicationAdvice, doctorAdvice
	if len(args) != 38 {
		return shim.Error("参数个数不满足")
	}
	doctorID := args[0]           // 医生id
	patientID := args[1]          // 患者id
	recordType := args[2]         // 病历类型
	fileHash := args[3]           // 文件哈希
	fileName := args[4]           // 文件名
	filePath := args[5]           // 文件路径
	symptomDescription := args[6] // 兼容旧字段-症状描述
	doctorDiagnosis := args[7]    // 兼容旧字段-医生诊断
	diagnosis := args[8]          // 兼容旧字段
	drugName := args[9]           // 兼容旧字段-药品名
	drugAmount := args[10]        // 兼容旧字段-药品数量
	hospitalID := args[11]        // 医院ID
	comment := args[12]           // 备注

	patientName := args[13]
	patientGender := args[14]
	patientAge := args[15]
	patientIDCardNo := args[16]
	patientPhone := args[17]
	insuranceCardNo := args[18]
	hospitalName := args[19]
	department := args[20]
	visitDoctorName := args[21]
	chiefComplaint := args[22]
	presentIllness := args[23]
	pastHistory := args[24]
	allergyHistory := args[25]
	familyHistory := args[26]
	temperature := args[27]
	pulse := args[28]
	bloodPressure := args[29]
	respiration := args[30]
	physicalExam := args[31]
	labExam := args[32]
	imagingExam := args[33]
	diagnosisResult := args[34]
	treatmentPlan := args[35]
	medicationAdvice := args[36]
	doctorAdvice := args[37]

	if doctorID == "" || patientID == "" || recordType == "" || fileHash == "" || hospitalID == "" {
		return shim.Error("参数存在空值")
	}
	if recordType != "门诊病历" && recordType != "住院病历" && recordType != "急诊病历" && recordType != "检查报告" {
		return shim.Error("病历类型不合法")
	}
	if chiefComplaint == "" || presentIllness == "" || diagnosisResult == "" {
		return shim.Error("主诉/现病史/诊断结果不能为空")
	}
	if len(symptomDescription) > 500 || len(doctorDiagnosis) > 500 || len(comment) > 500 || len(chiefComplaint) > 2000 || len(presentIllness) > 4000 || len(diagnosisResult) > 2000 || len(allergyHistory) > 1000 {
		return shim.Error("文本长度超出限制")
	}

	// 参数数据格式转换（兼容旧字段）
	var drugs []model.Drug
	if drugName != "" && drugAmount != "" {
		drugNames := strings.Split(drugName, ",")
		drugAmounts := strings.Split(drugAmount, ",")
		for i, v := range drugNames {
			amount := ""
			if i < len(drugAmounts) {
				amount = drugAmounts[i]
			}
			drug := model.Drug{
				Name:   v,
				Amount: amount,
			}
			drugs = append(drugs, drug)
		}
	}

	// 判断是否为医生操作
	resultsAccount, err := utils.GetStateByPartialCompositeKeys(stub, model.AccountV2Key, []string{doctorID})
	if err != nil || len(resultsAccount) != 1 {
		return shim.Error(fmt.Sprintf("操作人权限验证失败%s", err))
	}
	var account model.AccountV2
	if err = json.Unmarshal(resultsAccount[0], &account); err != nil {
		return shim.Error(fmt.Sprintf("查询操作人信息-反序列化出错: %s", err))
	}
	if account.Role != "doctor" && account.AccountName != "医生" {
		return shim.Error("操作人权限不足")
	}

	// 判断患者是否存在
	resultsPatient, err := utils.GetStateByPartialCompositeKeys(stub, model.AccountV2Key, []string{patientID})
	if err != nil || len(resultsPatient) != 1 {
		return shim.Error(fmt.Sprintf("患者信息验证失败%s", err))
	}

	if diagnosis == "" {
		diagnosis = doctorDiagnosis
	}

	prescription := &model.Prescription{
		ID:                 stub.GetTxID()[:16],
		Patient:            patientID,
		RecordType:         recordType,
		FileHash:           fileHash,
		FileName:           fileName,
		FilePath:           filePath,
		SymptomDescription: symptomDescription,
		DoctorDiagnosis:    doctorDiagnosis,
		Diagnosis:          diagnosis,
		Drug:               drugs,
		Doctor:             doctorID,
		Hospital:           hospitalID,
		Created:            time.Now().Format("2006-01-02 15:04:05"),
		Comment:            comment,

		PatientName:      patientName,
		PatientGender:    patientGender,
		PatientAge:       patientAge,
		PatientIDCardNo:  patientIDCardNo,
		PatientPhone:     patientPhone,
		InsuranceCardNo:  insuranceCardNo,
		HospitalName:     hospitalName,
		Department:       department,
		VisitDoctorName:  visitDoctorName,
		ChiefComplaint:   chiefComplaint,
		PresentIllness:   presentIllness,
		PastHistory:      pastHistory,
		AllergyHistory:   allergyHistory,
		FamilyHistory:    familyHistory,
		Temperature:      temperature,
		Pulse:            pulse,
		BloodPressure:    bloodPressure,
		Respiration:      respiration,
		PhysicalExam:     physicalExam,
		LabExam:          labExam,
		ImagingExam:      imagingExam,
		DiagnosisResult:  diagnosisResult,
		TreatmentPlan:    treatmentPlan,
		MedicationAdvice: medicationAdvice,
		DoctorAdvice:     doctorAdvice,
	}
	// 写入账本
	if err := utils.WriteLedger(prescription, stub, model.PrescriptionKey, []string{prescription.Patient, prescription.ID}); err != nil {
		return shim.Error(fmt.Sprintf("%s", err))
	}
	prescriptionByte, err := json.Marshal(prescription)
	if err != nil {
		return shim.Error(fmt.Sprintf("序列化成功创建的信息出错: %s", err))
	}
	return shim.Success(prescriptionByte)
}

// QueryPrescription 查询处方(可查询所有，也可根据所有人查询名下处方)
func QueryPrescription(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var prescriptionList []model.Prescription
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.PrescriptionKey, args)
	if err != nil {
		return shim.Error(fmt.Sprintf("%s", err))
	}
	for _, v := range results {
		if v != nil {
			var p model.Prescription
			err := json.Unmarshal(v, &p)
			if err != nil {
				return shim.Error(fmt.Sprintf("QueryPrescription-反序列化出错: %s", err))
			}
			prescriptionList = append(prescriptionList, p)
		}
	}
	prescriptionByte, err := json.Marshal(prescriptionList)
	if err != nil {
		return shim.Error(fmt.Sprintf("QueryPrescription-序列化出错: %s", err))
	}
	return shim.Success(prescriptionByte)
}

// QueryPatient 查询患者
func QueryPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var patientList []model.Patient
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.PatientKey, args)
	if err != nil {
		return shim.Error(fmt.Sprintf("%s", err))
	}
	for _, v := range results {
		if v != nil {
			var p model.Patient
			err := json.Unmarshal(v, &p)
			if err != nil {
				return shim.Error(fmt.Sprintf("QueryPatient-反序列化出错: %s", err))
			}
			patientList = append(patientList, p)
		}
	}
	patientByte, err := json.Marshal(patientList)
	if err != nil {
		return shim.Error(fmt.Sprintf("QueryPatient-序列化出错: %s", err))
	}
	return shim.Success(patientByte)
}
