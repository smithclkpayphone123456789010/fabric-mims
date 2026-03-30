package routers

import (
	v2 "application/api/v2"
	"application/middleware"
	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由信息
func InitRouter() *gin.Engine {
	r := gin.Default()

	// 添加审计中间件
	r.Use(middleware.AuditMiddleware())

	apiV2 := r.Group("/api/v2")
	{
		apiV2.GET("/hello", v2.Hello)
		apiV2.POST("/createAccountV2", v2.CreateAccountV2)
		apiV2.POST("/queryAccountV2List", v2.QueryAccountV2List)
		apiV2.POST("/createPrescription", v2.CreatePrescription)
		apiV2.POST("/queryPrescription", v2.QueryPrescriptionList)
		apiV2.GET("/previewPrescriptionFile", v2.PreviewPrescriptionFile)
		apiV2.POST("/createInsuranceCover", v2.CreateInsuranceCover)
		apiV2.POST("/queryInsuranceCoverList", v2.QueryInsuranceCoverList)
		apiV2.POST("/updateInsuranceCover", v2.UpdateInsuranceCover)
		apiV2.POST("/deleteInsuranceCover", v2.DeleteInsuranceCover)
		apiV2.POST("/createDrugOrder", v2.CreateDrugOrder)
		apiV2.POST("/queryDrugOrderList", v2.QueryDrugOrderList)
		apiV2.POST("/grantRecordAuthorization", v2.GrantRecordAuthorization)
		apiV2.POST("/renewRecordAuthorization", v2.RenewRecordAuthorization)
		apiV2.POST("/revokeRecordAuthorization", v2.RevokeRecordAuthorization)
		apiV2.POST("/queryMyAuthorizations", v2.QueryMyAuthorizations)
		apiV2.POST("/queryAccessibleRecordsByDoctor", v2.QueryAccessibleRecordsByDoctor)
		apiV2.POST("/checkRecordAccess", v2.CheckRecordAccess)

		apiV2.POST("/outpatient/registration/create", v2.CreateOutpatientRegistration)
		apiV2.POST("/outpatient/registration/cancel", v2.CancelOutpatientRegistration)
		apiV2.GET("/outpatient/registration/list", v2.QueryOutpatientRegistration)

		apiV2.POST("/outpatient/slot/create", v2.CreateScheduleSlot)
		apiV2.GET("/outpatient/slot/list", v2.QueryScheduleSlot)

		apiV2.GET("/outpatient/payment/list", v2.QueryOutpatientPayment)
		apiV2.POST("/outpatient/payment/pay", v2.PayOutpatientOrder)

		apiV2.GET("/outpatient/queue/current", v2.QueryOutpatientQueue)
		apiV2.POST("/outpatient/queue/start", v2.StartOutpatientVisit)
		apiV2.POST("/outpatient/queue/finish", v2.FinishOutpatientVisit)

		apiV2.GET("/outpatient/record/list", v2.QueryOutpatientRecord)

		apiV2.POST("/ai/chat", v2.AIChat)
		apiV2.POST("/ai/triage", v2.AITriage)
		apiV2.POST("/ai/rehab-companion", v2.AIRehabCompanion)
		apiV2.POST("/ai/report-translator", v2.AIReportTranslator)
		apiV2.GET("/ai/sessions", v2.AIGetSessions)
		apiV2.GET("/ai/session/:id/messages", v2.AIGetSessionMessages)

		// ---------------------- 审计监控模块 ----------------------
		// 日志采集
		apiV2.POST("/audit/events/manual", v2.CreateAuditEventManual)
		apiV2.GET("/audit/collector/health", v2.GetAuditCollectorHealth)

		// 审计检索
		apiV2.GET("/audit/events", v2.GetAuditEvents)
		apiV2.GET("/audit/events/stats", v2.GetAuditEventStats)
		apiV2.GET("/audit/events/:id", v2.GetAuditEventDetail)

		// 告警模块
		apiV2.GET("/audit/alerts", v2.GetAuditAlerts)
		apiV2.GET("/audit/alerts/stats", v2.GetAuditAlertStats)
		apiV2.GET("/audit/alerts/:id", v2.GetAuditAlertDetail)
		apiV2.POST("/audit/alerts/:id/ack", v2.AckAuditAlert)
		apiV2.POST("/audit/alerts/:id/resolve", v2.ResolveAuditAlert)

		// 导出模块
		apiV2.POST("/audit/reports/export", v2.CreateAuditExport)
		apiV2.GET("/audit/reports/tasks", v2.GetAuditExportTasks)
		apiV2.GET("/audit/reports/tasks/:id", v2.GetAuditExportTaskDetail)
	}
	return r
}
