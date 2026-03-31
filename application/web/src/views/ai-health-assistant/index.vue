<template>
  <div class="ai-page">
    <div class="ai-wrap">
      <el-card class="session-card" shadow="never">
        <div class="session-title">会话历史</div>
        <div class="session-sub">您的咨询记录</div>
        <div class="session-list">
          <div
            v-for="s in sessions"
            :key="s.session_id"
            class="session-item"
            :class="{ active: s.session_id === sessionId }"
            @click="switchSession(s.session_id)"
          >
            <div class="name">{{ s.title || '新会话' }}</div>
            <div class="sub">{{ s.updated_at || '-' }}</div>
          </div>
        </div>
      </el-card>

      <div class="main-col">
        <el-card class="chat-card" shadow="never">
          <div class="head">
            <div>
              <h3>AI健康助手</h3>
              <p>问症状、看流程、做导诊</p>
            </div>
          </div>

          <div class="chat-body" ref="chatBody">
            <div v-for="m in messages" :key="m.id" class="msg" :class="m.role==='user'?'user':'assistant'">
              <div v-if="m.role==='assistant'" class="avatar">AI</div>
              <div class="bubble">
                <template v-if="m.role==='assistant' && m.isLoading">
                  <span class="typing-indicator">
                    AI正在思考
                    <span class="ellipsis" aria-hidden="true">
                      <i class="dot dot1" />
                      <i class="dot dot2" />
                      <i class="dot dot3" />
                    </span>
                  </span>
                </template>
                <template v-else>
                  {{ m.content }}
                </template>
              </div>
            </div>
          </div>

          <div class="quick-row">
            <span class="quick-label">快捷咨询：</span>
            <el-button size="mini" :type="focusMode==='triage' ? 'primary' : 'default'" :plain="focusMode!=='triage'" @click="toggleMode('triage')">症状导诊</el-button>
            <el-button size="mini" :type="focusMode==='rehab' ? 'primary' : 'default'" :plain="focusMode!=='rehab'" @click="toggleMode('rehab')">个性化康复伴行</el-button>
            <el-button size="mini" :type="focusMode==='translator' ? 'primary' : 'default'" :plain="focusMode!=='translator'" @click="toggleMode('translator')">检查报告翻译官</el-button>
            <el-button size="mini" type="success" :plain="!isClinicMode" @click="toggleClinicMode">
              <i class="el-icon-plus" /> AI诊室
            </el-button>
          </div>

          <div class="guide-card">
            <div class="guide-title">
              <span>💡 功能说明与提问引导</span>
              <i class="el-icon-arrow-up" />
            </div>
            <div v-if="focusMode==='triage'" class="guide-text">
              <b>症状导诊 · 功能作用：</b><br>
              根据你描述的症状，评估风险等级，并给出可执行建议与推荐就诊科室，帮助你更快判断“该挂哪个科、是否需要尽快就医”。<br><br>
              <b>建议这样提问：</b><br>
              1) 主要症状（例如：胸闷、发热、头晕）<br>
              2) 持续时间（多久了）<br>
              3) 严重程度（轻/中/重，是否加重）<br>
              4) 伴随症状（如气短、恶心）<br>
              5) 既往病史与用药（如高血压、糖尿病）<br>
              6) 特殊情况（孕期、过敏史、近期手术）
            </div>
            <div v-else-if="focusMode==='rehab'" class="guide-text">
              <b>个性化康复伴行 · 功能作用：</b><br>
              基于你选择的病历，生成用药提醒、复诊建议和注意事项说明，帮助你在就诊后持续跟踪恢复情况。<br><br>
              <b>建议这样提问：</b><br>
              1) 当前最不舒服的症状<br>
              2) 症状持续多久、是否加重<br>
              3) 当前正在服用的药物与频次<br>
              4) 最近一次复诊/检查时间<br>
              5) 你最关心的问题（如“什么时候复诊最合适”）
            </div>
            <div v-else-if="focusMode==='translator'" class="guide-text">
              <b>检查报告“翻译官” · 功能作用：</b><br>
              对你选定病历中的诊断信息、检查结果和治疗方案进行非诊断型大白话解释，帮你看懂报告在说什么。<br><br>
              <b>建议这样提问：</b><br>
              1) 你想重点看哪部分（诊断结果/处方用药/治疗方案）<br>
              2) 哪些术语看不懂<br>
              3) 你最关心的问题（例如“严重吗？要不要马上复查？”）<br>
              4) 是否有医生当面给过补充说明
            </div>
            <div v-else class="guide-text normal-chat">
              当前未选中快捷功能，将按普通AI对话模式回复。<br>
              你可以直接输入任何问题与AI交流，也可再次点击上方按钮启用对应功能。
            </div>

            <div class="example-head">
              <span>✍️ 示例：</span>
              <el-button v-if="focusMode!=='triage'" size="mini" @click="openRecordDialog">选择病历</el-button>
            </div>
            <div class="example-text" v-if="focusMode==='triage'">“我发热38.5℃两天，伴随咳嗽和喉咙痛，夜间加重，无胸痛，既往有哮喘，正在用布地奈德。请问建议挂什么科？”</div>
            <div class="example-text" v-else-if="focusMode==='rehab'">请在下方输入框直接输入你的康复问题，例如：“我最关心夜间咳嗽和复诊时间”。</div>
            <div class="example-text" v-else>请在下方输入框直接输入你的翻译问题，例如：“请重点解释诊断结果+处方用药+治疗方案”。</div>

            <div class="guide-actions" v-if="focusMode!=='triage'">
              <span class="chosen" v-if="selectedRecordId">已选病历：{{ selectedRecordId }}</span>
              <span class="chosen warn" v-else>请先选择病历</span>
            </div>
          </div>

          <div class="chat-foot">
            <el-input
              type="textarea"
              :rows="3"
              v-model="input"
              placeholder="描述您的症状或问题...（按 Enter 发送，Shift + Enter 换行）"
              @keyup.enter.native="onEnterSend"
            />
            <div class="foot-actions">
              <el-button @click="manualRefresh">刷新会话</el-button>
              <el-button
                v-if="chatStatus==='loading' || chatStatus==='streaming'"
                type="danger"
                size="mini"
                @click="stopGenerate"
              >停止生成</el-button>
              <el-button
                v-else
                size="mini"
                :disabled="!canRegenerate"
                @click="regenerate"
              >重新生成</el-button>
              <el-button type="primary" :loading="chatStatus==='loading'" :disabled="isBusy" @click="send">发送</el-button>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <el-dialog :visible.sync="recordDialogVisible" width="720px" custom-class="record-dialog" :close-on-click-modal="false">
      <div slot="title" class="record-title">选择病历</div>
      <div class="record-sub">请选择一份病历，系统将自动提取相关信息用于AI分析</div>
      <div class="record-list">
        <div v-for="r in records" :key="r.id" class="record-item" :class="{ active: selectedRecordId===r.id }" @click="selectRecord(r.id)">
          <div class="ri-head">
            <div class="ri-name">门诊病历</div>
            <div class="ri-time">{{ r.create_time || r.created_time || '-' }}</div>
          </div>
          <div class="ri-line"><b>就诊原因：</b>{{ r.chief_complaint || '-' }}</div>
          <div class="ri-line"><b>诊断结果：</b>{{ r.diagnosis_result || r.doctor_diagnosis || '-' }}</div>
          <div class="ri-line"><b>处方：</b>{{ r.medication_advice || '-' }}</div>
        </div>
      </div>
      <span slot="footer">
        <el-button @click="recordDialogVisible=false">取消</el-button>
        <el-button type="primary" @click="recordDialogVisible=false">确认</el-button>
      </span>
    </el-dialog>

    <!-- AI诊室气泡框 -->
    <div v-if="showClinicDialog" class="clinic-overlay" @click.self="closeClinicDialog">
      <div class="clinic-dialog">
        <div class="clinic-header">
          <div class="clinic-title">
            <i class="el-icon-s-help" /> AI智能诊室
          </div>
          <div class="clinic-subtitle" v-if="!clinicComplete">您已进入AI诊室，系统将引导您完成健康咨询</div>
          <div class="clinic-subtitle" v-else>
            <span class="complete-tag"><i class="el-icon-circle-check" /> 问诊已完成</span>
          </div>
          <el-button class="clinic-close" type="text" @click="closeClinicDialog"><i class="el-icon-close" /></el-button>
        </div>

        <!-- 进度条 -->
        <div class="clinic-progress" v-if="!clinicComplete">
          <div class="progress-info">
            <span class="step-name">{{ clinicStepName }}</span>
            <span class="step-count">第 {{ clinicCurrentStep }}/6 轮</span>
          </div>
          <el-progress :percentage="clinicProgress" :show-text="false" :stroke-width="8" color="#10b981"></el-progress>
        </div>

        <!-- 问诊流程说明 -->
        <div class="clinic-steps" v-if="!clinicComplete && clinicCurrentStep === 0">
          <div class="steps-title">问诊流程</div>
          <div class="step-item" v-for="(step, idx) in clinicFlowSteps" :key="idx" :class="{ active: idx === 0, done: idx < 0 }">
            <span class="step-num">{{ idx + 1 }}</span>
            <span class="step-text">{{ step }}</span>
          </div>
          <div class="clinic-enter-hint">点击下方"开始问诊"按钮进入第1轮问询</div>
        </div>

        <!-- 问诊消息区域 -->
        <div class="clinic-messages" ref="clinicMessages">
          <div v-for="m in clinicMessages" :key="m.id" class="clinic-msg" :class="m.role">
            <div class="clinic-avatar" v-if="m.role === 'assistant'">AI</div>
            <div class="clinic-bubble">{{ m.content }}</div>
          </div>
          <div v-if="clinicLoading" class="clinic-msg assistant">
            <div class="clinic-avatar">AI</div>
            <div class="clinic-bubble typing">
              <span class="typing-indicator">
                AI正在思考
                <span class="ellipsis">
                  <i class="dot dot1" /><i class="dot dot2" /><i class="dot dot3" />
                </span>
              </span>
            </div>
          </div>
        </div>

        <!-- 诊断报告 - 优化版 -->
        <div v-if="clinicComplete" class="clinic-report">
          <!-- 主报告区域 -->
          <div class="report-main">
            <!-- 报告标题 -->
            <div class="report-header">
              <div class="report-badge">AI诊断报告已生成</div>
              <p class="report-intro">以下是基于您提供信息的详细分析报告：</p>
            </div>

            <!-- 诊断结论卡片 -->
            <div class="report-card conclusion-card" v-if="clinicReport">
              <div class="card-header">
                <i class="el-icon-document"></i>
                <span>诊断结论</span>
              </div>
              <div class="card-body">{{ clinicReport }}</div>
            </div>

            <!-- 诊断摘要卡片 -->
            <div class="report-card summary-card" v-if="clinicSummary">
              <div class="card-header">
                <i class="el-icon-notebook-2"></i>
                <span>诊断摘要</span>
              </div>
              <div class="card-body">{{ clinicSummary }}</div>
            </div>

            <!-- 建议卡片 -->
            <div class="report-card suggestion-card" v-if="clinicSuggestions && clinicSuggestions.length">
              <div class="card-header warning">
                <i class="el-icon-warning"></i>
                <span>健康建议</span>
              </div>
              <ul class="suggestion-list">
                <li v-for="(s, idx) in clinicSuggestions" :key="idx">{{ s }}</li>
              </ul>
            </div>
          </div>

          <!-- 底部操作栏 -->
          <div class="report-footer">
            <el-button type="primary" class="btn-register" @click="goToRegister">
              <i class="el-icon-s-order"></i> 立即挂号
            </el-button>
            <el-button type="success" class="btn-download" @click="downloadReport">
              <i class="el-icon-download"></i> 下载报告
            </el-button>
          </div>

          <!-- 免责声明 -->
          <div class="report-disclaimer">
            <i class="el-icon-warning"></i>
            <span>本报告由AI系统生成，仅供参考，不能替代专业医生的诊断。如症状严重或持续不缓解，请及时就医。紧急情况请拨打120急救电话。</span>
          </div>
        </div>

        <!-- 输入区域 -->
        <div class="clinic-foot" v-if="!clinicComplete">
          <el-input
            v-if="clinicCurrentStep > 0"
            type="textarea"
            :rows="2"
            v-model="clinicInput"
            :placeholder="clinicInputPlaceholder"
          />
          <div class="clinic-actions">
            <el-button v-if="clinicCurrentStep === 0" type="primary" size="medium" @click="startClinic">
              <i class="el-icon-video-play" /> 开始问诊
            </el-button>
            <template v-else>
              <el-button size="small" @click="resetClinic">重新开始</el-button>
              <el-button type="primary" :loading="clinicLoading" @click="submitClinicAnswer">
                提交回答
              </el-button>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { aiSessions, aiSessionMessages, aiTriage, aiRehabCompanion, aiReportTranslator,
  aiClinicStart, aiClinicChat, aiClinicReset, aiClinicReport } from '@/api/aiAssistant'
