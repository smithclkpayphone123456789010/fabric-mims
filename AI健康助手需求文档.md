# AI健康助手需求文档

## 1. 文档说明

- 文档名称：AI健康助手需求文档
- 适用项目：`fabric-mims`
- 目标：在既有门诊管理系统中落地患者端 AI 健康助手，形成“咨询 -> 导诊 -> 挂号 -> 支付 -> 排队 -> 就诊后管理”的闭环。

---

## 2. PRD摘要（V1）

### 2.1 产品目标

1. 提升患者就诊前后咨询效率与满意度。
2. 提升导诊到挂号转化率。
3. 降低人工客服重复咨询压力。
4. 提供安全、可追溯、可审计的健康问答能力。

### 2.2 V1 功能范围

1. 智能导诊：根据症状推荐科室、风险等级、是否急诊。
2. 门诊流程助手：围绕预约/支付/排队提供实时问答。
3. 个性化康复伴行：基于医生新增病历生成用药提醒、复诊建议、注意事项说明。
4. 检查报告“翻译官”：基于患者选定病历，对诊断信息与检查治疗部分进行非诊断型大白话解释。
5. 对话中业务动作：跳转挂号、支付、排队、我的预约。
6. 多模型接入：千问 + DeepSeek，支持策略路由与降级。

### 2.3 不在V1范围

- 自动诊断与自动开药。
- 影像自动判读。
- 语音/视频问诊。

---

## 3. 技术拆解版任务清单（前后端按文件级）

## 3.1 后端任务（Go）

### A. API 层

1. 新增文件：`application/server/api/v2/ai_assistant.go`
   - 提供路由处理函数：
     - `Chat()`
     - `Triage()`
     - `RehabCompanion()`
     - `ReportTranslator()`
     - `GetSessions()`
     - `GetSessionMessages()`
   - 统一处理请求参数校验、错误码映射、返回结构。

2. 修改文件：`application/server/routers/router.go`
   - 新增路由：
     - `POST /api/v2/ai/chat`
     - `POST /api/v2/ai/triage`
     - `POST /api/v2/ai/rehab-companion`
     - `POST /api/v2/ai/report-translator`
     - `GET /api/v2/ai/sessions`
     - `GET /api/v2/ai/session/:id/messages`

### B. Service 层（核心）

1. 新增目录：`application/server/service/ai/`

2. 新增文件：`application/server/service/ai/router.go`
   - 路由策略：
     - 高风险问题 -> 千问优先
     - 普通流程问答 -> DeepSeek优先
     - 超时/失败 -> 自动回退另一模型

3. 新增文件：`application/server/service/ai/providers/qwen.go`
   - 封装千问请求：鉴权、超时、重试、响应解析。

4. 新增文件：`application/server/service/ai/providers/deepseek.go`
   - 封装 DeepSeek 请求：鉴权、超时、重试、响应解析。

5. 新增文件：`application/server/service/ai/prompt_builder.go`
   - 拼接系统提示词 + 用户输入 + 业务上下文摘要。

6. 新增文件：`application/server/service/ai/safety_guard.go`
   - 前置风控策略：
     - 高危关键词拦截（胸痛、呼吸困难、昏迷等）
     - 医疗边界限制（禁止诊断/开药）

7. 新增文件：`application/server/service/ai/context_builder.go`
   - 聚合门诊上下文：预约、支付、排队、门诊记录摘要。

8. 新增文件：`application/server/service/ai/session_store.go`
   - 会话与消息存储（可先DB，后续可加缓存）。

9. 新增文件：`application/server/service/ai/rehab_companion.go`
   - 实现“个性化康复伴行”服务：
     - 拉取患者最近病历（或指定病历）
     - 生成用药提醒、复诊建议、注意事项说明
     - 严格按康复伴行提示词模板输出

10. 新增文件：`application/server/service/ai/report_translator.go`
   - 实现“检查报告翻译官”服务：
     - 读取患者选定病历中的诊断与检查治疗字段
     - 生成非诊断型通俗解释（大白话）
     - 严格按翻译官提示词模板输出

### C. 配置与模型

1. 新增/修改配置：`application/server/conf/config.yaml`（或现有配置文件）
   - 新增：
     - `ai.providers.qwen.base_url`
     - `ai.providers.qwen.api_key`
     - `ai.providers.deepseek.base_url`
     - `ai.providers.deepseek.api_key`
     - `ai.timeout_ms`
     - `ai.fallback_enabled`

