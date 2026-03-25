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

const authTimeLayout = "2006-01-02 15:04:05"

func nowAuthStatus(auth *model.RecordAuthorization, now time.Time) string {
	if auth.Status == "revoked" {
		return "revoked"
	}
	endAt, err := time.ParseInLocation(authTimeLayout, auth.EndTime, time.Local)
	if err != nil {
		return "expired"
	}
	if now.After(endAt) {
		return "expired"
	}
	return "active"
}

func doctorEndTimeIndexValue(endTime string) string {
	endAt, err := time.ParseInLocation(authTimeLayout, endTime, time.Local)
	if err != nil {
		return "00000000000000"
	}
	return endAt.Format("20060102150405")
}

func putAuthorizationIndexes(stub shim.ChaincodeStubInterface, auth *model.RecordAuthorization) error {
	if err := utils.WriteLedger(auth.ID, stub, model.AuthorizationPatientIndexKey, []string{auth.PatientID, auth.RecordID, auth.ID}); err != nil {
		return err
	}
	statusForIndex := auth.Status
	if statusForIndex != "revoked" {
		statusForIndex = "active"
	}
	return utils.WriteLedger(auth.ID, stub, model.AuthorizationDoctorIndexKey, []string{auth.DoctorID, statusForIndex, doctorEndTimeIndexValue(auth.EndTime), auth.ID})
}

func deleteAuthorizationDoctorIndexes(stub shim.ChaincodeStubInterface, auth *model.RecordAuthorization) {
	_ = utils.DelLedger(stub, model.AuthorizationDoctorIndexKey, []string{auth.DoctorID, "active", doctorEndTimeIndexValue(auth.EndTime), auth.ID})
	_ = utils.DelLedger(stub, model.AuthorizationDoctorIndexKey, []string{auth.DoctorID, "revoked", doctorEndTimeIndexValue(auth.EndTime), auth.ID})
}

// GrantRecordAuthorization 授权病历查看权限
// args: patientID, recordID, doctorID, hospitalName, department, endTime, remark
func GrantRecordAuthorization(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 7 {
		return shim.Error("参数个数不满足")
	}
	patientID := args[0]
	recordID := args[1]
	doctorID := args[2]
	hospitalName := args[3]
	department := args[4]
	endTimeStr := args[5]
	remark := args[6]

	if patientID == "" || recordID == "" || doctorID == "" || endTimeStr == "" {
		return shim.Error("参数存在空值")
	}

	endTime, err := time.ParseInLocation(authTimeLayout, endTimeStr, time.Local)
	if err != nil {
		return shim.Error("授权截止时间格式错误，应为2006-01-02 15:04:05")
	}
	now := time.Now()
	if !endTime.After(now) {
		return shim.Error("授权截止时间必须晚于当前时间")
	}

	patientBytes, err := utils.GetStateByPartialCompositeKeys(stub, model.AccountV2Key, []string{patientID})
	if err != nil || len(patientBytes) != 1 {
		return shim.Error(fmt.Sprintf("患者身份校验失败: %v", err))
	}
	var patient model.AccountV2
	if err = json.Unmarshal(patientBytes[0], &patient); err != nil {
		return shim.Error(fmt.Sprintf("患者信息解析失败: %v", err))
	}
	if patient.Role != "patient" {
		return shim.Error("仅患者可发起授权")
	}

	doctorBytes, err := utils.GetStateByPartialCompositeKeys(stub, model.AccountV2Key, []string{doctorID})
	if err != nil || len(doctorBytes) != 1 {
		return shim.Error(fmt.Sprintf("医生身份校验失败: %v", err))
	}
	var doctor model.AccountV2
	if err = json.Unmarshal(doctorBytes[0], &doctor); err != nil {
		return shim.Error(fmt.Sprintf("医生信息解析失败: %v", err))
	}
	if doctor.Role != "doctor" {
		return shim.Error("被授权对象不是医生")
	}

	prescriptionBytes, err := utils.GetStateByPartialCompositeKeys2(stub, model.PrescriptionKey, []string{patientID})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询病历失败: %v", err))
	}
	found := false
	for _, b := range prescriptionBytes {
		var p model.Prescription
		if err = json.Unmarshal(b, &p); err != nil {
			continue
		}
		if p.ID == recordID {
			found = true
			break
		}
	}
	if !found {
		return shim.Error("该病历不属于当前患者或不存在")
	}

	auth := &model.RecordAuthorization{
		ID:           stub.GetTxID()[:16],
		RecordID:     recordID,
		PatientID:    patientID,
		DoctorID:     doctorID,
		HospitalName: hospitalName,
		Department:   department,
		Scope:        "read",
		Status:       "active",
		StartTime:    now.Format(authTimeLayout),
		EndTime:      endTime.Format(authTimeLayout),
		CreatedTime:  now.Format(authTimeLayout),
		UpdatedTime:  now.Format(authTimeLayout),
		Remark:       remark,
	}

	if err := utils.WriteLedger(auth, stub, model.AuthorizationKey, []string{auth.PatientID, auth.RecordID, auth.ID}); err != nil {
		return shim.Error(fmt.Sprintf("写入授权失败: %v", err))
	}
	if err := putAuthorizationIndexes(stub, auth); err != nil {
		return shim.Error(fmt.Sprintf("写入授权索引失败: %v", err))
	}

	payload, _ := json.Marshal(auth)
	return shim.Success(payload)
}