import { queryPrescriptionList } from '@/api/prescription'

const API_BASE = process.env.VUE_APP_BASE_API || '/api/v2'

export default {
  name: 'AIHealthAssistant',
  data() {
    return {
      sessions: [],
      sessionId: '',
      messages: [],
      input: '',
      chatStatus: 'idle', // idle / loading / streaming / done
      streamAbortController: null,
      typingTimeout: null,
      typingStopRequested: false,
      generationToken: 0,
      lastRisk: '',
      records: [],
      selectedRecordId: '',
      // 用于“重新生成”
      lastUserInput: '',
      lastFocusMode: 'triage',
      lastAssistantMessageId: '',
      lastUserMessageId: '',
      focusMode: '',
      recordDialogVisible: false,

      // AI诊室相关状态
      showClinicDialog: false,
      isClinicMode: false,
      clinicSessionId: '',
      clinicCurrentStep: 0,
      clinicStepName: '',
      clinicProgress: 0,
      clinicMessages: [],
      clinicInput: '',
      clinicLoading: false,
      clinicComplete: false,
      clinicReport: '',
      clinicSummary: '',
      clinicSuggestions: [],
      clinicFlowSteps: [
        '主诉症状收集',
        '症状详情追问',
        '过往病史询问',
        '家族遗传病史',
        '用药情况了解',
        '个人健康状况'
      ],
      clinicInputPlaceholder: '请输入您的回答...'
    }
  },
  computed: {
    ...mapGetters(['account_id']),
    isBusy() {
      return this.chatStatus === 'loading' || this.chatStatus === 'streaming'
    },
    canRegenerate() {
      return !!this.lastUserInput && !!this.lastAssistantMessageId && this.chatStatus !== 'loading' && this.chatStatus !== 'streaming'
    }
  },
  created() {
    this.refreshSessions()
    this.loadRecords()
  },
  methods: {
    onEnterSend(e) {
      if (e.shiftKey) return
      e.preventDefault()
      this.send()
    },
    toggleMode(mode) {
      this.focusMode = this.focusMode === mode ? '' : mode
    },
    scheduleScrollBottom() {
      if (this._scrollRaf) return
      this._scrollRaf = requestAnimationFrame(() => {
        this._scrollRaf = null
        this.scrollBottom()
      })
    },
    getMessageIndex(messageId) {
      return this.messages.findIndex(m => m.id === messageId)
    },
    setAssistantLoading(messageId, isLoading) {
      const idx = this.getMessageIndex(messageId)
      if (idx >= 0) this.messages[idx].isLoading = isLoading
    },
    appendAssistantContent(messageId, delta) {
      if (!delta) return
      const idx = this.getMessageIndex(messageId)
      if (idx < 0) return
      if (this.messages[idx].isLoading) this.messages[idx].isLoading = false
      this.messages[idx].content += delta
      this.scheduleScrollBottom()
    },
    typeText(messageId, fullText) {
      // 20~50ms/字 打字机效果（按 Unicode 码点逐个输出）
      return new Promise(resolve => {
        if (this.typingTimeout) clearTimeout(this.typingTimeout)
        this.typingStopRequested = false
        const idx = this.getMessageIndex(messageId)
        if (idx < 0) return resolve()

        this.chatStatus = 'streaming'
        this.messages[idx].content = ''
        this.messages[idx].isLoading = false

        const chars = Array.from(fullText || '')
        let i = 0
        const step = () => {
          if (this.typingStopRequested) {
            this.chatStatus = 'done'
            this.messages[idx].isLoading = false
            return resolve()
          }
          if (i >= chars.length) {
            this.chatStatus = 'done'
            this.messages[idx].isLoading = false
            return resolve()
          }
          this.messages[idx].content += chars[i]
          i++
          this.scheduleScrollBottom()
          const delay = Math.floor(20 + Math.random() * 31)
          this.typingTimeout = setTimeout(step, delay)
        }
        step()
      })
    },
    stopGenerate() {
      // 让所有异步分支失效（避免“停止后请求返回仍继续打字”）
      this.generationToken++

      // 1) 停止 SSE
      if (this.streamAbortController) {
        try { this.streamAbortController.abort() } catch (e) {}
        this.streamAbortController = null
      }
      // 2) 停止打字机
      if (this.typingTimeout) {
        clearTimeout(this.typingTimeout)
        this.typingTimeout = null
      }
      this.typingStopRequested = true
      if (this.lastAssistantMessageId) this.setAssistantLoading(this.lastAssistantMessageId, false)
      this.chatStatus = 'done'
    },
    async refreshSessions() {
      const res = await aiSessions().catch(() => ({ items: [] }))
      this.sessions = res.items || []
      if (!this.sessionId && this.sessions.length) this.switchSession(this.sessions[0].session_id)
    },
    async loadRecords() {
      const list = await queryPrescriptionList({ patient: this.account_id }).catch(() => [])
      this.records = list || []
      if (!this.selectedRecordId && this.records.length) this.selectedRecordId = this.records[0].id
    },
    openRecordDialog() {
      this.recordDialogVisible = true
    },
    selectRecord(id) {
      this.selectedRecordId = id
    },
    async switchSession(id) {
      this.sessionId = id
      const res = await aiSessionMessages(id).catch(() => ({ messages: [] }))
      this.messages = res.messages || []
      this.$nextTick(this.scrollBottom)
    },
    async regenerate() {
      if (!this.canRegenerate) return
      // 移除旧的 AI 消息，重新生成到一条新占位消息
      const oldIdx = this.getMessageIndex(this.lastAssistantMessageId)
      if (oldIdx >= 0) this.messages.splice(oldIdx, 1)

      const text = this.lastUserInput
      const mode = this.lastFocusMode
      this.focusMode = mode
      // 重新生成：不再插入用户消息（保持列表上下文）
      if (mode === 'triage') {
        await this.generateTriage(text, { regenerate: true })
      } else if (mode === 'rehab') {
        await this.generateRehab(text, { regenerate: true })
      } else if (mode === 'translator') {
        await this.generateTranslator(text, { regenerate: true })
      } else {
        await this.generateChat(text, { regenerate: true })
      }
    },
    async send() {
      if (!this.input.trim() || this.isBusy) return
      const text = this.input.trim()
      this.lastUserInput = text
      this.lastFocusMode = this.focusMode

      if (this.focusMode === 'triage') {
        this.lastUserInput = text
        this.input = ''
        await this.generateTriage(text, { regenerate: false })
        return
      }
      if (this.focusMode === 'rehab') {
        this.input = ''
        await this.generateRehab(text, { regenerate: false })
        return
      }
      if (this.focusMode === 'translator') {
        this.input = ''
        await this.generateTranslator(text, { regenerate: false })
        return
      }

      this.input = ''
      await this.generateChat(text, { regenerate: false })
    },
    async generateTriage(rawInput, { regenerate } = { regenerate: false }) {
      const myToken = ++this.generationToken
      this.focusMode = 'triage'
      const symptom = (rawInput || '').trim() || '发热 咳嗽'

      if (!regenerate) {
        this.lastUserMessageId = `triage_q_${Date.now()}`
        this.messages.push({ id: this.lastUserMessageId, role: 'user', content: symptom })
      }

      const tempAiId = `triage_ai_${Date.now()}`
      this.lastAssistantMessageId = tempAiId
      this.messages.push({ id: tempAiId, role: 'assistant', content: '', isLoading: true })
      this.chatStatus = 'loading'
      this.scheduleScrollBottom()

      const res = await aiTriage({ session_id: this.sessionId, symptoms: symptom.split(/\s+/), extra: symptom }).catch(() => null)
      if (myToken !== this.generationToken) return
      const out = (res && res.triage_output) || `- 结论摘要：信息不足，建议完善症状描述后再试。\n- 风险等级：${(res && res.risk_level) || '中'}\n`
      if (!res) {
        if (myToken !== this.generationToken) return
        // 失败时也用打字机节奏展示错误信息
        await this.typeText(tempAiId, '导诊请求失败，请稍后重试。')
        return
      }

      if (myToken !== this.generationToken) return
      await this.typeText(tempAiId, out)
    },
    async generateRehab(promptInput, { regenerate } = { regenerate: false }) {
      const myToken = ++this.generationToken
      if (!this.selectedRecordId) return this.$message.warning('请先选择指定病历')
      const prompt = (promptInput || '').trim()
      if (!prompt) return this.$message.warning('请在下方输入框先输入问题')

      if (!regenerate) {
        this.focusMode = 'rehab'
        this.lastUserMessageId = `rehab_q_${Date.now()}`
        this.messages.push({ id: this.lastUserMessageId, role: 'user', content: prompt })
      }

      const tempAiId = `rehab_ai_${Date.now()}`
      this.lastAssistantMessageId = tempAiId
      this.messages.push({ id: tempAiId, role: 'assistant', content: '', isLoading: true })
      this.chatStatus = 'loading'
      this.scheduleScrollBottom()

      const res = await aiRehabCompanion({
        session_id: this.sessionId,
        patient_id: this.account_id,
        record_id: this.selectedRecordId,
        user_prompt: prompt
      }).catch(() => null)
      if (myToken !== this.generationToken) return

      if (!res) {
        await this.typeText(tempAiId, '康复建议生成失败，请稍后重试。')
        return
      }
      if (myToken !== this.generationToken) return
      await this.typeText(tempAiId, res.rehab_output || '康复建议生成失败')
    },
    async generateTranslator(promptInput, { regenerate } = { regenerate: false }) {
      const myToken = ++this.generationToken
      if (!this.selectedRecordId) return this.$message.warning('请先选择指定病历')
      const prompt = (promptInput || '').trim()
      if (!prompt) return this.$message.warning('请在下方输入框先输入问题')

      if (!regenerate) {
        this.focusMode = 'translator'
        this.lastUserMessageId = `trans_q_${Date.now()}`
        this.messages.push({ id: this.lastUserMessageId, role: 'user', content: prompt })
      }

      const tempAiId = `trans_ai_${Date.now()}`
      this.lastAssistantMessageId = tempAiId
      this.messages.push({ id: tempAiId, role: 'assistant', content: '', isLoading: true })
      this.chatStatus = 'loading'
      this.scheduleScrollBottom()

      const res = await aiReportTranslator({
        session_id: this.sessionId,
        patient_id: this.account_id,
        record_id: this.selectedRecordId,
        user_question: prompt
      }).catch(() => null)
      if (myToken !== this.generationToken) return

      if (!res) {
        await this.typeText(tempAiId, '翻译解释生成失败，请稍后重试。')
        return
      }
      if (myToken !== this.generationToken) return
      await this.typeText(tempAiId, res.translator_output || '翻译解释生成失败')
    },
    async generateChat(message, { regenerate } = { regenerate: false }) {
      const myToken = ++this.generationToken
      const text = (message || '').trim()
      if (!text) return

      if (!regenerate) {
        this.lastUserMessageId = `tmp_u_${Date.now()}`
        this.messages.push({ id: this.lastUserMessageId, role: 'user', content: text })
      }

      const tempAiId = `tmp_ai_${Date.now()}`
      this.lastAssistantMessageId = tempAiId
      this.messages.push({ id: tempAiId, role: 'assistant', content: '', isLoading: true })
      this.chatStatus = 'loading'
      this.scheduleScrollBottom()

      const sid = this.sessionId || `sess_${Date.now()}`
      this.sessionId = sid
      const url = `${API_BASE}/ai/chat?stream=1`

      let donePayload = null
      let receivedDelta = false

      try {
        if (this.streamAbortController) {
          try { this.streamAbortController.abort() } catch (e) {}
        }
        this.streamAbortController = new AbortController()

        const resp = await fetch(url, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ session_id: sid, message: text }),
          signal: this.streamAbortController.signal
        })

        if (!resp.ok) throw new Error(`status ${resp.status}`)
        const contentType = (resp.headers.get('content-type') || '').toLowerCase()

        // SSE：逐步追加；非 SSE：直接取完整答案后打字机展示
        if (contentType.includes('text/event-stream') && resp.body) {
          this.chatStatus = 'streaming'
          const reader = resp.body.getReader()
          const decoder = new TextDecoder('utf-8')
          let buf = ''
          while (true) {
            const { done, value } = await reader.read()
            if (done) break
            buf += decoder.decode(value, { stream: true })
            const chunks = buf.split('\n\n')
            buf = chunks.pop() || ''
            for (const chunk of chunks) {
              const lines = chunk.split('\n')
              let eventName = ''
              let dataText = ''
              for (const line of lines) {
                if (line.startsWith('event:')) eventName = line.replace('event:', '').trim()
                if (line.startsWith('data:')) dataText += line.replace('data:', '').trim()
              }
              if (!dataText) continue

              let data = null
              try { data = JSON.parse(dataText) } catch (e) { continue }

              if (eventName === 'start') this.lastRisk = data.risk_level || ''
              else if (eventName === 'delta') {
                receivedDelta = true
                this.appendAssistantContent(tempAiId, data.content || '')
              } else if (eventName === 'done') {
                donePayload = data
              }
            }
          }
        } else {
          const json = await resp.json().catch(() => null)
          const data = (json && json.data) ? json.data : json || {}
          const answer = data.answer || '暂时无法回答，请稍后重试。'
          donePayload = data
          await this.typeText(tempAiId, answer)
        }
      } catch (e) {
        // 用户主动停止 or 网络错误：保留当前已输出内容
        const idxTmp = this.getMessageIndex(tempAiId)
        const hasContent = idxTmp >= 0 && !!this.messages[idxTmp].content
        if (!hasContent) {
          this.setAssistantLoading(tempAiId, false)
          this.chatStatus = 'done'
          if (idxTmp >= 0) this.messages[idxTmp].content = '请求失败，请稍后重试。'
        } else {
          this.chatStatus = 'done'
          this.setAssistantLoading(tempAiId, false)
        }
      }

      // SSE done 时需要补齐 reply_id/session_id/answer
      if (myToken !== this.generationToken) return
      const idx = this.getMessageIndex(tempAiId)
      if (idx >= 0 && donePayload) {
        if (donePayload.reply_id) this.messages[idx].id = donePayload.reply_id
        if (!this.messages[idx].content && donePayload.answer) this.messages[idx].content = donePayload.answer
        this.lastRisk = donePayload.risk_level || this.lastRisk
        this.sessionId = donePayload.session_id || this.sessionId
      } else if (!receivedDelta) {
        // 如果没有收到 delta，且也没有走非 SSE typing，那么尽量兜底提示
        if (idx >= 0 && !this.messages[idx].content) {
          this.messages[idx].content = '请求处理中，请点击“刷新会话”同步结果。'
        }
      }

      this.chatStatus = 'done'
      this.setAssistantLoading(tempAiId, false)

      await this.refreshSessions()
      this.scheduleScrollBottom()
    },
    async manualRefresh() {
      await this.refreshSessions()
      if (this.sessionId) await this.switchSession(this.sessionId)
    },
    scrollBottom() {
      const el = this.$refs.chatBody
      if (el) el.scrollTop = el.scrollHeight
    },

    // ==================== AI诊室功能 ====================
    toggleClinicMode() {
      this.isClinicMode = !this.isClinicMode
      if (this.isClinicMode) {
        this.showClinicDialog = true
      }
    },
    closeClinicDialog() {
      this.showClinicDialog = false
      this.isClinicMode = false
    },
    async startClinic() {
      try {
        this.clinicLoading = true
        const res = await aiClinicStart({ session_id: this.clinicSessionId })
        if (res && res.session_id) {
          this.clinicSessionId = res.session_id
          this.clinicCurrentStep = res.current_step
          this.clinicStepName = res.step_name
          this.clinicProgress = res.progress || 0
          this.clinicMessages = [{
            id: res.message_id,
            role: 'assistant',
            content: res.content
          }]
        }
      } catch (e) {
        this.$message.error('启动AI诊室失败，请重试')
      } finally {
        this.clinicLoading = false
      }
    },
    async submitClinicAnswer() {
      if (!this.clinicInput.trim()) {
        this.$message.warning('请输入您的回答')
        return
      }
      const answer = this.clinicInput.trim()
      this.clinicLoading = true

      // 添加用户消息
      this.clinicMessages.push({
        id: `user_${Date.now()}`,
        role: 'user',
        content: answer
      })
      this.clinicInput = ''
      this.scrollClinicBottom()

      try {
        const res = await aiClinicChat({
          session_id: this.clinicSessionId,
          current_step: this.clinicCurrentStep,
          user_response: answer,
          action: 'continue'
        })

        if (res && res.session_id) {
          this.clinicCurrentStep = res.current_step
          this.clinicStepName = res.step_name
          this.clinicProgress = res.progress || 0

          if (res.is_complete) {
            // 问诊完成
            this.clinicComplete = true
            this.clinicReport = res.report || ''
            this.clinicSummary = res.summary || ''
            // 解析建议列表
            this.clinicSuggestions = this.parseSuggestions(res.suggestions || res.report || '')
            this.clinicMessages.push({
              id: `report_${Date.now()}`,
              role: 'assistant',
              content: res.report || '诊断报告生成完成'
            })
          } else {
            // 下一轮问询
            this.clinicMessages.push({
              id: res.message_id,
              role: 'assistant',
              content: res.ai_message || res.content
            })
            this.updateClinicPlaceholder()
          }
        }
      } catch (e) {
        this.$message.error('提交失败，请重试')
      } finally {
        this.clinicLoading = false
        this.scrollClinicBottom()
      }
    },
    updateClinicPlaceholder() {
      const placeholders = {
        1: '请描述您的主要不适症状（如头痛、胸闷、咳嗽等）',
        2: '请详细描述症状的具体表现',
        3: '请描述您的过往病史',
        4: '请描述您的家族遗传病史',
        5: '请描述您当前的用药情况',
        6: '请描述您的生活习惯和工作环境'
      }
      this.clinicInputPlaceholder = placeholders[this.clinicCurrentStep] || '请输入您的回答'
    },
    scrollClinicBottom() {
      this.$nextTick(() => {
        const el = this.$refs.clinicMessages
        if (el) el.scrollTop = el.scrollHeight
      })
    },
    async resetClinic() {
      try {
        await aiClinicReset({ session_id: this.clinicSessionId })
      } catch (e) {}
      this.clinicCurrentStep = 0
      this.clinicProgress = 0
      this.clinicMessages = []
      this.clinicInput = ''
      this.clinicComplete = false
      this.clinicReport = ''
      this.clinicSummary = ''
      this.clinicSuggestions = []
    },
    async startNewClinic() {
      await this.resetClinic()
      await this.startClinic()
    },
    goToRegister() {
      this.$router.push('/outpatient/register')
    },
    parseSuggestions(text) {
      // 从报告文本中提取建议列表
      if (!text) return []
      const suggestions = []
      const lines = text.split('\n')
      for (const line of lines) {
        const trimmed = line.trim()
        // 匹配常见的建议模式
        if (trimmed.startsWith('•') || trimmed.startsWith('-') || trimmed.startsWith('·')) {
          suggestions.push(trimmed.replace(/^[•\-\·]\s*/, ''))
        } else if (/建议|注意|多喝|避免|及时|请/.test(trimmed) && trimmed.length < 100) {
          suggestions.push(trimmed)
        }
      }
      return suggestions.slice(0, 5) // 最多显示5条
    },
    downloadReport() {
      // 生成并下载报告文本
      const reportContent = `
=== AI诊室诊断报告 ===

诊断结论：
${this.clinicReport}

健康建议：
${this.clinicSuggestions.map(s => '• ' + s).join('\n')}

咨询小结：
${this.clinicSummary}

---
免责声明：
本报告由AI系统生成，仅供参考，不能替代专业医生的诊断。
如症状严重或持续不缓解，请及时就医。紧急情况请拨打120急救电话。
      `.trim()

      const blob = new Blob([reportContent], { type: 'text/plain;charset=utf-8' })
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `AI诊断报告_${new Date().toLocaleDateString()}.txt`
      link.click()
      URL.revokeObjectURL(url)
      this.$message.success('报告下载成功')
    }
  }
}
</script>

