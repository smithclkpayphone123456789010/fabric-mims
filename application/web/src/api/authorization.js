import request from '@/utils/request'

export function grantRecordAuthorization(data) {
  return request({
    url: '/grantRecordAuthorization',
    method: 'post',
    data
  })
}

export function renewRecordAuthorization(data) {
  return request({
    url: '/renewRecordAuthorization',
    method: 'post',
    data
  })
}

export function revokeRecordAuthorization(data) {
  return request({
    url: '/revokeRecordAuthorization',
    method: 'post',
    data
  })
}

export function queryMyAuthorizations(data) {
  return request({
    url: '/queryMyAuthorizations',
    method: 'post',
    data
  })
}

export function queryAccessibleRecordsByDoctor(data) {
  return request({
    url: '/queryAccessibleRecordsByDoctor',
    method: 'post',
    data
  })
}

export function checkRecordAccess(data) {
  return request({
    url: '/checkRecordAccess',
    method: 'post',
    data
  })
}






