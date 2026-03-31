import request from '@/utils/request'

export const aiChat = (data) => request({ url: '/ai/chat', method: 'post', data, timeout: 30000 })
export const aiTriage = (data) => request({ url: '/ai/triage', method: 'post', data, timeout: 20000 })
export const aiSessions = () => request({ url: '/ai/sessions', method: 'get', timeout: 15000 })
export const aiSessionMessages = (id) => request({ url: `/ai/session/${id}/messages`, method: 'get', timeout: 15000 })

export const aiRehabCompanion = (data) => request({ url: '/ai/rehab-companion', method: 'post', data, timeout: 30000 })
export const aiReportTranslator = (data) => request({ url: '/ai/report-translator', method: 'post', data, timeout: 30000 })

// AI诊室相关API（增加超时时间到60秒，因为报告生成需要更长时间）
export const aiClinicStart = (data) => request({ url: '/ai/clinic/start', method: 'post', data, timeout: 60000 })
export const aiClinicChat = (data) => request({ url: '/ai/clinic/chat', method: 'post', data, timeout: 60000 })
export const aiClinicReset = (data) => request({ url: '/ai/clinic/reset', method: 'post', data, timeout: 15000 })
export const aiClinicSession = (id) => request({ url: `/ai/clinic/session/${id}`, method: 'get', timeout: 15000 })
export const aiClinicReport = (id) => request({ url: `/ai/clinic/report/${id}`, method: 'get', timeout: 15000 })
