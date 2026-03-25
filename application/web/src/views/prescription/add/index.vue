<template>
  <div class="add-record-page">
    <div class="page-head">
      <div class="head-left">
        <div class="head-icon"><i class="el-icon-document-add" /></div>
        <div>
          <div class="head-title">新增病历</div>
          <div class="head-subtitle">录入完整的患者病历信息并上链存储</div>
        </div>
      </div>
    </div>

    <div class="section-tabs">
      <div v-for="(tab, idx) in tabs" :key="tab.key" class="tab-item" :class="{ active: currentStep === idx }" @click="currentStep = idx">
        <i :class="tab.icon" />{{ tab.label }}
      </div>
    </div>

    <el-form ref="ruleForm" v-loading="loading" :model="ruleForm" :rules="rules" class="record-form">
      <el-card v-show="currentStep === 0" class="section-card">
        <div slot="header" class="section-header section-header--blue"><i class="el-icon-user" />患者基本信息</div>
        <div class="grid-3">
          <el-form-item label="患者姓名" prop="patient" required>
            <el-select v-model="ruleForm.patient" filterable clearable placeholder="请选择病人（支持搜索）" @change="onPatientChange">
              <el-option v-for="item in patientList" :key="item.account_id" :label="item.account_name" :value="item.account_id" />
            </el-select>
          </el-form-item>
          <el-form-item label="患者ID" prop="patient" required><el-input v-model="ruleForm.patient" disabled /></el-form-item>
          <el-form-item label="性别" prop="patient_gender" required>
            <el-select v-model="ruleForm.patient_gender" placeholder="请选择性别">
              <el-option label="男" value="男" />
              <el-option label="女" value="女" />
              <el-option label="其他" value="其他" />
            </el-select>
          </el-form-item>
          <el-form-item label="年龄" prop="patient_age" required><el-input v-model.trim="ruleForm.patient_age" /></el-form-item>
          <el-form-item label="身份证号"><el-input v-model.trim="ruleForm.patient_id_card_no" /></el-form-item>
          <el-form-item label="联系电话"><el-input v-model.trim="ruleForm.patient_phone" /></el-form-item>
          <el-form-item label="医保卡号"><el-input v-model.trim="ruleForm.insurance_card_no" /></el-form-item>
        </div>
      </el-card>

      <el-card v-show="currentStep === 1" class="section-card">
        <div slot="header" class="section-header section-header--green"><i class="el-icon-office-building" />就诊信息</div>
        <div class="grid-2">
          <el-form-item label="就诊医院" prop="hospital_name" required>
            <el-select v-model="ruleForm.hospital_name" filterable allow-create default-first-option placeholder="请选择或输入医院">
              <el-option v-for="name in hospitalOptions" :key="name" :label="name" :value="name" />
            </el-select>
          </el-form-item>
          <el-form-item label="就诊科室" prop="department" required><el-input v-model.trim="ruleForm.department" /></el-form-item>
          <el-form-item label="接诊医生" prop="visit_doctor_name" required><el-input v-model.trim="ruleForm.visit_doctor_name" /></el-form-item>
          <el-form-item label="病历类型" prop="record_type" required>
            <el-select v-model="ruleForm.record_type" placeholder="请选择病历类型">
              <el-option label="门诊病历" value="门诊病历" />
              <el-option label="住院病历" value="住院病历" />
              <el-option label="急诊病历" value="急诊病历" />
              <el-option label="检查报告" value="检查报告" />
            </el-select>
          </el-form-item>
        </div>
      </el-card>

      <el-card v-show="currentStep === 2" class="section-card">
        <div slot="header" class="section-header section-header--pink"><i class="el-icon-edit-outline" />诊断信息</div>
        <el-form-item label="主诉" prop="chief_complaint" required><el-input v-model="ruleForm.chief_complaint" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="现病史" prop="present_illness" required><el-input v-model="ruleForm.present_illness" type="textarea" :rows="4" /></el-form-item>
        <el-form-item label="既往史"><el-input v-model="ruleForm.past_history" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="过敏史"><el-input v-model="ruleForm.allergy_history" type="textarea" :rows="2" /><div class="danger-tip">请务必如实填写过敏史，以确保用药安全</div></el-form-item>
        <el-form-item label="家族史"><el-input v-model="ruleForm.family_history" type="textarea" :rows="2" /></el-form-item>
      </el-card>

      <el-card v-show="currentStep === 3" class="section-card">
        <div slot="header" class="section-header section-header--teal"><i class="el-icon-data-analysis" />检查与治疗</div>
        <div class="grid-4">
          <el-form-item label="体温(℃)"><el-input v-model.trim="ruleForm.temperature" /></el-form-item>
          <el-form-item label="脉搏(次/分)"><el-input v-model.trim="ruleForm.pulse" /></el-form-item>
          <el-form-item label="血压(mmHg)"><el-input v-model.trim="ruleForm.blood_pressure" /></el-form-item>
          <el-form-item label="呼吸(次/分)"><el-input v-model.trim="ruleForm.respiration" /></el-form-item>
        </div>
        <el-form-item label="体格检查详情"><el-input v-model="ruleForm.physical_exam" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="实验室检查"><el-input v-model="ruleForm.lab_exam" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="影像学检查"><el-input v-model="ruleForm.imaging_exam" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="诊断结果" prop="diagnosis_result" required><el-input v-model="ruleForm.diagnosis_result" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="治疗方案"><el-input v-model="ruleForm.treatment_plan" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="处方用药"><el-input v-model="ruleForm.medication_advice" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="医嘱"><el-input v-model="ruleForm.doctor_advice" type="textarea" :rows="3" /></el-form-item>
      </el-card>

      <el-card v-show="currentStep === 4" class="section-card">
        <div slot="header" class="section-header section-header--amber"><i class="el-icon-folder-opened" />附件上传</div>
        <el-form-item label="病历附件（选填）">
          <el-upload
            ref="recordUpload"
            drag
            action=""
            :auto-upload="false"
            :show-file-list="false"
            :before-upload="beforeFileUpload"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :file-list="fileList"
            class="upload-box"
          >
            <i class="el-icon-upload" />
            <div class="el-upload__text">点击上传或拖拽文件到此区域</div>
            <div class="el-upload__tip">支持 JPG / PNG / PDF / DOC / DOCX，单文件不超过200MB</div>
          </el-upload>
          <div v-if="fileList.length" class="file-row">
            <span>{{ fileList[0].name }}（{{ formatFileSize(fileList[0].size) }}）</span>
            <el-progress :percentage="uploadPercent" :stroke-width="8" :show-text="false" />
          </div>
        </el-form-item>
      </el-card>

      <div class="footer-actions">
        <div>
          <el-button @click="prevStep" :disabled="currentStep===0">上一步</el-button>
          <el-button @click="nextStep" v-if="currentStep<4">下一步</el-button>
        </div>
        <div>
          <el-button icon="el-icon-refresh-left" @click="resetForm('ruleForm')">重置</el-button>
          <el-button type="primary" icon="el-icon-document-checked" :loading="loading" @click="submitForm('ruleForm')">保存病历并上链</el-button>
        </div>
      </div>
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
      loading: false,
      accountList: [],
      patientList: [],
      currentStep: 0,
      tabs: [
        { key: 'patient', label: '患者信息', icon: 'el-icon-user' },
        { key: 'visit', label: '就诊信息', icon: 'el-icon-office-building' },
        { key: 'diagnosis', label: '诊断信息', icon: 'el-icon-edit-outline' },
        { key: 'treat', label: '检查与治疗', icon: 'el-icon-data-analysis' },
        { key: 'file', label: '附件上传', icon: 'el-icon-folder-opened' }
      ],
      selectedFile: null,
      fileList: [],
      uploadPercent: 0,
      ruleForm: {
        patient: '', patient_name: '', patient_gender: '', patient_age: '', patient_id_card_no: '', patient_phone: '', insurance_card_no: '',
        hospital_name: '', department: '', visit_doctor_name: '', record_type: '门诊病历',
        chief_complaint: '', present_illness: '', past_history: '', allergy_history: '', family_history: '',
        temperature: '', pulse: '', blood_pressure: '', respiration: '', physical_exam: '', lab_exam: '', imaging_exam: '',
        diagnosis_result: '', treatment_plan: '', medication_advice: '', doctor_advice: '', comment: ''
      },
      rules: {
        patient: [{ required: true, message: '请选择患者', trigger: 'change' }],
        patient_gender: [{ required: true, message: '请选择性别', trigger: 'change' }],
        patient_age: [{ required: true, message: '请输入年龄', trigger: 'blur' }],
        hospital_name: [{ required: true, message: '请输入就诊医院', trigger: 'change' }],
        department: [{ required: true, message: '请输入就诊科室', trigger: 'blur' }],
        visit_doctor_name: [{ required: true, message: '请输入接诊医生', trigger: 'blur' }],
        record_type: [{ required: true, message: '请选择病历类型', trigger: 'change' }],
        chief_complaint: [{ required: true, message: '主诉不能为空', trigger: 'blur' }],
        present_illness: [{ required: true, message: '现病史不能为空', trigger: 'blur' }],
        diagnosis_result: [{ required: true, message: '诊断结果不能为空', trigger: 'blur' }]
      }
    }
  },
  computed: {
    ...mapGetters(['account_id', 'account_name']),
    hospitalOptions() {
      const s = new Set()
      this.accountList.filter(i => i.role === 'doctor' && i.hospital_name).forEach(i => s.add(i.hospital_name))
      if (this.ruleForm.hospital_name) s.add(this.ruleForm.hospital_name)
      return Array.from(s)
    }
  },
  created() {
    queryAccountList().then((response) => {
      this.accountList = response || []
      this.patientList = this.accountList.filter(item => item.role === 'patient')
      const me = this.accountList.find(i => i.account_id === this.account_id)
      this.ruleForm.visit_doctor_name = (me && me.account_name) || this.account_name || ''
      this.ruleForm.hospital_name = (me && me.hospital_name) || ''
      this.ruleForm.department = (me && me.department) || ''
    })
  },
  methods: {
    nextStep() { this.currentStep = Math.min(4, this.currentStep + 1) },
    prevStep() { this.currentStep = Math.max(0, this.currentStep - 1) },
    onPatientChange(patientId) {
      const p = this.patientList.find(i => i.account_id === patientId)
      if (!p) return
      this.ruleForm.patient_name = p.account_name || ''
      this.ruleForm.patient_gender = p.gender || ''
      this.ruleForm.patient_age = p.age || ''
      this.ruleForm.patient_id_card_no = p.id_card_no || ''
      this.ruleForm.patient_phone = p.phone || ''
      this.ruleForm.insurance_card_no = p.insurance_card_no || ''
    },
    beforeFileUpload(file) {
      const ext = `.${file.name.split('.').pop().toLowerCase()}`
      if (!ALLOWED_FILE_EXT.includes(ext)) {
        this.$message.error('文件格式不支持，请上传 PDF/Word/图片')
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
        return
      }
      this.selectedFile = file.raw
      this.fileList = fileList.slice(-1)
      this.uploadPercent = 0
    },
    handleFileRemove() { this.selectedFile = null; this.fileList = []; this.uploadPercent = 0 },
    formatFileSize(size) {
      if (!size) return '0 B'
      if (size < 1024) return `${size} B`
      if (size < 1024 * 1024) return `${(size / 1024).toFixed(2)} KB`
      return `${(size / 1024 / 1024).toFixed(2)} MB`
    },
    submitForm(formName) {
      this.$refs[formName].validate(valid => {
        if (!valid) return this.$message.warning('请先完善必填项后再提交')
        this.loading = true
        const fd = new FormData()
        fd.append('doctor', this.account_id)
        fd.append('patient', this.ruleForm.patient)
        fd.append('record_type', this.ruleForm.record_type)
        fd.append('file_name', this.selectedFile ? this.selectedFile.name : '')
        fd.append('symptom_description', this.ruleForm.chief_complaint)
        fd.append('doctor_diagnosis', this.ruleForm.diagnosis_result)
        fd.append('diagnosis', this.ruleForm.diagnosis_result)
        fd.append('drug_name', this.ruleForm.medication_advice)
        fd.append('drug_amount', '')
        fd.append('hospital', this.ruleForm.hospital_name || '0feceb66ffc1')
        fd.append('comment', this.ruleForm.comment || '')

        fd.append('patient_name', this.ruleForm.patient_name)
        fd.append('patient_gender', this.ruleForm.patient_gender)
        fd.append('patient_age', this.ruleForm.patient_age)
        fd.append('patient_id_card_no', this.ruleForm.patient_id_card_no)
        fd.append('patient_phone', this.ruleForm.patient_phone)
        fd.append('insurance_card_no', this.ruleForm.insurance_card_no)
        fd.append('hospital_name', this.ruleForm.hospital_name)
        fd.append('department', this.ruleForm.department)
        fd.append('visit_doctor_name', this.ruleForm.visit_doctor_name || this.account_name)
        fd.append('chief_complaint', this.ruleForm.chief_complaint)
        fd.append('present_illness', this.ruleForm.present_illness)
        fd.append('past_history', this.ruleForm.past_history)
        fd.append('allergy_history', this.ruleForm.allergy_history)
        fd.append('family_history', this.ruleForm.family_history)
        fd.append('temperature', this.ruleForm.temperature)
        fd.append('pulse', this.ruleForm.pulse)
        fd.append('blood_pressure', this.ruleForm.blood_pressure)
        fd.append('respiration', this.ruleForm.respiration)
        fd.append('physical_exam', this.ruleForm.physical_exam)
        fd.append('lab_exam', this.ruleForm.lab_exam)
        fd.append('imaging_exam', this.ruleForm.imaging_exam)
        fd.append('diagnosis_result', this.ruleForm.diagnosis_result)
        fd.append('treatment_plan', this.ruleForm.treatment_plan)
        fd.append('medication_advice', this.ruleForm.medication_advice)
        fd.append('doctor_advice', this.ruleForm.doctor_advice)
        if (this.selectedFile) fd.append('record_file', this.selectedFile)

        createPrescription(fd, evt => {
          if (!evt || !evt.total) return
          this.uploadPercent = Math.min(100, Math.round((evt.loaded / evt.total) * 100))
        }).then(() => {
          this.$message.success('创建成功')
          this.loading = false
          this.$router.push('/prescription/all')
        }).catch((err) => {
          this.$message.error((err && err.message) || '创建失败')
          this.loading = false
        })
      })
    },
    resetForm(formName) {
      this.$refs[formName].resetFields()
      this.selectedFile = null
      this.fileList = []
      this.uploadPercent = 0
      this.currentStep = 0
      if (this.$refs.recordUpload) this.$refs.recordUpload.clearFiles()
    }
  }
}
</script>

