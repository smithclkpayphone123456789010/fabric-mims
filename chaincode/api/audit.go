package api

import (
	"chaincode/model"
	"chaincode/pkg/utils"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const auditTimeLayout = "2006-01-02 15:04:05"

// generateEventID 生成审计事件ID
func generateEventID(stub shim.ChaincodeStubInterface) string {
	txID := stub.GetTxID()
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("EVT%d%s", timestamp%1000000, txID[:8])
}

// generateAlertID 生成告警ID
func generateAlertID() string {
	return fmt.Sprintf("ALT%d", time.Now().UnixNano()%1000000)
}

// calculateHash 计算审计事件哈希
func calculateAuditHash(event *model.AuditEvent) string {
	data := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%s",
		event.ID, event.EventType, event.EventTime, event.ActorID,
		event.ActionResult, event.TargetPatientID, event.TargetRecordID,
		event.DetailJSON, event.HashPrev)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// getLastHash 获取哈希链最后一条哈希
func getLastHash(stub shim.ChaincodeStubInterface) (string, error) {
	results, err := utils.GetStateByPartialCompositeKeys(stub, model.AuditEventHashChainKey, []string{"last"})
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "GENESIS", nil
	}
	return string(results[0]), nil
}

// updateHashChain 更新哈希链
func updateHashChain(stub shim.ChaincodeStubInterface, hash string) error {
	return utils.WriteLedger(hash, stub, model.AuditEventHashChainKey, []string{"last"})
}

// putAuditEventIndexes 写入审计事件索引
func putAuditEventIndexes(stub shim.ChaincodeStubInterface, event *model.AuditEvent) error {
	// 时间索引
	if err := utils.WriteLedger(event.ID, stub, model.AuditEventTimeIndexKey,
		[]string{event.EventTime, event.ID}); err != nil {
		return err
	}
	// 类型索引
	if err := utils.WriteLedger(event.ID, stub, model.AuditEventTypeIndexKey,
		[]string{event.EventType, event.EventTime, event.ID}); err != nil {
		return err
	}
	// 操作者索引
	if err := utils.WriteLedger(event.ID, stub, model.AuditEventActorIndexKey,
		[]string{event.ActorID, event.EventTime, event.ID}); err != nil {
		return err
	}
	// 患者索引
	if event.TargetPatientID != "" {
		if err := utils.WriteLedger(event.ID, stub, model.AuditEventPatientIndexKey,
			[]string{event.TargetPatientID, event.EventTime, event.ID}); err != nil {
			return err
		}
	}
	// 病历索引
	if event.TargetRecordID != "" {
		if err := utils.WriteLedger(event.ID, stub, model.AuditEventRecordIndexKey,
			[]string{event.TargetRecordID, event.EventTime, event.ID}); err != nil {
			return err
		}
	}
	// 交易索引
	if event.TxID != "" {
		if err := utils.WriteLedger(event.ID, stub, model.AuditEventTxIndexKey,
			[]string{event.TxID, event.ID}); err != nil {
			return err
		}
	}
	return nil
}

// CreateAuditEvent 创建审计事件
// args: [AuditEvent JSON]
func CreateAuditEvent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("参数个数不满足")
	}

	var event model.AuditEvent
	if err := json.Unmarshal([]byte(args[0]), &event); err != nil {
		return shim.Error(fmt.Sprintf("审计事件解析失败: %v", err))
	}

	// 使用服务器端传入的ID（若未提供则生成）
	if event.ID == "" {
		event.ID = generateEventID(stub)
	}

	// 规范化时间格式
	if event.EventTime == "" {
		event.EventTime = time.Now().Format(auditTimeLayout)
	}

	// 写入主键
	if err := utils.WriteLedger(event, stub, model.AuditEventKey, []string{event.ID}); err != nil {
		return shim.Error(fmt.Sprintf("写入审计事件失败: %v", err))
	}

	// 写入复合索引
	if err := putAuditEventIndexes(stub, &event); err != nil {
		return shim.Error(fmt.Sprintf("写入索引失败: %v", err))
	}

	payload, _ := json.Marshal(event)
	return shim.Success(payload)
}