// RenewRecordAuthorization 患者续期授权
// args: patientID, authID, endTime
func RenewRecordAuthorization(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("参数个数不满足")
	}
	patientID := args[0]
	authID := args[1]
	endTimeStr := args[2]
	if patientID == "" || authID == "" || endTimeStr == "" {
		return shim.Error("参数存在空值")
	}

	newEndAt, err := time.ParseInLocation(authTimeLayout, endTimeStr, time.Local)
	if err != nil {
		return shim.Error("授权截止时间格式错误，应为2006-01-02 15:04:05")
	}
	if !newEndAt.After(time.Now()) {
		return shim.Error("授权截止时间必须晚于当前时间")
	}

	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuthorizationKey, []string{patientID})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询授权失败: %v", err))
	}
	for _, v := range results {
		var a model.RecordAuthorization
		if err = json.Unmarshal(v, &a); err != nil {
			continue
		}
		if a.ID != authID {
			continue
		}
		if a.Status == "revoked" {
			return shim.Error("授权已撤销，无法续期")
		}
		deleteAuthorizationDoctorIndexes(stub, &a)
		a.EndTime = newEndAt.Format(authTimeLayout)
		a.Status = "active"
		a.UpdatedTime = time.Now().Format(authTimeLayout)
		if err := utils.WriteLedger(&a, stub, model.AuthorizationKey, []string{a.PatientID, a.RecordID, a.ID}); err != nil {
			return shim.Error(fmt.Sprintf("续期失败: %v", err))
		}
		if err := putAuthorizationIndexes(stub, &a); err != nil {
			return shim.Error(fmt.Sprintf("写入授权索引失败: %v", err))
		}
		payload, _ := json.Marshal(a)
		return shim.Success(payload)
	}
	return shim.Error("未找到授权记录")
}

// QueryRecordAuthorizationsByPatient 查询患者授权列表
// args: patientID
func QueryRecordAuthorizationsByPatient(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("参数个数不满足")
	}
	patientID := args[0]
	if patientID == "" {
		return shim.Error("参数存在空值")
	}
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuthorizationKey, []string{patientID})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询授权失败: %v", err))
	}
	var list []model.RecordAuthorization
	now := time.Now()
	for _, v := range results {
		var a model.RecordAuthorization
		if err = json.Unmarshal(v, &a); err == nil {
			a.Status = nowAuthStatus(&a, now)
			list = append(list, a)
		}
	}
	payload, _ := json.Marshal(list)
	return shim.Success(payload)
}

