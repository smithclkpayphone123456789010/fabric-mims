<template>
  <div class="app-container">
    <el-form
      ref="ruleForm"
      v-loading="loading"
      :model="ruleForm"
      :rules="rules"
      label-width="80px"
      class="record-form"
    >
      <el-form-item label="病人" prop="patient">
        <el-select v-model="ruleForm.patient" placeholder="请选择病人（支持模糊搜索）" class="full-width" filterable clearable @change="selectGet">
          <el-option
            v-for="item in accountList"
            :key="item.account_id"
            :label="item.account_name"
            :value="item.account_id"
          >
            <span style="float: left">{{ item.account_name }}</span>
            <span style="float: right; color: #8492a6; font-size: 13px">{{ item.account_id }}</span>
          </el-option>
        </el-select>
      </el-form-item>

      <el-form-item label="病历类型" prop="record_type">
        <el-select v-model="ruleForm.record_type" class="full-width" placeholder="请选择病历类型">
          <el-option label="EMR（电子病历）" value="EMR" />
          <el-option label="REPORT（检查报告）" value="REPORT" />
          <el-option label="PRESCRIPTION（处方）" value="PRESCRIPTION" />
        </el-select>
      </el-form-item>

      <el-form-item label="病历文件" prop="record_file">
        <el-upload
          ref="recordUpload"
          class="upload-box"
          drag
          action=""
          :auto-upload="false"
          :show-file-list="false"
          :before-upload="beforeFileUpload"
          :on-change="handleFileChange"
          :on-remove="handleFileRemove"
          :file-list="fileList"
        >
          <i class="el-icon-upload" />
          <div class="el-upload__text">点击或拖拽文件到此区域</div>
        </el-upload>
        <div class="upload-tip">支持PDF、Word、图片等格式，单个文件不超过200MB</div>

        <div v-if="fileList.length" class="file-list">
          <div v-for="file in fileList" :key="file.uid" class="file-row">
            <div class="file-meta">
              <span class="file-name">{{ file.name }}</span>
              <span class="file-size">{{ formatFileSize(file.size) }}</span>
              <i v-if="uploadPercent === 100" class="el-icon-success success-icon" />
            </div>
            <el-progress :percentage="uploadPercent" :stroke-width="8" :show-text="false" />
          </div>
        </div>
      </el-form-item>

      <el-form-item label="症状描述" prop="symptom_description">
        <el-input
          v-model="ruleForm.symptom_description"
          type="textarea"
          :rows="4"
          maxlength="500"
          show-word-limit
          class="full-width"
          placeholder="请输入症状描述"
        />
      </el-form-item>

      <el-form-item label="医生诊断" prop="doctor_diagnosis">
        <el-input
          v-model="ruleForm.doctor_diagnosis"
          type="textarea"
          :rows="4"
          maxlength="500"
          show-word-limit
          class="full-width"
          placeholder="请输入医生诊断"
        />
      </el-form-item>

      <el-form-item label="备注" prop="comment">
        <el-input
          v-model="ruleForm.comment"
          type="textarea"
          :rows="3"
          maxlength="500"
          show-word-limit
          class="full-width"
          placeholder="请输入备注（选填）"
        />
      </el-form-item>

      <el-form-item>
        <el-button type="primary" @click="submitForm('ruleForm')">立即创建</el-button>
        <el-button @click="resetForm('ruleForm')">重置</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { queryAccountList } from '@/api/accountV2'
import { createPrescription } from '@/api/prescription'

const MAX_FILE_SIZE = 200 * 1024 * 1024
const ALLOWED_FILE_EXT = ['.pdf', '.doc', '.docx', '.jpg', '.jpeg', '.png']

