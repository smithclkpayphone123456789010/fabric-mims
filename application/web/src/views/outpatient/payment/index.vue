<template>
  <div class="pay-page">
    <div class="head">
      <div>
        <h2>缴费管理</h2>
        <p>在线完成门诊费用支付，支持查看历史交易</p>
      </div>
      <div class="date-pill"><i class="el-icon-time" /> 今日日期：{{ beijingDate }}</div>
    </div>

    <el-card class="filter-card" shadow="never">
      <el-form inline>
        <el-form-item label="订单状态">
          <el-select v-model="q.status" clearable class="w180">
            <el-option label="全部" value=""/>
            <el-option label="待支付" value="UNPAID"/>
            <el-option label="已支付" value="PAID"/>
          </el-select>
        </el-form-item>
        <el-form-item label="订单类型">
          <el-select v-model="q.order_type" clearable class="w180">
            <el-option label="全部" value=""/>
            <el-option label="挂号费" value="REG_FEE"/>
            <el-option label="药品费" value="DRUG_FEE"/>
          </el-select>
        </el-form-item>
        <el-button type="primary" :loading="loading" @click="load">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </el-form>
    </el-card>

    <div class="kpi-grid">
      <el-card shadow="never" class="kpi"><div class="kpi-label">全部订单</div><div class="kpi-value">{{ list.length }}</div></el-card>
      <el-card shadow="never" class="kpi"><div class="kpi-label">待支付</div><div class="kpi-value orange">{{ unpaidCount }}</div></el-card>
      <el-card shadow="never" class="kpi"><div class="kpi-label">已支付</div><div class="kpi-value green">{{ paidCount }}</div></el-card>
      <el-card shadow="never" class="kpi"><div class="kpi-label">累计金额</div><div class="kpi-value blue">¥ {{ totalAmount }}</div></el-card>
    </div>

    <el-card class="table-card" shadow="never">
      <div slot="header" class="table-head">
        <span>订单列表</span>
        <el-tag size="mini" type="info">链上支付记录可追溯</el-tag>
      </div>
      <el-table :data="displayList" border stripe v-loading="loading">
        <el-table-column prop="id" label="订单号" min-width="130"/>
        <el-table-column label="类型" width="120">
          <template slot-scope="s">
            <el-tag v-if="s.row.order_type==='REG_FEE'" size="mini">挂号费</el-tag>
            <el-tag v-else size="mini" type="success">药品费</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="金额" width="100">
          <template slot-scope="s"><span class="money">¥ {{ s.row.amount }}</span></template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template slot-scope="s">
            <el-tag v-if="s.row.status==='UNPAID'" type="warning" size="mini">待支付</el-tag>
            <el-tag v-else type="success" size="mini">已支付</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_time" label="创建时间" min-width="150"/>
        <el-table-column prop="paid_time" label="支付时间" min-width="150"/>
        <el-table-column prop="tx_id" label="交易ID" min-width="200"/>
        <el-table-column label="操作" width="130" fixed="right">
          <template slot-scope="s">
            <el-button v-if="s.row.status==='UNPAID'" size="mini" type="primary" @click="openPay(s.row)">去支付</el-button>
            <el-button v-else size="mini" @click="openDetail(s.row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && displayList.length===0" description="暂无订单"/>
    </el-card>

    <el-dialog :visible.sync="payDialogVisible" width="460px" custom-class="pay-dialog" :close-on-click-modal="false">
      <div slot="title" class="dialog-title">确认支付</div>
      <div v-if="currentOrder" class="pay-body">
        <div class="row"><span>订单号</span><b>{{ currentOrder.id }}</b></div>
        <div class="row"><span>订单类型</span><b>{{ currentOrder.order_type==='REG_FEE'?'挂号费':'药品费' }}</b></div>
        <div class="row"><span>支付金额</span><b class="fee">¥ {{ currentOrder.amount }}</b></div>
        <div class="channel-title">选择支付方式</div>
        <el-radio-group v-model="payChannel" size="mini">
          <el-radio-button label="mock">余额支付</el-radio-button>
          <el-radio-button label="wechat">微信支付</el-radio-button>
          <el-radio-button label="alipay">支付宝</el-radio-button>
        </el-radio-group>
      </div>
      <span slot="footer">
        <el-button @click="payDialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="paying" @click="confirmPay">确认支付</el-button>
      </span>
    </el-dialog>

    <el-dialog :visible.sync="detailVisible" width="520px" custom-class="pay-detail-dialog">
      <div slot="title" class="dialog-title">订单详情</div>
      <div v-if="currentOrder" class="pay-body">
        <div class="row"><span>订单号</span><b>{{ currentOrder.id }}</b></div>
        <div class="row"><span>状态</span><b>{{ currentOrder.status }}</b></div>
        <div class="row"><span>交易ID</span><b class="tx">{{ currentOrder.tx_id || '-' }}</b></div>
      </div>
      <span slot="footer"><el-button type="primary" @click="detailVisible=false">关闭</el-button></span>
    </el-dialog>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { queryPaymentList, payOrder } from '@/api/outpatient'

