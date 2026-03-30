import request from '@/utils/request'

// ---------------------- 审计检索 ----------------------

// 获取审计事件列表
export function getAuditEvents(params) {
    return request({
        url: '/audit/events',
        method: 'get',
        params
    })
}

// 获取审计事件详情
export function getAuditEventDetail(eventId) {
    return request({
        url: `/audit/events/${eventId}`,
        method: 'get'
    })
}

// 获取审计事件统计
export function getAuditEventStats() {
    return request({
        url: '/audit/events/stats',
        method: 'get'
    })
}

// 手工创建审计事件（测试用）
export function createAuditEventManual(data) {
    return request({
        url: '/audit/events/manual',
        method: 'post',
        data
    })
}

// 获取采集模块健康状态
export function getAuditCollectorHealth() {
    return request({
        url: '/audit/collector/health',
        method: 'get'
    })
}

// ---------------------- 告警模块 ----------------------

// 获取告警列表
export function getAuditAlerts(params) {
    return request({
        url: '/audit/alerts',
        method: 'get',
        params
    })
}

// 获取告警详情
export function getAuditAlertDetail(alertId) {
    return request({
        url: `/audit/alerts/${alertId}`,
        method: 'get'
    })
}

// 获取告警统计
export function getAuditAlertStats() {
    return request({
        url: '/audit/alerts/stats',
        method: 'get'
    })
}

// 确认告警
export function ackAuditAlert(alertId) {
    return request({
        url: `/audit/alerts/${alertId}/ack`,
        method: 'post'
    })
}

// 关闭告警
export function resolveAuditAlert(alertId, data) {
    return request({
        url: `/audit/alerts/${alertId}/resolve`,
        method: 'post',
        data
    })
}

// ---------------------- 导出模块 ----------------------

// 创建导出任务
export function createAuditExport(data) {
    return request({
        url: '/audit/reports/export',
        method: 'post',
        data
    })
}

// 获取导出任务列表
export function getAuditExportTasks(params) {
    return request({
        url: '/audit/reports/tasks',
        method: 'get',
        params
    })
}

// 获取导出任务详情
export function getAuditExportTaskDetail(taskId) {
    return request({
        url: `/audit/reports/tasks/${taskId}`,
        method: 'get'
    })
}
