<template>
  <div class="queue-page">
    <div class="queue-head">
      <div>
        <h2>就诊队列管理</h2>
        <p>管理今日患者就诊流程</p>
      </div>
      <div class="date-pill"><i class="el-icon-time" /> 今日日期：{{ beijingDate }}</div>
    </div>

    <div class="kpi-grid">
      <div class="kpi kpi-blue"><div class="kpi-title">等待就诊</div><div class="kpi-value">{{ waitingList.length }}</div></div>
      <div class="kpi kpi-orange"><div class="kpi-title">正在就诊</div><div class="kpi-value">{{ currentItem ? 1 : 0 }}</div></div>
      <div class="kpi kpi-green"><div class="kpi-title">已完成</div><div class="kpi-value">{{ doneList.length }}</div></div>
      <div class="kpi kpi-purple"><div class="kpi-title">总患者数</div><div class="kpi-value">{{ list.length }}</div></div>
    </div>

    <el-card class="section section-current" shadow="never">
      <div slot="header" class="section-head section-head-current">
        <div><i class="el-icon-caret-right" /> 当前就诊患者</div>
        <el-tag type="warning" size="mini">正在就诊</el-tag>
      </div>
      <div v-if="currentItem" class="current-row">
        <div class="queue-no current">{{ currentItem.queue_no }}</div>
        <div class="info-cell"><label>患者姓名</label><b>{{ currentItem.patient_name || '-' }}</b></div>
        <div class="info-cell"><label>性别 / 年龄</label><b>{{ currentItem.gender || '-' }} / {{ currentItem.age || '-' }}</b></div>
        <div class="info-cell"><label>预约时间</label><b>{{ currentItem.appointment_time || '-' }}</b></div>
        <div class="info-cell complaint"><label>主诉</label><b>{{ currentItem.complaint || '-' }}</b></div>
        <el-button type="success" size="mini" icon="el-icon-circle-check" @click="finish(currentItem)">完成就诊</el-button>
      </div>
      <el-empty v-else description="当前暂无就诊患者" :image-size="70"/>
    </el-card>

    <el-card class="section" shadow="never">
      <div slot="header" class="section-head section-head-wait">
        <div><i class="el-icon-user" /> 等待队列</div>
        <el-tag type="primary" size="mini">{{ waitingList.length }} 位患者等待中</el-tag>
      </div>
      <div v-for="it in waitingList" :key="it.registration_id" class="wait-row" :class="{ first: firstWaiting && firstWaiting.registration_id===it.registration_id }">
        <div class="queue-no">{{ it.queue_no }}</div>
        <div class="info-cell"><label>患者姓名</label><b>{{ it.patient_name || '-' }}</b></div>
        <div class="info-cell"><label>性别 / 年龄</label><b>{{ it.gender || '-' }} / {{ it.age || '-' }}</b></div>
        <div class="info-cell"><label>预约时间</label><b>{{ it.appointment_time || '-' }}</b></div>
        <div class="info-cell complaint"><label>主诉</label><b>{{ it.complaint || '-' }}</b></div>
        <div class="actions">
          <el-button size="mini" @click="openDetail(it)">详情</el-button>
          <el-button type="primary" size="mini" icon="el-icon-caret-right" :disabled="!!currentItem || (firstWaiting && firstWaiting.registration_id!==it.registration_id)" @click="start(it)">开始就诊</el-button>
        </div>
      </div>
      <el-empty v-if="!waitingList.length" description="暂无等待患者" :image-size="70"/>
    </el-card>

    <el-card class="section" shadow="never">
      <div slot="header" class="section-head section-head-done"><div><i class="el-icon-success" /> 已完成就诊</div></div>
      <div v-for="it in doneList" :key="it.registration_id" class="done-row">
        <div class="queue-no done">{{ it.queue_no }}</div>
        <div class="info-cell"><label>患者姓名</label><b>{{ it.patient_name || '-' }}</b></div>
        <div class="info-cell"><label>性别 / 年龄</label><b>{{ it.gender || '-' }} / {{ it.age || '-' }}</b></div>
        <div class="info-cell"><label>预约时间</label><b>{{ it.appointment_time || '-' }}</b></div>
        <el-tag size="mini" type="success">已完成</el-tag>
      </div>
      <el-empty v-if="!doneList.length" description="暂无完成记录" :image-size="70"/>
    </el-card>

    <el-dialog :visible.sync="detailVisible" width="760px" custom-class="q-detail-dialog">
      <div slot="title" class="dialog-title"><i class="el-icon-document" /> 患者详情</div>
      <div v-if="detailItem" class="detail-wrap">
        <div class="detail-grid">
          <div class="d-item"><span>患者姓名</span><b>{{ detailItem.patient_name || '-' }}</b></div>
          <div class="d-item"><span>患者ID</span><b>{{ detailItem.patient_id || '-' }}</b></div>
          <div class="d-item"><span>性别 / 年龄</span><b>{{ detailItem.gender || '-' }} / {{ detailItem.age || '-' }}</b></div>
          <div class="d-item"><span>预约时间</span><b>{{ detailItem.appointment_time || '-' }}</b></div>
          <div class="d-item"><span>排队号</span><b>{{ detailItem.queue_no || '-' }}</b></div>
          <div class="d-item"><span>状态</span><b>{{ statusText(detailItem.status) }}</b></div>
          <div class="d-item"><span>叫号时间</span><b>{{ detailItem.called_time || '-' }}</b></div>
          <div class="d-item"><span>完成时间</span><b>{{ detailItem.finished_time || '-' }}</b></div>
          <div class="d-item"><span>挂号ID</span><b>{{ detailItem.registration_id || '-' }}</b></div>
        </div>
        <div class="detail-block">
          <label>主诉</label>
          <div>{{ detailItem.complaint || '-' }}</div>
        </div>
        <div class="detail-block">
          <label>交易ID</label>
          <div>{{ detailItem.tx_id || '-' }}</div>
        </div>
      </div>
      <span slot="footer"><el-button @click="detailVisible=false">关闭</el-button></span>
    </el-dialog>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { queryQueueCurrent, startVisit, finishVisit, queryRegistrationList, querySlotList } from '@/api/outpatient'
