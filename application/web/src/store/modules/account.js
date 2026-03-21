import {
  login
} from '@/api/accountV2'
import {
  getToken,
  setToken,
  removeToken
} from '@/utils/auth'
import {
  resetRouter
} from '@/router'

const getDefaultState = () => {
  return {
    token: getToken(),
    account_id: '',
    account_name: '',
    roles: []
  }
}

const state = getDefaultState()

const mutations = {
  RESET_STATE: (state) => {
    Object.assign(state, getDefaultState())
  },
  SET_TOKEN: (state, token) => {
    state.token = token
  },
  SET_ACCOUNTID: (state, account_id) => {
    state.account_id = account_id
  },
  SET_USERNAME: (state, account_name) => {
    state.account_name = account_name
  },
  SET_ROLES: (state, roles) => {
    state.roles = roles
  }
}

const actions = {
  login({
    commit
  }, account_id) {
    return new Promise((resolve, reject) => {
      login({
        args: [{
          account_id: account_id
        }]
      }).then(response => {
        commit('SET_TOKEN', response[0].account_id)
        setToken(response[0].account_id)
        resolve()
      }).catch(error => {
        reject(error)
      })
    })
  },
  // get user info
  getInfo({
    commit,
    state
  }) {
    return new Promise((resolve, reject) => {
      login({
        args: [{
          account_id: state.token
        }]
      }).then(response => {
        const account = response[0] || {}
        let role = account.role

        // 兼容旧数据（没有 role 字段时按名称推断）
        if (!role) {
          if (/管理员/.test(account.account_name || '')) {
            role = 'admin'
          } else if (/医生/.test(account.account_name || '')) {
            role = 'doctor'
          } else if (/病人|患者/.test(account.account_name || '')) {
            role = 'patient'
          } else if (/药店/.test(account.account_name || '')) {
            role = 'drugstore'
          } else if (/保险/.test(account.account_name || '')) {
            role = 'insurance'
          }
        }

        const roles = role ? [role] : ['patient']
        commit('SET_ROLES', roles)
        commit('SET_ACCOUNTID', account.account_id || '')
        commit('SET_USERNAME', account.account_name || '')
        resolve(roles)
      }).catch(error => {
        reject(error)
      })
    })
  },
  logout({
    commit
  }) {
    return new Promise(resolve => {
      removeToken()
      resetRouter()
      commit('RESET_STATE')
      resolve()
    })
  },

  resetToken({
    commit
  }) {
    return new Promise(resolve => {
      removeToken()
      commit('RESET_STATE')
      resolve()
    })
  }
}

export default {
  namespaced: true,
  state,
  mutations,
  actions
}
