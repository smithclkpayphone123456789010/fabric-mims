<template>
  <div class="container">
    <el-alert type="success" :closable="false">
      <p>账户ID: {{ account_id }}</p>
      <p>用户名: {{ account_name }}</p>
    </el-alert>

    <div v-if="prescriptionList.length === 0" style="text-align: center; margin-top: 12px;">
      <el-alert title="查询不到数据" type="warning" :closable="false" />
    </div>

    <el-row v-loading="loading" :gutter="20">
      <el-col v-for="(val, index) in prescriptionList" :key="index" :span="8">
        <el-card class="prescription-card">
          <div slot="header" class="clearfix">
            病历ID:
            <span class="id-text">{{ val.id }}</span>
          </div>

          <div class="item"><el-tag>病人ID</el-tag><span class="ml">{{ val.patient }}</span></div>
          <div class="item"><el-tag type="success">患者姓名</el-tag><span class="ml">{{ getPatientInfo(val.patient).account_name || '-' }}</span></div>
          <div class="item"><el-tag type="warning">患者身份证号</el-tag><span class="ml">{{ getPatientInfo(val.patient).id_card_no || '-' }}</span></div>
          <div class="item"><el-tag type="success">医生ID</el-tag><span class="ml">{{ val.doctor }}</span></div>
          <div class="item"><el-tag type="danger">病历类型</el-tag><span class="ml">{{ val.record_type || '-' }}</span></div>
          <div class="item"><el-tag type="info">创建时间</el-tag><span class="ml">{{ val.created }}</span></div>

          <div class="actions">
            <el-button
              type="primary"
              size="mini"
              :disabled="roles[0] === 'doctor' && val.doctor !== account_id"
              @click="openDetail(val)"
            >
              查看详情
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog title="病历详情" :visible.sync="detailVisible" width="760px">
      <el-descriptions :column="2" border size="small" v-if="detailItem">
        <el-descriptions-item label="病历ID">{{ detailItem.id }}</el-descriptions-item>
        <el-descriptions-item label="病历类型">{{ detailItem.record_type || '-' }}</el-descriptions-item>
        <el-descriptions-item label="医生用户名">{{ doctorInfo.account_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="所属医院">{{ doctorInfo.hospital_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="所属科室">{{ doctorInfo.department || '-' }}</el-descriptions-item>
        <el-descriptions-item label="病人ID">{{ detailItem.patient || '-' }}</el-descriptions-item>
        <el-descriptions-item label="病人用户名">{{ patientInfo.account_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="年龄">{{ patientInfo.age || '-' }}</el-descriptions-item>
        <el-descriptions-item label="身份证">{{ patientInfo.id_card_no || '-' }}</el-descriptions-item>
        <el-descriptions-item label="病历文件hash">{{ detailItem.file_hash || '-' }}</el-descriptions-item>
        <el-descriptions-item label="诊断" :span="2">{{ detailItem.doctor_diagnosis || detailItem.diagnosis || '-' }}</el-descriptions-item>
        <el-descriptions-item label="备注" :span="2">{{ detailItem.comment || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间" :span="2">{{ detailItem.created || '-' }}</el-descriptions-item>
      </el-descriptions>

      <div class="preview-row" v-if="detailItem && detailItem.file_path">
        <el-button type="primary" plain size="mini" @click="openFilePreview(detailItem)">病历文件详情</el-button>
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
import { queryAccountList } from '@/api/accountV2'
import { queryPrescriptionList } from '@/api/prescription'
import { checkRecordAccess } from '@/api/authorization'

export default {
  name: 'Prescription',
  data() {
    return {
      loading: true,
      prescriptionList: [],
      accountList: [],
      detailVisible: false,
      previewVisible: false,
      detailItem: null,
      previewUrl: '',
      currentPreviewFileName: ''
    }
  },
  computed: {
    ...mapGetters([
      'account_id',
      'roles',
      'account_name'
    ]),
    doctorInfo() {
      if (!this.detailItem) return {}
      return this.accountList.find(item => item.account_id === this.detailItem.doctor) || {}
    },
    patientInfo() {
      if (!this.detailItem) return {}
      return this.accountList.find(item => item.account_id === this.detailItem.patient) || {}
    }
  },
  created() {
    Promise.all([
      queryAccountList(),
      this.roles[0] === 'admin'
        ? queryPrescriptionList()
        : this.roles[0] === 'doctor'
          ? queryPrescriptionList()
          : queryPrescriptionList({ patient: this.account_id })
    ]).then(([accounts, prescriptions]) => {
      this.accountList = accounts || []
      this.prescriptionList = prescriptions || []
      this.loading = false
    }).catch(() => {
      this.loading = false
    })
  },
  methods: {
    getPatientInfo(patientId) {
      return this.accountList.find(item => item.account_id === patientId) || {}
    },
    openDetail(item) {
      if (this.roles[0] !== 'doctor') {
        this.detailItem = item
        this.detailVisible = true
        return
      }
      checkRecordAccess({ doctor_id: this.account_id, record_id: item.id }).then(() => {
        this.detailItem = item
        this.detailVisible = true
      }).catch(() => {
        this.$message.error('无权限查看该病历详情')
      })
    },
    openFilePreview(item) {
      const doPreview = () => {
        const query = `doctor_id=${encodeURIComponent(this.roles[0] === 'doctor' ? this.account_id : '')}&record_id=${encodeURIComponent(item.id || '')}&file_path=${encodeURIComponent(item.file_path || '')}&file_name=${encodeURIComponent(item.file_name || '')}`
        this.currentPreviewFileName = item.file_name || ''
        this.previewUrl = `${process.env.VUE_APP_BASE_API}/previewPrescriptionFile?${query}`
        this.previewVisible = true
      }
      if (this.roles[0] !== 'doctor') {
        doPreview()
        return
      }
      checkRecordAccess({ doctor_id: this.account_id, record_id: item.id }).then(() => {
        doPreview()
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
.container {
  width: 100%;
  min-height: 100%;
  overflow: hidden;
  font-size: 15px;
}

.item {
  font-size: 13px;
  margin-bottom: 10px;
  color: #606266;
}

.ml {
  margin-left: 8px;
}

.id-text {
  color: #f56c6c;
}

.prescription-card {
  margin: 12px 0;
  min-height: 250px;
}

.actions {
  margin-top: 12px;
  text-align: right;
}

.preview-row {
  margin-top: 16px;
  text-align: right;
}

.preview-wrap {
  min-height: 380px;
}

.preview-image {
  width: 100%;
  max-height: 70vh;
  object-fit: contain;
}

.preview-frame {
  width: 100%;
  height: 70vh;
  border: none;
}

.empty-preview {
  color: #909399;
  text-align: center;
  padding: 40px 0;
}
</style>
