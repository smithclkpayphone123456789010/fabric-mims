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
            <el-button type="primary" size="mini" @click="openDetail(val)">查看详情</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog title="病历详情" :visible.sync="detailVisible" width="760px">
      <div v-if="detailItem" class="detail-vertical">
        <div class="detail-row"><span class="label">病历ID：</span><span>{{ detailItem.id || '-' }}</span></div>
        <div class="detail-row"><span class="label">医生用户名：</span><span>{{ doctorInfo.account_name || '-' }}</span></div>
        <div class="detail-row"><span class="label">所属医院：</span><span>{{ doctorInfo.hospital_name || '-' }}</span></div>
        <div class="detail-row"><span class="label">所属科室：</span><span>{{ doctorInfo.department || '-' }}</span></div>
        <div class="detail-row"><span class="label">病人ID：</span><span>{{ detailItem.patient || '-' }}</span></div>
        <div class="detail-row"><span class="label">病人用户名：</span><span>{{ patientInfo.account_name || '-' }}</span></div>
        <div class="detail-row"><span class="label">年龄：</span><span>{{ patientInfo.age || '-' }}</span></div>
        <div class="detail-row"><span class="label">身份证：</span><span>{{ patientInfo.id_card_no || '-' }}</span></div>
        <div class="detail-row"><span class="label">诊断：</span><span>{{ detailItem.doctor_diagnosis || detailItem.diagnosis || '-' }}</span></div>
        <div class="detail-row"><span class="label">病历文件hash：</span><span>{{ detailItem.file_hash || '-' }}</span></div>
        <div class="detail-row"><span class="label">备注：</span><span>{{ detailItem.comment || '-' }}</span></div>
        <div class="detail-row"><span class="label">创建时间：</span><span>{{ detailItem.created || '-' }}</span></div>
      </div>

      <div class="preview-row" v-if="detailItem && detailItem.file_path">
        <el-button type="primary" plain size="mini" @click="openSecondPasswordDialog">病历文件详情</el-button>
      </div>
    </el-dialog>

    <el-dialog title="二级密码校验" :visible.sync="pwdDialogVisible" width="420px">
      <el-form label-width="90px">
        <el-form-item label="二级密码">
          <el-input v-model="secondPassword" show-password placeholder="请输入二级密码" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="pwdDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="verifySecondPassword">确认</el-button>
      </span>
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

export default {
  name: 'PrescriptionMine',
  data() {
    return {
      loading: true,
      prescriptionList: [],
      accountList: [],
      detailVisible: false,
      previewVisible: false,
      pwdDialogVisible: false,
      detailItem: null,
      previewUrl: '',
      currentPreviewFileName: '',
      secondPassword: ''
    }
  },
  computed: {
    ...mapGetters(['account_id', 'roles', 'account_name']),
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
      this.roles[0] === 'doctor' ? queryPrescriptionList({ doctor_id: this.account_id }) : queryPrescriptionList({ patient: this.account_id })
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
      this.detailItem = item
      this.detailVisible = true
    },
    openSecondPasswordDialog() {
      this.secondPassword = ''
      this.pwdDialogVisible = true
    },
    verifySecondPassword() {
      if (this.secondPassword !== '123') {
        this.$message.error('二级密码错误')
        return
      }
      this.pwdDialogVisible = false
      this.openFilePreview(this.detailItem)
    },
    openFilePreview(item) {
      if (!item) return
      const query = `doctor_id=${encodeURIComponent(this.roles[0] === 'doctor' ? this.account_id : '')}&record_id=${encodeURIComponent(item.id || '')}&file_path=${encodeURIComponent(item.file_path || '')}&file_name=${encodeURIComponent(item.file_name || '')}`
      this.currentPreviewFileName = item.file_name || ''
      this.previewUrl = `${process.env.VUE_APP_BASE_API}/previewPrescriptionFile?${query}`
      this.previewVisible = true
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
.item { font-size: 13px; margin-bottom: 10px; color: #606266; }
.ml { margin-left: 8px; }
.id-text { color: #f56c6c; }
.prescription-card { margin: 12px 0; min-height: 250px; }
.actions { margin-top: 12px; text-align: right; }
.detail-vertical { display: flex; flex-direction: column; gap: 10px; }
.detail-row { line-height: 1.6; word-break: break-all; }
.detail-row .label { color: #606266; font-weight: 600; }
.preview-row { margin-top: 16px; text-align: right; }
.preview-wrap { min-height: 380px; }
.preview-image { width: 100%; max-height: 70vh; object-fit: contain; }
.preview-frame { width: 100%; height: 70vh; border: none; }
.empty-preview { color: #909399; text-align: center; padding: 40px 0; }
</style>