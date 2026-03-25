<template>
  <div class="mine-page">
    <div class="page-head">
      <div class="head-left">
        <div class="head-icon"><i class="el-icon-view" /></div>
        <div>
          <div class="head-title">可访问病历</div>
          <div class="head-subtitle">查看患者已授权给您的病历记录</div>
        </div>
      </div>
      <div class="head-right">
        <div class="mini-label">医生ID</div>
        <div class="mini-value">{{ account_id }}</div>
      </div>
    </div>

    <el-card class="filter-card" shadow="always">
      <div class="filter-row">
        <el-input v-model.trim="filters.patient_name_keyword" placeholder="患者姓名" class="w220" clearable>
          <i slot="prefix" class="el-input__icon el-icon-user" />
        </el-input>
        <el-input v-model.trim="filters.id_card_keyword" placeholder="身份证号" class="w220" clearable>
          <i slot="prefix" class="el-input__icon el-icon-postcard" />
        </el-input>
        <el-select v-model="filters.record_type_keyword" placeholder="类型筛选" class="w220" clearable>
          <el-option label="全部类型" value="" />
          <el-option label="门诊病历" value="门诊病历" />
          <el-option label="住院病历" value="住院病历" />
          <el-option label="急诊病历" value="急诊病历" />
          <el-option label="检查报告" value="检查报告" />
        </el-select>
        <el-date-picker
          v-model="filters.createdRange"
          type="daterange"
          range-separator="-"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          value-format="yyyy-MM-dd"
          unlink-panels
          class="w280"
        />
        <el-button type="primary" icon="el-icon-search" @click="fetchList">查询</el-button>
        <el-button icon="el-icon-refresh-left" @click="onReset">重置</el-button>
      </div>
    </el-card>

    <div v-if="pagedList.length === 0 && !loading" class="empty-wrap">
      <el-empty description="暂无可访问病历" />
    </div>

    <el-row v-else v-loading="loading" :gutter="16">
      <el-col v-for="item in pagedList" :key="item.record.id" :xs="24" :sm="12">
        <el-card class="record-card" shadow="always">
          <div class="card-head">
            <div class="id-line">{{ item.record.id }}</div>
            <el-tag size="mini" class="record-type-tag">{{ item.record.record_type || '-' }}</el-tag>
          </div>
          <div class="line"><i class="el-icon-user" /> {{ item.patient.account_name || '-' }}</div>
          <div class="line"><i class="el-icon-postcard" /> {{ item.patient.id_card_no || '-' }}</div>
          <div class="line"><i class="el-icon-date" /> {{ item.record.created || '-' }}</div>
          <div class="line"><i class="el-icon-time" /> 授权到期：{{ item.authorization.end_time || '-' }}</div>
          <div class="line"><i class="el-icon-help" /> 诊断：{{ item.record.diagnosis_result || item.record.doctor_diagnosis || item.record.diagnosis || '-' }}</div>
          <div class="card-foot">
            <span>{{ item.record.file_hash ? '有附件' : '无附件' }}</span>
            <el-button type="primary" size="mini" @click="openDetail(item)">查看详情</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <div class="pagination-wrap" v-if="currentList.length > 0">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :total="currentList.length"
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
            <el-col :span="8"><div class="kv"><span>患者姓名</span><b>{{ detailItem.record.patient_name || detailItem.patient.account_name || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>患者ID</span><b>{{ detailItem.record.patient || detailItem.patient.account_id || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>性别</span><b>{{ detailItem.record.patient_gender || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>年龄</span><b>{{ detailItem.record.patient_age || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>身份证号</span><b>{{ detailItem.record.patient_id_card_no || detailItem.patient.id_card_no || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>医保卡号</span><b>{{ detailItem.record.insurance_card_no || '-' }}</b></div></el-col>
          </el-row>
        </el-card>

        <el-card v-show="detailTab === 0" shadow="never" class="dcard dcard-green">
          <div slot="header" class="dheader"><i class="el-icon-office-building" /> 就诊信息</div>
          <el-row :gutter="10" class="dgrid">
            <el-col :span="8"><div class="kv"><span>就诊医院</span><b>{{ detailItem.record.hospital_name || detailItem.authorization.hospital_name || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>就诊科室</span><b>{{ detailItem.record.department || detailItem.authorization.department || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>接诊医生</span><b>{{ detailItem.record.visit_doctor_name || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>病历类型</span><b>{{ detailItem.record.record_type || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>就诊时间</span><b>{{ detailItem.record.created || '-' }}</b></div></el-col>
            <el-col :span="8"><div class="kv"><span>授权到期</span><b>{{ detailItem.authorization.end_time || '-' }}</b></div></el-col>
          </el-row>
        </el-card>

        <el-card v-show="detailTab === 1" shadow="never" class="dcard dcard-pink">
          <div slot="header" class="dheader"><i class="el-icon-edit-outline" /> 诊断信息</div>
          <div class="block"><label>主诉</label><div>{{ detailItem.record.chief_complaint || detailItem.record.symptom_description || '-' }}</div></div>
          <div class="block"><label>现病史</label><div>{{ detailItem.record.present_illness || '-' }}</div></div>
          <div class="block"><label>既往史</label><div>{{ detailItem.record.past_history || '-' }}</div></div>
          <div class="block"><label>过敏史</label><div>{{ detailItem.record.allergy_history || '-' }}</div></div>
          <div class="block"><label>家族史</label><div>{{ detailItem.record.family_history || '-' }}</div></div>
        </el-card>

        <el-card v-show="detailTab === 2" shadow="never" class="dcard dcard-teal">
          <div slot="header" class="dheader"><i class="el-icon-data-analysis" /> 检查治疗</div>
          <div class="vitals">
            <div class="vital"><span>体温</span><b>{{ detailItem.record.temperature || '-' }}℃</b></div>
            <div class="vital"><span>脉搏</span><b>{{ detailItem.record.pulse || '-' }}次/分</b></div>
            <div class="vital"><span>血压</span><b>{{ detailItem.record.blood_pressure || '-' }}</b></div>
            <div class="vital"><span>呼吸</span><b>{{ detailItem.record.respiration || '-' }}次/分</b></div>
          </div>
          <div class="block"><label>体格检查</label><div>{{ detailItem.record.physical_exam || '-' }}</div></div>
          <div class="block"><label>实验室检查</label><div>{{ detailItem.record.lab_exam || '-' }}</div></div>
          <div class="block"><label>影像学检查</label><div>{{ detailItem.record.imaging_exam || '-' }}</div></div>
          <div class="block"><label>诊断结果</label><div>{{ detailItem.record.diagnosis_result || detailItem.record.doctor_diagnosis || detailItem.record.diagnosis || '-' }}</div></div>
          <div class="block"><label>治疗方案</label><div>{{ detailItem.record.treatment_plan || '-' }}</div></div>
          <div class="block"><label>处方用药</label><div>{{ detailItem.record.medication_advice || '-' }}</div></div>
          <div class="block"><label>医嘱</label><div>{{ detailItem.record.doctor_advice || '-' }}</div></div>
        </el-card>

        <el-card v-show="detailTab === 3" shadow="never" class="dcard dcard-amber">
          <div slot="header" class="dheader"><i class="el-icon-folder-opened" /> 附件资料</div>
          <div class="block"><label>文件名</label><div>{{ detailItem.record.file_name || '-' }}</div></div>
          <div class="block"><label>文件哈希</label><div>{{ detailItem.record.file_hash || '-' }}</div></div>
          <div class="preview-row" v-if="detailItem.record.file_path">
            <el-button type="primary" plain size="mini" @click="openFilePreview(detailItem.record)">病历文件详情</el-button>
          </div>
        </el-card>
      </div>
      <span slot="footer">
        <el-button @click="detailVisible = false">关闭</el-button>
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
      detailTab: 0,
      detailTabs: [
        { key: 'base', label: '基本信息' },
        { key: 'diag', label: '诊断信息' },
        { key: 'check', label: '检查治疗' },
        { key: 'file', label: '附件资料' }
      ],
      previewUrl: '',
      currentPreviewFileName: '',
      pager: {
        page: 1,
        pageSize: 4
      },
      filters: {
        patient_name_keyword: '',
        id_card_keyword: '',
        record_type_keyword: '',
        createdRange: []
      }
    }
  },
  computed: {
    ...mapGetters(['account_id', 'account_name']),
    pagedList() {
      const start = (this.pager.page - 1) * this.pager.pageSize
      return this.currentList.slice(start, start + this.pager.pageSize)
    }
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
        this.currentList = (data || []).sort((a, b) => String((b.record && b.record.created) || '').localeCompare(String((a.record && a.record.created) || '')))
        this.pager.page = 1
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
        this.detailTab = 0
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
.record-card { border-radius: 12px; margin-bottom: 14px; min-height: 200px; }
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