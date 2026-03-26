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
  </div>
</template>

<script>
import { mapGetters } from 'vuex'
import { aiSessions, aiSessionMessages, aiTriage, aiRehabCompanion, aiReportTranslator } from '@/api/aiAssistant'
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
      recordDialogVisible: false
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
</style>