2. 新增文件：`application/server/model/ai_model.go`
   - 定义：会话、消息、风控日志、请求响应 DTO。

### D. 数据层与审计

1. 新增文件：`application/server/repository/ai_repository.go`
   - 提供 `session/message/audit` 数据写入与查询。

2. 新增表（或链下存储结构）
   - `ai_sessions`
   - `ai_messages`
   - `ai_audit_logs`

3. 安全审计字段建议
   - `user_id`
   - `session_id`
   - `provider`
   - `model`
   - `risk_level`
   - `prompt_hash`
   - `action_type`
   - `created_at`

### E. 后端联调与验证

1. 新增测试文件：`application/server/api/v2/ai_assistant_test.go`
2. 新增测试文件：`application/server/service/ai/router_test.go`
3. 增加压测脚本（可选）：`application/server/test/ai_chat_bench.go`

---

## 3.2 前端任务（Vue）

### A. 页面与路由

1. 新增页面：`application/web/src/views/ai-health-assistant/index.vue`
   - 主聊天界面（消息流 + 输入框 + 建议动作区）

2. 新增页面：`application/web/src/views/ai-health-assistant/history.vue`
   - 会话历史列表与会话切换

3. 修改路由：`application/web/src/router/index.js`
   - 新增路由：`/ai-health-assistant`
   - 新增子路由：`/ai-health-assistant/history`

### B. API 与状态管理

1. 新增文件：`application/web/src/api/aiAssistant.js`
   - `chat()`
   - `triage()`
   - `rehabCompanion()`
   - `reportTranslator()`
   - `getSessions()`
   - `getSessionMessages()`

2. 新增（可选）store：`application/web/src/store/modules/aiAssistant.js`
   - 管理当前会话、消息列表、发送状态、建议动作。

### C. 组件拆分

1. 新增组件：`application/web/src/components/ai-assistant/ChatMessage.vue`
2. 新增组件：`application/web/src/components/ai-assistant/ActionChips.vue`
3. 新增组件：`application/web/src/components/ai-assistant/RiskBanner.vue`
4. 新增组件：`application/web/src/components/ai-assistant/TypingIndicator.vue`

### D. 与门诊模块联动

1. 对话动作映射：
   - `go_register` -> `outpatient/register`
   - `go_payment` -> `outpatient/payment`
   - `go_queue` -> `outpatient/queue`
   - `go_my_registration` -> `outpatient/my-registration`

---

## 4. UI设计（可直接给设计/前端执行）

## 4.1 设计风格

- 视觉关键词：医疗、可信、清晰、低干扰。
- 主色：`#2563EB`（医疗蓝）
- 成功：`#16A34A`
- 警告：`#EA580C`
- 风险提示背景：`#FFF7E6`
- 页面背景：`#F5F7FB`
- 卡片圆角：`10px`
- 基础阴影：`0 2px 8px rgba(15, 23, 42, 0.06)`

## 4.2 信息架构

1. 顶部标题区
   - 标题：AI健康助手
   - 副标题：问症状、看流程、做导诊

2. 左侧（可选）会话列表
   - 今日会话
   - 历史会话

3. 主区域
   - 消息流（用户/助手）
   - 风险提示条（高危时固定置顶）
   - 建议动作按钮（去挂号/去支付/查看排队）

4. 底部输入区
   - 多行输入框
   - “发送”按钮
   - 常见问题快捷入口（症状导诊、个性化康复伴行、检查报告翻译官）

5. 功能卡片区（新增）
   - 个性化康复伴行：
     - 入口文案：根据最新病历生成康复计划
     - 一键触发：用药提醒 + 复诊建议 + 注意事项
   - 检查报告翻译官：
     - 入口文案：把专业术语翻译成大白话
     - 选择病历后触发解释

## 4.3 关键状态UI

1. Loading：助手“思考中”气泡 + 点动效
2. 错误：红色浅底错误条 + “重试”按钮
3. 高风险：橙色警示卡 + “立即拨打急救电话”提示
4. 无数据：空态插画 + 引导问题按钮

## 4.4 响应式建议

- 桌面端：左右分栏（会话列表 + 聊天主区）
- 平板/移动端：单栏，历史会话抽屉式展开

---

## 5. 可直接落地的 API JSON 契约

## 5.1 `POST /api/v2/ai/chat`

### 请求

