<template>
  <div class="register-page">
    <div class="head">
      <div>
        <h2>挂号预约</h2>
        <p>请选择医院、科室和就诊时间段，完成预约挂号</p>
      </div>
      <div class="date-pill"><i class="el-icon-time" /> 今日日期：{{ beijingDate }}</div>
    </div>

    <el-card class="panel" shadow="never">
      <el-form inline>
        <el-form-item label="选择医院">
          <el-select v-model="q.hospital_name" filterable clearable placeholder="搜索医院" class="w220">
            <el-option v-for="h in hospitalOptions" :key="h" :label="h" :value="h" />
          </el-select>
        </el-form-item>
        <el-form-item label="就诊科室">
          <el-select v-model="q.department_id" filterable clearable placeholder="搜索科室" class="w220">
            <el-option v-for="d in departmentOptions" :key="d" :label="d" :value="d" />
          </el-select>
        </el-form-item>
        <el-form-item label="就诊日期">
          <el-date-picker v-model="q.visit_date" type="date" value-format="yyyy-MM-dd" class="w220"/>
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="load">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </el-form>
    </el-card>

    <el-row :gutter="14" v-loading="loading">
      <el-col v-for="item in doctorCards" :key="item.doctor_id" :xs="24" :sm="12" :md="8">
        <el-card class="doctor-card" shadow="hover">
          <div class="card-head">
            <div class="avatar">医</div>
            <div>
              <div class="name">{{ item.doctor_name }}</div>
              <div class="sub">{{ item.department }} · {{ item.hospital_name || '门诊部' }}</div>
            </div>
          </div>
          <div class="section-title">可预约时间段</div>
          <div class="time-grid">
            <div
              v-for="s in item.slots"
              :key="s.id"
              class="time-item"
              :class="{ disabled: remains(s)<=0 }"
              @click="openBookDialog(s, item)"
            >
              <div>{{ s.start_time }} - {{ s.end_time }}</div>
              <small>余 {{ remains(s) }}</small>
            </div>
          </div>
          <el-button type="primary" class="quick-book" size="mini" :disabled="!item.slots.length" @click="openBookDialog(item.slots[0], item)">立即挂号</el-button>
        </el-card>
      </el-col>
    </el-row>
    <el-empty v-if="!loading && doctorCards.length===0" description="暂无可预约号源"/>

    <el-dialog :visible.sync="bookDialogVisible" width="460px" custom-class="book-dialog" :close-on-click-modal="false">
      <div slot="title" class="dialog-title">确认挂号信息</div>
      <div class="dialog-sub">请您核实挂号时间并确认挂号信息</div>
      <div v-if="selectedSlot" class="dialog-body">
        <div class="card-line"><span>医生姓名</span><b>{{ selectedDoctor.doctor_name }}</b></div>
        <div class="card-line"><span>就诊医院</span><b>{{ selectedDoctor.hospital_name || '-' }}</b></div>
        <div class="card-line"><span>就诊科室</span><b>{{ selectedDoctor.department }}</b></div>
        <div class="card-line"><span>就诊日期</span><b>{{ selectedSlot.visit_date }}</b></div>
        <div class="card-line"><span>就诊时间段</span><b>{{ selectedSlot.start_time }} - {{ selectedSlot.end_time }}</b></div>
        <div class="card-line"><span>挂号费用</span><b class="fee">¥50</b></div>
      </div>
      <div class="dialog-tip"><i class="el-icon-info" /> 提交后将写入区块链，不可篡改</div>
      <span slot="footer" class="dialog-footer">
        <el-button @click="bookDialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="confirmBook">确认挂号</el-button>
      </span>
    </el-dialog>

    <el-dialog :visible.sync="successDialogVisible" width="420px" custom-class="success-dialog" :show-close="false" :close-on-click-modal="false">
      <div class="success-wrap">
        <div class="ok-top-icon"><span>✓</span></div>
        <div class="ok-title">挂号成功</div>
        <div class="ok-sub">您的预约信息已确认，请按时就诊</div>

        <div class="ok-card">
          <div class="ok-row"><span>就诊日期</span><b>{{ successInfo.date || '-' }}</b></div>
          <div class="ok-row"><span>就诊时段</span><b>{{ successInfo.timeRange || '-' }}</b></div>
          <div class="ok-row"><span>交易ID</span><b class="tx">{{ successInfo.txid || '-' }}</b></div>
        </div>
      </div>
      <span slot="footer" class="success-footer">
        <el-button type="primary" class="success-confirm" @click="successDialogVisible=false">完成</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { querySlotList, createRegistration } from '@/api/outpatient'
import { queryAccountList } from '@/api/accountV2'

