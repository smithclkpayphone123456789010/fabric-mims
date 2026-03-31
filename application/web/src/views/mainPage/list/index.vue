<template>
  <div class="home-container">
    <!-- 顶部欢迎区域 -->
    <div class="welcome-section">
      <div class="welcome-info">
        <h1 class="welcome-title">您好，{{ account_name }}</h1>
        <p class="welcome-sub">欢迎使用区块链医院管理系统，祝您身体健康！</p>
      </div>
      <div class="welcome-time">
        <span class="date">{{ currentDate }}</span>
        <span class="weekday">{{ weekday }}</span>
      </div>
    </div>

    <!-- 快捷功能入口 -->
    <div class="section-title">
      <i class="el-icon-menu"></i>
      <span>快捷功能</span>
    </div>
    <div class="quick-actions">
      <div class="action-card" v-for="action in quickActions" :key="action.path" @click="$router.push(action.path)">
        <div class="action-icon" :style="{ background: action.bgColor }">
          <i :class="action.icon"></i>
        </div>
        <div class="action-text">
          <span class="action-title">{{ action.title }}</span>
          <span class="action-desc">{{ action.desc }}</span>
        </div>
      </div>
    </div>

    <!-- 健康数据概览 -->
    <div class="section-title">
      <i class="el-icon-data-line"></i>
      <span>健康数据概览</span>
    </div>
    <div class="overview-cards">
      <div class="overview-card" v-for="item in overviewData" :key="item.label" :class="item.type">
        <div class="overview-icon">
          <i :class="item.icon"></i>
        </div>
        <div class="overview-info">
          <span class="overview-value">{{ item.value }}</span>
          <span class="overview-label">{{ item.label }}</span>
        </div>
        <div class="overview-trend" v-if="item.trend">
          <i :class="item.trend > 0 ? 'el-icon-top' : 'el-icon-bottom'" :style="{ color: item.trend > 0 ? '#67c23a' : '#f56c6c' }"></i>
          <span :style="{ color: item.trend > 0 ? '#67c23a' : '#f56c6c' }">{{ Math.abs(item.trend) }}%</span>
        </div>
      </div>
    </div>

    <!-- 消息通知 -->
    <div class="section-title">
      <i class="el-icon-bell"></i>
      <span>待办事项</span>
      <span class="notice-count" v-if="noticeList.length > 0">{{ noticeList.length }}</span>
    </div>
    <div class="notice-list">
      <div class="notice-item" v-for="notice in noticeList" :key="notice.id" @click="handleNotice(notice)">
        <div class="notice-icon" :class="notice.type">
          <i :class="notice.icon"></i>
        </div>
        <div class="notice-content">
          <span class="notice-title">{{ notice.title }}</span>
          <span class="notice-desc">{{ notice.desc }}</span>
        </div>
        <div class="notice-action">
          <el-tag size="mini" :type="notice.status === 'pending' ? 'warning' : 'success'">
            {{ notice.status === 'pending' ? '待处理' : '已完成' }}
          </el-tag>
        </div>
      </div>
      <div class="notice-empty" v-if="noticeList.length === 0">
        <i class="el-icon-circle-check"></i>
        <span>暂无待办事项</span>
      </div>
    </div>

    <!-- 健康档案摘要 -->
    <div class="section-title">
      <i class="el-icon-user"></i>
      <span>健康档案</span>
    </div>
    <div class="health-profile">
      <div class="profile-header">
        <div class="profile-avatar">
          <i class="el-icon-user-solid"></i>
        </div>
        <div class="profile-info">
          <span class="profile-name">{{ account_name }}</span>
          <span class="profile-role">{{ roleText }}</span>
        </div>
        <el-button type="primary" plain size="small" @click="$router.push('/ai-health-assistant')">
          <i class="el-icon-chat-dot-round"></i> AI健康助手
        </el-button>
      </div>
      <div class="profile-content">
        <div class="profile-item" v-for="item in profileItems" :key="item.label">
          <span class="profile-label">{{ item.label }}</span>
          <span class="profile-value">{{ item.value }}</span>
        </div>
      </div>
    </div>

    <!-- 温馨提示 -->
    <div class="tips-section">
      <div class="tips-header">
        <i class="el-icon-info"></i>
        <span>健康提示</span>
      </div>
      <div class="tips-content">
        <div class="tip-item" v-for="tip in healthTips" :key="tip">
          <i class="el-icon-medal"></i>
          <span>{{ tip }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { queryPaymentList } from '@/api/outpatient'
import { queryRegistrationList } from '@/api/outpatient'
import { queryQueueCurrent } from '@/api/outpatient'

export default {
  name: 'HomePage',
  data() {
    return {
      currentDate: '',
      weekday: '',
      overviewData: [
        { label: '待缴费', value: 0, icon: 'el-icon-wallet', type: 'warning', trend: null },
        { label: '排队中', value: 0, icon: 'el-icon-timer', type: 'info', trend: null },
        { label: '未读消息', value: 0, icon: 'el-icon-message', type: 'primary', trend: null },
        { label: '最近就诊', value: 0, icon: 'el-icon-first-aid-kit', type: 'success', trend: null }
      ],
      noticeList: [],
      profileItems: [
        { label: '账户ID', value: '' },
        { label: '用户角色', value: '' },
        { label: '最近就诊', value: '暂无记录' },
        { label: '健康状态', value: '良好' }
      ],
      healthTips: [
        '定期体检是预防疾病的重要手段，建议每年进行一次全面体检',
        '保持规律作息，每天保证7-8小时睡眠时间',
        '均衡饮食，多吃蔬菜水果，少吃油腻辛辣食物',
        '适度运动，每周保持3-5次有氧运动'
      ]
    }
  },
  computed: {
    ...mapGetters([
      'account_id',
      'account_name',
      'roles'
    ]),
    roleText() {
      const roleMap = {
        'admin': '管理员',
        'doctor': '医生',
        'patient': '患者',
        'drugstore': '药店',
        'insurance': '保险机构'
      }
      const role = this.roles && this.roles.length > 0 ? this.roles[0] : 'patient'
      return roleMap[role] || '用户'
    },
    quickActions() {
      const role = this.roles && this.roles.length > 0 ? this.roles[0] : 'patient'
      const actions = {
        patient: [
          { title: '挂号预约', desc: '快速预约门诊', icon: 'el-icon-s-order', path: '/outpatient/register', bgColor: '#e8f4ff' },
          { title: '我的预约', desc: '查看预约记录', icon: 'el-icon-tickets', path: '/outpatient/my-registration', bgColor: '#fef0e8' },
          { title: '缴费管理', desc: '待缴费用', icon: 'el-icon-wallet', path: '/outpatient/payment', bgColor: '#e8fdf0' },
          { title: '排队叫号', desc: '实时排队', icon: 'el-icon-timer', path: '/outpatient/queue', bgColor: '#fff7e8' }
        ],
        doctor: [
          { title: '就诊队列', desc: '当前患者', icon: 'el-icon-user', path: '/outpatient/doctor-queue', bgColor: '#e8f4ff' },
          { title: '我的病历', desc: '查看病历', icon: 'el-icon-document', path: '/prescription/mine', bgColor: '#e8fdf0' },
          { title: '新增病历', desc: '创建病历', icon: 'el-icon-plus', path: '/prescription/add', bgColor: '#fef0e8' },
          { title: 'AI健康助手', desc: '智能问诊', icon: 'el-icon-chat-dot-round', path: '/ai-health-assistant', bgColor: '#f3e8ff' }
        ],
        admin: [
          { title: '门诊数据', desc: '数据统计', icon: 'el-icon-data-analysis', path: '/outpatient/statistics', bgColor: '#e8f4ff' },
          { title: '账号管理', desc: '账户列表', icon: 'el-icon-user', path: '/account/all', bgColor: '#e8fdf0' },
          { title: '审计监控', desc: '查看日志', icon: 'el-icon-view', path: '/audit/search', bgColor: '#fef0e8' },
          { title: 'AI健康助手', desc: '智能问诊', icon: 'el-icon-chat-dot-round', path: '/ai-health-assistant', bgColor: '#f3e8ff' }
        ]
      }
      return actions[role] || actions.patient
    }
  },
  created() {
    this.updateDateTime()
    this.loadData()
    this.initProfile()
  },
  methods: {
    updateDateTime() {
      const now = new Date()
      const weeks = ['星期日', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']
      this.weekday = weeks[now.getDay()]
      this.currentDate = `${now.getFullYear()}年${now.getMonth() + 1}月${now.getDate()}日`
    },
    async loadData() {
      // 根据角色加载不同数据
      const role = this.roles && this.roles.length > 0 ? this.roles[0] : 'patient'
      
      if (role === 'patient' || role === 'admin') {
        // 加载待缴费数据
        try {
          const payments = await queryPaymentList({ patient_id: this.account_id })
          this.overviewData[0].value = payments?.length || 0
        } catch (e) {}
        
        // 加载排队数据
        try {
          const queues = await queryQueueCurrent({ patient_id: this.account_id })
          this.overviewData[1].value = queues ? 1 : 0
        } catch (e) {}
        
        // 加载预约数据
        try {
          const registrations = await queryRegistrationList({ patient_id: this.account_id })
          this.overviewData[3].value = registrations?.length || 0
          this.buildNotices(registrations)
        } catch (e) {}
      }
      
      if (role === 'doctor') {
        // 加载就诊队列
        try {
          const queues = await queryQueueCurrent({ doctor_id: this.account_id })
          this.overviewData[1].value = queues ? 1 : 0
        } catch (e) {}
      }
    },
    buildNotices(registrations) {
      const notices = []
      if (registrations && registrations.length > 0) {
        registrations.forEach(reg => {
          if (reg.status === 'pending') {
            notices.push({
              id: reg.id,
              title: '待就诊提醒',
              desc: `您有预约：${reg.department || '门诊'} ${reg.doctor_name || ''} ${reg.date || ''}`,
              type: 'warning',
              icon: 'el-icon-s-order',
              status: 'pending'
            })
          }
        })
      }
      
      if (this.overviewData[0].value > 0) {
        notices.push({
          id: 'payment',
          title: '待缴费提醒',
          desc: `您有 ${this.overviewData[0].value} 项待缴费项目`,
          type: 'danger',
          icon: 'el-icon-wallet',
          status: 'pending'
        })
      }
      
      if (this.overviewData[1].value > 0) {
        notices.push({
          id: 'queue',
          title: '排队提醒',
          desc: `当前排队人数：${this.overviewData[1].value} 人`,
          type: 'info',
          icon: 'el-icon-timer',
          status: 'pending'
        })
      }
      
      this.noticeList = notices.slice(0, 5)
    },
    initProfile() {
      this.profileItems[0].value = this.account_id || '-'
      this.profileItems[1].value = this.roleText
    },
    handleNotice(notice) {
      if (notice.id === 'payment') {
        this.$router.push('/outpatient/payment')
      } else if (notice.id === 'queue') {
        this.$router.push('/outpatient/queue')
      } else if (notice.id === 'registration') {
        this.$router.push('/outpatient/my-registration')
      }
    }
  }
}
</script>

<style scoped>
.home-container {
  padding: 20px 24px;
  background: #f5f7fa;
  min-height: calc(100vh - 84px);
}

/* 欢迎区域 */
.welcome-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, #10b981, #059669);
  border-radius: 16px;
  padding: 24px 32px;
  margin-bottom: 24px;
  color: #fff;
}

.welcome-title {
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 8px 0;
}

.welcome-sub {
  font-size: 14px;
  opacity: 0.9;
  margin: 0;
}

.welcome-time {
  text-align: right;
}

.welcome-time .date {
  display: block;
  font-size: 18px;
  font-weight: 600;
}

.welcome-time .weekday {
  font-size: 14px;
  opacity: 0.8;
}

/* 区块标题 */
.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 16px;
  margin-top: 24px;
}