// QueryAccessibleRecordsByDoctor 查询医生被授权可访问病历
// args: doctorID, patientNameKeyword, idCardKeyword, recordTypeKeyword, createdStart, createdEnd
func QueryAccessibleRecordsByDoctor(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 6 {
		return shim.Error("参数个数不满足")
	}
	doctorID := strings.TrimSpace(args[0])
	patientNameKeyword := strings.TrimSpace(strings.ToLower(args[1]))
	idCardKeyword := strings.TrimSpace(strings.ToLower(args[2]))
	recordTypeKeyword := strings.TrimSpace(strings.ToLower(args[3]))
	createdStart := strings.TrimSpace(args[4])
	createdEnd := strings.TrimSpace(args[5])
	if doctorID == "" {
		return shim.Error("参数存在空值")
	}

	allAuths, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuthorizationKey, []string{})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询授权失败: %v", err))
	}

	accountBytes, err := utils.GetStateByPartialCompositeKeys2(stub, model.AccountV2Key, []string{})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询账户失败: %v", err))
	}
	patientMap := make(map[string]model.AccountV2)
	for _, b := range accountBytes {
		var a model.AccountV2
		if err = json.Unmarshal(b, &a); err == nil {
			patientMap[a.AccountId] = a
		}
	}

	prescriptionBytes, err := utils.GetStateByPartialCompositeKeys2(stub, model.PrescriptionKey, []string{})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询病历失败: %v", err))
	}
	recordMap := make(map[string]model.Prescription)
	for _, b := range prescriptionBytes {
		var p model.Prescription
		if err = json.Unmarshal(b, &p); err == nil {
			recordMap[p.ID] = p
		}
	}

	now := time.Now()
	result := make([]map[string]interface{}, 0)
	for _, b := range allAuths {
		var auth model.RecordAuthorization
		if err = json.Unmarshal(b, &auth); err != nil {
			continue
		}
		if auth.DoctorID != doctorID || nowAuthStatus(&auth, now) != "active" {
			continue
		}

		record, ok := recordMap[auth.RecordID]
		if !ok {
			continue
		}
		patient := patientMap[auth.PatientID]

		patientName := strings.ToLower(patient.AccountName)
		idCard := strings.ToLower(patient.IDCardNo)
		recordType := strings.ToLower(record.RecordType)
		if patientNameKeyword != "" && !strings.Contains(patientName, patientNameKeyword) {
			continue
		}
		if idCardKeyword != "" && !strings.Contains(idCard, idCardKeyword) {
			continue
		}
		if recordTypeKeyword != "" && !strings.Contains(recordType, recordTypeKeyword) {
			continue
		}
		createdDate := ""
		if len(record.Created) >= 10 {
			createdDate = record.Created[:10]
		}
		if createdStart != "" && createdDate < createdStart {
			continue
		}
		if createdEnd != "" && createdDate > createdEnd {
			continue
		}
		item := map[string]interface{}{
			"authorization": auth,
			"record":        record,
			"patient": map[string]interface{}{
				"account_id":   patient.AccountId,
				"account_name": patient.AccountName,
				"id_card_no":   patient.IDCardNo,
			},
		}
		result = append(result, item)
	}
	payload, _ := json.Marshal(result)
	return shim.Success(payload)
}

// RevokeRecordAuthorization 撤销授权
// args: patientID, authID
func RevokeRecordAuthorization(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("参数个数不满足")
	}
	patientID := args[0]
	authID := args[1]
	if patientID == "" || authID == "" {
		return shim.Error("参数存在空值")
	}

	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuthorizationKey, []string{patientID})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询授权失败: %v", err))
	}

	for _, v := range results {
		var a model.RecordAuthorization
		if err = json.Unmarshal(v, &a); err != nil {
			continue
		}
		if a.ID == authID {
			deleteAuthorizationDoctorIndexes(stub, &a)
			a.Status = "revoked"
			a.UpdatedTime = time.Now().Format(authTimeLayout)
			if err := utils.WriteLedger(&a, stub, model.AuthorizationKey, []string{a.PatientID, a.RecordID, a.ID}); err != nil {
				return shim.Error(fmt.Sprintf("撤销授权失败: %v", err))
			}
			if err := putAuthorizationIndexes(stub, &a); err != nil {
				return shim.Error(fmt.Sprintf("写入授权索引失败: %v", err))
			}
			payload, _ := json.Marshal(a)
			return shim.Success(payload)
		}
	}
	return shim.Error("未找到授权记录")
}

// CheckRecordAccess 校验医生是否有某病历访问权限
// args: doctorID, recordID
func CheckRecordAccess(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("参数个数不满足")
	}
	doctorID := args[0]
	recordID := args[1]
	if doctorID == "" || recordID == "" {
		return shim.Error("参数存在空值")
	}

	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuthorizationKey, []string{})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询授权失败: %v", err))
	}

	now := time.Now()
	for _, v := range results {
		var a model.RecordAuthorization
		if err = json.Unmarshal(v, &a); err != nil {
			continue
		}
		if a.DoctorID != doctorID || a.RecordID != recordID {
			continue
		}
		if nowAuthStatus(&a, now) == "active" {
			return shim.Success([]byte(`{"allowed":true,"reason":"authorized"}`))
		}
	}

	return shim.Success([]byte(`{"allowed":false,"reason":"unauthorized_or_expired"}`))
}
