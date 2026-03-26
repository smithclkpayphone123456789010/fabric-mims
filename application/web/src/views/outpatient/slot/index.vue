<template>
  <div class="slot-page">
    <div class="title-wrap">
      <h2>号源管理</h2>
      <p>发布和管理您的出诊时间安排</p>
    </div>

    <div class="top-tabs">
      <div class="tab" :class="{ active: mode === 'publish' }" @click="mode='publish'"><i class="el-icon-plus" /> 发布号源</div>
      <div class="tab" :class="{ active: mode === 'history' }" @click="mode='history'"><i class="el-icon-time" /> 历史记录</div>
    </div>

    <template v-if="mode==='publish'">
      <div class="content-grid">
        <div>
          <el-card shadow="never" class="card card-blue">
            <div slot="header" class="card-head">
              <div class="head-title"><i class="el-icon-date" /> 基础信息</div>
              <div class="head-sub">自动读取当前医生所属医院与科室</div>
            </div>
            <div class="base-grid">
              <div class="base-item"><span>医生ID</span><b>{{ form.doctor_id || '--' }}</b></div>
              <div class="base-item"><span>所属医院</span><b>{{ form.hospital_name || '--' }}</b></div>
              <div class="base-item"><span>就诊科室</span><b>{{ form.department_id || '--' }}</b></div>
            </div>
            <div style="margin-top: 12px;">
              <el-date-picker v-model="form.visit_date" type="date" value-format="yyyy-MM-dd" placeholder="选择日期" class="full" />
            </div>
          </el-card>

          <el-card shadow="never" class="card card-green">
            <div slot="header" class="card-head row-between">
              <div>
                <div class="head-title"><i class="el-icon-alarm-clock" /> 选择出诊时间段</div>
                <div class="head-sub">选择开始时间与结束时间（30分钟粒度）</div>
              </div>
              <el-tag size="mini" type="success">{{ form.start_time && form.end_time ? '已选择' : '未选择' }}</el-tag>
            </div>
            <div class="range-wrap">
              <el-time-select v-model="form.start_time" :picker-options="{ start: '08:00', step: '00:30', end: '18:00' }" placeholder="开始时间" class="w160" />
              <span class="to">至</span>
              <el-time-select v-model="form.end_time" :picker-options="{ start: '08:30', step: '00:30', end: '18:30', minTime: form.start_time }" placeholder="结束时间" class="w160" />
            </div>
            <div class="quick-row">
              <el-button size="mini" @click="pickRange('09:00','12:00')">上午门诊</el-button>
              <el-button size="mini" @click="pickRange('14:00','17:30')">下午门诊</el-button>
              <el-button size="mini" @click="pickRange('09:00','17:30')">全天门诊</el-button>
            </div>
          </el-card>

          <el-card shadow="never" class="card card-orange">
            <div slot="header" class="card-head">
              <div class="head-title">设置接诊人数</div>
              <div class="head-sub">该时间段可接诊的最大患者数量</div>
            </div>
            <div class="cap-row">
              <el-input-number v-model="form.capacity" :min="1" :max="99" />
              <span class="unit">人</span>
              <el-button size="mini" @click="form.capacity=3">3人</el-button>
              <el-button size="mini" @click="form.capacity=5">5人</el-button>
              <el-button size="mini" @click="form.capacity=10">10人</el-button>
            </div>
          </el-card>
        </div>

        <el-card shadow="never" class="summary">
          <div class="summary-title"><i class="el-icon-odometer" /> 号源预览</div>
          <div class="summary-item"><span>出诊日期</span><b>{{ form.visit_date || '--' }}</b></div>
          <div class="summary-item"><span>所属医院</span><b>{{ form.hospital_name || '--' }}</b></div>
          <div class="summary-item"><span>就诊科室</span><b>{{ form.department_id || '--' }}</b></div>
          <div class="summary-item"><span>时间段</span><b>{{ form.start_time || '--' }} - {{ form.end_time || '--' }}</b></div>
          <div class="summary-item"><span>接诊人数</span><b>{{ form.capacity }} 人</b></div>
          <el-button type="primary" class="publish-btn" :loading="submitting" @click="publish">发布号源</el-button>
          <div class="warn">提示：发布后号源写入区块链不可篡改，请确认后提交</div>
        </el-card>
      </div>
    </template>

    <template v-else>
      <el-card class="table-card" shadow="never">
        <div slot="header" class="history-head">已发布的号源</div>
        <div v-loading="loading">
          <div v-for="group in historyGroups" :key="group.date" class="history-card">
            <div class="history-top">
              <div class="date"><i class="el-icon-date" /> {{ group.date }}</div>
              <el-tag size="mini" type="success">已发布</el-tag>
            </div>
            <div class="stat-grid">
              <div class="stat"><span>时间段</span><b>{{ group.count }}个</b></div>
              <div class="stat"><span>每段人数(最大)</span><b>{{ group.capacity }}人</b></div>
              <div class="stat"><span>总可约</span><b class="blue">{{ group.total }}人</b></div>
              <div class="stat"><span>已预约</span><b class="green">{{ group.booked }}人</b></div>
            </div>
            <div class="times">
              <el-tag v-for="t in group.times" :key="t" size="mini" effect="plain">{{ t }}</el-tag>
            </div>
            <div class="tx">TxID: {{ group.txid }}</div>
          </div>
          <el-empty v-if="!historyGroups.length" description="暂无历史记录"/>
        </div>
      </el-card>
    </template>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { createSlot, querySlotList } from '@/api/outpatient'
