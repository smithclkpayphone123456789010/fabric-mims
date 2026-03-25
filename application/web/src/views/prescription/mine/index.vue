<template>
  <div class="mine-page">
    <div class="page-head">
      <div class="head-left">
        <div class="head-icon"><i class="el-icon-document-copy" /></div>
        <div>
          <div class="head-title">我的病历</div>
          <div class="head-subtitle">查看管理您就诊的所有病历记录</div>
        </div>
      </div>
      <div class="head-right">
        <div class="mini-label">患者ID</div>
        <div class="mini-value">{{ account_id }}</div>
      </div>
    </div>

    <el-card class="filter-card" shadow="always">
      <div class="filter-row">
        <el-input v-model.trim="filters.keyword" placeholder="关键词搜索" class="w280" clearable>
          <i slot="prefix" class="el-input__icon el-icon-search" />
        </el-input>
        <el-select v-model="filters.hospital" placeholder="医院筛选" class="w220" clearable>
          <el-option label="全部医院" value="" />
          <el-option v-for="h in hospitalOptions" :key="h" :label="h" :value="h" />
        </el-select>
        <el-select v-model="filters.recordType" placeholder="类型筛选" class="w220" clearable>
          <el-option label="全部类型" value="" />
          <el-option label="门诊病历" value="门诊病历" />
          <el-option label="住院病历" value="住院病历" />
          <el-option label="急诊病历" value="急诊病历" />
          <el-option label="检查报告" value="检查报告" />
        </el-select>
      </div>
    </el-card>

    <div v-if="filteredList.length === 0" class="empty-wrap">
      <el-empty description="暂无病历记录" />
    </div>

    <el-row v-else v-loading="loading" :gutter="16">
      <el-col v-for="item in pagedList" :key="item.id" :xs="24" :sm="12">
        <el-card class="record-card" shadow="always">
          <div class="card-head">
            <div class="id-line">{{ item.id }}</div>
            <el-tag size="mini" class="record-type-tag">{{ item.record_type || '-' }}</el-tag>
          </div>
          <div class="line"><i class="el-icon-date" /> {{ item.created || '-' }}</div>
          <div class="line"><i class="el-icon-office-building" /> {{ item.hospital_name || doctorOf(item).hospital_name || '-' }}</div>
          <div class="line"><i class="el-icon-user" /> 主诉：{{ item.chief_complaint || item.symptom_description || '-' }}</div>
          <div class="line"><i class="el-icon-help" /> 诊断：{{ item.diagnosis_result || item.doctor_diagnosis || item.diagnosis || '-' }}</div>
          <div class="card-foot">
            <span>{{ item.file_hash ? '有附件' : '无附件' }}</span>
            <el-button type="primary" size="mini" @click="openDetail(item)">查看详情</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="pagination-wrap" v-if="filteredList.length > 0">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="filteredList.length"
        :current-page.sync="pager.page"
        :page-size="pager.pageSize"
      />
    </div>

    <el-dialog :visible.sync="detailVisible" width="820px" custom-class="record-detail-dialog">
      <div slot="title" class="dialog-title-wrap">
        <i class="el-icon-document" />
        <span>病历详情</span>
      </div>
      <div v-if="detailItem" class="detail-body">
        <div class="tabs-line">
          <div v-for="(tab, idx) in detailTabs" :key="tab.key" class="dtab" :class="{ active: detailTab === idx }" @click="detailTab = idx">{{ tab.label }}</div>
        </div>

        <el-card v-show="detailTab === 0" shadow="never" class="dcard dcard-blue">
          <div slot="header" class="dheader"><i class="el-icon-user" /> 患者基本信息</div>
          <el-row :gutter="10" class="dgrid">
            <el-col :span="8"><div class="kv"><span>患者姓名</span><b>{{ detailItem.patient_name || patientOf(detailItem).account_name || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>患者ID</span><b>{{ detailItem.patient || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>性别</span><b>{{ detailItem.patient_gender || patientOf(detailItem).gender || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>年龄</span><b>{{ detailItem.patient_age || patientOf(detailItem).age || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>身份证号</span><b>{{ detailItem.patient_id_card_no || patientOf(detailItem).id_card_no || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>医保卡号</span><b>{{ detailItem.insurance_card_no || patientOf(detailItem).insurance_card_no || '-' }}</b></div></el-col>
          </el-row>
        </el-card>

        <el-card v-show="detailTab === 0" shadow="never" class="dcard dcard-green">
          <div slot="header" class="dheader"><i class="el-icon-office-building" /> 就诊信息</div>
          <el-row :gutter="10" class="dgrid">
            <el-col :span="8"><div class="kv"><span>就诊医院</span><b>{{ detailItem.hospital_name || doctorOf(detailItem).hospital_name || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>就诊科室</span><b>{{ detailItem.department || doctorOf(detailItem).department || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>接诊医生</span><b>{{ detailItem.visit_doctor_name || doctorOf(detailItem).account_name || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>病历类型</span><b>{{ detailItem.record_type || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>就诊时间</span><b>{{ detailItem.created || '-' }}</b></div></el-col>
          </el-row>
        </el-card>

        <el-card v-show="detailTab === 1" shadow="never" class="dcard dcard-pink">
          <div slot="header" class="dheader"><i class="el-icon-edit-outline" /> 诊断信息</div>
          <div class="block"><label>主诉</label><div>{{ detailItem.chief_complaint || detailItem.symptom_description || '-' }}</div></div>
          <div class="block"><label>现病史</label><div>{{ detailItem.present_illness || '-' }}</div></div>
          <div class="block"><label>既往史</label><div>{{ detailItem.past_history || '-' }}</div></div>
          <div class="block"><label>过敏史</label><div>{{ detailItem.allergy_history || '-' }}</div></div>
          <div class="block"><label>家族史</label><div>{{ detailItem.family_history || '-' }}</div></div>
        </el-card>

        <el-card v-show="detailTab === 2" shadow="never" class="dcard dcard-teal">
          <div slot="header" class="dheader"><i class="el-icon-data-analysis" /> 检查治疗</div>
          <div class="vitals">
            <div class="vital"><span>体温</span><b>{{ detailItem.temperature || '-' }}℃</b></div>
            <div class="vital"><span>脉搏</span><b>{{ detailItem.pulse || '-' }}次/分</b></div>
            <div class="vital"><span>血压</span><b>{{ detailItem.blood_pressure || '-' }}</b></div>
            <div class="vital"><span>呼吸</span><b>{{ detailItem.respiration || '-' }}次/分</b></div>
          </div>
          <div class="block"><label>体格检查</label><div>{{ detailItem.physical_exam || '-' }}</div></div>
          <div class="block"><label>实验室检查</label><div>{{ detailItem.lab_exam || '-' }}</div></div>
          <div class="block"><label>影像学检查</label><div>{{ detailItem.imaging_exam || '-' }}</div></div>
          <div class="block"><label>诊断结果</label><div>{{ detailItem.diagnosis_result || detailItem.doctor_diagnosis || detailItem.diagnosis || '-' }}</div></div>
          <div class="block"><label>治疗方案</label><div>{{ detailItem.treatment_plan || '-' }}</div></div>
          <div class="block"><label>处方用药</label><div>{{ detailItem.medication_advice || '-' }}</div></div>
          <div class="block"><label>医嘱</label><div>{{ detailItem.doctor_advice || '-' }}</div></div>
        </el-card>

        <el-card v-show="detailTab === 3" shadow="never" class="dcard dcard-amber">
          <div slot="header" class="dheader"><i class="el-icon-folder-opened" /> 附件资料</div>
          <div class="block"><label>文件名</label><div>{{ detailItem.file_name || '-' }}</div></div>
          <div class="block"><label>文件哈希</label><div>{{ detailItem.file_hash || '-' }}</div></div>
          <div class="preview-row" v-if="detailItem.file_path">
            <el-button type="primary" plain size="mini" @click="openSecondPasswordDialog">病历文件详情</el-button>
          </div>
        </el-card>
      </div>
      <span slot="footer">
        <el-button @click="detailVisible = false">关闭</el-button>
      </span>
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
      detailTab: 0,
      detailTabs: [
        { key: 'base', label: '基本信息' },
        { key: 'diag', label: '诊断信息' },
        { key: 'check', label: '检查治疗' },
        { key: 'file', label: '附件资料' }
      ],
      previewUrl: '',
      currentPreviewFileName: '',
      secondPassword: '',
      filters: {
        keyword: '',
        hospital: '',
        recordType: ''
      },
      pager: {
        page: 1,
        pageSize: 3
      }
    }
  },
  computed: {
    ...mapGetters(['account_id', 'roles', 'account_name']),
    filteredList() {
      const f = this.filters
      return this.prescriptionList.filter(i => {
        const hospital = i.hospital_name || this.doctorOf(i).hospital_name || ''
        const byHospital = !f.hospital || hospital === f.hospital
        const byType = !f.recordType || i.record_type === f.recordType
        const text = `${i.id || ''} ${i.chief_complaint || ''} ${i.symptom_description || ''} ${i.diagnosis_result || ''} ${i.doctor_diagnosis || ''}`.toLowerCase()
        const byKeyword = !f.keyword || text.includes(String(f.keyword).toLowerCase())
        return byHospital && byType && byKeyword
      }).sort((a, b) => String(b.created || '').localeCompare(String(a.created || '')))
    },
    pagedList() {
      const start = (this.pager.page - 1) * this.pager.pageSize
      return this.filteredList.slice(start, start + this.pager.pageSize)
    },
    hospitalOptions() {
      const s = new Set()
      this.prescriptionList.forEach(i => {
        const h = i.hospital_name || this.doctorOf(i).hospital_name
        if (h) s.add(h)
      })
      return Array.from(s)
    }
  },
  watch: {
    filters: {
      deep: true,
      handler() {
        this.pager.page = 1
      }
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
    doctorOf(item) {
      return this.accountList.find(a => a.account_id === item.doctor) || {}
    },
    patientOf(item) {
      return this.accountList.find(a => a.account_id === item.patient) || {}
    },
    openDetail(item) {
      this.detailItem = item
      this.detailTab = 0
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
.mine-page { width: 100%; min-height: 100%; background: linear-gradient(120deg, #eef4ff, #f7f9fc); padding: 16px; }
.page-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.head-left { display: flex; align-items: center; gap: 10px; }
.head-icon { width: 34px; height: 34px; border-radius: 10px; color: #fff; background: #295dff; display: flex; align-items: center; justify-content: center; }
.head-title { font-size: 22px; font-weight: 700; color: #111827; }
.head-subtitle { font-size: 12px; color: #6b7280; }
.head-right { text-align: right; }
.mini-label { font-size: 11px; color: #6b7280; }
.mini-value { font-weight: 700; color: #111827; }
.filter-card { margin-bottom: 14px; border-radius: 12px; }
.filter-row { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
.w220 { width: 220px; }
.w280 { width: 280px; }
.record-card { border-radius: 12px; margin-bottom: 14px; min-height: 190px; }
.card-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.id-line { font-size: 18px; font-weight: 700; color: #111827; }
.record-type-tag { background: #dbeafe; color: #1d4ed8; border: none; }
.line { color: #4b5563; font-size: 12px; margin: 8px 0; }
.line i { color: #64748b; margin-right: 4px; }
.card-foot { margin-top: 10px; display: flex; justify-content: space-between; align-items: center; color: #9ca3af; font-size: 12px; }
.pagination-wrap { margin-top: 6px; display: flex; justify-content: center; }
.empty-wrap { background: #fff; border-radius: 12px; padding: 30px 0; }
.dialog-title-wrap { display: flex; align-items: center; gap: 8px; font-size: 16px; font-weight: 700; }
.tabs-line { display: grid; grid-template-columns: repeat(4, 1fr); gap: 6px; background: #f3f4f6; border-radius: 9px; padding: 4px; margin-bottom: 10px; }
.dtab { text-align: center; font-size: 12px; line-height: 30px; border-radius: 8px; cursor: pointer; }
.dtab.active { background: #fff; font-weight: 700; color: #111827; }
.dcard { border-radius: 10px; margin-bottom: 10px; }
.dheader { font-size: 13px; font-weight: 700; }
.dcard-blue .dheader { color: #1e40af; }
.dcard-green .dheader { color: #166534; }
.dcard-pink .dheader { color: #7e22ce; }
.dcard-teal .dheader { color: #0f766e; }
.dcard-amber .dheader { color: #b45309; }
.kv { margin-bottom: 10px; }
.kv span { display: block; font-size: 11px; color: #6b7280; }
.kv b { font-size: 12px; color: #111827; }
.block { margin-bottom: 10px; }
.block label { display: block; font-size: 12px; color: #374151; margin-bottom: 4px; font-weight: 600; }
.block > div { border: 1px solid #e5e7eb; border-radius: 8px; padding: 8px; font-size: 12px; color: #111827; background: #fafafa; }
.vitals { display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px; margin-bottom: 10px; }
.vital { border: 1px solid #e5e7eb; border-radius: 8px; text-align: center; padding: 8px; }
.vital span { display: block; color: #6b7280; font-size: 11px; }
.vital b { font-size: 13px; color: #111827; }
.preview-row { margin-top: 12px; text-align: right; }
.preview-wrap { min-height: 380px; }
.preview-image { width: 100%; max-height: 70vh; object-fit: contain; }
.preview-frame { width: 100%; height: 70vh; border: none; }
.empty-preview { color: #909399; text-align: center; padding: 40px 0; }
</style>