// GetAuditEventsByCompositeKey 按复合条件查询审计事件
// args: [keyType, key1, key2, key3, page, size]
func GetAuditEventsByCompositeKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 6 {
		return shim.Error("参数个数不满足")
	}

	keyType := args[0]
	key1 := args[1]
	key2 := args[2]
	key3 := args[3]
	page, _ := strconv.Atoi(args[4])
	size, _ := strconv.Atoi(args[5])
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	var results [][]byte
	var err error

	switch keyType {
	case "time":
		// 时间范围查询: key1=startTime, key2=endTime
		results, err = queryAuditEventsByTimeRange(stub, key1, key2, page, size)
	case "type":
		// 类型查询: key1=eventType, key2=startTime, key3=endTime
		results, err = queryAuditEventsByType(stub, key1, key2, key3, page, size)
	case "actor":
		// 操作者查询: key1=actorID, key2=startTime, key3=endTime
		results, err = queryAuditEventsByActor(stub, key1, key2, key3, page, size)
	case "patient":
		// 患者查询: key1=patientID, key2=startTime, key3=endTime
		results, err = queryAuditEventsByPatient(stub, key1, key2, key3, page, size)
	case "record":
		// 病历查询: key1=recordID, key2=startTime, key3=endTime
		results, err = queryAuditEventsByRecord(stub, key1, key2, key3, page, size)
	case "all":
		// 全量查询
		results, err = queryAllAuditEvents(stub, page, size)
	default:
		return shim.Error("不支持的索引类型")
	}

	if err != nil {
		return shim.Error(fmt.Sprintf("查询失败: %v", err))
	}

	var events []model.AuditEvent
	for _, v := range results {
		var event model.AuditEvent
		if err := json.Unmarshal(v, &event); err != nil {
			continue
		}
		events = append(events, event)
	}

	payload, _ := json.Marshal(events)
	return shim.Success(payload)
}

// queryAuditEventsByTimeRange 按时间范围查询
func queryAuditEventsByTimeRange(stub shim.ChaincodeStubInterface, startTime, endTime string, page, size int) ([][]byte, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditEventTimeIndexKey, []string{})
	if err != nil {
		return nil, err
	}

	// 收集符合条件的事件ID
	var eventIDs []string
	for _, v := range results {
		eventID := string(v)
		// 根据ID查询完整事件
		eventData, err := getAuditEventByID(stub, eventID)
		if err != nil {
			continue
		}
		// 时间过滤
		if startTime != "" && eventData.EventTime < startTime {
			continue
		}
		if endTime != "" && eventData.EventTime > endTime {
			continue
		}
		eventIDs = append(eventIDs, eventID)
	}

	// 分页
	start := (page - 1) * size
	end := start + size
	if start >= len(eventIDs) {
		return [][]byte{}, nil
	}
	if end > len(eventIDs) {
		end = len(eventIDs)
	}

	resultBytes := make([][]byte, 0)
	for i := start; i < end; i++ {
		eventData, _ := getAuditEventByID(stub, eventIDs[i])
		data, _ := json.Marshal(eventData)
		resultBytes = append(resultBytes, data)
	}

	return resultBytes, nil
}

// getAuditEventByID 根据ID获取审计事件
func getAuditEventByID(stub shim.ChaincodeStubInterface, eventID string) (*model.AuditEvent, error) {
	key, _ := stub.CreateCompositeKey(model.AuditEventKey, []string{eventID})
	bytes, err := stub.GetState(key)
	if err != nil || len(bytes) == 0 {
		return nil, fmt.Errorf("事件不存在")
	}
	var event model.AuditEvent
	if err := json.Unmarshal(bytes, &event); err != nil {
		return nil, err
	}
	return &event, nil
}