import { queryAccountList } from '@/api/accountV2'

export default {
  name: 'OutpatientDoctorQueue',
  data() {
    return {
      list: [],
      detailVisible: false,
      detailItem: null,
      timer: null,
      accountMap: {},
      regMap: {},
      slotMap: {}
    }
  },
  computed: {
    ...mapGetters(['account_id']),
    beijingDate() {
      const formatter = new Intl.DateTimeFormat('zh-CN', {
        timeZone: 'Asia/Shanghai',
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
      })
      const parts = formatter.formatToParts(new Date())
      const y = parts.find(p => p.type === 'year').value
      const m = parts.find(p => p.type === 'month').value
      const d = parts.find(p => p.type === 'day').value
      return `${y}-${m}-${d}`
    },
    currentItem() { return this.list.find(i => i.status === 'IN_PROGRESS') || null },
    waitingList() { return this.list.filter(i => i.status === 'WAITING').sort((a, b) => String(a.queue_no).localeCompare(String(b.queue_no))) },
    doneList() { return this.list.filter(i => i.status === 'DONE').sort((a, b) => String(a.queue_no).localeCompare(String(b.queue_no))) },
    firstWaiting() { return this.waitingList[0] || null }
  },
  created() {
    this.loadAll()
    this.timer = setInterval(() => this.loadAll(), 5000)
  },
  beforeDestroy() {
    if (this.timer) clearInterval(this.timer)
  },
  methods: {
    statusText(s) {
      if (s === 'WAITING') return '等待中'
      if (s === 'IN_PROGRESS') return '就诊中'
      if (s === 'DONE') return '已完成'
      return s || '-'
    },
    async loadAll() {
      const [queue, accounts, regs, slots] = await Promise.all([
        queryQueueCurrent({ doctor_id: this.account_id }).catch(() => []),
        queryAccountList().catch(() => []),
        queryRegistrationList({ doctor_id: this.account_id }).catch(() => []),
        querySlotList({ doctor_id: this.account_id }).catch(() => [])
      ])

      this.accountMap = {}
      ;(accounts || []).forEach(a => { this.accountMap[a.account_id] = a })

      this.regMap = {}
      ;(regs || []).forEach(r => { this.regMap[r.id] = r })

      this.slotMap = {}
      ;(slots || []).forEach(s => { this.slotMap[s.id] = s })

      this.list = (queue || []).map(q => this.enrichQueueItem(q))
    },
    enrichQueueItem(q) {
      const patient = this.accountMap[q.patient_id] || {}
      const reg = this.regMap[q.registration_id] || {}
      const slot = this.slotMap[reg.schedule_slot_id] || {}
      const date = reg.visit_date || slot.visit_date || ''
      const t = slot.start_time || ''
      return {
        ...q,
        patient_name: patient.account_name || '',
        gender: patient.gender || reg.patient_gender || '',
        age: patient.age || reg.patient_age || '',
        complaint: reg.chief_complaint || reg.symptom_description || '',
        appointment_time: (date && t) ? `${date} ${t}` : (date || t || '-')
      }
    },
    openDetail(it) { this.detailItem = it; this.detailVisible = true },
    start(it) {
      this.$confirm('确认开始该患者就诊？', '开始就诊').then(() => startVisit({ registration_id: it.registration_id, doctor_id: this.account_id })).then(res => {
        this.$notify({ title: '开始就诊成功', message: `交易ID: ${(res.tx_id || '-').slice(0, 18)}...`, type: 'success', duration: 2300, position: 'top-right' })
        this.loadAll()
      })
    },
    finish(it) {
      this.$confirm('确认完成该患者就诊？', '完成就诊').then(() => finishVisit({ registration_id: it.registration_id, doctor_id: this.account_id })).then(res => {
        this.$notify({ title: '完成就诊成功', message: `交易ID: ${(res.tx_id || '-').slice(0, 18)}...`, type: 'success', duration: 2300, position: 'top-right' })
        this.loadAll()
      })
    }
  }
}
</script>

