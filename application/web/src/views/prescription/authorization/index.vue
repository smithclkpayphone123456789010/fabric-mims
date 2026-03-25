<template>
  <div class="auth-page">
    <el-alert type="success" :closable="false" class="mb16 account-alert">
      <p>账户ID: {{ account_id }}</p>
      <p>用户名: {{ account_name }}</p>
    </el-alert>

    <el-row :gutter="16" class="mb16">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card stat-total">
          <div class="stat-title">授权总数</div>
          <div class="stat-value">{{ authList.length }}</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card stat-active">
          <div class="stat-title">生效中</div>
          <div class="stat-value">{{ stats.active }}</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card stat-revoked">
          <div class="stat-title">已撤销</div>
          <div class="stat-value">{{ stats.revoked }}</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="stat-card stat-expired">
          <div class="stat-title">已过期</div>
          <div class="stat-value">{{ stats.expired }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="section-card" shadow="always">
      <div slot="header" class="section-title">可授权病历</div>

      <div class="filter-row filter-row--figma">
        <el-input v-model.trim="recordFilters.id" clearable placeholder="请输入病历ID" class="w220" @keyup.enter.native="applyRecordFilters">
          <i slot="prefix" class="el-input__icon el-icon-tickets" />
        </el-input>
        <el-input v-model.trim="recordFilters.doctor" clearable placeholder="请输入医生ID" class="w220" @keyup.enter.native="applyRecordFilters">
          <i slot="prefix" class="el-input__icon el-icon-user" />
        </el-input>
        <el-input v-model.trim="recordFilters.name" clearable placeholder="请输入医生姓名" class="w220" @keyup.enter.native="applyRecordFilters">
          <i slot="prefix" class="el-input__icon el-icon-user-solid" />
        </el-input>
        <el-input v-model.trim="recordFilters.hospital" clearable placeholder="请选择医院" class="w220" @keyup.enter.native="applyRecordFilters">
          <i slot="prefix" class="el-input__icon el-icon-office-building" />
        </el-input>
        <el-input v-model.trim="recordFilters.department" clearable placeholder="请选择科室" class="w220" @keyup.enter.native="applyRecordFilters">
          <i slot="prefix" class="el-input__icon el-icon-collection-tag" />
        </el-input>
        <el-date-picker
          v-model="recordFilters.createdRange"
          type="daterange"
          value-format="yyyy-MM-dd"
          range-separator="-"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          class="w280"
          unlink-panels
        />
        <el-button type="primary" icon="el-icon-search" @click="applyRecordFilters">查询</el-button>
        <el-button icon="el-icon-refresh-left" @click="resetRecordFilters">重置</el-button>
      </div>

      <el-table :data="recordPagedList" stripe border class="figma-table" :header-cell-style="headerCellStyle" :cell-style="cellStyle">
        <el-table-column prop="id" label="病历ID" min-width="184">
          <template slot-scope="scope">
            <span class="id-link">{{ scope.row.id }}</span>
          </template>
        </el-table-column>
        <el-table-column label="医院" min-width="196">
          <template slot-scope="scope">
            <i class="el-icon-office-building row-icon" />{{ getRecordHospital(scope.row) || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="科室" min-width="132">
          <template slot-scope="scope">{{ getRecordDepartment(scope.row) || '-' }}</template>
        </el-table-column>
        <el-table-column label="医生姓名" min-width="136">
          <template slot-scope="scope">
            <i class="el-icon-user row-icon" />{{ getDoctorInfo(scope.row.doctor).account_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created" label="创建时间" min-width="182" />
        <el-table-column label="病历类型" min-width="126" align="center">
          <template slot-scope="scope"><el-tag class="type-tag" size="mini">{{ scope.row.record_type || '门诊病历' }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="128" align="center" fixed="right">
          <template slot-scope="scope">
            <el-button type="success" size="mini" icon="el-icon-circle-check" class="btn-grant" @click="openGrantDialog(scope.row)">授权</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="recordFilteredList.length"
          :current-page.sync="recordPager.page"
          :page-size="recordPager.pageSize"
        />
      </div>
    </el-card>

    <el-card class="section-card" shadow="always">
      <div slot="header" class="section-title">授权记录</div>

      <div class="filter-row filter-row--figma">
        <el-input v-model.trim="authFilters.record_id" clearable placeholder="病历ID" class="w220" @keyup.enter.native="applyAuthFilters">
          <i slot="prefix" class="el-input__icon el-icon-tickets" />
        </el-input>
        <el-input v-model.trim="authFilters.doctor_name" clearable placeholder="医生姓名" class="w220" @keyup.enter.native="applyAuthFilters">
          <i slot="prefix" class="el-input__icon el-icon-user" />
        </el-input>
        <el-select v-model="authFilters.status" clearable placeholder="状态" class="w220">
          <el-option label="全部" value="" />
          <el-option label="授权中" value="active" />
          <el-option label="已撤销" value="revoked" />
          <el-option label="已过期" value="expired" />
        </el-select>
        <el-button type="primary" icon="el-icon-search" @click="applyAuthFilters">查询</el-button>
        <el-button icon="el-icon-refresh-left" @click="resetAuthFilters">重置</el-button>
      </div>

      <el-table :data="authPagedList" stripe border class="figma-table" :default-sort="{ prop: 'updated_time', order: 'descending' }" :header-cell-style="headerCellStyle" :cell-style="cellStyle" :row-class-name="authRowClassName">
        <el-table-column prop="record_id" label="病历ID" min-width="176" />
        <el-table-column label="医生姓名" min-width="132">
          <template slot-scope="scope">
            <i class="el-icon-user row-icon" />{{ getDoctorInfo(scope.row.doctor_id).account_name || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="医院" min-width="188">
          <template slot-scope="scope">
            <i class="el-icon-office-building row-icon" />{{ getAuthHospital(scope.row) || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="科室" min-width="128">
          <template slot-scope="scope">{{ getAuthDepartment(scope.row) || '-' }}</template>
        </el-table-column>
        <el-table-column label="状态" width="112" align="center">
          <template slot-scope="scope">
            <span :class="['status-pill', `status-pill--${scope.row.status || 'unknown'}`]">{{ statusLabel(scope.row.status) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="start_time" label="授权时间" min-width="168" sortable="custom" />
        <el-table-column prop="updated_time" label="更新时间" min-width="168" sortable="custom" />
        <el-table-column prop="end_time" label="到期时间" min-width="168" />
        <el-table-column label="操作" width="214" align="center" fixed="right">
          <template slot-scope="scope">
            <el-button v-if="scope.row.status === 'active'" size="mini" type="primary" plain icon="el-icon-refresh-left" @click="openRenewDialog(scope.row)">续期</el-button>
            <el-button v-if="scope.row.status === 'active'" size="mini" type="danger" plain icon="el-icon-close" @click="onRevoke(scope.row)">撤销</el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          background
          layout="total, prev, pager, next"
          :total="authFilteredList.length"
          :current-page.sync="authPager.page"
          :page-size="authPager.pageSize"
        />
      </div>
    </el-card>

    <el-dialog
      :visible.sync="grantVisible"
      width="560px"
      custom-class="auth-dialog"
      :close-on-click-modal="false"
      :destroy-on-close="true"
    >
      <div slot="title" class="dialog-title-wrap">
        <i class="el-icon-medal icon-green" />
        <span class="dialog-title">病历授权</span>
      </div>
      <div class="dialog-subtitle">为病历 {{ selectedRecord.id || '-' }} 授权访问权限</div>

      <div class="dialog-section-title">选择授权医生</div>
      <div class="doctor-search-grid">
        <el-input v-model.trim="doctorSearch.id" placeholder="输入医生ID搜索" clearable @input="filterDoctors" />
        <el-input v-model.trim="doctorSearch.name" placeholder="输入医生姓名搜索" clearable @input="filterDoctors" />
        <el-input v-model.trim="doctorSearch.hospital" placeholder="输入医院搜索" clearable @input="filterDoctors" />
        <el-input v-model.trim="doctorSearch.department" placeholder="输入科室搜索" clearable @input="filterDoctors" />
      </div>

      <div class="dialog-section-title mt16">搜索结果</div>
      <div class="doctor-list">
        <div
          v-for="item in doctorResultList"
          :key="item.account_id"
          class="doctor-item"
          :class="{ active: grantForm.doctor_id === item.account_id }"
          @click="selectDoctor(item)"
        >
          <div class="doctor-name">{{ item.account_name || '-' }}</div>
          <div class="doctor-desc">ID: {{ item.account_id }} | {{ item.hospital_name || '-' }} - {{ item.department || '-' }}</div>
        </div>
        <div v-if="doctorResultList.length === 0" class="empty-text">未找到匹配医生</div>
      </div>

      <div class="dialog-section-title mt16">设置到期时间</div>
      <el-date-picker
        v-model="grantForm.end_time"
        type="datetime"
        value-format="yyyy-MM-dd HH:mm:ss"
        placeholder="选择到期日期时间"
        class="full-width"
      />

      <span slot="footer">
        <el-button @click="grantVisible = false">取消</el-button>
        <el-button type="success" :loading="grantLoading" @click="onGrant">确认授权</el-button>
      </span>
    </el-dialog>

    <el-dialog
      :visible.sync="renewVisible"
      width="560px"
      custom-class="auth-dialog"
      :close-on-click-modal="false"
      :destroy-on-close="true"
    >
      <div slot="title" class="dialog-title-wrap">
        <i class="el-icon-refresh-left icon-blue" />
        <span class="dialog-title">授权续期</span>
      </div>
      <div class="dialog-subtitle">为授权记录 {{ renewItem.id || '-' }} 设置新的到期时间</div>

      <div class="renew-info-box">
        <div class="renew-row"><span>病历ID:</span><b>{{ renewItem.record_id || '-' }}</b></div>
        <div class="renew-row"><span>医生:</span><b>{{ getDoctorInfo(renewItem.doctor_id).account_name || '-' }}</b></div>
        <div class="renew-row"><span>当前到期时间:</span><b>{{ renewItem.end_time || '-' }}</b></div>
      </div>

      <div class="dialog-section-title">选择新的到期时间</div>
      <el-date-picker
        v-model="renewForm.end_time"
        type="datetime"
        value-format="yyyy-MM-dd HH:mm:ss"
        placeholder="选择到期日期时间"
        class="full-width"
      />

      <span slot="footer">
        <el-button @click="renewVisible = false">取消</el-button>
        <el-button type="primary" :loading="renewLoading" @click="onRenew">确认续期</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { queryAccountList } from '@/api/accountV2'
import { queryPrescriptionList } from '@/api/prescription'
import {
  grantRecordAuthorization,
  queryMyAuthorizations,
  revokeRecordAuthorization,
  renewRecordAuthorization
} from '@/api/authorization'

export default {
  name: 'PrescriptionAuthorization',
  data() {
    return {
      allDoctors: [],
      patientRecords: [],
      authList: [],

      recordFilters: { id: '', doctor: '', name: '', hospital: '', department: '', createdRange: [] },
      authFilters: { record_id: '', doctor_name: '', status: '' },

      recordFilteredList: [],
      authFilteredList: [],

      recordPager: { page: 1, pageSize: 5 },
      authPager: { page: 1, pageSize: 5 },

      grantVisible: false,
      renewVisible: false,
      grantLoading: false,
      renewLoading: false,

      selectedRecord: {},
      renewItem: {},

      doctorSearch: { id: '', name: '', hospital: '', department: '' },
      doctorResultList: [],
      grantForm: {
        doctor_id: '',
        hospital_name: '',
        department: '',
        end_time: '',
        remark: ''
      },
      renewForm: {
        end_time: ''
      }
    }
  },
  computed: {
    ...mapGetters(['account_id', 'account_name']),
    recordPagedList() {
      const start = (this.recordPager.page - 1) * this.recordPager.pageSize
      return this.recordFilteredList.slice(start, start + this.recordPager.pageSize)
    },
    authPagedList() {
      const start = (this.authPager.page - 1) * this.authPager.pageSize
      return this.authFilteredList.slice(start, start + this.authPager.pageSize)
    },
    stats() {
      return this.authList.reduce((acc, item) => {
        if (item.status === 'active') acc.active += 1
        if (item.status === 'revoked') acc.revoked += 1
        if (item.status === 'expired') acc.expired += 1
        return acc
      }, { active: 0, revoked: 0, expired: 0 })
    }
  },
  created() {
    this.fetchData()
  },
  methods: {
    headerCellStyle() {
      return {
        height: '44px',
        padding: '10px 14px',
        fontSize: '12px',
        color: '#475569',
        background: '#f8fafc',
        fontWeight: 600,
        borderColor: 'rgba(148, 163, 184, 0.22)'
      }
    },
    cellStyle() {
      return {
        height: '46px',
        padding: '11px 14px',
        fontSize: '12px',
        color: '#0f172a',
        borderColor: 'rgba(148, 163, 184, 0.18)'
      }
    },
    fetchData() {
      Promise.all([
        queryAccountList(),
        queryPrescriptionList({ patient: this.account_id }),
        queryMyAuthorizations({ patient_id: this.account_id })
      ]).then(([accounts, records, auths]) => {
        this.allDoctors = (accounts || []).filter(i => i.role === 'doctor')
        this.patientRecords = records || []
        this.authList = (auths || []).sort((a, b) => String(b.updated_time || '').localeCompare(String(a.updated_time || '')))

        this.applyRecordFilters()
        this.applyAuthFilters()
      })
    },

    getDoctorInfo(id) {
      return this.allDoctors.find(i => i.account_id === id) || {}
    },
    getRecordHospital(record) {
      return record.hospital_name || this.getDoctorInfo(record.doctor).hospital_name || ''
    },
    getRecordDepartment(record) {
      return record.department || this.getDoctorInfo(record.doctor).department || ''
    },
    getAuthHospital(auth) {
      return auth.hospital_name || this.getDoctorInfo(auth.doctor_id).hospital_name || ''
    },
    getAuthDepartment(auth) {
      return auth.department || this.getDoctorInfo(auth.doctor_id).department || ''
    },

    applyRecordFilters() {
      const f = this.recordFilters
      const [start, end] = f.createdRange || []
      this.recordPager.page = 1
      this.recordFilteredList = this.patientRecords.filter(item => {
        const doc = this.getDoctorInfo(item.doctor)
        const createdDate = item.created && item.created.length >= 10 ? item.created.slice(0, 10) : ''
        const byId = !f.id || String(item.id || '').toLowerCase().includes(String(f.id).toLowerCase())
        const byDoctor = !f.doctor || String(item.doctor || '').toLowerCase().includes(String(f.doctor).toLowerCase())
        const byName = !f.name || String(doc.account_name || '').toLowerCase().includes(String(f.name).toLowerCase())
        const byHospital = !f.hospital || String(this.getRecordHospital(item) || '').toLowerCase().includes(String(f.hospital).toLowerCase())
        const byDept = !f.department || String(this.getRecordDepartment(item) || '').toLowerCase().includes(String(f.department).toLowerCase())
        const byStart = !start || (createdDate && createdDate >= start)
        const byEnd = !end || (createdDate && createdDate <= end)
        return byId && byDoctor && byName && byHospital && byDept && byStart && byEnd
      })
    },
    resetRecordFilters() {
      this.recordFilters = { id: '', doctor: '', name: '', hospital: '', department: '', createdRange: [] }
      this.applyRecordFilters()
    },

    applyAuthFilters() {
      const f = this.authFilters
      this.authPager.page = 1
      this.authFilteredList = this.authList.filter(item => {
        const docName = this.getDoctorInfo(item.doctor_id).account_name || ''
        const byRecord = !f.record_id || String(item.record_id || '').toLowerCase().includes(String(f.record_id).toLowerCase())
        const byDoctorName = !f.doctor_name || String(docName).toLowerCase().includes(String(f.doctor_name).toLowerCase())
        const byStatus = !f.status || item.status === f.status
        return byRecord && byDoctorName && byStatus
      }).sort((a, b) => String(b.updated_time || '').localeCompare(String(a.updated_time || '')))
    },
    resetAuthFilters() {
      this.authFilters = { record_id: '', doctor_name: '', status: '' }
      this.applyAuthFilters()
    },

    openGrantDialog(record) {
      this.selectedRecord = record || {}
      this.grantVisible = true
      this.grantForm = {
        doctor_id: '',
        hospital_name: '',
        department: '',
        end_time: '',
        remark: ''
      }
      this.doctorSearch = { id: '', name: '', hospital: '', department: '' }
      this.doctorResultList = [...this.allDoctors]
    },
    filterDoctors() {
      const s = this.doctorSearch
      this.doctorResultList = this.allDoctors.filter(i => {
        const byId = !s.id || String(i.account_id || '').includes(s.id)
        const byName = !s.name || String(i.account_name || '').includes(s.name)
        const byHospital = !s.hospital || String(i.hospital_name || '').includes(s.hospital)
        const byDept = !s.department || String(i.department || '').includes(s.department)
        return byId && byName && byHospital && byDept
      })
    },
    selectDoctor(item) {
      this.grantForm.doctor_id = item.account_id
      this.grantForm.hospital_name = item.hospital_name || this.getRecordHospital(this.selectedRecord) || ''
      this.grantForm.department = item.department || this.getRecordDepartment(this.selectedRecord) || ''
    },
    onGrant() {
      if (!this.selectedRecord.id) {
        this.$message.error('缺少病历信息')
        return
      }
      if (!this.grantForm.doctor_id) {
        this.$message.warning('请选择授权医生')
        return
      }
      if (!this.grantForm.end_time) {
        this.$message.warning('请选择到期时间')
        return
      }
      if (new Date(this.grantForm.end_time).getTime() <= Date.now()) {
        this.$message.warning('到期时间必须晚于当前时间')
        return
      }

      this.grantLoading = true
      grantRecordAuthorization({
        patient_id: this.account_id,
        record_id: this.selectedRecord.id,
        doctor_id: this.grantForm.doctor_id,
        hospital_name: this.grantForm.hospital_name,
        department: this.grantForm.department,
        end_time: this.grantForm.end_time,
        remark: this.grantForm.remark || ''
      }).then(() => {
        this.$message.success('授权成功')
        this.grantLoading = false
        this.grantVisible = false
        this.fetchData()
      }).catch(() => {
        this.grantLoading = false
      })
    },

    openRenewDialog(row) {
      this.renewItem = row
      this.renewForm.end_time = ''
      this.renewVisible = true
    },
    onRenew() {
      if (!this.renewItem.id) {
        this.$message.error('缺少授权信息')
        return
      }
      if (!this.renewForm.end_time) {
        this.$message.warning('请选择续期时间')
        return
      }
      if (new Date(this.renewForm.end_time).getTime() <= Date.now()) {
        this.$message.warning('续期时间必须晚于当前时间')
        return
      }

      this.renewLoading = true
      renewRecordAuthorization({
        patient_id: this.account_id,
        auth_id: this.renewItem.id,
        end_time: this.renewForm.end_time
      }).then(() => {
        this.$message.success('续期成功')
        this.renewLoading = false
        this.renewVisible = false
        this.fetchData()
      }).catch(() => {
        this.renewLoading = false
      })
    },

    onRevoke(row) {
      this.$confirm('确认撤销该授权吗？', '提示', {
        type: 'warning'
      }).then(() => {
        revokeRecordAuthorization({ patient_id: this.account_id, auth_id: row.id }).then(() => {
          this.$message.success('撤销成功')
          this.fetchData()
        })
      }).catch(() => {})
    },
    authRowClassName({ row }) {
      if (row.status === 'active') return 'active-row'
      return ''
    },

    statusLabel(status) {
      if (status === 'active') return '授权中'
      if (status === 'revoked') return '已撤销'
      if (status === 'expired') return '已过期'
      return '-'
    }
  }
}
</script>

<style scoped>
.auth-page {
  width: 100%;
  min-height: 100%;
  background: #f6f8fc;
}
.account-alert /deep/ .el-alert__content {
  padding: 2px 0;
}
.mb16 { margin-bottom: 16px; }
.section-card {
  border-radius: 10px;
  margin-bottom: 16px;
}
.section-title {
  font-size: 16px;
  font-weight: 700;
  color: #1f2d3d;
}
.stat-card {
  border-radius: 10px;
}
.stat-title {
  color: #6b7280;
  font-size: 13px;
}
.stat-value {
  margin-top: 8px;
  font-size: 28px;
  line-height: 1;
  font-weight: 700;
  color: #111827;
}
.stat-total { border-left: 4px solid #3b82f6; }
.stat-active { border-left: 4px solid #22c55e; }
.stat-revoked { border-left: 4px solid #ef4444; }
.stat-expired { border-left: 4px solid #9ca3af; }

.filter-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}
.filter-row--figma /deep/ .el-input__inner,
.filter-row--figma /deep/ .el-range-input,
.filter-row--figma /deep/ .el-range-separator,
.filter-row--figma /deep/ .el-select .el-input__inner {
  height: 36px;
  line-height: 36px;
  font-size: 12px;
}
.filter-row--figma /deep/ .el-input__prefix,
.filter-row--figma /deep/ .el-input__icon {
  height: 36px;
  line-height: 36px;
  color: #94a3b8;
  font-size: 13px;
}
.filter-row--figma /deep/ .el-date-editor .el-range-separator {
  width: 22px;
  color: #94a3b8;
}
.filter-row--figma /deep/ .el-button {
  height: 36px;
  padding: 0 16px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}
.w220 { width: 220px; }
.w280 { width: 280px; }
.pagination-wrap {
  margin-top: 12px;
  padding: 6px 0 2px;
  display: flex;
  justify-content: flex-end;
}
.row-icon {
  color: #64748b;
  margin-right: 6px;
  font-size: 12px;
}
.id-link {
  color: #2563eb;
  font-weight: 500;
}
.type-tag {
  border: none;
  color: #1d4ed8;
  background: #dbeafe;
  border-radius: 999px;
  padding: 0 8px;
}
.btn-grant {
  border-radius: 10px;
  padding: 7px 12px;
  box-shadow: 0 1px 2px rgba(22, 163, 74, 0.2);
}
.status-pill {
  display: inline-block;
  min-width: 56px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 11px;
  line-height: 18px;
  font-weight: 600;
}
.status-pill--active {
  color: #15803d;
  background: #dcfce7;
}
.status-pill--revoked {
  color: #dc2626;
  background: #fee2e2;
}
.status-pill--expired,
.status-pill--unknown {
  color: #6b7280;
  background: #f3f4f6;
}
.figma-table /deep/ .el-table__header th {
  border-bottom: 1px solid rgba(148, 163, 184, 0.22) !important;
}
.figma-table /deep/ .el-table__body td {
  line-height: 20px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.14);
}
.figma-table /deep/ .el-table__body tr {
  transition: background-color .16s ease;
}
.figma-table /deep/ .el-table__body tr:hover > td {
  background: #f8fbff !important;
}
.figma-table /deep/ .el-table__body tr.current-row > td,
.figma-table /deep/ .el-table__body tr.active-row > td {
  background: #eef6ff !important;
}

.dialog-title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 1px;
}
.dialog-title {
  font-size: 20px;
  line-height: 28px;
  letter-spacing: 0.1px;
  font-weight: 700;
  color: #111827;
}
.dialog-subtitle {
  color: #717182;
  font-size: 12px;
  line-height: 18px;
  margin: 2px 0 12px;
}
.icon-green { color: #16a34a; }
.icon-blue { color: #2563eb; }
.dialog-section-title {
  font-size: 16px;
  line-height: 22px;
  font-weight: 700;
  color: #111827;
  margin-bottom: 8px;
}
.mt16 { margin-top: 16px; }
.full-width { width: 100%; }

.doctor-search-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}
.doctor-list {
  max-height: 260px;
  border: 1px solid rgba(15, 23, 42, 0.1);
  border-radius: 12px;
  background: #fff;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.9), 0 2px 8px rgba(15,23,42,0.04);
  overflow-y: auto;
}
.doctor-item {
  padding: 12px 14px;
  border-bottom: 1px solid rgba(148,163,184,0.2);
  cursor: pointer;
  transition: background-color .14s ease;
}
.doctor-item:hover {
  background: #f8fbff;
}
.doctor-item:last-child {
  border-bottom: 0;
}
.doctor-item.active {
  background: #eef6ff;
}
.doctor-name {
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
}
.doctor-desc {
  margin-top: 4px;
  color: #64748b;
  font-size: 12px;
}
.empty-text {
  text-align: center;
  color: #909399;
  padding: 24px 0;
}

.renew-info-box {
  background: linear-gradient(180deg, #fbfdff 0%, #f8fafc 100%);
  border: 1px solid rgba(15,23,42,0.1);
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(15,23,42,0.05);
  padding: 14px 16px;
  margin-bottom: 14px;
}
.renew-row {
  display: flex;
  justify-content: space-between;
  line-height: 1.85;
  color: #475569;
  font-size: 12px;
}
.renew-row b {
  color: #0f172a;
  font-size: 12px;
  font-weight: 600;
}
</style>