// queryAuditEventsByType 按事件类型查询
func queryAuditEventsByType(stub shim.ChaincodeStubInterface, eventType, startTime, endTime string, page, size int) ([][]byte, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditEventTypeIndexKey, []string{eventType})
	if err != nil {
		return nil, err
	}

	// 收集符合条件的事件ID
	var eventIDs []string
	for _, v := range results {
		eventID := string(v)
		eventData, err := getAuditEventByID(stub, eventID)
		if err != nil {
			continue
		}
		if startTime != "" && eventData.EventTime < startTime {
			continue
		}
		if endTime != "" && eventData.EventTime > endTime {
			continue
		}
		eventIDs = append(eventIDs, eventID)
	}

	// 分页
	start := (page - 1) * size
	end := start + size
	if start >= len(eventIDs) {
		return [][]byte{}, nil
	}
	if end > len(eventIDs) {
		end = len(eventIDs)
	}

	resultBytes := make([][]byte, 0)
	for i := start; i < end; i++ {
		eventData, _ := getAuditEventByID(stub, eventIDs[i])
		data, _ := json.Marshal(eventData)
		resultBytes = append(resultBytes, data)
	}

	return resultBytes, nil
}

// queryAuditEventsByActor 按操作者查询
func queryAuditEventsByActor(stub shim.ChaincodeStubInterface, actorID, startTime, endTime string, page, size int) ([][]byte, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditEventActorIndexKey, []string{actorID})
	if err != nil {
		return nil, err
	}

	// 收集符合条件的事件ID
	var eventIDs []string
	for _, v := range results {
		eventID := string(v)
		eventData, err := getAuditEventByID(stub, eventID)
		if err != nil {
			continue
		}
		if startTime != "" && eventData.EventTime < startTime {
			continue
		}
		if endTime != "" && eventData.EventTime > endTime {
			continue
		}
		eventIDs = append(eventIDs, eventID)
	}

	// 分页
	start := (page - 1) * size
	end := start + size
	if start >= len(eventIDs) {
		return [][]byte{}, nil
	}
	if end > len(eventIDs) {
		end = len(eventIDs)
	}

	resultBytes := make([][]byte, 0)
	for i := start; i < end; i++ {
		eventData, _ := getAuditEventByID(stub, eventIDs[i])
		data, _ := json.Marshal(eventData)
		resultBytes = append(resultBytes, data)
	}

	return resultBytes, nil
}

// queryAuditEventsByPatient 按患者查询
func queryAuditEventsByPatient(stub shim.ChaincodeStubInterface, patientID, startTime, endTime string, page, size int) ([][]byte, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditEventPatientIndexKey, []string{patientID})
	if err != nil {
		return nil, err
	}

	// 收集符合条件的事件ID
	var eventIDs []string
	for _, v := range results {
		eventID := string(v)
		eventData, err := getAuditEventByID(stub, eventID)
		if err != nil {
			continue
		}
		if startTime != "" && eventData.EventTime < startTime {
			continue
		}
		if endTime != "" && eventData.EventTime > endTime {
			continue
		}
		eventIDs = append(eventIDs, eventID)
	}

	// 分页
	start := (page - 1) * size
	end := start + size
	if start >= len(eventIDs) {
		return [][]byte{}, nil
	}
	if end > len(eventIDs) {
		end = len(eventIDs)
	}

	resultBytes := make([][]byte, 0)
	for i := start; i < end; i++ {
		eventData, _ := getAuditEventByID(stub, eventIDs[i])
		data, _ := json.Marshal(eventData)
		resultBytes = append(resultBytes, data)
	}

	return resultBytes, nil
}

// queryAuditEventsByRecord 按病历查询
func queryAuditEventsByRecord(stub shim.ChaincodeStubInterface, recordID, startTime, endTime string, page, size int) ([][]byte, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditEventRecordIndexKey, []string{recordID})
	if err != nil {
		return nil, err
	}

	// 收集符合条件的事件ID
	var eventIDs []string
	for _, v := range results {
		eventID := string(v)
		eventData, err := getAuditEventByID(stub, eventID)
		if err != nil {
			continue
		}
		if startTime != "" && eventData.EventTime < startTime {
			continue
		}
		if endTime != "" && eventData.EventTime > endTime {
			continue
		}
		eventIDs = append(eventIDs, eventID)
	}

	// 分页
	start := (page - 1) * size
	end := start + size
	if start >= len(eventIDs) {
		return [][]byte{}, nil
	}
	if end > len(eventIDs) {
		end = len(eventIDs)
	}

	resultBytes := make([][]byte, 0)
	for i := start; i < end; i++ {
		eventData, _ := getAuditEventByID(stub, eventIDs[i])
		data, _ := json.Marshal(eventData)
		resultBytes = append(resultBytes, data)
	}

	return resultBytes, nil
}

