<template>
  <div class="login-container">
    <div class="login-mask" />

    <el-form ref="loginForm" class="login-form" auto-complete="on" label-position="left">
      <div class="header">
        <div class="logo-wrap">H+</div>
        <div>
          <h3 class="title">基于区块链的医院患者管理系统</h3>
          <p class="sub-title">登录</p>
        </div>
      </div>

      <div class="role-tabs">
        <div
          v-for="role in roleDefs"
          :key="role.key"
          :class="['role-tab', { active: activeRole === role.key }]"
          @click="switchRole(role.key)"
        >
          {{ role.label }}
        </div>
      </div>

      <el-select
        v-model="value"
        placeholder="请选择用户账号"
        class="form-item"
        filterable
        clearable
      >
        <el-option
          v-for="item in filteredAccounts"
          :key="item.account_id"
          :label="item.account_name"
          :value="item.account_id"
        >
          <span style="float: left">{{ item.account_name }}</span>
          <span style="float: right; color: #8492a6; font-size: 13px">{{ item.account_id }}</span>
        </el-option>
      </el-select>

      <el-input v-model="form.password" placeholder="密码" show-password class="form-item" />

      <el-select
        v-if="showAffiliation"
        v-model="form.affiliation"
        :placeholder="activeRole === 'doctor' ? '请选择所属医院' : '请选择所属药店'"
        class="form-item"
      >
        <el-option
          v-for="item in currentAffiliationOptions"
          :key="item"
          :label="item"
          :value="item"
        />
      </el-select>

      <el-button :loading="loading" type="primary" class="login-btn" @click.native.prevent="handleLogin">登录</el-button>

      <div class="tips">请选择不同的用户角色</div>
    </el-form>
  </div>
</template>

<script>
import { queryAccountList } from '@/api/accountV2'

export default {
  name: 'Login',
  data() {
    return {
      loading: false,
      redirect: undefined,
      accountList: [],
      value: '',
      form: {
        password: '',
        affiliation: ''
      },
      activeRole: 'admin',
      roleDefs: [
        { key: 'admin', label: '管理员', roleValue: 'admin', keywords: ['管理员'] },
        { key: 'doctor', label: '医生', roleValue: 'doctor', keywords: ['医生'] },
        { key: 'patient', label: '患者', roleValue: 'patient', keywords: ['病人', '患者'] }
      ],
      hospitals: ['北京协和医院', '华西医院', '上海交通大学医学院附属瑞金医院', '广东省人民医院'],
      pharmacies: ['国大药房', '益丰大药房', '老百姓大药房', '大参林药店']
    }
  },
  computed: {
    normalizedAccounts() {
      return this.accountList.map(item => {
        if (item.role) return item
        const name = item.account_name || ''
        let role = 'patient'
        if (/管理员/.test(name)) role = 'admin'
        else if (/医生/.test(name)) role = 'doctor'
        else if (/药店/.test(name)) role = 'drugstore'
        else if (/保险/.test(name)) role = 'insurance'
        return { ...item, role }
      })
    },
    filteredAccounts() {
      const currentRole = this.roleDefs.find(item => item.key === this.activeRole)
      if (!currentRole) return this.normalizedAccounts
      return this.normalizedAccounts.filter(item => item.role === currentRole.roleValue)
    },
    showAffiliation() {
      return this.activeRole === 'doctor' || this.activeRole === 'drugstore'
    },
    currentAffiliationOptions() {
      return this.activeRole === 'doctor' ? this.hospitals : this.pharmacies
    }
  },
  watch: {
    $route: {
      handler: function(route) {
        this.redirect = route.query && route.query.redirect
      },
      immediate: true
    }
  },
  created() {
    this.loadAccounts()
  },
  methods: {
    loadAccounts() {
      queryAccountList().then(response => {
        if (response !== null) {
          this.accountList = response
        }
      })
    },
    switchRole(roleKey) {
      this.activeRole = roleKey
      this.value = ''
      this.form.affiliation = ''
      this.loadAccounts()
    },
    handleLogin() {
      if (this.value) {
        this.loading = true
        this.$store.dispatch('account/login', this.value).then(() => {
          this.$router.push({ path: '/' })
          this.loading = false
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.$message('请选择用户账号')
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.login-container {
  min-height: 100%;
  width: 100%;
  background: linear-gradient(135deg, rgba(224, 240, 255, 0.95), rgba(224, 255, 243, 0.95)), url('https://images.unsplash.com/photo-1586773860418-d37222d8fce3?auto=format&fit=crop&w=1920&q=80') center/cover no-repeat;
  overflow: hidden;
  position: relative;

  .login-mask {
    position: absolute;
    inset: 0;
    background: rgba(245, 249, 255, 0.5);
    backdrop-filter: blur(2px);
  }

  .login-form {
    position: relative;
    z-index: 1;
    width: 500px;
    max-width: 92%;
    margin: 90px auto 0;
    padding: 0 0 24px;
    background: #fff;
    border-radius: 10px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
    overflow: hidden;
  }

  .header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 20px;
    background: linear-gradient(90deg, #eef4ff, #def8ee);

    .logo-wrap {
      width: 42px;
      height: 42px;
      border-radius: 8px;
      background: #2e7be7;
      color: #fff;
      font-weight: 700;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .title {
      margin: 0;
      font-size: 28px;
      color: #1f2d3d;
      font-weight: 700;
    }

    .sub-title {
      margin: 4px 0 0;
      color: #5c6b77;
      font-size: 20px;
    }
  }

  .role-tabs {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 6px;
    padding: 16px 20px 8px;

    .role-tab {
      text-align: center;
      padding: 8px 4px;
      border: 1px solid #dcdfe6;
      border-radius: 4px;
      color: #333;
      cursor: pointer;
      font-size: 14px;

      &.active {
        background: #1f4f8b;
        color: #fff;
        border-color: #1f4f8b;
      }
    }
  }

  .form-item {
    display: block;
    width: calc(100% - 40px);
    margin: 12px 20px;
  }

  .login-btn {
    width: calc(100% - 40px);
    margin: 18px 20px 14px;
    background: linear-gradient(90deg, #1a74df, #1f92ff);
    border: none;
  }

  .tips {
    text-align: center;
    color: #74808f;
    font-size: 13px;
  }
}
</style>