export default {
  name: 'OutpatientRegister',
  data() {
    return {
      loading: false,
      submitting: false,
      q: { hospital_name: '', department_id: '', visit_date: '' },
      slots: [],
      accounts: [],
      bookDialogVisible: false,
      successDialogVisible: false,
      selectedSlot: null,
      selectedDoctor: {},
      successInfo: { date: '', timeRange: '', txid: '' }
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
      return Array.from(new Set(this.accounts.filter(a => a.role === 'doctor' && a.hospital_name).map(a => a.hospital_name)))
    },
    departmentOptions() {
      const list = this.accounts.filter(a => a.role === 'doctor' && (!this.q.hospital_name || a.hospital_name === this.q.hospital_name))
      return Array.from(new Set(list.map(a => a.department).filter(Boolean)))
    },
    doctorCards() {
      const map = {}
      this.slots.forEach(s => {
        const did = s.doctor_id || 'unknown'
        const acc = this.accounts.find(a => a.account_id === did) || {}
        if (this.q.hospital_name && acc.hospital_name !== this.q.hospital_name) return
        if (this.q.department_id && s.department_id !== this.q.department_id) return
        if (!map[did]) {
          map[did] = {
            doctor_id: did,
            doctor_name: acc.account_name || `医生 ${did}`,
            department: s.department_id || acc.department || '-',
            hospital_name: acc.hospital_name || '',
            slots: []
          }
        }
        map[did].slots.push(s)
      })
      return Object.values(map)
    }
  },
  created() {
    this.q.visit_date = this.beijingDate
    Promise.all([queryAccountList().catch(() => []), this.load()]).then(([a]) => {
      this.accounts = a || []
    })
  },
  methods: {
    remains(s) { return Math.max(0, Number(s.capacity || 0) - Number(s.booked_count || 0)) },
    load() {
      this.loading = true
      return querySlotList({ visit_date: this.q.visit_date, department_id: this.q.department_id }).then(r => {
        this.slots = (r || []).sort((a, b) => `${a.visit_date} ${a.start_time}`.localeCompare(`${b.visit_date} ${b.start_time}`))
      }).finally(() => { this.loading = false })
    },
    reset() {
      this.q = { hospital_name: '', department_id: '', visit_date: this.beijingDate }
      this.load()
    },
    openBookDialog(slot, doctor) {
      if (!slot || this.remains(slot) <= 0) return
      this.selectedSlot = slot
      this.selectedDoctor = doctor
      this.bookDialogVisible = true
    },
    confirmBook() {
      if (!this.selectedSlot) return
      this.submitting = true
      createRegistration({
        patient_id: this.account_id,
        doctor_id: this.selectedSlot.doctor_id,
        department_id: this.selectedSlot.department_id,
        slot_id: this.selectedSlot.id,
        visit_date: this.selectedSlot.visit_date
      }).then(res => {
        this.bookDialogVisible = false
        this.successInfo = {
          date: this.selectedSlot.visit_date,
          timeRange: `${this.selectedSlot.start_time} - ${this.selectedSlot.end_time}`,
          txid: res.tx_id || '-'
        }
        this.successDialogVisible = true
        this.load()
      }).finally(() => { this.submitting = false })
    }
  }
}
</script>

<style scoped>
.register-page{padding:16px;background:#f3f5f9;min-height:100%}
.head{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:12px}
.head h2{margin:0;font-size:28px;font-weight:700;color:#111827}
.head p{margin:4px 0 0;color:#6b7280;font-size:13px}
.date-pill{margin-top:6px;font-size:12px;color:#3b82f6;background:#eaf2ff;border:1px solid #bfd6ff;padding:6px 10px;border-radius:999px}
.panel{margin-bottom:12px;border-radius:10px}.w220{width:220px}
.doctor-card{border-radius:12px;margin-bottom:12px}.card-head{display:flex;gap:10px;align-items:center}
.avatar{width:36px;height:36px;border-radius:50%;background:#e8f0ff;color:#1d4ed8;font-weight:700;display:flex;align-items:center;justify-content:center}
.name{font-size:15px;font-weight:700;color:#111827}.sub{font-size:12px;color:#6b7280;margin-top:2px}
.section-title{margin:12px 0 8px;font-size:12px;color:#374151;font-weight:600}
.time-grid{display:grid;grid-template-columns:repeat(2,1fr);gap:8px}
.time-item{border:1px solid #dbe4f3;background:#f8fbff;border-radius:8px;padding:8px;cursor:pointer}
.time-item small{display:block;color:#6b7280;margin-top:4px}.time-item.disabled{opacity:.5;cursor:not-allowed}
.quick-book{margin-top:10px;width:100%}
.dialog-title{font-size:16px;font-weight:700}
.dialog-sub{font-size:12px;color:#6b7280;margin-bottom:10px}
.card-line{display:flex;justify-content:space-between;align-items:center;border:1px solid #edf0f5;border-radius:8px;padding:10px 12px;background:#fafbfc;margin-bottom:8px}
.card-line span{color:#6b7280;font-size:12px}.card-line b{font-size:13px;color:#111827}.fee{color:#2563eb}
.dialog-tip{margin-top:8px;background:#fff7e6;border:1px solid #ffe2b2;color:#b45309;border-radius:8px;padding:8px 10px;font-size:12px}
.success-wrap{text-align:center;padding:6px 0 0}
.ok-top-icon{width:64px;height:64px;margin:2px auto 10px;border-radius:50%;background:#111827;display:flex;align-items:center;justify-content:center}
.ok-top-icon span{color:#fff;font-size:28px;line-height:1}
.ok-title{font-size:20px;font-weight:700;color:#111827;line-height:1.2}
.ok-sub{font-size:12px;color:#6b7280;margin-top:8px;line-height:1.6}
.ok-card{margin-top:12px;background:#f8fafc;border:1px solid #e5e7eb;border-radius:10px;padding:10px}
.ok-row{display:flex;justify-content:space-between;align-items:flex-start;padding:7px 2px;border-bottom:1px dashed #e5e7eb;font-size:12px}
.ok-row:last-child{border-bottom:none}
.ok-row span{color:#6b7280}
.ok-row b{color:#111827;font-weight:600}
.ok-row .tx{max-width:210px;text-align:right;word-break:break-all}
.success-footer{display:block;text-align:center;padding-top:2px}
.success-confirm{width:90%;height:36px;border-radius:8px;font-weight:600}
</style>