// queryAllAuditEvents 查询所有审计事件
func queryAllAuditEvents(stub shim.ChaincodeStubInterface, page, size int) ([][]byte, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditEventKey, []string{})
	if err != nil {
		return nil, err
	}

	var events []model.AuditEvent
	for _, v := range results {
		var event model.AuditEvent
		if err := json.Unmarshal(v, &event); err != nil {
			continue
		}
		events = append(events, event)
	}

	// 分页
	start := (page - 1) * size
	end := start + size
	if start >= len(events) {
		return [][]byte{}, nil
	}
	if end > len(events) {
		end = len(events)
	}

	resultBytes := make([][]byte, 0)
	for i := start; i < end; i++ {
		data, _ := json.Marshal(events[i])
		resultBytes = append(resultBytes, data)
	}

	return resultBytes, nil
}

// GetAuditEventByID 根据ID查询审计事件
// args: [eventID]
func GetAuditEventByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("参数个数不满足")
	}

	eventID := args[0]
	results, err := utils.GetStateByPartialCompositeKeys(stub, model.AuditEventKey, []string{eventID})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询失败: %v", err))
	}
	if len(results) == 0 {
		return shim.Error("审计事件不存在")
	}

	var event model.AuditEvent
	if err := json.Unmarshal(results[0], &event); err != nil {
		return shim.Error(fmt.Sprintf("解析失败: %v", err))
	}

	payload, _ := json.Marshal(event)
	return shim.Success(payload)
}

// GetAuditEventStats 获取审计事件统计
// args: []
func GetAuditEventStats(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditEventKey, []string{})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询失败: %v", err))
	}

	var totalCount, successCount, failCount, l2Count, l3Count int
	today := time.Now().Format("2006-01-02")

	for _, v := range results {
		var event model.AuditEvent
		if err := json.Unmarshal(v, &event); err != nil {
			continue
		}
		totalCount++
		if event.ActionResult == "SUCCESS" {
			successCount++
		} else if event.ActionResult == "FAIL" {
			failCount++
		}
		if event.EventLevel == "L2" {
			l2Count++
		} else if event.EventLevel == "L3" {
			l3Count++
		}
		_ = today // 后续可用于今日统计
	}

	stats := map[string]int{
		"total_count":   totalCount,
		"success_count": successCount,
		"fail_count":    failCount,
		"l2_count":      l2Count,
		"l3_count":      l3Count,
	}

	payload, _ := json.Marshal(stats)
	return shim.Success(payload)
}

// ---------------------- 告警相关 ----------------------

// putAuditAlertIndexes 写入告警索引
func putAuditAlertIndexes(stub shim.ChaincodeStubInterface, alert *model.AuditAlert) error {
	// 状态索引
	if err := utils.WriteLedger(alert.ID, stub, model.AuditAlertStatusIndexKey,
		[]string{alert.Status, alert.TriggerTime, alert.ID}); err != nil {
		return err
	}
	// 级别索引
	if err := utils.WriteLedger(alert.ID, stub, model.AuditAlertLevelIndexKey,
		[]string{alert.Level, alert.TriggerTime, alert.ID}); err != nil {
		return err
	}
	return nil
}

// deleteAuditAlertIndexes 删除告警索引
func deleteAuditAlertIndexes(stub shim.ChaincodeStubInterface, alert *model.AuditAlert) {
	_ = utils.DelLedger(stub, model.AuditAlertStatusIndexKey,
		[]string{alert.Status, alert.TriggerTime, alert.ID})
	_ = utils.DelLedger(stub, model.AuditAlertLevelIndexKey,
		[]string{alert.Level, alert.TriggerTime, alert.ID})
}

// getAlertByID 根据ID获取告警
func getAlertByID(stub shim.ChaincodeStubInterface, alertID string) (*model.AuditAlert, error) {
	results, err := utils.GetStateByPartialCompositeKeys(stub, model.AuditAlertKey, []string{alertID})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("告警不存在")
	}

	var alert model.AuditAlert
	if err := json.Unmarshal(results[0], &alert); err != nil {
		return nil, err
	}
	return &alert, nil
}

