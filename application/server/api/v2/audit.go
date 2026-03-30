package v2

import (
	bc "application/blockchain"
	"application/model"
	"application/pkg/app"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ---------------------- 日志采集模块 ----------------------

// CreateAuditEventManual 手工写入测试审计事件
// POST /api/v2/audit/events/manual
func CreateAuditEventManual(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.AuditEventRequestBody)

	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错: %s", err.Error()))
		return
	}

	// 构建审计事件（使用服务器生成的ID，确保所有peer背书一致）
	eventID := fmt.Sprintf("EVT%d%06d", time.Now().UnixNano()/1e6, rand.Intn(999999))
	eventTime := time.Now().Format("2006-01-02 15:04:05")

	event := map[string]interface{}{
		"id":                eventID,
		"event_type":        body.EventType,
		"event_level":       body.EventLevel,
		"event_time":        eventTime,
		"action_result":     body.ActionResult,
		"actor_id":          "0feceb66ffc1", // admin账号
		"target_patient_id": body.TargetPatientID,
		"target_record_id":  body.TargetRecordID,
		"detail_json":       body.DetailJSON,
		"fail_reason":       body.FailReason,
	}

	eventBytes, _ := json.Marshal(event)

	// 调用链码
	resp, err := bc.ChannelExecute("createAuditEvent", [][]byte{eventBytes})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	var result map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &result); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"event_id": result["id"],
		"tx_id":    result["tx_id"],
	})
}

// GetAuditCollectorHealth 获取采集模块健康状态
// GET /api/v2/audit/collector/health
func GetAuditCollectorHealth(c *gin.Context) {
	appG := app.Gin{C: c}

	// 获取最新一条审计事件作为最后事件ID
	_, err := bc.ChannelQuery("getAuditEventByID", [][]byte{[]byte("last")})
	if err != nil {
		// 如果查询失败，返回降级状态
		appG.Response(http.StatusOK, "成功", map[string]interface{}{
			"status":        "degraded",
			"last_event_id": "",
			"hash_chain_ok": false,
			"pending_count": 0,
		})
		return
	}

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"status":        "healthy",
		"last_event_id": "",
		"hash_chain_ok": true,
		"pending_count": 0,
	})
}

// ---------------------- 审计检索模块 ----------------------

// GetAuditEvents 获取审计事件列表
// GET /api/v2/audit/events
func GetAuditEvents(c *gin.Context) {
	appG := app.Gin{C: c}

	// 解析查询参数
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	eventType := c.Query("event_type")
	_ = c.Query("event_level")   // 保留字段，后续扩展用
	_ = c.Query("action_result") // 保留字段，后续扩展用
	targetPatientID := c.Query("target_patient_id")
	targetRecordID := c.Query("target_record_id")
	_ = c.Query("tx_id") // 保留字段，后续扩展用
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	// 参数校验
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	// 确定查询类型
	keyType := "all"
	key1 := ""
	key2 := ""
	key3 := ""

	if startTime != "" && endTime != "" {
		keyType = "time"
		key1 = startTime
		key2 = endTime
	} else if eventType != "" {
		keyType = "type"
		key1 = eventType
	} else if targetPatientID != "" {
		keyType = "patient"
		key1 = targetPatientID
	} else if targetRecordID != "" {
		keyType = "record"
		key1 = targetRecordID
	}

	// 调用链码
	args := [][]byte{
		[]byte(keyType),
		[]byte(key1),
		[]byte(key2),
		[]byte(key3),
		[]byte(strconv.Itoa(page)),
		[]byte(strconv.Itoa(size)),
	}

	resp, err := bc.ChannelQuery("getAuditEventsByCompositeKey", args)
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	var events []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &events); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	// 转换为前端需要的格式
	var list []model.AuditEventItem
	for _, e := range events {
		item := model.AuditEventItem{
			ID:              getString(e, "id"),
			EventType:       getString(e, "event_type"),
			EventLevel:      getString(e, "event_level"),
			EventTime:       getString(e, "event_time"),
			ActorID:         getString(e, "actor_id"),
			TargetPatientID: getString(e, "target_patient_id"),
			TargetRecordID:  getString(e, "target_record_id"),
			ActionResult:    getString(e, "action_result"),
			TxID:            getString(e, "tx_id"),
			RequestPath:     getString(e, "request_path"),
			Message:         getString(e, "detail_json"),
		}
		list = append(list, item)
	}

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"list":  list,
		"total": len(list),
		"page":  page,
		"size":  size,
	})
}

