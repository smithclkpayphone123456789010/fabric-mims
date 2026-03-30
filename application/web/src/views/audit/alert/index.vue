<template>
  <div class="audit-alert">
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-cards">
      <el-col :span="8">
        <el-card>
          <div class="stat-item">
            <div class="stat-value">{{ stats.today_count || 0 }}</div>
            <div class="stat-label">今日告警</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <div class="stat-item">
            <div class="stat-value warning">{{ stats.unresolved || 0 }}</div>
            <div class="stat-label">未处理</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <div class="stat-item">
            <div class="stat-value danger">{{ stats.high_level_count || 0 }}</div>
            <div class="stat-label">高危告警</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选区 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filterForm">
        <el-form-item label="告警级别">
          <el-select v-model="filterForm.level" placeholder="请选择" clearable>
            <el-option label="高危" value="HIGH" />
            <el-option label="中危" value="MEDIUM" />
            <el-option label="低危" value="LOW" />
          </el-select>
        </el-form-item>
        <el-form-item label="告警状态">
          <el-select v-model="filterForm.status" placeholder="请选择" clearable>
            <el-option label="新建" value="NEW" />
            <el-option label="已确认" value="ACKED" />
            <el-option label="已关闭" value="RESOLVED" />
          </el-select>
        </el-form-item>
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
      </el-form>
      <div class="filter-actions">
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>
    </el-card>

    <!-- 告警列表 -->
    <el-card class="table-card">
      <el-table :data="tableData" stripe border v-loading="loading">
        <el-table-column prop="trigger_time" label="触发时间" width="180" sortable />
        <el-table-column prop="rule_code" label="规则编码" width="180">
          <template slot-scope="{ row }">
            <el-tooltip :content="getRuleName(row.rule_code)" placement="top">
              <el-tag>{{ row.rule_code }}</el-tag>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column prop="level" label="级别" width="80">
          <template slot-scope="{ row }">
            <el-tag :type="getLevelTag(row.level)" size="small">{{ row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template slot-scope="{ row }">
            <el-tag :type="getStatusTag(row.status)">{{ getStatusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="actor_id" label="操作人" width="120" />
        <el-table-column prop="target_patient_id" label="患者ID" width="120" show-overflow-tooltip />
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template slot-scope="{ row }">
            <el-button type="text" size="small" @click="handleDetail(row)">详情</el-button>
            <el-button
              v-if="row.status === 'NEW'"
              type="success"
              size="small"
              @click="handleAck(row)"
            >
              确认
            </el-button>
            <el-button
              v-if="row.status !== 'RESOLVED'"
              type="warning"
              size="small"
              @click="handleResolve(row)"
            >
              关闭
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        class="pagination"
        :current-page="pagination.page"
        :page-size="pagination.size"
        :total="pagination.total"
        layout="total, prev, pager, next"
        @current-change="handlePageChange"
      />
    </el-card>

    <!-- 详情弹窗 -->
    <el-dialog title="告警详情" :visible.sync="detailVisible" width="700px">
      <el-descriptions :column="2" border v-if="detail">
        <el-descriptions-item label="告警ID" :span="2">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="规则编码">{{ detail.rule_code }}</el-descriptions-item>
        <el-descriptions-item label="规则名称">{{ getRuleName(detail.rule_code) }}</el-descriptions-item>
        <el-descriptions-item label="级别">
          <el-tag :type="getLevelTag(detail.level)">{{ detail.level }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusTag(detail.status)">{{ getStatusName(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="触发时间">{{ detail.trigger_time }}</el-descriptions-item>
        <el-descriptions-item label="关联事件ID">
          <el-button type="text" @click="handleViewEvent(detail.event_id)">
            {{ detail.event_id }}
          </el-button>
        </el-descriptions-item>
        <el-descriptions-item label="操作人">{{ detail.actor_id }}</el-descriptions-item>
        <el-descriptions-item label="患者ID">{{ detail.target_patient_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="病历ID">{{ detail.target_record_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="描述" :span="2">{{ detail.description }}</el-descriptions-item>
        <el-descriptions-item label="处理时间">{{ detail.handle_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="处理备注" :span="2">{{ detail.handle_note || '-' }}</el-descriptions-item>
      </el-descriptions>
      <div slot="footer">
        <el-button @click="detailVisible = false">关闭</el-button>
      </div>
    </el-dialog>

    <!-- 关闭告警弹窗 -->
    <el-dialog title="关闭告警" :visible.sync="resolveVisible" width="500px">
      <el-form :model="resolveForm" :rules="resolveRules" ref="resolveForm">
        <el-form-item label="处理备注" prop="handleNote">
          <el-input
            v-model="resolveForm.handleNote"
            type="textarea"
            :rows="4"
            placeholder="请输入处理备注（必填）"
          />
        </el-form-item>
      </el-form>
      <div slot="footer">
        <el-button @click="resolveVisible = false">取消</el-button>
        <el-button type="primary" @click="handleResolveSubmit">确定关闭</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { getAuditAlerts, getAuditAlertDetail, getAuditAlertStats, ackAuditAlert, resolveAuditAlert } from '@/api/audit'

export default {
  name: 'AuditAlert',
  data() {
    return {
      loading: false,
      tableData: [],
      stats: {},
      filterForm: {
        level: '',
        status: '',
        start_time: '',
        end_time: '',
        dateRange: []
      },
      pagination: {
        page: 1,
        size: 20,
        total: 0
      },
      detailVisible: false,
      detail: null,
      resolveVisible: false,
      resolveForm: {
        handleNote: ''
      },
      resolveRules: {
        handleNote: [
          { required: true, message: '处理备注不能为空', trigger: 'blur' }
        ]
      },
      currentAlert: null
    }
  },
  created() {
    this.loadStats()
    this.loadData()
  },
  methods: {
    loadStats() {
      getAuditAlertStats().then(res => {
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
      if (this.filterForm.level) params.level = this.filterForm.level
      if (this.filterForm.status) params.status = this.filterForm.status
      if (this.filterForm.start_time) params.start_time = this.filterForm.start_time
      if (this.filterForm.end_time) params.end_time = this.filterForm.end_time

      getAuditAlerts(params).then(res => {
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
        level: '',
        status: '',
        start_time: '',
        end_time: '',
        dateRange: []
      }
      this.pagination.page = 1
      this.loadData()
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
    handlePageChange(page) {
      this.pagination.page = page
      this.loadData()
    },
    handleDetail(row) {
      getAuditAlertDetail(row.id).then(res => {
        // request.js 成功时已返回 res.data
        this.detail = res
        this.detailVisible = true
      }).catch(() => {
        this.$message.error('获取详情失败')
      })
    },
    handleViewEvent(eventId) {
      this.$router.push({ path: '/audit/search', query: { event_id: eventId } })
    },
    handleAck(row) {
      this.$confirm('确认要处理此告警吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        ackAuditAlert(row.id).then(() => {
          this.$message.success('确认成功')
          this.loadData()
          this.loadStats()
        }).catch(err => {
          this.$message.error(err.message || '确认失败')
        })
      }).catch(() => {})
    },
    handleResolve(row) {
      this.currentAlert = row
      this.resolveForm.handleNote = ''
      this.resolveVisible = true
    },
    handleResolveSubmit() {
      this.$refs.resolveForm.validate(valid => {
        if (valid) {
          resolveAuditAlert(this.currentAlert.id, { handle_note: this.resolveForm.handleNote }).then(() => {
            this.$message.success('关闭成功')
            this.resolveVisible = false
            this.loadData()
            this.loadStats()
          }).catch(err => {
            this.$message.error(err.message || '关闭失败')
          })
        }
      })
    },
    getLevelTag(level) {
      const map = { 'HIGH': 'danger', 'MEDIUM': 'warning', 'LOW': 'info' }
      return map[level] || 'info'
    },
    getStatusTag(status) {
      const map = { 'NEW': 'danger', 'ACKED': 'warning', 'RESOLVED': 'success' }
      return map[status] || 'info'
    },
    getStatusName(status) {
      const map = { 'NEW': '新建', 'ACKED': '已确认', 'RESOLVED': '已关闭' }
      return map[status] || status
    },
    getRuleName(code) {
      const map = {
        'UNAUTHORIZED_ACCESS': '未授权访问',
        'HIGH_FREQ_ACCESS': '高频访问',
        'CONTINUOUS_FAIL': '连续失败',
        'PRIVILEGE_ESCALATION': '权限提升',
        'BULK_EXPORT': '批量导出'
      }
      return map[code] || code
    }
  }
}
</script>

<style scoped>
.audit-alert {
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

.stat-value.warning {
  color: #e6a23c;
}

.stat-value.danger {
  color: #f56c6c;
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
</style>