.section-title i {
  color: #10b981;
  font-size: 18px;
}

.notice-count {
  background: #f56c6c;
  color: #fff;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: normal;
}

/* 快捷功能 */
.quick-actions {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.action-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  cursor: pointer;
  transition: all 0.3s;
  border: 1px solid #ebeef5;
}

.action-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.action-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  color: #333;
}

.action-text {
  display: flex;
  flex-direction: column;
}

.action-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.action-desc {
  font-size: 12px;
  color: #909399;
}

/* 数据概览 */
.overview-cards {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.overview-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 16px;
  border: 1px solid #ebeef5;
}

.overview-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
}

.overview-card.warning .overview-icon {
  background: #fef0e8;
  color: #e6a23c;
}

.overview-card.info .overview-icon {
  background: #e8f4ff;
  color: #409eff;
}

.overview-card.primary .overview-icon {
  background: #f3e8ff;
  color: #9c6ade;
}

.overview-card.success .overview-icon {
  background: #e8fdf0;
  color: #67c23a;
}

.overview-info {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.overview-value {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
}

.overview-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}

.overview-trend {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
}

/* 消息通知 */
.notice-list {
  background: #fff;
  border-radius: 12px;
  border: 1px solid #ebeef5;
  overflow: hidden;
}

.notice-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  cursor: pointer;
  transition: background 0.2s;
  border-bottom: 1px solid #ebeef5;
}