// CreateAuditAlert 创建告警
// args: [AuditAlert JSON]
func CreateAuditAlert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("参数个数不满足")
	}

	var alert model.AuditAlert
	if err := json.Unmarshal([]byte(args[0]), &alert); err != nil {
		return shim.Error(fmt.Sprintf("告警解析失败: %v", err))
	}

	if alert.ID == "" {
		alert.ID = generateAlertID()
	}
	if alert.TriggerTime == "" {
		alert.TriggerTime = time.Now().Format(auditTimeLayout)
	}
	alert.Status = "NEW"

	// 写入主键
	if err := utils.WriteLedger(alert, stub, model.AuditAlertKey, []string{alert.ID}); err != nil {
		return shim.Error(fmt.Sprintf("创建告警失败: %v", err))
	}

	// 写入索引
	if err := putAuditAlertIndexes(stub, &alert); err != nil {
		return shim.Error(fmt.Sprintf("写入索引失败: %v", err))
	}

	payload, _ := json.Marshal(alert)
	return shim.Success(payload)
}

// GetAuditAlertsByCompositeKey 按条件查询告警
// args: [keyType, key1, key2, key3, page, size]
func GetAuditAlertsByCompositeKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 4 {
		return shim.Error("参数个数不满足")
	}

	keyType := args[0]
	key1 := args[1]
	_ = args[2] // key2 reserved for future use
	page, _ := strconv.Atoi(args[3])
	size, _ := strconv.Atoi(args[4])
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	var results [][]byte
	var err error

	switch keyType {
	case "status":
		results, err = queryAlertsByStatus(stub, key1, page, size)
	case "level":
		results, err = queryAlertsByLevel(stub, key1, page, size)
	case "all":
		results, err = queryAllAlerts(stub, page, size)
	default:
		return shim.Error("不支持的索引类型")
	}

	if err != nil {
		return shim.Error(fmt.Sprintf("查询失败: %v", err))
	}

	var alerts []model.AuditAlert
	for _, v := range results {
		var alert model.AuditAlert
		if err := json.Unmarshal(v, &alert); err != nil {
			continue
		}
		alerts = append(alerts, alert)
	}

	payload, _ := json.Marshal(alerts)
	return shim.Success(payload)
}

// queryAlertsByStatus 按状态查询告警
func queryAlertsByStatus(stub shim.ChaincodeStubInterface, status string, page, size int) ([][]byte, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditAlertStatusIndexKey, []string{status})
	if err != nil {
		return nil, err
	}

	// 分页
	start := (page - 1) * size
	end := start + size
	if start >= len(results) {
		return [][]byte{}, nil
	}
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// queryAlertsByLevel 按级别查询告警
func queryAlertsByLevel(stub shim.ChaincodeStubInterface, level string, page, size int) ([][]byte, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditAlertLevelIndexKey, []string{level})
	if err != nil {
		return nil, err
	}

	// 分页
	start := (page - 1) * size
	end := start + size
	if start >= len(results) {
		return [][]byte{}, nil
	}
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// queryAllAlerts 查询所有告警
func queryAllAlerts(stub shim.ChaincodeStubInterface, page, size int) ([][]byte, error) {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditAlertKey, []string{})
	if err != nil {
		return nil, err
	}

	// 分页
	start := (page - 1) * size
	end := start + size
	if start >= len(results) {
		return [][]byte{}, nil
	}
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// GetAuditAlertByID 根据ID查询告警
// args: [alertID]
func GetAuditAlertByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("参数个数不满足")
	}

	alertID := args[0]
	alert, err := getAlertByID(stub, alertID)
	if err != nil {
		return shim.Error(err.Error())
	}

	payload, _ := json.Marshal(alert)
	return shim.Success(payload)
}