```json
{
  "session_id": "sess_202603260001",
  "message": "我最近胸闷，应该挂什么科？",
  "context_scope": {
    "include_registration": true,
    "include_payment": true,
    "include_queue": true,
    "include_records": true
  },
  "client_meta": {
    "page": "outpatient_home",
    "trace_id": "trace-abc-001"
  }
}
```

### 响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "session_id": "sess_202603260001",
    "reply_id": "msg_ai_10001",
    "answer": "建议优先挂心内科，若伴随持续胸痛、呼吸困难请立即急诊。",
    "risk_level": "high",
    "risk_notice": "若出现胸痛持续超过15分钟，请立即拨打120。",
    "suggestions": [
      "优先挂号：心内科",
      "准备近期心电图/既往病史",
      "避免剧烈活动"
    ],
    "actions": [
      {
        "type": "go_register",
        "label": "去挂号",
        "target": "/outpatient/register",
        "params": {
          "department_id": "心内科"
        }
      }
    ],
    "provider": "qwen",
    "model": "qwen-max",
    "tokens": {
      "prompt": 850,
      "completion": 210,
      "total": 1060
    },
    "created_at": "2026-03-26T10:21:00+08:00"
  }
}
```

---

## 5.2 `POST /api/v2/ai/triage`

### 请求

```json
{
  "session_id": "sess_202603260001",
  "symptoms": ["发热", "咳嗽", "喉咙痛"],
  "duration_days": 3,
  "age": 31,
  "gender": "male",
  "chronic_diseases": ["高血压"],
  "extra": "夜间咳嗽加重"
}
```

### 响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "risk_level": "medium",
    "triage_output": "- 结论摘要：...\n- 风险等级：中\n- 建议行动（最多3条）：...\n- 可选就诊科室：...\n- 免责声明：本建议仅供健康管理参考，不构成医疗诊断。",
    "is_emergency_recommended": false
  }
}
```

---

## 5.3 `POST /api/v2/ai/rehab-companion`

### 请求

```json
{
  "session_id": "sess_202603260101",
  "patient_id": "patient_001",
  "record_id": "record_20260326001",
  "focus": "夜间咳嗽和用药依从性"
}
```

### 响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "risk_level": "medium",
    "rehab_output": "- 康复阶段判断：...\n- 用药提醒：...\n- 复诊建议：...\n- 注意事项：...\n- 预警信号：...\n- 免责声明：本建议仅供康复管理参考，不构成诊断或处方调整。"
  }
}
```

---

## 5.4 `POST /api/v2/ai/report-translator`

### 请求

```json
{
  "session_id": "sess_202603260102",
  "patient_id": "patient_001",
  "record_id": "record_20260326001",
  "user_question": "这份报告最需要关注什么？"
}
```

### 响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "risk_level": "medium",
    "translator_output": "- 一句话总结：...\n- 术语翻译：...\n- 检查/治疗在说什么：...\n- 我现在该做什么：...\n- 什么时候尽快就医：...\n- 免责声明：本解释用于信息理解，不构成诊断或治疗方案。"
  }
}
```

---

## 5.3 `GET /api/v2/ai/sessions`

### 响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "items": [
      {
        "session_id": "sess_202603260001",
        "title": "胸闷导诊咨询",
        "last_message": "建议优先挂心内科...",
        "updated_at": "2026-03-26T10:21:00+08:00"
      }
    ],
    "total": 1
  }
}
```

---

## 5.4 `GET /api/v2/ai/session/:id/messages`

### 响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "session_id": "sess_202603260001",
    "messages": [
      {
        "id": "msg_u_1",
        "role": "user",
        "content": "我最近胸闷，应该挂什么科？",
        "created_at": "2026-03-26T10:20:32+08:00"
      },
      {
        "id": "msg_ai_1",
        "role": "assistant",
        "content": "建议优先挂心内科...",
        "risk_level": "high",
        "created_at": "2026-03-26T10:21:00+08:00"
      }
    ]
  }
}
```

---

## 6. 安全提示词模板（千问 / DeepSeek 通用）

## 6.1 System Prompt（通用主模板）

```text
你是“门诊AI健康助手”，服务于互联网医院门诊系统。
你的任务是提供健康管理与就医流程建议，而不是医疗诊断。

【安全原则】
1. 绝不做确诊结论、绝不替代医生开药决策。
2. 如用户出现高危症状（胸痛持续、呼吸困难、意识障碍、抽搐、大出血等），必须优先建议立即急诊/拨打120。
3. 回答要简洁、可执行、风险优先。
4. 对不确定内容明确说明“建议线下医生进一步评估”。
5. 禁止提供危险偏方、违规医疗建议。

【输出格式】
请严格按以下结构输出：
- 结论摘要：...
- 风险等级：低/中/高
- 建议行动（最多3条）：
  1) ...
  2) ...
  3) ...
- 可选就诊科室：...
- 免责声明：本建议仅供健康管理参考，不构成医疗诊断。
```

