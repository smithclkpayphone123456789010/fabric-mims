<template>
  <div class="audit-export">
    <!-- 创建导出区 -->
    <el-card class="create-card">
      <div slot="header">
        <span>创建导出任务</span>
      </div>
      <el-form :model="exportForm" label-width="120px">
        <el-form-item label="筛选条件">
          <div class="filter-preview">
            <el-tag v-if="hasFilters" type="info">已设置筛选条件</el-tag>
            <el-tag v-else>默认全部</el-tag>
            <el-button type="text" @click="showFilterDialog = true">查看/修改</el-button>
          </div>
        </el-form-item>
        <el-form-item label="导出格式">
          <el-radio-group v-model="exportForm.format">
            <el-radio label="csv">CSV</el-radio>
            <el-radio label="xlsx">XLSX</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="数据脱敏">
          <el-switch v-model="exportForm.mask_sensitive" />
          <span class="hint">开启后将隐藏患者姓名、身份证号等敏感信息</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="creating" @click="handleCreateExport">
            创建导出任务
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 任务列表 -->
    <el-card class="task-card">
      <div slot="header">
        <span>导出任务列表</span>
        <el-button style="float: right;" type="text" @click="handleRefresh">刷新</el-button>
      </div>
      <el-table :data="taskList" stripe border v-loading="loading">
        <el-table-column prop="id" label="任务ID" width="150" show-overflow-tooltip />
        <el-table-column prop="create_time" label="创建时间" width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template slot-scope="{ row }">
            <el-tag :type="getStatusTag(row.status)">{{ getStatusName(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="format" label="格式" width="80" />
        <el-table-column prop="file_name" label="文件名" min-width="200" show-overflow-tooltip />
        <el-table-column prop="finish_time" label="完成时间" width="180" />
        <el-table-column prop="fail_reason" label="失败原因" min-width="150" show-overflow-tooltip />
        <el-table-column label="操作" width="180" fixed="right">
          <template slot-scope="{ row }">
            <el-button
              v-if="row.status === 'SUCCESS'"
              type="success"
              size="small"
              @click="handleDownload(row)"
            >
              下载
            </el-button>
            <el-button type="text" size="small" @click="handleViewDetail(row)">详情</el-button>
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

    <!-- 筛选条件弹窗 -->
    <el-dialog title="筛选条件" :visible.sync="showFilterDialog" width="600px">
      <el-form :inline="true" :model="filterForm">
        <el-form-item label="时间范围">
          <el-date-picker
            v-model="filterForm.dateRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            value-format="yyyy-MM-dd HH:mm:ss"
          />
        </el-form-item>
        <el-form-item label="事件类型">
          <el-select v-model="filterForm.event_type" placeholder="请选择" clearable>
            <el-option label="登录" value="LOGIN" />
            <el-option label="查询病历" value="QUERY_RECORD" />
            <el-option label="创建病历" value="CREATE_RECORD" />
            <el-option label="授权操作" value="GRANT_RECORD_AUTH" />
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
      </el-form>
      <div slot="footer">
        <el-button @click="showFilterDialog = false">取消</el-button>
        <el-button type="primary" @click="showFilterDialog = false">确定</el-button>
      </div>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog title="导出任务详情" :visible.sync="detailVisible" width="600px">
      <el-descriptions :column="1" border v-if="detail">
        <el-descriptions-item label="任务ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusTag(detail.status)">{{ getStatusName(detail.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="导出格式">{{ detail.format }}</el-descriptions-item>
        <el-descriptions-item label="文件名">{{ detail.file_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="文件哈希">{{ detail.file_hash || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ detail.create_time }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ detail.finish_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="筛选条件">
          <pre>{{ formatJSON(detail.filter_json) }}</pre>
        </el-descriptions-item>
        <el-descriptions-item label="失败原因" v-if="detail.fail_reason">
          <span class="fail-reason">{{ detail.fail_reason }}</span>
        </el-descriptions-item>
      </el-descriptions>
      <div slot="footer">
        <el-button @click="detailVisible = false">关闭</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { createAuditExport, getAuditExportTasks, getAuditExportTaskDetail } from '@/api/audit'

export default {
  name: 'AuditExport',
  data() {
    return {
      creating: false,
      loading: false,
      taskList: [],
      pagination: {
        page: 1,
        size: 10,
        total: 0
      },
      exportForm: {
        format: 'csv',
        mask_sensitive: false,
        filter_json: ''
      },
      filterForm: {
        dateRange: [],
        event_type: '',
        action_result: '',
        target_patient_id: ''
      },
      showFilterDialog: false,
      detailVisible: false,
      detail: null
    }
  },
  computed: {
    hasFilters() {
      return this.filterForm.event_type ||
             this.filterForm.action_result ||
             this.filterForm.target_patient_id ||
             (this.filterForm.dateRange && this.filterForm.dateRange.length > 0)
    }
  },
  created() {
    // 从路由获取筛选条件
    if (this.$route.query) {
      if (this.$route.query.event_type) this.filterForm.event_type = this.$route.query.event_type
      if (this.$route.query.action_result) this.filterForm.action_result = this.$route.query.action_result
      if (this.$route.query.target_patient_id) this.filterForm.target_patient_id = this.$route.query.target_patient_id
    }
    this.loadTasks()
  },
  methods: {
    loadTasks() {
      this.loading = true
      getAuditExportTasks({
        page: this.pagination.page,
        size: this.pagination.size
      }).then(res => {
        this.loading = false
        // request.js 成功时已返回 res.data
        this.taskList = res?.list || []
        this.pagination.total = res?.total || 0
      }).catch(() => {
        this.loading = false
      })
    },
    handleCreateExport() {
      this.$confirm('确认要创建导出任务吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info'
      }).then(() => {
        this.creating = true
        const data = {
          format: this.exportForm.format,
          mask_sensitive: this.exportForm.mask_sensitive,
          filter_json: JSON.stringify(this.filterForm)
        }
        createAuditExport(data).then(() => {
          this.creating = false
          this.$message.success('导出任务创建成功')
          this.loadTasks()
        }).catch(err => {
          this.creating = false
          this.$message.error(err.message || '创建失败')
        })
      }).catch(() => {})
    },
    handleRefresh() {
      this.loadTasks()
    },
    handlePageChange(page) {
      this.pagination.page = page
      this.loadTasks()
    },
    handleDownload(row) {
      this.$message.info('下载功能开发中，请稍后')
      // TODO: 实现下载功能
    },
    handleViewDetail(row) {
      getAuditExportTaskDetail(row.id).then(res => {
        // request.js 成功时已返回 res.data
        this.detail = res
        this.detailVisible = true
      }).catch(() => {
        this.$message.error('获取详情失败')
      })
    },
    getStatusTag(status) {
      const map = {
        'PENDING': 'info',
        'RUNNING': 'warning',
        'SUCCESS': 'success',
        'FAIL': 'danger'
      }
      return map[status] || 'info'
    },
    getStatusName(status) {
      const map = {
        'PENDING': '等待中',
        'RUNNING': '生成中',
        'SUCCESS': '已完成',
        'FAIL': '失败'
      }
      return map[status] || status
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
.audit-export {
  padding: 20px;
}

.create-card {
  margin-bottom: 20px;
}

.filter-preview {
  display: flex;
  align-items: center;
  gap: 12px;
}

.hint {
  margin-left: 12px;
  color: #909399;
  font-size: 12px;
}

.task-card {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

.fail-reason {
  color: #f56c6c;
}
</style>