<style scoped>
.add-record-page { background: linear-gradient(149deg, #eff6ff 0%, #fff 50%, #eef2ff 100%); padding: 24px; }
.page-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.head-left { display: flex; align-items: center; gap: 12px; }
.head-icon { width: 48px; height: 48px; border-radius: 14px; background: linear-gradient(135deg, #155dfc 0%, #4f39f6 100%); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 24px; }
.head-title { font-size: 24px; font-weight: 700; color: #101828; line-height: 32px; }
.head-subtitle { font-size: 14px; color: #6a7282; }
.section-tabs { display: grid; grid-template-columns: repeat(5, 1fr); gap: 8px; background: #fff; border-radius: 14px; padding: 4px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,.08); }
.tab-item { height: 44px; display: flex; align-items: center; justify-content: center; gap: 8px; border-radius: 12px; font-size: 14px; cursor: pointer; }
.tab-item.active { background: #fff; box-shadow: 0 1px 2px rgba(0,0,0,.08); font-weight: 600; color: #155dfc; }
.section-card { border-radius: 14px; margin-bottom: 16px; }
.section-header { font-size: 16px; font-weight: 600; display: flex; align-items: center; gap: 8px; }
.section-header--blue { color: #1d4ed8; }
.section-header--green { color: #15803d; }
.section-header--pink { color: #9333ea; }
.section-header--teal { color: #0f766e; }
.section-header--amber { color: #c2410c; }
.grid-2,.grid-3,.grid-4 { display: grid; gap: 16px 24px; }
.grid-2 { grid-template-columns: repeat(2, 1fr); }
.grid-3 { grid-template-columns: repeat(3, 1fr); }
.grid-4 { grid-template-columns: repeat(4, 1fr); }
.upload-box { width: 100%; }
.file-row { margin-top: 12px; }
.footer-actions { background: rgba(255,255,255,.95); border: 1px solid rgba(229,231,235,.8); border-radius: 14px; padding: 14px 16px; display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.danger-tip { margin-top: 6px; color: #e7000b; font-size: 12px; }
.record-form /deep/ .el-form-item__label { font-size: 14px; color: #0a0a0a; }
.record-form /deep/ .el-input__inner,
.record-form /deep/ .el-textarea__inner { border-radius: 8px; }
</style>