package audit

import (
	"application/blockchain"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// 事件类型常量
const (
	EventTypeLogin              = "LOGIN"
	EventTypeQueryRecord        = "QUERY_RECORD"
	EventTypeCreateRecord       = "CREATE_RECORD"
	EventTypeGrantRecordAuth    = "GRANT_RECORD_AUTH"
	EventTypeRenewRecordAuth    = "RENEW_RECORD_AUTH"
	EventTypeRevokeRecordAuth   = "REVOKE_RECORD_AUTH"
	EventTypeCheckRecordAccess  = "CHECK_RECORD_ACCESS"
	EventTypeExportReportCreate = "EXPORT_REPORT_CREATE"
	EventTypeExportReportDown   = "EXPORT_REPORT_DOWNLOAD"
	EventTypeAlertAck           = "ALERT_ACK"
	EventTypeAlertResolve       = "ALERT_RESOLVE"
	EventTypeAPIError           = "API_ERROR"
	EventTypeChaincodeError     = "CHAINCODE_ERROR"
	EventTypeOutpatientReg      = "OUTPATIENT_REGISTRATION"
	EventTypeOutpatientVisit    = "OUTPATIENT_VISIT"
	EventTypeDrugOrder          = "DRUG_ORDER"
	EventTypeInsurance          = "INSURANCE_CLAIM"
)

// 事件级别
const (
	EventLevelL1 = "L1"
	EventLevelL2 = "L2"
	EventLevelL3 = "L3"
)

// AuditContext 审计上下文，存储请求相关信息
type AuditContext struct {
	EventType       string
	EventLevel      string
	ActorID         string
	TargetPatientID string
	TargetRecordID  string
	ChaincodeFunc   string
	DetailJSON      string
	FailReason      string
	RequestPath     string
	RequestMethod   string
	ClientIP        string
	UserAgent       string
}

// generateEventID 生成审计事件ID
func generateEventID() string {
	timestamp := time.Now().UnixNano() / 1e6
	randNum := rand.Intn(999999)
	return fmt.Sprintf("EVT%d%06d", timestamp, randNum)
}

// WriteAuditEvent 写入审计事件到链上
func WriteAuditEvent(ctx *AuditContext) error {
	eventID := generateEventID()
	eventTime := time.Now().Format("2006-01-02 15:04:05")

	event := map[string]interface{}{
		"id":                eventID,
		"event_type":        ctx.EventType,
		"event_level":       ctx.EventLevel,
		"event_time":        eventTime,
		"actor_id":          ctx.ActorID,
		"target_patient_id": ctx.TargetPatientID,
		"target_record_id":  ctx.TargetRecordID,
		"chaincode_func":    ctx.ChaincodeFunc,
		"request_path":      ctx.RequestPath,
		"request_method":    ctx.RequestMethod,
		"client_ip":         ctx.ClientIP,
		"user_agent":        ctx.UserAgent,
		"action_result":     "SUCCESS",
		"detail_json":       ctx.DetailJSON,
	}

	eventBytes, _ := json.Marshal(event)
	_, err := blockchain.ChannelExecute("createAuditEvent", [][]byte{eventBytes})
	return err
}

// WriteAuditEventWithFail 写入失败审计事件到链上
func WriteAuditEventWithFail(ctx *AuditContext) error {
	eventID := generateEventID()
	eventTime := time.Now().Format("2006-01-02 15:04:05")

	event := map[string]interface{}{
		"id":                eventID,
		"event_type":        ctx.EventType,
		"event_level":       ctx.EventLevel,
		"event_time":        eventTime,
		"actor_id":          ctx.ActorID,
		"target_patient_id": ctx.TargetPatientID,
		"target_record_id":  ctx.TargetRecordID,
		"chaincode_func":    ctx.ChaincodeFunc,
		"request_path":      ctx.RequestPath,
		"request_method":    ctx.RequestMethod,
		"client_ip":         ctx.ClientIP,
		"user_agent":        ctx.UserAgent,
		"action_result":     "FAIL",
		"fail_reason":       ctx.FailReason,
		"detail_json":       ctx.DetailJSON,
	}

	eventBytes, _ := json.Marshal(event)
	_, err := blockchain.ChannelExecute("createAuditEvent", [][]byte{eventBytes})
	return err
}

// GetEventLevelByType 根据事件类型获取事件级别
func GetEventLevelByType(eventType string) string {
	// L2级别：敏感操作
	l2Types := []string{
		EventTypeGrantRecordAuth,
		EventTypeRenewRecordAuth,
		EventTypeRevokeRecordAuth,
		EventTypeExportReportCreate,
		EventTypeExportReportDown,
		EventTypeAlertAck,
		EventTypeAlertResolve,
	}
	for _, t := range l2Types {
		if eventType == t {
			return EventLevelL2
		}
	}
	// L3级别：错误和异常
	l3Types := []string{
		EventTypeAPIError,
		EventTypeChaincodeError,
	}
	for _, t := range l3Types {
		if eventType == t {
			return EventLevelL3
		}
	}
	// 默认L1：普通操作
	return EventLevelL1
}

// GetChaincodeFuncByPath 根据请求路径推断链码函数
func GetChaincodeFuncByPath(path, method string) string {
	path = strings.ToLower(path)

	// 病历相关
	if strings.Contains(path, "prescription") {
		if method == "POST" && strings.Contains(path, "create") {
			return "createPrescription"
		}
		if method == "POST" && strings.Contains(path, "query") {
			return "queryPrescription"
		}
	}

	// 授权相关
	if strings.Contains(path, "authorization") {
		if strings.Contains(path, "grant") {
			return "grantRecordAuthorization"
		}
		if strings.Contains(path, "renew") {
			return "renewRecordAuthorization"
		}
		if strings.Contains(path, "revoke") {
			return "revokeRecordAuthorization"
		}
		if strings.Contains(path, "check") {
			return "checkRecordAccess"
		}
		if strings.Contains(path, "accessible") {
			return "queryAccessibleRecordsByDoctor"
		}
	}

	// 门诊相关
	if strings.Contains(path, "outpatient") {
		if strings.Contains(path, "registration") {
			if strings.Contains(path, "create") {
				return "createOutpatientRegistration"
			}
			if strings.Contains(path, "cancel") {
				return "cancelOutpatientRegistration"
			}
			return "queryOutpatientRegistration"
		}
		if strings.Contains(path, "slot") {
			if strings.Contains(path, "create") {
				return "createScheduleSlot"
			}
			return "queryScheduleSlot"
		}
		if strings.Contains(path, "payment") {
			if strings.Contains(path, "pay") {
				return "payOutpatientOrder"
			}
			return "queryOutpatientPayment"
		}
		if strings.Contains(path, "queue") {
			if strings.Contains(path, "start") {
				return "startOutpatientVisit"
			}
			if strings.Contains(path, "finish") {
				return "finishOutpatientVisit"
			}
			return "queryOutpatientQueue"
		}
		if strings.Contains(path, "record") {
			return "queryOutpatientRecord"
		}
	}

	// 保险相关
	if strings.Contains(path, "insurance") {
		if method == "POST" && strings.Contains(path, "create") {
			return "createInsuranceCover"
		}
		if strings.Contains(path, "update") {
			return "updateInsuranceCover"
		}
		if strings.Contains(path, "delete") {
			return "deleteInsuranceCover"
		}
	}

	// 药品订单
	if strings.Contains(path, "drugorder") || strings.Contains(path, "drug-order") {
		return "createDrugOrder"
	}

	// 审计相关
	if strings.Contains(path, "audit") {
		if strings.Contains(path, "export") {
			if strings.Contains(path, "download") {
				return "downloadAuditReport"
			}
			return "createAuditExportTask"
		}
		if strings.Contains(path, "alert") {
			if strings.Contains(path, "ack") {
				return "ackAuditAlert"
			}
			if strings.Contains(path, "resolve") {
				return "resolveAuditAlert"
			}
		}
	}

	return ""
}

// GetEventTypeByPath 根据请求路径推断事件类型
func GetEventTypeByPath(path, method string) string {
	path = strings.ToLower(path)

	// 登录
	if strings.Contains(path, "login") || strings.Contains(path, "auth") {
		return EventTypeLogin
	}

	// 病历查询
	if strings.Contains(path, "prescription") && method == "POST" && strings.Contains(path, "query") {
		return EventTypeQueryRecord
	}

	// 创建病历
	if strings.Contains(path, "prescription") && method == "POST" && strings.Contains(path, "create") {
		return EventTypeCreateRecord
	}

	// 授权操作
	if strings.Contains(path, "authorization") {
		if strings.Contains(path, "grant") {
			return EventTypeGrantRecordAuth
		}
		if strings.Contains(path, "renew") {
			return EventTypeRenewRecordAuth
		}
		if strings.Contains(path, "revoke") {
			return EventTypeRevokeRecordAuth
		}
		if strings.Contains(path, "check") {
			return EventTypeCheckRecordAccess
		}
	}

	// 门诊挂号
	if strings.Contains(path, "outpatient") && strings.Contains(path, "registration") {
		if strings.Contains(path, "create") {
			return EventTypeOutpatientReg
		}
	}

	// 门诊就诊
	if strings.Contains(path, "outpatient") && (strings.Contains(path, "start") || strings.Contains(path, "finish")) {
		return EventTypeOutpatientVisit
	}

	// 导出报表
	if strings.Contains(path, "audit") && strings.Contains(path, "export") {
		if strings.Contains(path, "download") {
			return EventTypeExportReportDown
		}
		return EventTypeExportReportCreate
	}

	// 告警处置
	if strings.Contains(path, "audit") && strings.Contains(path, "alert") {
		if strings.Contains(path, "ack") {
			return EventTypeAlertAck
		}
		if strings.Contains(path, "resolve") {
			return EventTypeAlertResolve
		}
	}

	// 保险
	if strings.Contains(path, "insurance") {
		return EventTypeInsurance
	}

	// 药品订单
	if strings.Contains(path, "drugorder") || strings.Contains(path, "drug-order") {
		return EventTypeDrugOrder
	}

	return ""
}

// MaskSensitiveData 脱敏敏感数据
func MaskSensitiveData(data string) string {
	if len(data) <= 4 {
		return "****"
	}
	return data[:2] + "****" + data[len(data)-2:]
}

// ExtractPatientIDFromRequest 从请求体中提取患者ID
func ExtractPatientIDFromRequest(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}
	if patientID, ok := data["patient"].(string); ok {
		return patientID
	}
	if patientID, ok := data["patient_id"].(string); ok {
		return patientID
	}
	return ""
}

// ExtractRecordIDFromRequest 从请求体中提取病历ID
func ExtractRecordIDFromRequest(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}
	if recordID, ok := data["record_id"].(string); ok {
		return recordID
	}
	if prescription, ok := data["prescription"].(string); ok {
		return prescription
	}
	return ""
}

// HashRequest 生成请求摘要哈希
func HashRequest(data []byte) string {
	hash := md5.Sum(data)
	return fmt.Sprintf("%x", hash)[:16]
}