<style scoped>
.queue-page{padding:16px;background:#f2f4f8;min-height:100%}
.queue-head{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:12px}
.queue-head h2{margin:0;font-size:28px;font-weight:700;color:#111827}
.queue-head p{margin:4px 0 0;color:#6b7280;font-size:13px}
.date-pill{margin-top:6px;font-size:12px;color:#3b82f6;background:#eaf2ff;border:1px solid #bfd6ff;padding:6px 10px;border-radius:999px}
.kpi-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin-bottom:12px}
.kpi{border-radius:10px;padding:12px;border:1px solid transparent}
.kpi-title{font-size:12px;color:#374151}.kpi-value{margin-top:6px;font-size:34px;line-height:1;font-weight:700}
.kpi-blue{background:#e8f1ff;border-color:#bfd7ff}.kpi-blue .kpi-value{color:#2563eb}
.kpi-orange{background:#fff1e6;border-color:#ffd3aa}.kpi-orange .kpi-value{color:#ea580c}
.kpi-green{background:#eaf9ef;border-color:#bfe8ca}.kpi-green .kpi-value{color:#16a34a}
.kpi-purple{background:#f3e8ff;border-color:#dfc6ff}.kpi-purple .kpi-value{color:#9333ea}
.section{border-radius:10px;margin-bottom:12px}
.section-head{display:flex;justify-content:space-between;align-items:center;font-size:13px;font-weight:700}
.section-head-current{color:#b45309}.section-head-wait{color:#1d4ed8}.section-head-done{color:#16a34a}
.current-row,.wait-row,.done-row{display:grid;grid-template-columns:56px 1.1fr 1fr 1fr 1.4fr auto;gap:10px;align-items:center;padding:10px;border:1px solid #edf0f5;border-radius:10px;background:#fff}
.wait-row{margin-bottom:8px}.wait-row.first{border-color:#9cc3ff;background:#f5f9ff}
.current-row{background:#fff8ef;border-color:#ffd9b0}
.done-row{margin-bottom:8px;background:#f8fbf9}
.queue-no{width:40px;height:40px;border-radius:50%;display:flex;align-items:center;justify-content:center;font-weight:700;background:#eef2f7;color:#4b5563}
.queue-no.current{background:#ff7a00;color:#fff}.queue-no.done{background:#86efac;color:#166534}
.info-cell label{display:block;font-size:10px;color:#9ca3af}.info-cell b{font-size:13px;color:#111827}
.complaint b{font-size:12px;color:#4b5563}
.actions{display:flex;gap:6px}
.dialog-title{font-size:15px;font-weight:700;display:flex;align-items:center;gap:6px}
.detail-wrap{padding-top:4px}
.detail-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;margin-bottom:10px}
.d-item{background:#f8fafc;border:1px solid #e5e7eb;border-radius:8px;padding:8px}
.d-item span{display:block;font-size:11px;color:#6b7280}.d-item b{font-size:13px;color:#111827}
.detail-block{margin-top:8px}
.detail-block label{display:block;font-size:12px;color:#6b7280;margin-bottom:4px}
.detail-block div{border:1px solid #e5e7eb;border-radius:8px;background:#fafafa;padding:8px;font-size:12px;word-break:break-all}
</style>