import { queryAccountList } from '@/api/accountV2'

export default {
  name: 'OutpatientSlot',
  data() {
    return {
      mode: 'publish',
      loading: false,
      submitting: false,
      list: [],
      doctorProfile: null,
      form: {
        doctor_id: '',
        hospital_name: '',
        department_id: '',
        visit_date: new Date().toISOString().slice(0, 10),
        start_time: '',
        end_time: '',
        capacity: 5
      }
    }
  },
  computed: {
    ...mapGetters(['account_id']),
    historyGroups() {
      const g = {}
      this.list.forEach(i => {
        const d = i.visit_date || '--'
        if (!g[d]) g[d] = { date: d, count: 0, capacity: 0, total: 0, booked: 0, times: [], txid: i.tx_id || '-' }
        g[d].count += 1
        g[d].capacity = Math.max(g[d].capacity, Number(i.capacity || 0))
        g[d].total += Number(i.capacity || 0)
        g[d].booked += Number(i.booked_count || 0)
        g[d].times.push(i.start_time)
      })
      return Object.values(g).sort((a, b) => String(b.date).localeCompare(String(a.date)))
    }
  },
  created() {
    this.form.doctor_id = this.account_id
    this.initDoctorProfile()
  },
  methods: {
    initDoctorProfile() {
      queryAccountList().then(list => {
        const me = (list || []).find(i => i.account_id === this.account_id)
        this.doctorProfile = me || null
        this.form.hospital_name = (me && me.hospital_name) || ''
        this.form.department_id = (me && me.department) || this.form.department_id || ''
      }).finally(() => {
        this.load()
      })
    },
    load() {
      this.loading = true
      querySlotList({ doctor_id: this.form.doctor_id, department_id: this.form.department_id }).then(r => { this.list = r || [] }).finally(() => { this.loading = false })
    },
    pickRange(start, end) {
      this.form.start_time = start
      this.form.end_time = end
    },
    publish() {
      if (!this.form.visit_date || !this.form.start_time || !this.form.end_time) return this.$message.warning('请先选择日期和时间段')
      if (this.form.end_time <= this.form.start_time) return this.$message.warning('结束时间需晚于开始时间')
      this.submitting = true
      createSlot(this.form).then(res => {
        this.$notify({
          title: '号源发布成功（已上链）',
          message: `交易ID: ${(res.tx_id || '-').slice(0, 18)}...`,
          type: 'success',
          duration: 2500,
          position: 'top-right'
        })
        this.mode = 'history'
        this.load()
      }).finally(() => { this.submitting = false })
    }
  }
}
</script>

<style scoped>
.slot-page{padding:16px;background:#f5f7fb;min-height:100%}
.title-wrap h2{margin:0;font-size:28px;font-weight:700;color:#111827}
.title-wrap p{margin:4px 0 12px;color:#6b7280;font-size:13px}
.top-tabs{display:flex;gap:8px;background:#eceff4;padding:4px;border-radius:12px;width:280px;margin-bottom:14px}
.tab{flex:1;height:32px;border-radius:10px;display:flex;align-items:center;justify-content:center;gap:6px;font-size:12px;color:#6b7280;cursor:pointer}
.tab.active{background:#fff;color:#111827;font-weight:600}
.content-grid{display:grid;grid-template-columns:1fr 290px;gap:12px;align-items:start}
.card{border-radius:10px;margin-bottom:12px}
.card-head{line-height:1.2}.row-between{display:flex;justify-content:space-between;align-items:center}
.head-title{font-size:14px;font-weight:700;color:#111827}.head-sub{font-size:12px;color:#6b7280;margin-top:4px}
.full{width:100%}.range-wrap{display:flex;align-items:center;gap:10px}.to{font-size:12px;color:#6b7280}.w160{width:160px}
.quick-row{margin-top:10px;display:flex;gap:8px}
.cap-row{display:flex;align-items:center;gap:8px}.unit{font-size:12px;color:#6b7280}
.base-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;background:#f8fafc;border:1px dashed #d1d5db;border-radius:8px;padding:10px}
.base-item span{display:block;font-size:11px;color:#6b7280}.base-item b{font-size:13px;color:#111827}
.summary{border-radius:10px;background:#f8fbff;border:1px solid #d8e6ff}
.summary-title{font-size:14px;font-weight:700;color:#1f2937;margin-bottom:12px}
.summary-item{display:flex;justify-content:space-between;margin-bottom:10px;font-size:12px}.summary-item span{color:#6b7280}
.publish-btn{width:100%;margin:10px 0 8px;background:#79a5ff;border-color:#79a5ff}
.warn{font-size:11px;color:#b45309;background:#fff7df;border:1px solid #f6e3a1;padding:6px 8px;border-radius:6px}
.table-card{border-radius:10px}
.history-head{font-weight:700}
.history-card{border:1px solid #e5e7eb;border-radius:10px;padding:12px;margin-bottom:10px;background:#fff}
.history-top{display:flex;justify-content:space-between;align-items:center;margin-bottom:10px}.date{font-weight:700}
.stat-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin-bottom:10px}
.stat{border:1px solid #edf0f5;background:#fafbfc;border-radius:8px;padding:8px}.stat span{display:block;font-size:11px;color:#6b7280}.stat b{font-size:18px}
.blue{color:#2563eb}.green{color:#16a34a}
.times{display:flex;gap:6px;flex-wrap:wrap;margin-bottom:6px}.tx{font-size:11px;color:#9ca3af;text-align:right}
</style>