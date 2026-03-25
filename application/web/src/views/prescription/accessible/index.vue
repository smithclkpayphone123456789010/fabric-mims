<template>
  <div class="container">
    <el-alert type="success" :closable="false" class="mb16">
      <p>账户ID: {{ account_id }}</p>
      <p>用户名: {{ account_name }}</p>
    </el-alert>

    <el-card shadow="never" class="mb16">
      <div class="filters">
        <el-input v-model="filters.patient_name_keyword" clearable placeholder="患者姓名（模糊）" class="w220" />
        <el-input v-model="filters.id_card_keyword" clearable placeholder="身份证号（模糊）" class="w220" />
        <el-input v-model="filters.record_type_keyword" clearable placeholder="病历类型（模糊）" class="w220" />
        <el-date-picker
          v-model="filters.createdRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          value-format="yyyy-MM-dd"
          unlink-panels
          class="w360"
        />
        <el-button type="primary" @click="fetchList">查询</el-button>
        <el-button @click="onReset">重置</el-button>
      </div>
    </el-card>

    <div v-if="currentList.length === 0" style="text-align:center;">
      <el-alert title="查询不到数据" type="warning" :closable="false" />
    </div>

    <el-row v-loading="loading" :gutter="20">
      <el-col v-for="(val, index) in currentList" :key="index" :span="8">
        <el-card class="prescription-card">
          <div slot="header" class="clearfix">
            病历ID:
            <span class="id-text">{{ val.record.id }}</span>
          </div>
          <div class="item"><el-tag>患者姓名</el-tag><span class="ml">{{ val.patient.account_name || '-' }}</span></div>
          <div class="item"><el-tag type="warning">患者身份证号</el-tag><span class="ml">{{ val.patient.id_card_no || '-' }}</span></div>
          <div class="item"><el-tag type="success">病历类型</el-tag><span class="ml">{{ val.record.record_type || '-' }}</span></div>
          <div class="item"><el-tag type="info">创建时间</el-tag><span class="ml">{{ val.record.created || '-' }}</span></div>
          <div class="item"><el-tag type="danger">授权到期</el-tag><span class="ml">{{ val.authorization.end_time || '-' }}</span></div>

          <div class="actions">
            <el-button type="primary" size="mini" @click="openDetail(val)">查看详情</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog title="病历详情" :visible.sync="detailVisible" width="760px">
      <el-descriptions :column="2" border size="small" v-if="detailItem">
        <el-descriptions-item label="病历ID">{{ detailItem.record.id }}</el-descriptions-item>
        <el-descriptions-item label="病历类型">{{ detailItem.record.record_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="患者姓名">{{ detailItem.patient.account_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="身份证">{{ detailItem.patient.id_card_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="病历文件hash">{{ detailItem.record.file_hash || '-' }}</el-descriptions-item>
        <el-descriptions-item label="授权到期">{{ detailItem.authorization.end_time || '-' }}</el-descriptions-item>
        <el-descriptions-item label="诊断" :span="2">{{ detailItem.record.doctor_diagnosis || detailItem.record.diagnosis || '-' }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ detailItem.record.comment || '-' }}</el-descriptions-item>
      </el-descriptions>

      <div class="preview-row" v-if="detailItem && detailItem.record.file_path">
        <el-button type="primary" plain size="mini" @click="openFilePreview(detailItem.record)">病历文件详情</el-button>
      </div>
    </el-dialog>

    <el-dialog title="病历文件预览" :visible.sync="previewVisible" width="900px">
      <div class="preview-wrap" v-if="previewUrl">
        <img v-if="isImageFile(currentPreviewFileName)" :src="previewUrl" alt="病历图片" class="preview-image">
        <iframe v-else :src="previewUrl" class="preview-frame" />
      </div>
      <div v-else class="empty-preview">暂无可预览文件</div>
    </el-dialog>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { queryAccessibleRecordsByDoctor, checkRecordAccess } from '@/api/authorization'

export default {
  name: 'PrescriptionAccessible',
  data() {
    return {
      loading: false,
      currentList: [],
      detailVisible: false,
      previewVisible: false,
      detailItem: null,
      previewUrl: '',
      currentPreviewFileName: '',
      filters: {
        patient_name_keyword: '',
        id_card_keyword: '',
        record_type_keyword: '',
        createdRange: []
      }
    }
  },
  computed: {
    ...mapGetters(['account_id', 'account_name'])
  },
  created() {
    this.fetchList()
  },
  methods: {
    fetchList() {
      this.loading = true
      const [start, end] = this.filters.createdRange || []
      queryAccessibleRecordsByDoctor({
        doctor_id: this.account_id,
        patient_name_keyword: this.filters.patient_name_keyword,
        id_card_keyword: this.filters.id_card_keyword,
        record_type_keyword: this.filters.record_type_keyword,
        created_start: start || '',
        created_end: end || ''
      }).then(data => {
        this.currentList = data || []
        this.loading = false
      }).catch(() => {
        this.loading = false
      })
    },
    onReset() {
      this.filters = {
        patient_name_keyword: '',
        id_card_keyword: '',
        record_type_keyword: '',
        createdRange: []
      }
      this.fetchList()
    },
    openDetail(item) {
      checkRecordAccess({ doctor_id: this.account_id, record_id: item.record.id }).then(() => {
        this.detailItem = item
        this.detailVisible = true
      }).catch(() => {
        this.$message.error('无权限查看该病历详情')
      })
    },
    openFilePreview(record) {
      checkRecordAccess({ doctor_id: this.account_id, record_id: record.id }).then(() => {
        const query = `doctor_id=${encodeURIComponent(this.account_id)}&record_id=${encodeURIComponent(record.id)}&file_path=${encodeURIComponent(record.file_path || '')}&file_name=${encodeURIComponent(record.file_name || '')}`
        this.currentPreviewFileName = record.file_name || ''
        this.previewUrl = `${process.env.VUE_APP_BASE_API}/previewPrescriptionFile?${query}`
        this.previewVisible = true
      }).catch(() => {
        this.$message.error('无权限预览该病历文件')
      })
    },
    isImageFile(name) {
      const lower = String(name || '').toLowerCase()
      return lower.endsWith('.jpg') || lower.endsWith('.jpeg') || lower.endsWith('.png')
    }
  }
}
</script>

<style scoped>
.container { width: 100%; min-height: 100%; overflow: hidden; font-size: 15px; }
.mb16 { margin-bottom: 16px; }
.filters { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
.w220 { width: 220px; }
.w360 { width: 360px; }
.item { font-size: 13px; margin-bottom: 10px; color: #606266; }
.ml { margin-left: 8px; }
.id-text { color: #f56c6c; }
.prescription-card { margin: 12px 0; min-height: 250px; }
.actions { margin-top: 12px; text-align: right; }
.preview-row { margin-top: 16px; text-align: right; }
.preview-wrap { min-height: 380px; }
.preview-image { width: 100%; max-height: 70vh; object-fit: contain; }
.preview-frame { width: 100%; height: 70vh; border: none; }
.empty-preview { color: #909399; text-align: center; padding: 40px 0; }
</style>