// AckAlert 确认告警
// args: [alertID]
func AckAlert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("参数个数不满足")
	}

	alertID := args[0]

	alert, err := getAlertByID(stub, alertID)
	if err != nil {
		return shim.Error(err.Error())
	}

	if alert.Status != "NEW" {
		return shim.Error(fmt.Sprintf("当前状态不允许确认: %s", alert.Status))
	}

	// 删除旧索引
	deleteAuditAlertIndexes(stub, alert)

	// 更新状态
	alert.Status = "ACKED"

	// 写回
	if err := utils.WriteLedger(alert, stub, model.AuditAlertKey, []string{alert.ID}); err != nil {
		return shim.Error(fmt.Sprintf("更新告警失败: %v", err))
	}

	// 写入新索引
	if err := putAuditAlertIndexes(stub, alert); err != nil {
		return shim.Error(fmt.Sprintf("写入索引失败: %v", err))
	}

	payload, _ := json.Marshal(alert)
	return shim.Success(payload)
}

// ResolveAlert 关闭/解决告警
// args: [alertID, handleNote]
func ResolveAlert(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("参数个数不满足")
	}

	alertID := args[0]
	handleNote := strings.TrimSpace(args[1])

	if handleNote == "" {
		return shim.Error("处理备注不能为空")
	}

	alert, err := getAlertByID(stub, alertID)
	if err != nil {
		return shim.Error(err.Error())
	}

	if alert.Status == "RESOLVED" {
		return shim.Error("告警已关闭")
	}

	// 删除旧索引
	deleteAuditAlertIndexes(stub, alert)

	// 更新状态
	alert.Status = "RESOLVED"
	alert.HandleTime = time.Now().Format(auditTimeLayout)
	alert.HandleNote = handleNote

	// 写回
	if err := utils.WriteLedger(alert, stub, model.AuditAlertKey, []string{alert.ID}); err != nil {
		return shim.Error(fmt.Sprintf("更新告警失败: %v", err))
	}

	// 写入新索引
	if err := putAuditAlertIndexes(stub, alert); err != nil {
		return shim.Error(fmt.Sprintf("写入索引失败: %v", err))
	}

	payload, _ := json.Marshal(alert)
	return shim.Success(payload)
}

// GetAuditAlertStats 获取告警统计
// args: []
func GetAuditAlertStats(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.AuditAlertKey, []string{})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询失败: %v", err))
	}

	var todayCount, unresolved, highLevelCount int
	today := time.Now().Format("2006-01-02")

	for _, v := range results {
		var alert model.AuditAlert
		if err := json.Unmarshal(v, &alert); err != nil {
			continue
		}
		// 今日统计
		if strings.HasPrefix(alert.TriggerTime, today) {
			todayCount++
		}
		// 未处理统计
		if alert.Status != "RESOLVED" {
			unresolved++
		}
		// 高危统计
		if alert.Level == "HIGH" {
			highLevelCount++
		}
	}

	stats := map[string]int{
		"today_count":      todayCount,
		"unresolved":       unresolved,
		"high_level_count": highLevelCount,
	}

	payload, _ := json.Marshal(stats)
	return shim.Success(payload)
}

// ---------------------- 导出任务相关 ----------------------

// CreateAuditExportTask 创建导出任务
// args: [AuditExportTask JSON]
func CreateAuditExportTask(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("参数个数不满足")
	}

	var task model.AuditExportTask
	if err := json.Unmarshal([]byte(args[0]), &task); err != nil {
		return shim.Error(fmt.Sprintf("任务解析失败: %v", err))
	}

	if task.ID == "" {
		task.ID = fmt.Sprintf("EXP%d%s", time.Now().UnixNano()%1000000, stub.GetTxID()[:8])
	}
	task.Status = "PENDING"
	task.CreateTime = time.Now().Format(auditTimeLayout)

	if err := utils.WriteLedger(task, stub, model.AuditExportTaskKey, []string{task.ID}); err != nil {
		return shim.Error(fmt.Sprintf("创建任务失败: %v", err))
	}

	// 写入状态索引
	if err := utils.WriteLedger(task.ID, stub, model.AuditExportTaskStatusIndexKey,
		[]string{task.Status, task.CreateTime, task.ID}); err != nil {
		return shim.Error(fmt.Sprintf("写入索引失败: %v", err))
	}

	payload, _ := json.Marshal(task)
	return shim.Success(payload)
}

