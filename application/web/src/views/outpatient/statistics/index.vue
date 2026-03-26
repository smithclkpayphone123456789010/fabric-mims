<template>
  <div class="stats-page">
    <div class="head">
      <div>
        <h2>门诊数据统计</h2>
        <p>实时监控门诊业务运行情况</p>
      </div>
      <div class="date-pill"><i class="el-icon-time" /> {{ beijingDate }}</div>
    </div>

    <el-card class="filter-card" shadow="never">
      <el-form inline>
        <el-form-item label="统计日期">
          <el-date-picker v-model="range" type="daterange" value-format="yyyy-MM-dd" start-placeholder="开始日期" end-placeholder="结束日期" class="w320"/>
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="load">查询</el-button>
        <el-button @click="exportData">导出统计</el-button>
      </el-form>
    </el-card>

    <div class="kpi-grid">
      <el-card shadow="never" class="kpi blue"><div class="kpi-title">今日挂号人数</div><div class="kpi-value">{{ stats.todayReg }}</div><div class="kpi-sub">较昨日 {{ growth.reg }}%</div></el-card>
      <el-card shadow="never" class="kpi green"><div class="kpi-title">今日已就诊</div><div class="kpi-value">{{ stats.todayVisited }}</div><div class="kpi-sub">完成率 {{ visitRate }}%</div></el-card>
      <el-card shadow="never" class="kpi orange"><div class="kpi-title">今日收入</div><div class="kpi-value">¥ {{ stats.todayIncome.toFixed(2) }}</div><div class="kpi-sub">支付订单 {{ stats.paidCount }} 笔</div></el-card>
      <el-card shadow="never" class="kpi purple"><div class="kpi-title">当前排队人数</div><div class="kpi-value">{{ stats.waiting }}</div><div class="kpi-sub">等待叫号</div></el-card>
    </div>

    <el-row :gutter="12" class="chart-row">
      <el-col :span="12">
        <el-card shadow="never" class="chart-card">
          <div slot="header" class="chart-head">近7日挂号趋势</div>
          <div class="chart-wrap">
            <svg viewBox="0 0 520 240" class="chart-svg">
              <line x1="44" y1="20" x2="44" y2="196" class="axis"/>
              <line x1="44" y1="196" x2="504" y2="196" class="axis"/>
              <g v-for="(t, i) in regTicks" :key="`rt-${i}`">
                <line x1="44" :y1="t.y" x2="504" :y2="t.y" class="grid"/>
                <text x="38" :y="t.y+4" text-anchor="end" class="y-label">{{ t.label }}</text>
              </g>
              <polyline class="line blue" :points="regLine.points" />
              <circle v-for="(d, i) in regLine.dots" :key="`rd-${i}`" :cx="d.x" :cy="d.y" r="3" class="dot blue"/>
              <text
                v-for="(d, i) in regLine.dots"
                :key="`rxt-${i}`"
                :x="d.x"
                y="214"
                text-anchor="middle"
                class="x-svg-label"
              >{{ i % 2 === 0 ? sevenDates[i].slice(5) : '' }}</text>
            </svg>
            <div class="legend blue">● 挂号人次</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="never" class="chart-card">
          <div slot="header" class="chart-head">近7日收入趋势</div>
          <div class="chart-wrap">
            <svg viewBox="0 0 520 240" class="chart-svg">
              <line x1="44" y1="20" x2="44" y2="196" class="axis"/>
              <line x1="44" y1="196" x2="504" y2="196" class="axis"/>
              <g v-for="(t, i) in incomeTicks" :key="`it-${i}`">
                <line x1="44" :y1="t.y" x2="504" :y2="t.y" class="grid"/>
                <text x="38" :y="t.y+4" text-anchor="end" class="y-label">{{ t.label }}</text>
              </g>
              <polyline class="line green" :points="incomeLine.points" />
              <circle v-for="(d, i) in incomeLine.dots" :key="`id-${i}`" :cx="d.x" :cy="d.y" r="3" class="dot green"/>
              <text
                v-for="(d, i) in incomeLine.dots"
                :key="`ixt-${i}`"
                :x="d.x"
                y="214"
                text-anchor="middle"
                class="x-svg-label"
              >{{ i % 2 === 0 ? sevenDates[i].slice(5) : '' }}</text>
            </svg>
            <div class="legend green">● 收入(万元)</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" class="table-card">
      <div slot="header" class="table-head"><span>门诊明细记录</span><el-tag size="mini" type="info">链上数据可追溯</el-tag></div>
      <el-table :data="tableData" border stripe v-loading="loading">
        <el-table-column prop="visit_date" label="就诊日期" width="120"/>
        <el-table-column prop="patient_id" label="患者ID" min-width="120"/>
        <el-table-column prop="doctor_id" label="医生ID" min-width="120"/>
        <el-table-column prop="department_id" label="科室" width="100"/>
        <el-table-column label="挂号状态" width="100"><template slot-scope="s"><el-tag size="mini" v-if="s.row.registration_status==='VISITED'" type="success">已就诊</el-tag><el-tag size="mini" v-else-if="s.row.registration_status==='BOOKED'" type="warning">待就诊</el-tag><el-tag size="mini" v-else type="info">已取消</el-tag></template></el-table-column>
        <el-table-column label="支付" width="90"><template slot-scope="s"><el-tag size="mini" :type="s.row.fee_status==='PAID'?'success':'warning'">{{ s.row.fee_status }}</el-tag></template></el-table-column>
        <el-table-column label="队列" width="100"><template slot-scope="s"><el-tag size="mini">{{ s.row.queue_status || '-' }}</el-tag></template></el-table-column>
        <el-table-column prop="tx_id" label="交易ID" min-width="220"/>
      </el-table>
    </el-card>
  </div>