export default {
  name: 'OutpatientPayment',
  data() {
    return {
      loading: false,
      paying: false,
      q: { status: '', order_type: '' },
      list: [],
      payDialogVisible: false,
      detailVisible: false,
      currentOrder: null,
      payChannel: 'mock'
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
      return this.list.filter(i => {
        if (this.q.status && i.status !== this.q.status) return false
        if (this.q.order_type && i.order_type !== this.q.order_type) return false
        return true
      })
    },
    unpaidCount() { return this.list.filter(i => i.status === 'UNPAID').length },
    paidCount() { return this.list.filter(i => i.status === 'PAID').length },
    totalAmount() { return this.list.reduce((s, i) => s + Number(i.amount || 0), 0).toFixed(2) }
  },
  created() { this.load() },
  methods: {
    load() {
      this.loading = true
      queryPaymentList({ patient_id: this.account_id, status: this.q.status }).then(r => {
        this.list = (r || []).sort((a, b) => String(b.created_time || '').localeCompare(String(a.created_time || '')))
      }).finally(() => { this.loading = false })
    },
    reset() {
      this.q = { status: '', order_type: '' }
      this.load()
    },
    openPay(row) {
      this.currentOrder = row
      this.payChannel = 'mock'
      this.payDialogVisible = true
    },
    openDetail(row) {
      this.currentOrder = row
      this.detailVisible = true
    },
    confirmPay() {
      if (!this.currentOrder) return
      this.paying = true
      payOrder({ payment_id: this.currentOrder.id, patient_id: this.account_id, pay_channel: this.payChannel }).then(res => {
        this.payDialogVisible = false
        this.$notify({ title: '支付成功（已上链）', message: `交易ID：${(res.tx_id || '-').slice(0, 18)}...`, type: 'success', duration: 2600, position: 'top-right' })
        this.load()
      }).finally(() => { this.paying = false })
    }
  }
}
</script>

<style scoped>
.pay-page{padding:16px;background:#f3f5f9;min-height:100%}
.head{display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:12px}
.head h2{margin:0;font-size:28px;font-weight:700;color:#111827}
.head p{margin:4px 0 0;color:#6b7280;font-size:13px}
.date-pill{margin-top:6px;font-size:12px;color:#3b82f6;background:#eaf2ff;border:1px solid #bfd6ff;padding:6px 10px;border-radius:999px}
.filter-card{margin-bottom:12px;border-radius:10px}.w180{width:180px}
.kpi-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin-bottom:12px}
.kpi{border-radius:10px}.kpi-label{font-size:12px;color:#6b7280}.kpi-value{margin-top:6px;font-size:28px;font-weight:700;color:#111827}
.kpi-value.orange{color:#ea580c}.kpi-value.green{color:#16a34a}.kpi-value.blue{color:#2563eb}
.table-card{border-radius:10px}.table-head{display:flex;justify-content:space-between;align-items:center}
.money{font-weight:700;color:#111827}
.dialog-title{font-size:16px;font-weight:700}
.pay-body .row{display:flex;justify-content:space-between;border:1px solid #edf0f5;background:#fafbfc;border-radius:8px;padding:10px;margin-bottom:8px}
.pay-body .row span{font-size:12px;color:#6b7280}.pay-body .row b{font-size:13px;color:#111827}
.pay-body .row .fee{color:#2563eb}.pay-body .row .tx{max-width:250px;word-break:break-all;text-align:right}
.channel-title{font-size:12px;color:#111827;margin:10px 0 8px}
</style>