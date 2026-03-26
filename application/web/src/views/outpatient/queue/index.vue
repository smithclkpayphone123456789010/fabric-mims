<template>
  <div class="queue-page">
    <div class="top">
      <div>
        <h2>排队叫号</h2>
        <p>实时查看叫号进度，合理安排候诊时间</p>
      </div>
      <div class="date-pill"><i class="el-icon-time" /> {{ beijingDate }}</div>
    </div>

    <el-card shadow="never" class="filter-card">
      <el-form inline>
        <el-form-item label="就诊医生">
          <el-select v-model="selectedDoctorId" filterable clearable placeholder="搜索医生姓名" class="w220" @change="onDoctorChange">
            <el-option v-for="d in filteredDoctors" :key="d.account_id" :label="d.account_name" :value="d.account_id" />
          </el-select>
        </el-form-item>
        <el-form-item label="医院">
          <el-select v-model="selectedHospital" filterable clearable placeholder="搜索医院" class="w220" @change="onHospitalChange">
            <el-option v-for="h in hospitalOptions" :key="h" :label="h" :value="h" />
          </el-select>
        </el-form-item>
        <el-button type="primary" :loading="loading" icon="el-icon-refresh" @click="load">刷新</el-button>
        <el-button :type="auto ? 'danger' : 'default'" @click="toggleAuto">{{ auto ? '关闭自动刷新' : '开启自动刷新' }}</el-button>
      </el-form>
    </el-card>

    <div class="summary-grid">
      <el-card class="summary current" shadow="never">
        <div class="label">当前叫号</div>
        <div class="value">{{ callingNo || '-' }}</div>
        <div class="sub">正在就诊号码</div>
      </el-card>
      <el-card class="summary mine" shadow="never">
        <div class="label">我的号码</div>
        <div class="value">{{ myNo || '-' }}</div>
        <div class="sub">当前预约队列号</div>
      </el-card>
      <el-card class="summary front" shadow="never">
        <div class="label">前方人数</div>
        <div class="value">{{ frontCount }}</div>
        <div class="sub">预计等待中</div>
      </el-card>
    </div>

    <el-card shadow="never" class="progress-card">
      <div class="progress-title">我的排队进度</div>
      <el-steps :active="stepActive" finish-status="success" align-center>
        <el-step title="已预约" />
        <el-step title="等待叫号" />
        <el-step title="就诊中" />
        <el-step title="就诊完成" />
      </el-steps>
    </el-card>

    <el-card shadow="never" class="table-card">
      <div slot="header" class="table-head">
        <span>候诊队列</span>
        <el-tag size="mini" type="info">每 5 秒可自动更新</el-tag>
      </div>
      <el-table :data="list" border stripe v-loading="loading" row-key="registration_id">
        <el-table-column label="号码" width="90">
          <template slot-scope="s"><b class="qno">{{ s.row.queue_no }}</b></template>
        </el-table-column>
        <el-table-column prop="patient_id" label="患者ID" min-width="160"/>
        <el-table-column label="状态" width="120">
          <template slot-scope="s">
            <el-tag v-if="s.row.status==='WAITING'" size="mini" type="info">等待中</el-tag>
            <el-tag v-else-if="s.row.status==='IN_PROGRESS'" size="mini" type="warning">就诊中</el-tag>
            <el-tag v-else size="mini" type="success">已完成</el-tag>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && list.length===0" description="当前暂无排队信息" />
    </el-card>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { queryQueueCurrent } from '@/api/outpatient'
import { queryAccountList } from '@/api/accountV2'