// GetAuditEventDetail 获取审计事件详情
// GET /api/v2/audit/events/:id
func GetAuditEventDetail(c *gin.Context) {
	appG := app.Gin{C: c}
	eventID := c.Param("id")

	if eventID == "" {
		appG.Response(http.StatusBadRequest, "失败", "事件ID不能为空")
		return
	}

	resp, err := bc.ChannelQuery("getAuditEventByID", [][]byte{[]byte(eventID)})
	if err != nil {
		appG.Response(http.StatusNotFound, "失败", "审计事件不存在")
		return
	}

	var result model.AuditEventDetailResponse
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &result); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	appG.Response(http.StatusOK, "成功", result)
}

// GetAuditEventStats 获取审计事件统计
// GET /api/v2/audit/events/stats
func GetAuditEventStats(c *gin.Context) {
	appG := app.Gin{C: c}

	resp, err := bc.ChannelQuery("getAuditEventStats", [][]byte{})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	var result model.AuditEventStatsResponse
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &result); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	appG.Response(http.StatusOK, "成功", result)
}

// ---------------------- 告警模块 ----------------------

// GetAuditAlerts 获取告警列表
// GET /api/v2/audit/alerts
func GetAuditAlerts(c *gin.Context) {
	appG := app.Gin{C: c}

	level := c.Query("level")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	keyType := "all"
	key1 := ""

	if status != "" {
		keyType = "status"
		key1 = status
	} else if level != "" {
		keyType = "level"
		key1 = level
	}

	args := [][]byte{
		[]byte(keyType),
		[]byte(key1),
		[]byte(""),
		[]byte(strconv.Itoa(page)),
		[]byte(strconv.Itoa(size)),
	}

	resp, err := bc.ChannelQuery("getAuditAlertsByCompositeKey", args)
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	var alerts []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &alerts); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	var list []model.AuditAlertItem
	for _, a := range alerts {
		item := model.AuditAlertItem{
			ID:              getString(a, "id"),
			RuleCode:        getString(a, "rule_code"),
			Level:           getString(a, "level"),
			Status:          getString(a, "status"),
			TriggerTime:     getString(a, "trigger_time"),
			ActorID:         getString(a, "actor_id"),
			TargetPatientID: getString(a, "target_patient_id"),
			TargetRecordID:  getString(a, "target_record_id"),
			Description:     getString(a, "description"),
		}
		list = append(list, item)
	}

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"list":  list,
		"total": len(list),
	})
}

// GetAuditAlertDetail 获取告警详情
// GET /api/v2/audit/alerts/:id
func GetAuditAlertDetail(c *gin.Context) {
	appG := app.Gin{C: c}
	alertID := c.Param("id")

	if alertID == "" {
		appG.Response(http.StatusBadRequest, "失败", "告警ID不能为空")
		return
	}

	resp, err := bc.ChannelQuery("getAuditAlertByID", [][]byte{[]byte(alertID)})
	if err != nil {
		appG.Response(http.StatusNotFound, "失败", "告警不存在")
		return
	}

	var result model.AuditAlertDetailResponse
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &result); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	appG.Response(http.StatusOK, "成功", result)
}

// AckAuditAlert 确认告警
// POST /api/v2/audit/alerts/:id/ack
func AckAuditAlert(c *gin.Context) {
	appG := app.Gin{C: c}
	alertID := c.Param("id")

	if alertID == "" {
		appG.Response(http.StatusBadRequest, "失败", "告警ID不能为空")
		return
	}

	_, err := bc.ChannelExecute("ackAuditAlert", [][]byte{[]byte(alertID)})
	if err != nil {
		if strings.Contains(err.Error(), "当前状态不允许确认") {
			appG.Response(http.StatusBadRequest, "失败", err.Error())
		} else {
			appG.Response(http.StatusInternalServerError, "失败", err.Error())
		}
		return
	}

	appG.Response(http.StatusOK, "成功", map[string]bool{"success": true})
}

