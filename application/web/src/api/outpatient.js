import request from '@/utils/request'

export const createRegistration = (data) => request({ url: '/outpatient/registration/create', method: 'post', data })
export const cancelRegistration = (data) => request({ url: '/outpatient/registration/cancel', method: 'post', data })
export const queryRegistrationList = (params) => request({ url: '/outpatient/registration/list', method: 'get', params })

export const createSlot = (data) => request({ url: '/outpatient/slot/create', method: 'post', data })
export const querySlotList = (params) => request({ url: '/outpatient/slot/list', method: 'get', params })

export const queryPaymentList = (params) => request({ url: '/outpatient/payment/list', method: 'get', params })
export const payOrder = (data) => request({ url: '/outpatient/payment/pay', method: 'post', data })

export const queryQueueCurrent = (params) => request({ url: '/outpatient/queue/current', method: 'get', params })
export const startVisit = (data) => request({ url: '/outpatient/queue/start', method: 'post', data })
export const finishVisit = (data) => request({ url: '/outpatient/queue/finish', method: 'post', data })

export const queryOutpatientRecordList = (params) => request({ url: '/outpatient/record/list', method: 'get', params })
