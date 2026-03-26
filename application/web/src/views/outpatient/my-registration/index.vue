<template>
  <div class="mine-page">
    <div class="hero">
      <div>
        <h2>我的预约</h2>
        <p>查看和管理您的预约记录，掌握就诊进度</p>
      </div>
      <div class="today-pill"><i class="el-icon-time" /> 今日：{{ beijingDate }}</div>
    </div>

    <el-card class="filter-card" shadow="never">
      <el-form inline>
        <el-form-item label="预约状态">
          <el-select v-model="q.status" clearable class="w180" placeholder="全部状态">
            <el-option label="全部" value="" />
            <el-option label="已预约" value="BOOKED"/>
            <el-option label="已取消" value="CANCELLED"/>
            <el-option label="已就诊" value="VISITED"/>
          </el-select>
        </el-form-item>
        <el-form-item label="就诊日期">
          <el-date-picker v-model="q.visit_date" type="date" value-format="yyyy-MM-dd" clearable class="w200" placeholder="选择日期"/>
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="load">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </el-form>
    </el-card>

    <div class="kpis">
      <div class="kpi blue"><span>全部预约</span><b>{{ list.length }}</b></div>
      <div class="kpi indigo"><span>待就诊</span><b>{{ bookedCount }}</b></div>
      <div class="kpi green"><span>已就诊</span><b>{{ visitedCount }}</b></div>
      <div class="kpi gray"><span>已取消</span><b>{{ cancelledCount }}</b></div>
    </div>

    <el-card class="table-card" shadow="never">
      <div slot="header" class="table-head">预约记录</div>
      <el-table :data="displayList" border stripe v-loading="loading" class="my-table">
        <el-table-column prop="id" label="预约编号" min-width="140"/>
        <el-table-column label="医生信息" min-width="150">
          <template slot-scope="s">
            <div class="doctor-cell">
              <div class="doctor-name">{{ s.row.doctor_name || s.row.doctor_id }}</div>
              <div class="doctor-sub">{{ s.row.hospital_name || '-' }} · {{ s.row.department_id || '-' }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="就诊时间" min-width="170">
          <template slot-scope="s">{{ s.row.visit_date || '-' }} {{ s.row.start_time || '--:--' }}-{{ s.row.end_time || '--:--' }}</template>
        </el-table-column>
        <el-table-column prop="queue_no" label="排队号" width="88"/>
        <el-table-column label="状态" width="100">
          <template slot-scope="s">
            <el-tag v-if="s.row.status==='BOOKED'" type="primary" size="mini">已预约</el-tag>
            <el-tag v-else-if="s.row.status==='VISITED'" type="success" size="mini">已就诊</el-tag>
            <el-tag v-else type="info" size="mini">已取消</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="tx_id" label="交易ID" min-width="190"/>
        <el-table-column label="操作" width="180" fixed="right">
          <template slot-scope="s">
            <el-button size="mini" @click="viewDetail(s.row)">详情</el-button>
            <el-button v-if="s.row.status==='BOOKED'" size="mini" type="danger" @click="cancel(s.row)">取消预约</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && displayList.length===0" description="暂无预约记录" :image-size="80"/>
    </el-card>

    <el-dialog :visible.sync="detailVisible" width="680px" custom-class="detail-dialog">
      <div slot="title" class="d-title"><i class="el-icon-document" /> 预约详情</div>
      <div v-if="detailItem" class="detail-grid">
        <div class="d-item"><span>预约编号</span><b>{{ detailItem.id }}</b></div>
        <div class="d-item"><span>医生</span><b>{{ detailItem.doctor_name || detailItem.doctor_id }}</b></div>
        <div class="d-item"><span>医院</span><b>{{ detailItem.hospital_name || '-' }}</b></div>
        <div class="d-item"><span>科室</span><b>{{ detailItem.department_id || '-' }}</b></div>
        <div class="d-item"><span>就诊日期</span><b>{{ detailItem.visit_date || '-' }}</b></div>
        <div class="d-item"><span>就诊时段</span><b>{{ detailItem.start_time || '--:--' }}-{{ detailItem.end_time || '--:--' }}</b></div>
        <div class="d-item"><span>排队号</span><b>{{ detailItem.queue_no || '-' }}</b></div>
        <div class="d-item"><span>状态</span><b>{{ detailItem.status }}</b></div>
        <div class="d-item"><span>交易ID</span><b class="tx">{{ detailItem.tx_id || '-' }}</b></div>
      </div>
      <span slot="footer"><el-button type="primary" @click="detailVisible=false">关闭</el-button></span>
    </el-dialog>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { queryRegistrationList, cancelRegistration, querySlotList } from '@/api/outpatient'
import { queryAccountList } from '@/api/accountV2'

export default {
  name: 'OutpatientMyRegistration',
  data() {
    return {
      loading: false,
      q: { status: '', visit_date: '' },
      list: [],
      accounts: [],
      slots: [],
      detailVisible: false,
      detailItem: null
    }
  },
  computed: {
    ...mapGetters(['account_id']),
    beijingDate() {
      const f = new Intl.DateTimeFormat('zh-CN', { timeZone: 'Asia/Shanghai', year: 'numeric', month: '2-digit', day: '2-digit' })
      const p = f.formatToParts(new Date())
      return `${p.find(i => i.type === 'year').value}-${p.find(i => i.type === 'month').value}-${p.find(i => i.type === 'day').value}`
    },
    displayList() {
      return this.list
        .filter(i => !this.q.visit_date || i.visit_date === this.q.visit_date)
        .map(i => {
          const doctor = this.accounts.find(a => a.account_id === i.doctor_id) || {}
          const slot = this.slots.find(s => s.id === i.schedule_slot_id)
          return {
            ...i,
            doctor_name: doctor.account_name,
            hospital_name: doctor.hospital_name,
            start_time: slot && slot.start_time,
            end_time: slot && slot.end_time
          }
        })
    },
    bookedCount() { return this.list.filter(i => i.status === 'BOOKED').length },
    visitedCount() { return this.list.filter(i => i.status === 'VISITED').length },
    cancelledCount() { return this.list.filter(i => i.status === 'CANCELLED').length }
  },
  created() {
    this.load()
  },
  methods: {
    async load() {
      this.loading = true
      const [regs, accounts, slots] = await Promise.all([
        queryRegistrationList({ patient_id: this.account_id, status: this.q.status }).catch(() => []),
        queryAccountList().catch(() => []),
        querySlotList({}).catch(() => [])
      ])
      this.list = regs || []
      this.accounts = accounts || []
      this.slots = slots || []
      this.loading = false
    },
    reset() {
      this.q = { status: '', visit_date: '' }
      this.load()
    },
    viewDetail(row) {
      this.detailItem = row
      this.detailVisible = true
    },
    cancel(row) {
      this.$confirm('确认取消预约？', '提示').then(() => {
        return cancelRegistration({ registration_id: row.id, operator_id: this.account_id })
      }).then(res => {
        this.$notify({ title: '取消成功', message: `交易ID：${(res.tx_id || '-').slice(0, 20)}...`, type: 'success', duration: 2300 })
        this.load()
      })
    }
  }
}
</script>

<style scoped>
.mine-page{padding:16px;background:#f3f5f9;min-height:100%}
.hero{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:12px}
.hero h2{margin:0;font-size:28px;font-weight:700;color:#111827}
.hero p{margin:4px 0 0;color:#6b7280;font-size:13px}
.today-pill{margin-top:6px;font-size:12px;color:#3b82f6;background:#eaf2ff;border:1px solid #bfd6ff;padding:6px 10px;border-radius:999px}
.filter-card{border-radius:10px;margin-bottom:12px}.w180{width:180px}.w200{width:200px}
.kpis{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin-bottom:12px}
.kpi{border-radius:10px;padding:10px 12px}.kpi span{font-size:12px;color:#6b7280}.kpi b{display:block;font-size:30px;line-height:1;margin-top:6px;font-weight:700}
.kpi.blue{background:#e8f1ff}.kpi.blue b{color:#2563eb}.kpi.indigo{background:#eef2ff}.kpi.indigo b{color:#4f46e5}.kpi.green{background:#eaf9ef}.kpi.green b{color:#16a34a}.kpi.gray{background:#f3f4f6}.kpi.gray b{color:#6b7280}
.table-card{border-radius:10px}.table-head{font-weight:700}.doctor-name{font-weight:600;color:#111827}.doctor-sub{font-size:11px;color:#6b7280}
.d-title{font-weight:700;display:flex;align-items:center;gap:6px}
.detail-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:10px}
.d-item{border:1px solid #e5e7eb;background:#fafafa;border-radius:8px;padding:8px}
.d-item span{display:block;font-size:11px;color:#6b7280}.d-item b{font-size:13px;color:#111827}.d-item .tx{word-break:break-all}
</style>