// ResolveAuditAlert 关闭告警
// POST /api/v2/audit/alerts/:id/resolve
func ResolveAuditAlert(c *gin.Context) {
	appG := app.Gin{C: c}
	alertID := c.Param("id")
	body := new(model.AuditAlertResolveRequestBody)

	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错: %s", err.Error()))
		return
	}

	if body.HandleNote == "" {
		appG.Response(http.StatusBadRequest, "失败", "处理备注不能为空")
		return
	}

	_, err := bc.ChannelExecute("resolveAuditAlert", [][]byte{[]byte(alertID), []byte(body.HandleNote)})
	if err != nil {
		if strings.Contains(err.Error(), "已关闭") {
			appG.Response(http.StatusBadRequest, "失败", err.Error())
		} else {
			appG.Response(http.StatusInternalServerError, "失败", err.Error())
		}
		return
	}

	appG.Response(http.StatusOK, "成功", map[string]bool{"success": true})
}

// GetAuditAlertStats 获取告警统计
// GET /api/v2/audit/alerts/stats
func GetAuditAlertStats(c *gin.Context) {
	appG := app.Gin{C: c}

	resp, err := bc.ChannelQuery("getAuditAlertStats", [][]byte{})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	var result model.AuditAlertStatsResponse
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &result); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	appG.Response(http.StatusOK, "成功", result)
}

// ---------------------- 导出模块 ----------------------

// CreateAuditExport 创建导出任务
// POST /api/v2/audit/reports/export
func CreateAuditExport(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.AuditExportCreateRequestBody)

	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错: %s", err.Error()))
		return
	}

	// 验证格式
	if body.Format != "csv" && body.Format != "xlsx" {
		appG.Response(http.StatusBadRequest, "失败", "导出格式仅支持 csv 或 xlsx")
		return
	}

	// 构建导出任务
	task := map[string]interface{}{
		"creator_id":  "0feceb66ffc1", // admin账号
		"format":      body.Format,
		"filter_json": body.FilterJSON,
	}

	taskBytes, _ := json.Marshal(task)

	resp, err := bc.ChannelExecute("createAuditExportTask", [][]byte{taskBytes})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	var result map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &result); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	appG.Response(http.StatusOK, "成功", map[string]string{
		"task_id": result["id"].(string),
	})
}

// GetAuditExportTasks 获取导出任务列表
// GET /api/v2/audit/reports/tasks
func GetAuditExportTasks(c *gin.Context) {
	appG := app.Gin{C: c}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 10
	}

	resp, err := bc.ChannelQuery("getAuditExportTasksByStatus", [][]byte{
		[]byte(status),
		[]byte(strconv.Itoa(page)),
		[]byte(strconv.Itoa(size)),
	})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	var tasks []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &tasks); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	var list []model.AuditExportTaskItem
	for _, t := range tasks {
		item := model.AuditExportTaskItem{
			ID:         getString(t, "id"),
			CreatorID:  getString(t, "creator_id"),
			CreateTime: getString(t, "create_time"),
			Status:     getString(t, "status"),
			Format:     getString(t, "format"),
			FileName:   getString(t, "file_name"),
			FinishTime: getString(t, "finish_time"),
			FailReason: getString(t, "fail_reason"),
		}
		list = append(list, item)
	}

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"list":  list,
		"total": len(list),
	})
}

// GetAuditExportTaskDetail 获取导出任务详情
// GET /api/v2/audit/reports/tasks/:id
func GetAuditExportTaskDetail(c *gin.Context) {
	appG := app.Gin{C: c}
	taskID := c.Param("id")

	if taskID == "" {
		appG.Response(http.StatusBadRequest, "失败", "任务ID不能为空")
		return
	}

	resp, err := bc.ChannelQuery("getAuditExportTaskByID", [][]byte{[]byte(taskID)})
	if err != nil {
		appG.Response(http.StatusNotFound, "失败", "导出任务不存在")
		return
	}

	var result model.AuditExportTaskDetailResponse
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &result); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	appG.Response(http.StatusOK, "成功", result)
}

// ---------------------- 辅助函数 ----------------------

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