export default {
  name: 'OutpatientQueue',
  data() {
    return {
      doctors: [],
      selectedDoctorId: '',
      selectedHospital: '',
      list: [],
      loading: false,
      auto: true,
      timer: null,
      callingNo: '',
      myNo: '-',
      myStatus: '',
      frontCount: 0
    }
  },
  computed: {
    ...mapGetters(['account_id']),
    beijingDate() {
      const f = new Intl.DateTimeFormat('zh-CN', { timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit' })
      const p = f.formatToParts(new Date())
      return `${p.find(i => i.type === 'year').value}-${p.find(i => i.type === 'month').value}-${p.find(i => i.type === 'day').value}`
    },
    hospitalOptions() {
      return Array.from(new Set(this.doctors.map(d => d.hospital_name).filter(Boolean)))
    },
    filteredDoctors() {
      return this.doctors.filter(d => !this.selectedHospital || d.hospital_name === this.selectedHospital)
    },
    stepActive() {
      if (!this.myNo || this.myNo === '-') return 0
      if (this.myStatus === 'DONE') return 4
      if (this.myStatus === 'IN_PROGRESS') return 3
      if (this.myStatus === 'WAITING') return 2
      return 1
    }
  },
  async created() {
    const accounts = await queryAccountList().catch(() => [])
    this.doctors = (accounts || []).filter(a => a.role === 'doctor')
    if (this.doctors.length) {
      this.selectedDoctorId = this.doctors[0].account_id
      this.selectedHospital = this.doctors[0].hospital_name || ''
    }
    this.load()
    this.startAuto()
  },
  beforeDestroy() {
    this.stopAuto()
  },
  methods: {
    onHospitalChange() {
      const first = this.filteredDoctors[0]
      this.selectedDoctorId = first ? first.account_id : ''
      this.load()
    },
    onDoctorChange() {
      const d = this.doctors.find(i => i.account_id === this.selectedDoctorId)
      if (d && d.hospital_name) this.selectedHospital = d.hospital_name
      this.load()
    },
    async load() {
      if (!this.selectedDoctorId) { this.list = []; return }
      this.loading = true
      const queue = await queryQueueCurrent({ doctor_id: this.selectedDoctorId }).catch(() => [])

      this.list = (queue || []).sort((a, b) => String(a.queue_no).localeCompare(String(b.queue_no)))
      const inP = this.list.find(i => i.status === 'IN_PROGRESS')
      this.callingNo = inP ? inP.queue_no : ''

      const myAll = this.list.filter(i => i.patient_id === this.account_id)
      const myActive = myAll.find(i => i.status === 'IN_PROGRESS') || myAll.find(i => i.status === 'WAITING')
      const myLatest = myAll.sort((a, b) => String(b.queue_no).localeCompare(String(a.queue_no)))[0]
      const myQueue = myActive || myLatest || null

      this.myNo = myQueue ? myQueue.queue_no : '-'
      this.myStatus = myQueue ? myQueue.status : ''

      if (this.myNo !== '-' && this.myStatus === 'WAITING') {
        const waiting = this.list.filter(i => i.status === 'WAITING').map(i => i.queue_no)
        this.frontCount = waiting.filter(n => String(n) < String(this.myNo)).length
      } else {
        this.frontCount = 0
      }
      this.loading = false
    },
    startAuto() {
      this.stopAuto()
      if (!this.auto) return
      this.timer = setInterval(() => this.load(), 5000)
    },
    stopAuto() {
      if (this.timer) {
        clearInterval(this.timer)
        this.timer = null
      }
    },
    toggleAuto() {
      this.auto = !this.auto
      if (this.auto) this.startAuto()
      else this.stopAuto()
    }
  }
}
</script>

<style scoped>
.queue-page{padding:16px;background:#f5f7fb;min-height:100%}
.top{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:12px}
.top h2{margin:0;font-size:28px;font-weight:700;color:#111827}
.top p{margin:4px 0 0;color:#6b7280;font-size:13px}
.date-pill{margin-top:6px;font-size:12px;color:#3b82f6;background:#eaf2ff;border:1px solid #bfd6ff;padding:6px 10px;border-radius:999px}
.filter-card{border-radius:10px;margin-bottom:12px}.w220{width:220px}
.summary-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;margin-bottom:12px}
.summary{border-radius:10px}.summary .label{font-size:12px;color:#6b7280}.summary .value{font-size:30px;font-weight:700;line-height:1.2;margin-top:6px}.summary .sub{font-size:11px;color:#9ca3af}
.summary.current .value{color:#ea580c}.summary.mine .value{color:#2563eb}.summary.front .value{color:#16a34a}
.progress-card{border-radius:10px;margin-bottom:12px}.progress-title{margin-bottom:10px;font-size:13px;font-weight:700;color:#111827}
.table-card{border-radius:10px}.table-head{display:flex;justify-content:space-between;align-items:center}
.qno{color:#1d4ed8;font-size:16px}
</style>