export default {
  name: 'AddPrescription',
  data() {
    return {
      ruleForm: {
        patient: '',
        record_type: 'EMR',
        symptom_description: '',
        doctor_diagnosis: '',
        comment: ''
      },
      selectedFile: null,
      fileList: [],
      uploadPercent: 0,
      accountList: [],
      rules: {
        patient: [{ required: true, message: '请选择病人', trigger: 'change' }],
        record_type: [{ required: true, message: '请选择病历类型', trigger: 'change' }],
        record_file: [{ validator: (rule, value, callback) => this.validateRecordFile(callback), trigger: 'change' }],
        symptom_description: [{ max: 500, message: '症状描述不能超过500字符', trigger: 'blur' }],
        doctor_diagnosis: [{ max: 500, message: '医生诊断不能超过500字符', trigger: 'blur' }],
        comment: [{ max: 500, message: '备注不能超过500字符', trigger: 'blur' }]
      },
      loading: false
    }
  },
  computed: {
    ...mapGetters([
      'account_id'
    ])
  },
  created() {
    queryAccountList().then(response => {
      if (response !== null) {
        this.accountList = response.filter(item => {
          if (item.role) return item.role === 'patient'
          return /病人|患者/.test(item.account_name || '')
        })
      }
    })
  },
  methods: {
    validateRecordFile(callback) {
      if (!this.selectedFile) {
        callback(new Error('请上传病历文件'))
        return
      }
      callback()
    },
    beforeFileUpload(file) {
      const ext = `.${file.name.split('.').pop().toLowerCase()}`
      if (!ALLOWED_FILE_EXT.includes(ext)) {
        this.$message.error('文件格式不支持，请上传PDF/Word/图片文件')
        return false
      }
      if (file.size > MAX_FILE_SIZE) {
        this.$message.error('文件大小不能超过200MB')
        return false
      }
      return true
    },
    handleFileChange(file, fileList) {
      const ext = `.${file.name.split('.').pop().toLowerCase()}`
      if (!ALLOWED_FILE_EXT.includes(ext) || file.size > MAX_FILE_SIZE) {
        this.fileList = []
        this.selectedFile = null
        this.$refs.ruleForm.validateField('record_file')
        return
      }
      this.selectedFile = file.raw
      this.fileList = fileList.slice(-1)
      this.uploadPercent = 0
      this.$refs.ruleForm.validateField('record_file')
    },
    handleFileRemove() {
      this.selectedFile = null
      this.fileList = []
      this.uploadPercent = 0
      this.$refs.ruleForm.validateField('record_file')
    },
    formatFileSize(size) {
      if (!size) return '0 B'
      if (size < 1024) return `${size} B`
      if (size < 1024 * 1024) return `${(size / 1024).toFixed(2)} KB`
      if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(2)} MB`
      return `${(size / 1024 / 1024 / 1024).toFixed(2)} GB`
    },
    submitForm(formName) {
      this.$refs[formName].validate((valid) => {
        if (!valid) return false

          this.$confirm('是否立即创建?', '提示', {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'success'
          }).then(() => {
            this.loading = true
          this.uploadPercent = 0

          const formData = new FormData()
          formData.append('doctor', this.account_id)
          formData.append('patient', this.ruleForm.patient)
          formData.append('record_type', this.ruleForm.record_type)
          formData.append('symptom_description', this.ruleForm.symptom_description)
          formData.append('doctor_diagnosis', this.ruleForm.doctor_diagnosis)
          formData.append('diagnosis', this.ruleForm.doctor_diagnosis)
          formData.append('hospital', '0feceb66ffc1')
          formData.append('comment', this.ruleForm.comment)
          formData.append('record_file', this.selectedFile)

          createPrescription(formData, (evt) => {
            if (!evt || !evt.total) return
            this.uploadPercent = Math.min(100, Math.round((evt.loaded / evt.total) * 100))
            }).then(response => {
              this.loading = false
            this.uploadPercent = 100
              if (response !== null) {
              this.$message({ type: 'success', message: '创建成功!' })
              } else {
              this.$message({ type: 'error', message: '创建失败!' })
              }
          }).catch(() => {
              this.loading = false
            this.uploadPercent = 0
            })
          }).catch(() => {
            this.loading = false
          this.$message({ type: 'info', message: '已取消创建' })
            })
      })
    },
    resetForm(formName) {
      this.$refs[formName].resetFields()
      this.selectedFile = null
      this.fileList = []
      this.uploadPercent = 0
      if (this.$refs.recordUpload) {
        this.$refs.recordUpload.clearFiles()
      }
    },
    selectGet(account_id) {
      this.ruleForm.patient = account_id
    }
  }
}
</script>

<style scoped>
.record-form {
  max-width: 780px;
}

.full-width {
  width: 100%;
}

.upload-box {
  width: 100%;
}

.upload-tip {
  margin-top: 8px;
  color: #909399;
  font-size: 12px;
}

.file-list {
  margin-top: 12px;
}

.file-row {
  padding: 10px 12px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  margin-bottom: 8px;
  background: #fafafa;
}

.file-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.file-name {
  font-weight: 500;
}

.file-size {
  color: #909399;
  font-size: 12px;
}

.success-icon {
  color: #67c23a;
}
</style>
