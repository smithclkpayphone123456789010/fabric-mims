<template>
  <div class="container">
    <el-alert type="success" :closable="false" class="mb16">
      <p>账户ID: {{ account_id }}</p>
      <p>用户名: {{ account_name }}</p>
    </el-alert>

    <el-card class="mb16" shadow="never">
      <div class="toolbar">
        <el-input
          v-model="keyword"
          clearable
          placeholder="按姓名/名称/手机号模糊搜索"
          class="search-input"
        />
        <el-select v-model="roleFilter" placeholder="按角色筛选" class="role-filter">
          <el-option label="全部角色" value="all" />
          <el-option label="医生" value="doctor" />
          <el-option label="患者" value="patient" />
          <el-option label="药店" value="drugstore" />
          <el-option label="保险机构" value="insurance" />
          <el-option label="管理员" value="admin" />
        </el-select>
      </div>
    </el-card>

    <el-tabs v-model="activeRole">
      <el-tab-pane
        v-for="item in roleTabs"
        :key="item.value"
        :label="`${item.label}（${countByRole(item.value)}）`"
        :name="item.value"
      />
    </el-tabs>

    <div v-if="currentList.length === 0" style="text-align: center;">
      <el-alert title="查询不到数据" type="warning" :closable="false" />
    </div>

    <el-row v-loading="loading" :gutter="20">
      <el-col v-for="(val, index) in currentList" :key="index" :span="8">
        <el-card class="account-card">
          <div slot="header" class="clearfix">
            账户ID:
            <span class="id-text">{{ val.account_id }}</span>
          </div>

          <div class="item"><el-tag size="mini">姓名/名称</el-tag><span class="value">{{ val.account_name }}</span></div>
          <div class="item"><el-tag size="mini" type="success">角色</el-tag><span class="value">{{ roleLabel(val.role) }}</span></div>

          <template v-if="val.role === 'doctor'">
            <div class="item"><el-tag size="mini" type="warning">医院</el-tag><span class="value">{{ val.hospital_name || '-' }}</span></div>
            <div class="item"><el-tag size="mini" type="warning">科室</el-tag><span class="value">{{ val.department || '-' }}</span></div>
          </template>

          <template v-else-if="val.role === 'patient'">
            <div class="item"><el-tag size="mini" type="info">性别</el-tag><span class="value">{{ val.gender || '-' }}</span></div>
            <div class="item"><el-tag size="mini" type="info">联系方式</el-tag><span class="value">{{ val.phone || '-' }}</span></div>
          </template>

          <template v-else-if="val.role === 'drugstore'">
            <div class="item"><el-tag size="mini" type="danger">所属医院</el-tag><span class="value">{{ val.hospital_name || '-' }}</span></div>
          </template>

          <template v-else-if="val.role === 'insurance'">
            <div class="item"><el-tag size="mini" type="danger">机构名称</el-tag><span class="value">{{ val.account_name || '-' }}</span></div>
          </template>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { queryAccountList } from '@/api/accountV2'

export default {
  name: 'Account',
  data() {
    return {
      loading: true,
      accountList: [],
      keyword: '',
      roleFilter: 'all',
      activeRole: 'all',
      roleTabs: [
        { label: '全部', value: 'all' },
        { label: '医生', value: 'doctor' },
        { label: '患者', value: 'patient' },
        { label: '药店', value: 'drugstore' },
        { label: '保险机构', value: 'insurance' },
        { label: '管理员', value: 'admin' }
      ]
    }
  },
  computed: {
    ...mapGetters([
      'account_id',
      'roles',
      'account_name'
    ]),
    normalizedList() {
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
    filteredList() {
      const kw = this.keyword.trim().toLowerCase()
      return this.normalizedList.filter(item => {
        const byRoleFilter = this.roleFilter === 'all' || item.role === this.roleFilter
        const byTab = this.activeRole === 'all' || item.role === this.activeRole
        const byKeyword = !kw ||
          String(item.account_name || '').toLowerCase().includes(kw) ||
          String(item.phone || '').toLowerCase().includes(kw)
        return byRoleFilter && byTab && byKeyword
      })
    },
    currentList() {
      return this.filteredList
    }
  },
  created() {
    queryAccountList().then(response => {
      if (response !== null) {
        this.accountList = response
      }
      this.loading = false
    }).catch(() => {
      this.loading = false
    })
  },
  methods: {
    countByRole(role) {
      if (role === 'all') return this.normalizedList.length
      return this.normalizedList.filter(item => item.role === role).length
    },
    roleLabel(role) {
      const map = {
        doctor: '医生',
        patient: '患者',
        drugstore: '药店',
        insurance: '保险机构',
        admin: '管理员'
      }
      return map[role] || role || '-'
    }
  }
}

</script>

<style scoped>
.container {
  width: 100%;
  min-height: 100%;
  overflow: hidden;
  font-size: 15px;
}

.mb16 {
  margin-bottom: 16px;
}

.toolbar {
  display: flex;
  gap: 12px;
}

.search-input {
  flex: 1;
}

.role-filter {
  width: 180px;
}

.item {
  font-size: 14px;
  margin-bottom: 12px;
  color: #606266;
}

.value {
  margin-left: 8px;
}

.id-text {
  color: #f56c6c;
}

.account-card {
  margin-bottom: 18px;
  min-height: 200px;
}
</style>
