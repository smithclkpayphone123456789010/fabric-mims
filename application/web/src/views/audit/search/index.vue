<template>
  <div class="audit-search">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-cards">
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-value">{{ stats.total_count || 0 }}</div>
            <div class="stat-label">总事件数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-value success">{{ stats.success_count || 0 }}</div>
            <div class="stat-label">成功</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-value danger">{{ stats.fail_count || 0 }}</div>
            <div class="stat-label">失败</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-value warning">{{ stats.l3_count || 0 }}</div>
            <div class="stat-label">L3级事件</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选区 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm">
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filterForm.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="yyyy-MM-dd HH:mm:ss"
            @change="handleDateChange"
          />
        </el-form-item>
        <el-form-item label="事件类型">
          <el-select v-model="filterForm.event_type" placeholder="请选择" clearable>
            <el-option label="登录" value="LOGIN" />
            <el-option label="查询病历" value="QUERY_RECORD" />
            <el-option label="创建病历" value="CREATE_RECORD" />
            <el-option label="授权操作" value="GRANT_RECORD_AUTH" />
            <el-option label="导出报表" value="EXPORT_REPORT_CREATE" />
            <el-option label="API错误" value="API_ERROR" />
            <el-option label="链码错误" value="CHAINCODE_ERROR" />
          </el-select>
        </el-form-item>
        <el-form-item label="事件级别">
          <el-select v-model="filterForm.event_level" placeholder="请选择" clearable>
            <el-option label="L1-普通" value="L1" />
            <el-option label="L2-警告" value="L2" />
            <el-option label="L3-严重" value="L3" />
          </el-select>
        </el-form-item>
        <el-form-item label="执行结果">
          <el-select v-model="filterForm.action_result" placeholder="请选择" clearable>
            <el-option label="成功" value="SUCCESS" />
            <el-option label="失败" value="FAIL" />
          </el-select>
        </el-form-item>
        <el-form-item label="患者ID">
          <el-input v-model="filterForm.target_patient_id" placeholder="请输入" clearable />
        </el-form-item>
        <el-form-item label="病历ID">
          <el-input v-model="filterForm.target_record_id" placeholder="请输入" clearable />
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="filterForm.keyword" placeholder="detail_json关键词" clearable />
        </el-form-item>
      </el-form>
      <div class="filter-actions">
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
        <el-button type="warning" @click="handleOnlyFail">仅看失败</el-button>
        <el-button type="info" @click="handleOnlyToday">仅看今日</el-button>
        <el-button type="success" @click="handleExport">导出当前结果</el-button>
      </div>
    </el-card>

    <!-- 列表区 -->
    <el-card class="table-card">
      <el-table :data="tableData" stripe border v-loading="loading">
        <el-table-column prop="event_time" label="时间" width="180" sortable />
        <el-table-column prop="event_type" label="事件类型" width="150">
          <template slot-scope="{ row }">
            <el-tag :type="getEventTypeTag(row.event_type)">{{ getEventTypeName(row.event_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="event_level" label="级别" width="80">
          <template slot-scope="{ row }">
            <el-tag :type="getLevelTag(row.event_level)" size="small">{{ row.event_level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="actor_id" label="操作人" width="120" />
        <el-table-column prop="target_patient_id" label="患者ID" width="120" show-overflow-tooltip />
        <el-table-column prop="target_record_id" label="病历ID" width="120" show-overflow-tooltip />
        <el-table-column prop="action_result" label="结果" width="80">
          <template slot-scope="{ row }">
            <el-tag :type="row.action_result === 'SUCCESS' ? 'success' : 'danger'" size="small">
              {{ row.action_result }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="tx_id" label="TxID" min-width="150" show-overflow-tooltip />
        <el-table-column prop="request_path" label="请求路径" min-width="150" show-overflow-tooltip />
        <el-table-column label="操作" width="100" fixed="right">
          <template slot-scope="{ row }">
            <el-button type="text" size="small" @click="handleDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        class="pagination"
        :current-page="pagination.page"
        :page-sizes="[10, 20, 50, 100]"
        :page-size="pagination.size"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog title="审计事件详情" :visible.sync="detailVisible" width="800px">
      <el-descriptions :column="2" border v-if="detail">
        <el-descriptions-item label="事件ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="事件类型">{{ getEventTypeName(detail.event_type) }}</el-descriptions-item>
        <el-descriptions-item label="事件级别">
          <el-tag :type="getLevelTag(detail.event_level)">{{ detail.event_level }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="执行结果">
          <el-tag :type="detail.action_result === 'SUCCESS' ? 'success' : 'danger'">{{ detail.action_result }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="事件时间">{{ detail.event_time }}</el-descriptions-item>
        <el-descriptions-item label="操作人ID">{{ detail.actor_id }}</el-descriptions-item>
        <el-descriptions-item label="目标患者ID">{{ detail.target_patient_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="目标病历ID">{{ detail.target_record_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="客户端IP">{{ detail.client_ip || '-' }}</el-descriptions-item>
        <el-descriptions-item label="请求方法">{{ detail.request_method || '-' }}</el-descriptions-item>
        <el-descriptions-item label="请求路径" :span="2">{{ detail.request_path || '-' }}</el-descriptions-item>
        <el-descriptions-item label="交易哈希" :span="2">{{ detail.tx_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="链码函数">{{ detail.chaincode_func || '-' }}</el-descriptions-item>
        <el-descriptions-item label="失败原因" :span="2">
          <span class="fail-reason">{{ detail.fail_reason || '-' }}</span>
        </el-descriptions-item>
        <el-descriptions-item label="请求追踪ID" :span="2">{{ detail.request_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="详情JSON" :span="2">
          <pre class="detail-json">{{ formatJSON(detail.detail_json) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="哈希链" :span="2">
          <div class="hash-info">
            <div>上一条: <span class="hash">{{ detail.hash_prev || 'GENESIS' }}</span></div>
            <div>当前: <span class="hash">{{ detail.hash_current }}</span></div>
          </div>
        </el-descriptions-item>
      </el-descriptions>
      <div slot="footer">
        <el-button @click="detailVisible = false">关闭</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getAuditEvents, getAuditEventDetail, getAuditEventStats } from '@/api/audit'

export default {
  name: 'AuditSearch',
  data() {
    return {
      loading: false,
      tableData: [],
      stats: {},
      filterForm: {
        dateRange: [],
        start_time: '',
        end_time: '',
        event_type: '',
        event_level: '',
        action_result: '',
        target_patient_id: '',
        target_record_id: '',
        keyword: ''
      },
      pagination: {
        page: 1,
        size: 20,
        total: 0
      },
      detailVisible: false,
      detail: null
    }
  },
  created() {
    this.loadStats()
    this.loadData()
  },
  methods: {
    loadStats() {
      getAuditEventStats().then(res => {
        // request.js 成功时已返回 res.data
        this.stats = res || {}
      }).catch(() => {})
    },
    loadData() {
      this.loading = true
      const params = {
        page: this.pagination.page,
        size: this.pagination.size
      }
      if (this.filterForm.start_time) params.start_time = this.filterForm.start_time
      if (this.filterForm.end_time) params.end_time = this.filterForm.end_time
      if (this.filterForm.event_type) params.event_type = this.filterForm.event_type
      if (this.filterForm.event_level) params.event_level = this.filterForm.event_level
      if (this.filterForm.action_result) params.action_result = this.filterForm.action_result
      if (this.filterForm.target_patient_id) params.target_patient_id = this.filterForm.target_patient_id
      if (this.filterForm.target_record_id) params.target_record_id = this.filterForm.target_record_id
      if (this.filterForm.keyword) params.keyword = this.filterForm.keyword

      getAuditEvents(params).then(res => {
        this.loading = false
        // request.js 成功时已返回 res.data
        this.tableData = res?.list || []
        this.pagination.total = res?.total || 0
      }).catch(() => {
        this.loading = false
      })
    },
    handleSearch() {
      this.pagination.page = 1
      this.loadData()
    },
    handleReset() {
      this.filterForm = {
        dateRange: [],
        start_time: '',
        end_time: '',
        event_type: '',
        event_level: '',
        action_result: '',
        target_patient_id: '',
        target_record_id: '',
        keyword: ''
      }
      this.pagination.page = 1
      this.loadData()
    },
    handleOnlyFail() {
      this.filterForm.action_result = 'FAIL'
      this.pagination.page = 1
      this.loadData()
    },
    handleOnlyToday() {
      const now = new Date()
      const start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0)
      const end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)
      this.filterForm.dateRange = [start, end]
      this.handleDateChange([start, end])
    },
    handleDateChange(val) {
      if (val && val.length === 2) {
        this.filterForm.start_time = val[0]
        this.filterForm.end_time = val[1]
      } else {
        this.filterForm.start_time = ''
        this.filterForm.end_time = ''
      }
    },
    handleExport() {
      this.$router.push({ path: '/audit/export', query: this.filterForm })
    },
    handleDetail(row) {
      getAuditEventDetail(row.id).then(res => {
        // request.js 成功时已返回 res.data
        this.detail = res
        this.detailVisible = true
      }).catch(() => {
        this.$message.error('获取详情失败')
      })
    },
    handleSizeChange(size) {
      this.pagination.size = size
      this.loadData()
    },
    handlePageChange(page) {
      this.pagination.page = page
      this.loadData()
    },
    getEventTypeTag(type) {
      const map = {
        'LOGIN': 'primary',
        'QUERY_RECORD': 'info',
        'CREATE_RECORD': 'success',
        'GRANT_RECORD_AUTH': 'warning',
        'EXPORT_REPORT_CREATE': 'warning',
        'API_ERROR': 'danger',
        'CHAINCODE_ERROR': 'danger'
      }
      return map[type] || 'info'
    },
    getEventTypeName(type) {
      const map = {
        'LOGIN': '登录',
        'QUERY_RECORD': '查询病历',
        'CREATE_RECORD': '创建病历',
        'GRANT_RECORD_AUTH': '授权操作',
        'RENEW_RECORD_AUTH': '续期授权',
        'REVOKE_RECORD_AUTH': '撤销授权',
        'CHECK_RECORD_ACCESS': '校验访问权限',
        'EXPORT_REPORT_CREATE': '创建导出任务',
        'EXPORT_REPORT_DOWNLOAD': '下载导出文件',
        'ALERT_ACK': '确认告警',
        'ALERT_RESOLVE': '关闭告警',
        'API_ERROR': 'API错误',
        'CHAINCODE_ERROR': '链码错误'
      }
      return map[type] || type
    },
    getLevelTag(level) {
      const map = { 'L1': 'info', 'L2': 'warning', 'L3': 'danger' }
      return map[level] || 'info'
    },
    formatJSON(str) {
      if (!str) return '-'
      try {
        return JSON.stringify(JSON.parse(str), null, 2)
      } catch {
        return str
      }
    }
  }
}
</script>

<style scoped>
.audit-search {
  padding: 20px;
}

.stats-cards {
  margin-bottom: 20px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.stat-value.success {
  color: #67c23a;
}

.stat-value.danger {
  color: #f56c6c;
}

.stat-value.warning {
  color: #e6a23c;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}

.filter-card {
  margin-bottom: 20px;
}

.filter-actions {
  margin-top: 16px;
  text-align: right;
}

.table-card {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

.fail-reason {
  color: #f56c6c;
}

.detail-json {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  max-height: 200px;
  overflow: auto;
  font-size: 12px;
  margin: 0;
}

.hash-info {
  font-size: 12px;
}

.hash {
  word-break: break-all;
  color: #409eff;
}
</style>