// GetAuditExportTasksByStatus 按状态查询导出任务
// args: [status, page, size]
func GetAuditExportTasksByStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 3 {
		return shim.Error("参数个数不满足")
	}

	status := args[0]
	page, _ := strconv.Atoi(args[1])
	size, _ := strconv.Atoi(args[2])
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 10
	}

	var results [][]byte
	var err error

	if status == "" {
		results, err = utils.GetStateByPartialCompositeKeys2(stub, model.AuditExportTaskKey, []string{})
	} else {
		results, err = utils.GetStateByPartialCompositeKeys2(stub, model.AuditExportTaskStatusIndexKey, []string{status})
	}

	if err != nil {
		return shim.Error(fmt.Sprintf("查询失败: %v", err))
	}

	// 分页
	start := (page - 1) * size
	end := start + size
	if start >= len(results) {
		var emptyTasks []model.AuditExportTask
		payload, _ := json.Marshal(emptyTasks)
		return shim.Success(payload)
	}
	if end > len(results) {
		end = len(results)
	}

	var tasks []model.AuditExportTask
	for _, v := range results[start:end] {
		var task model.AuditExportTask
		if err := json.Unmarshal(v, &task); err != nil {
			continue
		}
		tasks = append(tasks, task)
	}

	payload, _ := json.Marshal(tasks)
	return shim.Success(payload)
}

// GetAuditExportTaskByID 根据ID查询导出任务
// args: [taskID]
func GetAuditExportTaskByID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("参数个数不满足")
	}

	taskID := args[0]
	results, err := utils.GetStateByPartialCompositeKeys(stub, model.AuditExportTaskKey, []string{taskID})
	if err != nil {
		return shim.Error(fmt.Sprintf("查询失败: %v", err))
	}
	if len(results) == 0 {
		return shim.Error("导出任务不存在")
	}

	var task model.AuditExportTask
	if err := json.Unmarshal(results[0], &task); err != nil {
		return shim.Error(fmt.Sprintf("解析失败: %v", err))
	}

	payload, _ := json.Marshal(task)
	return shim.Success(payload)
}

// UpdateAuditExportTaskStatus 更新导出任务状态
// args: [taskID, status, fileName, fileHash, failReason]
func UpdateAuditExportTaskStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("参数个数不满足")
	}

	taskID := args[0]
	newStatus := args[1]
	fileName := args[2]
	fileHash := args[3]
	failReason := args[4]

	// 获取现有任务
	results, err := utils.GetStateByPartialCompositeKeys(stub, model.AuditExportTaskKey, []string{taskID})
	if err != nil || len(results) == 0 {
		return shim.Error("任务不存在")
	}

	var task model.AuditExportTask
	if err := json.Unmarshal(results[0], &task); err != nil {
		return shim.Error(fmt.Sprintf("任务解析失败: %v", err))
	}

	// 状态机校验
	validTransitions := map[string][]string{
		"PENDING": {"RUNNING", "FAIL"},
		"RUNNING": {"SUCCESS", "FAIL"},
		"SUCCESS": {},
		"FAIL":    {},
	}

	allowed := false
	for _, s := range validTransitions[task.Status] {
		if s == newStatus {
			allowed = true
			break
		}
	}
	if !allowed {
		return shim.Error(fmt.Sprintf("状态转换不允许: %s -> %s", task.Status, newStatus))
	}

	// 删除旧索引
	_ = utils.DelLedger(stub, model.AuditExportTaskStatusIndexKey,
		[]string{task.Status, task.CreateTime, task.ID})

	// 更新字段
	task.Status = newStatus
	if fileName != "" {
		task.FileName = fileName
	}
	if fileHash != "" {
		task.FileHash = fileHash
	}
	if failReason != "" {
		task.FailReason = failReason
	}
	if newStatus == "SUCCESS" || newStatus == "FAIL" {
		task.FinishTime = time.Now().Format(auditTimeLayout)
	}

	// 写回
	if err := utils.WriteLedger(task, stub, model.AuditExportTaskKey, []string{task.ID}); err != nil {
		return shim.Error(fmt.Sprintf("更新任务失败: %v", err))
	}

	// 写入新索引
	if err := utils.WriteLedger(task.ID, stub, model.AuditExportTaskStatusIndexKey,
		[]string{task.Status, task.CreateTime, task.ID}); err != nil {
		return shim.Error(fmt.Sprintf("写入索引失败: %v", err))
	}

	payload, _ := json.Marshal(task)
	return shim.Success(payload)
}