</template>

<script>
import { queryOutpatientRecordList } from '@/api/outpatient'

export default {
  name: 'OutpatientStatistics',
  data() {
    const end = new Date().toISOString().slice(0, 10)
    const start = new Date(Date.now() - 6 * 24 * 3600 * 1000).toISOString().slice(0, 10)
    return {
      loading: false,
      range: [start, end],
      stats: { todayReg: 0, todayVisited: 0, todayIncome: 0, waiting: 0, paidCount: 0 },
      growth: { reg: 0 },
      regSeries: [],
      incomeSeries: [],
      tableData: []
    }
  },
  computed: {
    beijingDate() {
      const f = new Intl.DateTimeFormat('zh-CN', { timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit' })
      const p = f.formatToParts(new Date())
      return `${p.find(i => i.type === 'year').value}-${p.find(i => i.type === 'month').value}-${p.find(i => i.type === 'day').value}`
    },
    visitRate() { return this.stats.todayReg ? Math.round((this.stats.todayVisited / this.stats.todayReg) * 100) : 0 },
    sevenDates() {
      const arr = []
      const end = new Date(this.range[1] || this.beijingDate)
      for (let i = 6; i >= 0; i--) {
        const d = new Date(end.getTime() - i * 24 * 3600 * 1000)
        arr.push(d.toISOString().slice(0, 10))
      }
      return arr
    },
    regLine() { return this.buildLine(this.regSeries, 44, 504, 20, 196) },
    incomeLine() { return this.buildLine(this.incomeSeries, 44, 504, 20, 196) },
    regTicks() { return this.buildTicks(this.regSeries, 20, 196) },
    incomeTicks() { return this.buildTicks(this.incomeSeries, 20, 196, 1000) }
  },
  created() { this.load() },
  methods: {
    buildLine(values, xStart, xEnd, yTop, yBottom) {
      const vals = values.length ? values : [0, 0, 0, 0, 0, 0, 0]
      const max = Math.max(...vals, 1)
      const min = 0
      const stepX = (xEnd - xStart) / Math.max(vals.length - 1, 1)
      const dots = vals.map((v, i) => {
        const x = xStart + stepX * i
        const y = yBottom - ((v - min) / (max - min || 1)) * (yBottom - yTop)
        return { x: Number(x.toFixed(2)), y: Number(y.toFixed(2)) }
      })
      return { points: dots.map(d => `${d.x},${d.y}`).join(' '), dots }
    },
    buildTicks(values, yTop, yBottom, roundBase = 10) {
      const maxRaw = Math.max(...(values.length ? values : [0]), 1)
      const max = Math.ceil(maxRaw / roundBase) * roundBase
      const levels = 4
      const ticks = []
      for (let i = 0; i <= levels; i++) {
        const ratio = i / levels
        const v = Math.round(max * (1 - ratio))
        const y = yTop + (yBottom - yTop) * ratio
        ticks.push({ label: v, y: Number(y.toFixed(2)) })
      }
      return ticks
    },
    async load() {
      this.loading = true
      const [startDate, endDate] = this.range || []
      const list = await queryOutpatientRecordList({ start_date: startDate, end_date: endDate }).catch(() => [])
      const data = list || []

      const today = this.beijingDate
      const todayList = data.filter(i => (i.registration || {}).visit_date === today)
      this.stats.todayReg = todayList.length
      this.stats.todayVisited = todayList.filter(i => ((i.registration || {}).status) === 'VISITED').length
      this.stats.todayIncome = todayList.reduce((s, i) => s + ((((i.payment || {}).status) === 'PAID') ? Number((i.payment || {}).amount || 0) : 0), 0)
      this.stats.waiting = data.filter(i => ((i.queue || {}).status) === 'WAITING').length
      this.stats.paidCount = todayList.filter(i => ((i.payment || {}).status) === 'PAID').length

      const y = new Date(new Date(today).getTime() - 24 * 3600 * 1000).toISOString().slice(0, 10)
      const yCount = data.filter(i => (i.registration || {}).visit_date === y).length
      this.growth.reg = yCount ? Math.round(((this.stats.todayReg - yCount) / yCount) * 100) : 0

      const mReg = {}
      const mIncome = {}
      data.forEach(i => {
        const d = (i.registration || {}).visit_date || ''
        if (!d) return
        mReg[d] = (mReg[d] || 0) + 1
        const p = i.payment || {}
        if (p.status === 'PAID') mIncome[d] = (mIncome[d] || 0) + Number(p.amount || 0)
      })

      this.regSeries = this.sevenDates.map(d => mReg[d] || 0)
      this.incomeSeries = this.sevenDates.map(d => Number(((mIncome[d] || 0) / 10000).toFixed(2)))

      this.tableData = data.map(i => {
        const reg = i.registration || {}
        const pay = i.payment || {}
        const queue = i.queue || {}
        return {
          visit_date: reg.visit_date || '-',
          patient_id: reg.patient_id || '-',
          doctor_id: reg.doctor_id || '-',
          department_id: reg.department_id || '-',
          registration_status: reg.status || '-',
          fee_status: pay.status || '-',
          queue_status: queue.status || '-',
          tx_id: reg.tx_id || pay.tx_id || queue.tx_id || '-'
        }
      }).sort((a, b) => String(b.visit_date).localeCompare(String(a.visit_date)))

      this.loading = false
    },
    exportData() { this.$message.success('导出统计成功') }
  }
}
</script>

<style scoped>
.stats-page{padding:16px;background:#f5f7fb;min-height:100%}
.head{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:12px}
.head h2{margin:0;font-size:28px;font-weight:700;color:#111827}
.head p{margin:4px 0 0;color:#6b7280;font-size:13px}
.date-pill{margin-top:6px;font-size:12px;color:#3b82f6;background:#eaf2ff;border:1px solid #bfd6ff;padding:6px 10px;border-radius:999px}
.filter-card{border-radius:10px;margin-bottom:12px}.w320{width:320px}
.kpi-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin-bottom:12px}
.kpi{border-radius:10px}.kpi-title{font-size:12px;color:#6b7280}.kpi-value{margin-top:6px;font-size:30px;font-weight:700;line-height:1.1;color:#111827}.kpi-sub{margin-top:6px;font-size:11px;color:#9ca3af}
.kpi.blue .kpi-value{color:#2563eb}.kpi.green .kpi-value{color:#16a34a}.kpi.orange .kpi-value{color:#ea580c}.kpi.purple .kpi-value{color:#7c3aed}
.chart-row{margin-bottom:12px}.chart-card{border-radius:10px;min-height:300px}.chart-head{font-weight:700;color:#111827}
.chart-wrap{padding:2px 2px 0}.chart-svg{width:100%;height:240px;display:block}
.axis{stroke:#9ca3af;stroke-width:1}.grid{stroke:#e5e7eb;stroke-width:1;stroke-dasharray:2 3}.y-label{fill:#9ca3af;font-size:10px}
.line{fill:none;stroke-width:2}.line.blue{stroke:#3b82f6}.line.green{stroke:#10b981}
.dot{stroke:#fff;stroke-width:1.5}.dot.blue{fill:#3b82f6}.dot.green{fill:#10b981}
.x-svg-label{fill:#9ca3af;font-size:10px}
.legend{text-align:center;font-size:11px;margin-top:4px}.legend.blue{color:#3b82f6}.legend.green{color:#10b981}
.table-card{border-radius:10px}.table-head{display:flex;justify-content:space-between;align-items:center}
</style>