<style scoped>
.ai-page{padding:0;background:#f3f4f6;min-height:100%}
.ai-wrap{display:grid;grid-template-columns:170px 1fr;min-height:calc(100vh - 84px)}
.session-card{border-radius:0;border:none;border-right:1px solid #e5e7eb;background:#fff;padding:6px 4px}
.session-title{font-size:16px;font-weight:700;color:#111827;margin:10px 8px 4px}
.session-sub{font-size:11px;color:#9ca3af;margin:0 8px 10px}
.session-list{padding:0 8px;overflow:auto;max-height:calc(100vh - 180px)}
.session-item{padding:8px 6px;border-radius:8px;margin-bottom:6px;cursor:pointer}
.session-item:hover{background:#f3f4f6}
.session-item.active{background:#eef4ff}
.name{font-size:12px;font-weight:600;color:#111827;line-height:1.4}
.sub{font-size:10px;color:#9ca3af;margin-top:2px}

.main-col{background:#fff}
.chat-card{border:none;border-radius:0;min-height:100%;padding:10px 14px}
.head h3{font-size:28px;color:#2563eb;margin:0}
.head p{font-size:12px;color:#6b7280;margin:4px 0 0}

.chat-body{margin-top:8px;background:#f3f6fb;border:1px solid #e9edf5;border-radius:0;min-height:260px;max-height:360px;overflow:auto;padding:14px}
.msg{display:flex;align-items:flex-start;gap:8px;margin-bottom:12px}
.msg.user{justify-content:flex-end}
.avatar{width:26px;height:26px;border-radius:50%;background:#3b82f6;color:#fff;display:flex;align-items:center;justify-content:center;font-size:12px;flex:0 0 auto}
.bubble{max-width:78%;padding:10px 12px;border-radius:14px;white-space:pre-wrap;font-size:12.5px;line-height:1.75}
.msg.assistant .bubble{background:#fff;border:1px solid #e5e7eb;color:#111827}
.msg.user .bubble{background:#2563eb;color:#fff;border:1px solid #2563eb}

.typing-indicator{display:inline-flex;align-items:center;gap:8px;color:#6b7280}
.ellipsis{display:inline-flex;align-items:center;gap:4px}
.dot{width:4px;height:4px;border-radius:50%;background:#9ca3af;display:inline-block;animation:blink 1.2s infinite}
.dot1{animation-delay:0s}
.dot2{animation-delay:0.15s}
.dot3{animation-delay:0.3s}
@keyframes blink{
  0%,20%{opacity:0.2;transform:translateY(0)}
  50%{opacity:1;transform:translateY(-1px)}
  100%{opacity:0.2;transform:translateY(0)}
}

.quick-row{display:flex;align-items:center;gap:6px;margin-top:8px;padding:8px 0;border-bottom:1px solid #eceff4}
.quick-label{font-size:12px;color:#6b7280}

.guide-card{margin-top:10px;background:#eef5ff;border:1px solid #cfe0ff;border-radius:8px;padding:10px}
.guide-title{display:flex;justify-content:space-between;align-items:center;color:#1d4ed8;font-weight:700;font-size:12px}
.guide-text{margin-top:8px;color:#1d4ed8;font-size:12px;line-height:1.9}
.example-head{margin-top:8px;padding-top:8px;border-top:1px solid #cfe0ff;display:flex;justify-content:space-between;align-items:center;color:#1d4ed8;font-size:12px;font-weight:600}
.example-text{margin-top:8px;background:#fff;border-radius:6px;border:1px solid #e5e7eb;color:#2563eb;font-size:12px;padding:8px;line-height:1.7}
.guide-actions{display:flex;align-items:center;gap:8px;margin-top:8px}
.chosen{font-size:11px;color:#6b7280}

.chat-foot{margin-top:10px;padding-top:10px;border-top:1px solid #eceff4}
.foot-actions{display:flex;justify-content:flex-end;gap:8px;margin-top:8px}

.record-title{font-size:16px;font-weight:700;color:#fff;background:linear-gradient(90deg,#3b82f6,#2563eb);margin:-20px -20px 0;padding:12px 14px;border-radius:8px 8px 0 0}
.record-sub{font-size:12px;color:#6b7280;margin:10px 0}
.record-list{max-height:390px;overflow:auto}
.record-item{border:1px solid #e5e7eb;border-radius:8px;padding:10px;margin-bottom:10px;cursor:pointer}
.record-item.active{border-color:#3b82f6;background:#eff6ff}
.ri-head{display:flex;justify-content:space-between;align-items:center;margin-bottom:6px}
.ri-name{font-size:13px;font-weight:700;color:#111827}
.ri-time{font-size:11px;color:#9ca3af}
.ri-line{font-size:12px;color:#374151;line-height:1.7}

/* AI诊室气泡框样式 - 优化版 */
.clinic-overlay{position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.5);z-index:2000;display:flex;align-items:center;justify-content:center;padding:20px}
.clinic-dialog{background:#fff;border-radius:16px;width:720px;max-width:95vw;max-height:90vh;display:flex;flex-direction:column;box-shadow:0 12px 40px rgba(0,0,0,0.2);overflow:hidden}
.clinic-header{background:linear-gradient(135deg,#10b981,#059669);color:#fff;padding:20px 24px;position:relative}
.clinic-title{font-size:20px;font-weight:700;display:flex;align-items:center;gap:10px}
.clinic-subtitle{font-size:13px;opacity:0.9;margin-top:6px}
.complete-tag{background:rgba(255,255,255,0.25);padding:6px 14px;border-radius:16px;font-size:13px;font-weight:500;display:inline-flex;align-items:center;gap:6px}
.clinic-close{position:absolute;top:16px;right:16px;color:#fff;padding:6px;border-radius:8px;background:rgba(255,255,255,0.15)}
.clinic-close:hover{color:#fff;background:rgba(255,255,255,0.25)}

.clinic-progress{padding:14px 24px;background:#f9fafb;border-bottom:1px solid #e5e7eb}
.progress-info{display:flex;justify-content:space-between;align-items:center;margin-bottom:10px;font-size:13px}
.step-name{color:#374151;font-weight:600}
.step-count{color:#6b7280}

.clinic-steps{padding:18px 24px;border-bottom:1px solid #e5e7eb;background:#f0fdf4}
.steps-title{font-size:14px;font-weight:700;color:#065f46;margin-bottom:12px}
.step-item{display:flex;align-items:center;gap:10px;padding:8px 0;font-size:13px;color:#6b7280}
.step-item.active{color:#059669;font-weight:600}
.step-num{width:24px;height:24px;border-radius:50%;background:#d1fae5;color:#059669;display:flex;align-items:center;justify-content:center;font-size:12px;font-weight:700}
.step-text{color:#374151}
.clinic-enter-hint{margin-top:14px;font-size:13px;color:#059669;text-align:center;padding:12px;background:#ecfdf5;border-radius:10px}

.clinic-messages{flex:1;overflow:auto;padding:20px 24px;max-height:400px;background:#f9fafb}
.clinic-msg{display:flex;gap:12px;margin-bottom:16px;align-items:flex-start}
.clinic-msg.user{flex-direction:row-reverse}
.clinic-avatar{width:36px;height:36px;border-radius:50%;background:linear-gradient(135deg,#10b981,#059669);color:#fff;display:flex;align-items:center;justify-content:center;font-size:14px;font-weight:700;flex-shrink:0}
.clinic-msg.user .clinic-avatar{background:linear-gradient(135deg,#3b82f6,#155dfc)}
.clinic-bubble{max-width:72%;padding:14px 18px;border-radius:14px;font-size:14px;line-height:1.7;white-space:pre-wrap}
.clinic-msg.assistant .clinic-bubble{background:#fff;border:1px solid #e5e7eb;color:#111827;box-shadow:0 2px 8px rgba(0,0,0,0.06)}
.clinic-msg.user .clinic-bubble{background:linear-gradient(135deg,#155dfc,#2563eb);color:#fff;border:none}
.clinic-bubble.typing{background:#f3f4f6;color:#6b7280}

/* AI诊室底部输入区域 */
.clinic-foot{padding:18px 24px;background:#fff;border-top:1px solid #e5e7eb}
.clinic-actions{display:flex;justify-content:center;gap:14px;margin-top:12px}

/* 诊断报告区域 - 优化版 */
.clinic-report{flex:1;overflow:auto;padding:0;background:#fff;border-top:1px solid #e5e7eb;max-height:none;display:flex;flex-direction:column}

/* 主报告区域 */
.report-main{flex:1;overflow:auto;padding:20px 24px}

/* 报告标题 */
.report-header{margin-bottom:16px}
.report-badge{background:linear-gradient(135deg,#10b981,#059669);color:#fff;padding:8px 16px;border-radius:8px;display:inline-flex;align-items:center;gap:8px;font-size:14px;font-weight:600;margin-bottom:10px}
.report-intro{font-size:14px;color:#6b7280;margin:0}

/* 报告卡片通用样式 */
.report-card{background:#fff;border:1px solid #e5e7eb;border-radius:12px;padding:16px;margin-bottom:14px}
.card-header{display:flex;align-items:center;gap:10px;font-size:15px;font-weight:700;color:#111827;margin-bottom:12px;padding-bottom:12px;border-bottom:1px solid #f3f4f6}
.card-header i{font-size:18px;color:#155dfc}
.card-header.warning i{color:#ea580c}
.card-body{font-size:14px;line-height:1.8;color:#374151;white-space:pre-wrap}

/* 诊断结论卡片 */
.conclusion-card{border-left:4px solid #155dfc}
.conclusion-card .card-header{color:#155dfc}

/* 诊断摘要卡片 */
.summary-card{border-left:4px solid #10b981}

/* 建议卡片 */
.suggestion-card{background:#fff7ed;border-color:#fed7aa;border-left:4px solid #ea580c}
.suggestion-card .card-header{color:#9a3412;border-bottom-color:#fed7aa}
.suggestion-list{list-style:none;padding:0;margin:0}
.suggestion-list li{display:flex;align-items:flex-start;gap:10px;padding:8px 0;font-size:14px;color:#9a3412;line-height:1.6;border-bottom:1px solid #fed7aa}
.suggestion-list li:last-child{border-bottom:none}
.suggestion-list li::before{content:"•";color:#ea580c;font-weight:bold;font-size:16px;flex-shrink:0}

/* 底部操作栏 */
.report-footer{display:flex;gap:12px;padding:16px 24px;background:#f9fafb;border-top:1px solid #e5e7eb}
.report-footer .el-button{height:40px;padding:0 24px;font-size:14px;font-weight:500}
.btn-register{background:#155dfc !important;border-color:#155dfc !important;color:#fff !important}
.btn-register:hover{background:#0d4fd9 !important;border-color:#0d4fd9 !important}
.btn-download{background:#00a63e !important;border-color:#00a63e !important;color:#fff !important}
.btn-download:hover{background:#008f36 !important;border-color:#008f36 !important}

/* 免责声明 */
.report-disclaimer{background:#f9fafb;border:1px solid #e5e7eb;border-radius:10px;padding:14px 18px;font-size:12px;color:#6b7280;display:flex;align-items:flex-start;gap:8px;line-height:1.6;margin:0 24px 20px}
.report-disclaimer i{color:#f59e0b;flex-shrink:0;margin-top:2px}
</style>