## 6.2 高危兜底模板（命中风控词时强制）

```text
检测到可能高风险症状，请立即采取以下措施：
1) 立即停止当前活动，保持平卧或半卧位。
2) 尽快前往最近急诊或拨打120。
3) 若有既往病史或正在用药，请携带相关资料。

本助手不能替代急救与临床诊断。
```

## 6.3 业务流程问答模板（非医疗高风险）

```text
你可以结合用户当前门诊状态（预约、支付、排队）给出下一步操作。
优先输出“可点击动作建议”：去挂号/去支付/查看排队/查看我的预约。
避免冗长解释，保持最多5行核心建议。
```

## 6.4 Prompt 输入拼装模板

```text
[系统规则]
{{system_prompt}}

[用户信息摘要]
- 年龄: {{age}}
- 性别: {{gender}}
- 慢病: {{chronic_diseases}}

[门诊上下文摘要]
- 最近预约: {{latest_registration_summary}}
- 支付状态: {{payment_summary}}
- 排队状态: {{queue_summary}}

[用户问题]
{{user_message}}
```

## 6.5 个性化康复伴行 Prompt 约束模板（新增）

```text
你是“个性化康复伴行”助手，基于医生新增病历为患者提供康复管理建议。
你不是临床诊断医生，不可修改处方，不可给出确诊。

【输入】
- 病历摘要（诊断、医嘱、用药、检查、治疗）
- 患者补充问题

【输出要求】
请严格按以下结构输出：
- 康复阶段判断：...
- 用药提醒：
  1) ...
  2) ...
  3) ...
- 复诊建议：...
- 注意事项：
  1) ...
  2) ...
  3) ...
- 预警信号（出现即尽快就医）：...
- 免责声明：本建议仅供康复管理参考，不构成诊断或处方调整。
```

## 6.6 检查报告“翻译官” Prompt 约束模板（新增）

```text
你是检查报告“翻译官”，负责把病历中的诊断信息、检查和治疗内容翻译成大白话。
你不是医生，不可做诊断结论，不可给出处方调整建议。

【输入】
- 病历中的诊断信息
- 病历中的检查与治疗内容
- 用户问题

【输出要求】
请严格按以下结构输出：
- 一句话总结：...
- 术语翻译（专业词 -> 大白话）：
  1) ...
  2) ...
  3) ...
- 检查/治疗在说什么：...
- 我现在该做什么（最多3条）：...
- 什么时候尽快就医：...
- 免责声明：本解释用于信息理解，不构成诊断或治疗方案。
```

## 6.7 提问引导提示词（提高回答质量）

### A. 个性化康复伴行 引导

```text
为了生成更贴合你的康复建议，请尽量补充：
1) 当前最不舒服的症状（例如夜间咳嗽、头晕）
2) 症状持续多久、是否加重
3) 当前正在服用的药物与频次
4) 最近一次复诊或检查时间
5) 你最关心的问题（如“什么时候复诊最合适”）
```

### B. 检查报告“翻译官” 引导

```text
为了翻译得更准确，请告诉我：
1) 你想重点看哪部分（诊断结果/检查结果/治疗方案/处方用药/医嘱）
2) 哪些术语看不懂
3) 你最关心的一个问题（例如“严重吗”“要不要马上复查”）
4) 是否有医生当面给过补充说明
```


---

## 7. 验收标准（AI助手）

1. 患者可完成至少3轮稳定对话。
2. 高危症状能触发急诊提示。
3. 对话中可一键跳转门诊业务页面。
4. 会话与审计日志可查询。
5. 双模型可切换且失败可回退。

---

## 8. 分期实施建议

- M1：后端AI网关 + Chat接口 + 前端聊天页骨架
- M2：导诊能力 + 动作跳转 + 会话历史
- M3：审计与风控强化 + 指标看板
- M4：提示词与路由策略优化（A/B）

---

## 9. 交付清单

1. 本文档：`AI健康助手需求文档.md`
2. 后端任务清单（文件级）
3. 前端任务清单（文件级）
4. API JSON 契约
5. 安全提示词模板（千问/DeepSeek通用）