.notice-item:last-child {
  border-bottom: none;
}

.notice-item:hover {
  background: #f5f7fa;
}

.notice-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.notice-icon.warning {
  background: #fef0e8;
  color: #e6a23c;
}

.notice-icon.danger {
  background: #fde2e2;
  color: #f56c6c;
}

.notice-icon.info {
  background: #e8f4ff;
  color: #409eff;
}

.notice-icon.success {
  background: #e8fdf0;
  color: #67c23a;
}

.notice-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.notice-title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.notice-desc {
  font-size: 12px;
  color: #909399;
}

.notice-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 40px;
  color: #909399;
  font-size: 14px;
}

.notice-empty i {
  font-size: 24px;
  color: #67c23a;
}

/* 健康档案 */
.health-profile {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #ebeef5;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 16px;
}

.profile-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #10b981, #059669);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #fff;
}

.profile-info {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.profile-name {
  font-size: 18px;
  font-weight: 700;
  color: #303133;
}

.profile-role {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}

.profile-content {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.profile-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.profile-label {
  font-size: 13px;
  color: #909399;
  min-width: 70px;
}

.profile-value {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

/* 健康提示 */
.tips-section {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #ebeef5;
  margin-top: 24px;
  margin-bottom: 24px;
}

.tips-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 700;
  color: #303133;
  margin-bottom: 16px;
}

.tips-header i {
  color: #409eff;
}

.tips-content {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.tip-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 13px;
  color: #606266;
  line-height: 1.6;
}

.tip-item i {
  color: #67c23a;
  margin-top: 2px;
  flex-shrink: 0;
}

/* 响应式 */
@media (max-width: 1200px) {
  .quick-actions,
  .overview-cards {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .quick-actions,
  .overview-cards {
    grid-template-columns: 1fr;
  }
  
  .profile-content,
  .tips-content {
    grid-template-columns: 1fr;
  }
  
  .welcome-section {
    flex-direction: column;
    text-align: center;
    gap: 16px;
  }
  
  .welcome-time {
    text-align: center;
  }
}
</style>
