package v2

import (
	bc "application/blockchain"
	"application/pkg/app"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type aiChatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type aiTriageRequest struct {
	SessionID string   `json:"session_id"`
	Symptoms  []string `json:"symptoms"`
	Extra     string   `json:"extra"`
}

type aiRehabCompanionRequest struct {
	SessionID  string `json:"session_id"`
	PatientID  string `json:"patient_id"`
	RecordID   string `json:"record_id"`
	UserPrompt string `json:"user_prompt"`
}

type aiReportTranslatorRequest struct {
	SessionID    string `json:"session_id"`
	PatientID    string `json:"patient_id"`
	RecordID     string `json:"record_id"`
	UserQuestion string `json:"user_question"`
}

// AI诊室相关数据结构
type aiClinicRequest struct {
	SessionID    string `json:"session_id"`
	CurrentStep  int    `json:"current_step"`  // 1-6表示第几轮，0表示开始
	UserResponse string `json:"user_response"` // 用户本轮回答
	Action       string `json:"action"`        // "start", "continue", "report"
}

type clinicCollectedData struct {
	// 第1轮：主诉症状
	ChiefComplaint  string `json:"chief_complaint"`  // 主要症状
	SymptomDuration string `json:"symptom_duration"` // 持续时间
	SymptomLocation string `json:"symptom_location"` // 不适部位

	// 第2轮：症状详情
	SymptomDetails     string `json:"symptom_details"`     // 具体表现
	AggravatingFactors string `json:"aggravating_factors"` // 加重因素
	RelievingFactors   string `json:"relieving_factors"`   // 缓解因素
	AssociatedSymptoms string `json:"associated_symptoms"` // 伴随症状

	// 第3轮：过往病史
	PastMedicalHistory     string `json:"past_medical_history"`    // 既往病史
	HospitalizationHistory string `json:"hospitalization_history"` // 住院史
	SurgeryHistory         string `json:"surgery_history"`         // 手术史

	// 第4轮：家族遗传病史
	FamilyHistory   string `json:"family_history"`   // 家族病史
	GeneticDiseases string `json:"genetic_diseases"` // 遗传病

	// 第5轮：用药情况
	CurrentMedications string `json:"current_medications"` // 当前用药
	MedicationDosage   string `json:"medication_dosage"`   // 剂量
	AllergyHistory     string `json:"allergy_history"`     // 过敏药物

	// 第6轮：个人健康状况
	LifestyleHabits     string `json:"lifestyle_habits"`      // 生活习惯（饮食/运动/睡眠/烟酒）
	WorkEnvironment     string `json:"work_environment"`      // 工作环境
	PersonalHealthNotes string `json:"personal_health_notes"` // 其他健康备注
}

type aiClinicSession struct {
	SessionID     string              `json:"session_id"`
	Title         string              `json:"title"`
	CurrentStep   int                 `json:"current_step"` // 1-6表示当前轮次
	Progress      float64             `json:"progress"`     // 进度百分比 0-100
	CollectedData clinicCollectedData `json:"collected_data"`
	Messages      []aiMessage         `json:"messages"`
	Report        string              `json:"report,omitempty"`  // 最终诊断报告
	Summary       string              `json:"summary,omitempty"` // AI诊室咨询小结
	StartedAt     string              `json:"started_at"`
	UpdatedAt     string              `json:"updated_at"`
}

type aiMessage struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	RiskLevel string `json:"risk_level,omitempty"`
	CreatedAt string `json:"created_at"`
}

type aiSession struct {
	SessionID   string      `json:"session_id"`
	Title       string      `json:"title"`
	LastMessage string      `json:"last_message"`
	UpdatedAt   string      `json:"updated_at"`
	Messages    []aiMessage `json:"messages"`
}

var (
	aiStoreMu   sync.RWMutex
	aiStore     = map[string]*aiSession{}
	clinicStore = map[string]*aiClinicSession{} // AI诊室会话存储
	dotenvOnce  sync.Once
)

const triageSystemPrompt = `你是“门诊AI健康助手”，服务于互联网医院门诊系统。
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
- 免责声明：本建议仅供健康管理参考，不构成医疗诊断。`

const rehabCompanionPrompt = `你是“个性化康复伴行”助手，基于医生新增病历为患者提供康复管理建议。
你不是临床诊断医生，不可修改处方，不可给出确诊。

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
- 免责声明：本建议仅供康复管理参考，不构成诊断或处方调整。`

const reportTranslatorPrompt = `你是检查报告“翻译官”，负责把病历中的诊断信息、检查和治疗内容翻译成大白话。
你不是医生，不可做诊断结论，不可给出处方调整建议。

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
- 免责声明：本解释用于信息理解，不构成诊断或治疗方案。`

// AI诊室六轮问询提示词
var clinicStepPrompts = map[int]string{
	1: `你是AI智能诊室的医生，请以专业、友好、耐心的态度与患者进行问诊。

【当前轮次】第1轮：主诉症状收集

【问询目标】
收集患者最主要的症状信息，包括：
1. 主要不适的部位或器官
2. 症状开始出现的时间
3. 症状持续了多久
4. 症状是否持续存在还是间歇性发作

【语气要求】
- 温和专业，像门诊医生一样问诊
- 一次只问1-2个关键问题
- 避免医学术语，用患者能理解的语言
- 如患者描述模糊，适当追问确认

【开场白】
请先说："您好，欢迎来到AI智能诊室。我将协助您进行健康咨询，为了更好地帮助您，请详细描述您的症状。"

然后开始第1轮问询，询问患者的主要不适症状。`,

	2: `你是AI智能诊室的医生，请以专业、友好、耐心的态度与患者进行问诊。

【当前轮次】第2轮：症状详情追问

【问询目标】
在收集到主诉症状后，进一步追问以下信息：
1. 症状的具体表现（如疼痛的性质：胀痛、刺痛、烧灼感等）
2. 症状的严重程度（轻/中/重度，0-10分评分）
3. 什么情况下症状会加重（如活动、情绪、饮食、体位变化等）
4. 什么情况下症状会缓解（如休息、服药、体位改变等）
5. 是否有其他伴随症状

【语气要求】
- 继续以门诊医生的口吻进行追问
- 一次问一个方面，不要一次性问太多
- 对患者回答给予适当回应和确认
- 如患者提到新的重要症状，及时记录并追问

【承接上一轮】
根据患者在第1轮描述的主诉症状，有针对性地追问症状的详细信息。`,

	3: `你是AI智能诊室的医生，请以专业、友好、耐心的态度与患者进行问诊。

【当前轮次】第3轮：过往病史询问

【问询目标】
了解患者既往的医疗经历：
1. 是否患有慢性疾病（如高血压、糖尿病、心脏病、哮喘等）
2. 过往是否有住院治疗经历（原因和时间）
3. 是否有过手术经历（手术类型和时间）
4. 既往检查发现的重要异常（如体检发现的结节、囊肿等）
5. 既往诊断过但可能已痊愈的疾病

【语气要求】
- 引导患者全面回忆过往病史
- 如患者不确定，可以给出常见慢性病选项供参考
- 注意区分"正在治疗中"和"曾经有过"
- 对于重大病史（如心梗、脑梗、肿瘤）要特别标注

【过渡语】
"好的，我已经了解了您的当前症状。接下来我想了解一下您的过往病史，这有助于我更全面地评估您的健康状况。"`,

	4: `你是AI智能诊室的医生，请以专业、友好、耐心的态度与患者进行问诊。

【当前轮次】第4轮：家族遗传病史

【问询目标】
了解患者直系亲属的健康状况和遗传倾向：
1. 父母、兄弟姐妹、子女的健康状况
2. 家族中是否有类似症状或疾病史
3. 家族中是否有确诊的遗传性疾病
4. 家族中是否有早发性疾病（50岁以前发病的心脑血管疾病等）
5. 家族中是否有肿瘤病史

【语气要求】
- 询问家族史要温和，尊重患者可能不太了解家族情况
- 如患者不确定，可以请其回家询问后再补充
- 对于有遗传倾向的疾病（如高血压、糖尿病、某些肿瘤）要特别说明
- 注意区分"不确定"和"确实没有"

【过渡语】
"感谢您分享过往病史。现在我想了解一下您的家族健康状况，家族病史有时候能够帮助判断某些疾病的遗传风险。"`,

	5: `你是AI智能诊室的医生，请以专业、友好、耐心的态度与患者进行问诊。

【当前轮次】第5轮：用药情况了解

【问询目标】
详细了解患者当前的用药情况：
1. 目前正在服用的所有药物（包括西药、中药、保健品）
2. 每种药物的名称和剂量
3. 服药的频率和时间
4. 是否有药物过敏史或不良反应史
5. 近期是否有新增或停用的药物
6. 是否有自行购药服用的习惯

【语气要求】
- 强调按医嘱用药的重要性
- 如患者不确定药物名称，可请其描述药物外观
- 对于多种用药的患者，帮助整理用药清单
- 特别关注可能与当前症状相关的药物

【过渡语】
"了解了您的病史和家族情况。现在我想了解一下您目前的用药情况，这对判断您的健康状况和用药安全非常重要。"`,

	6: `你是AI智能诊室的医生，请以专业、友好、耐心的态度与患者进行问诊。

【当前轮次】第6轮：个人健康状况

【问询目标】
了解患者整体健康状况和生活方式：
1. 饮食习惯（规律性、偏好类型、营养状况）
2. 运动锻炼情况（频率、强度、类型）
3. 睡眠质量（时长、质量、是否失眠）
4. 吸烟饮酒情况（如有，频率和量）
5. 工作环境和工作压力
6. 近期是否有明显的体重变化
7. 精神心理状态（是否焦虑、抑郁、压力大）

【语气要求】
- 像朋友一样关心患者的生活方式
- 不要评判患者的生活方式，而是引导其意识到可能的健康影响因素
- 对于明显不健康的生活方式，温和地指出并给出建议
- 可以适当给予健康生活方式的科普

【过渡语】
"您的用药情况很重要，我会记录下来。最后，我想了解一下您的日常生活习惯，这些信息有助于我给您更全面的健康建议。"`,

	7: `你是AI智能诊室的医生，请以专业、权威的态度，基于收集到的患者信息生成诊断建议。

【当前状态】已完成6轮问询，生成诊断报告

【收集到的患者信息】
请根据前面6轮收集的所有信息，进行综合分析。

【输出要求】
请严格按以下结构生成诊断报告：

===诊断结论===
【症状总结】简要总结患者描述的主要症状
【可能患有的疾病】根据症状分析可能的疾病（注意：仅供参考，不可替代医生诊断）
【缓解建议】症状缓解的一般建议
【何时需要就医】列出需要立即就医或尽快就医的指征
【建议就诊科室】根据症状推荐合适的就诊科室
【建议做的检查】推荐可能需要的检查项目
【建议使用的药物】推荐可能需要的药物类别（注意：具体用药需医生面诊后确定）
【注意事项】日常生活和就医的注意事项

===AI诊室咨询小结===
【患者基本信息】年龄（如提供）、性别（如提供）
【患者主诉】主要不适症状
【既往史】过往病史概述
【过敏史】过敏药物（如有）
【病情分析】AI对病情的分析
【建议科室】推荐就诊科室
【建议检查】推荐检查项目
【建议药物】推荐药物类别
【注意事项】重要注意事项

【重要声明】
1. 本报告仅供参考，不能替代医生面诊
2. 如有高危症状（如胸痛、呼吸困难、意识障碍等），请立即就医或拨打120
3. 具体诊断和用药方案请遵医嘱
4. 本AI系统不对诊疗结果承担责任`,
}

// 诊室步骤名称映射
var clinicStepNames = map[int]string{
	1: "主诉症状收集",
	2: "症状详情追问",
	3: "过往病史询问",
	4: "家族遗传病史",
	5: "用药情况了解",
	6: "个人健康状况",
	7: "生成诊断报告",
}

// AI诊室每轮的自然问询问题（只显示给用户的友好问题，不含提示词）
var clinicNaturalQuestions = map[int]string{
	1: "您好，欢迎来到AI智能诊室。我将协助您进行健康咨询。请详细描述您最主要的症状是什么？比如哪里不舒服、持续多久了？",
	2: "好的，我已经了解了您的主要症状。请问这个症状具体是怎样的？比如疼痛的性质（胀痛、刺痛等）、严重程度（0-10分），什么情况下会加重或缓解？",
	3: "感谢您的描述。接下来我想了解一下您的过往病史。您是否有慢性疾病（如高血压、糖尿病等）、住院或手术经历？",
	4: "了解了您的病史。请问您的直系亲属（父母、兄弟姐妹）中是否有人患有类似疾病或遗传性疾病？",
	5: "好的，现在我想了解一下您目前的用药情况。您正在服用哪些药物？是否有药物过敏史？",
	6: "最后，我想了解您的生活习惯。您平时的饮食、运动、睡眠情况如何？工作压力大吗？是否有吸烟饮酒习惯？",
}

func AIChat(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(aiChatRequest)
	if err := c.ShouldBindJSON(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if strings.TrimSpace(body.Message) == "" {
		appG.Response(http.StatusBadRequest, "失败", "message不能为空")
		return
	}

	sid := strings.TrimSpace(body.SessionID)
	if sid == "" {
		sid = fmt.Sprintf("sess_%d", time.Now().UnixNano())
	}
	risk := detectRisk(body.Message)

	if c.Query("stream") == "1" {
		aiChatStream(c, sid, body.Message, risk)
		return
	}

	answer := "我已收到您的问题。建议您描述症状持续时间、是否伴随发热/疼痛、既往病史，以便给出更准确建议。"
	if risk == "high" {
		answer = "检测到可能高风险症状，请立即前往急诊或拨打120。AI助手不能替代急救与临床诊断。"
	} else {
		if modelReply, ok := tryModelReply(body.Message, risk); ok {
			answer = modelReply
		}
	}

	aiMsg := saveChatMessages(sid, body.Message, answer, risk)
	data := map[string]interface{}{
		"session_id": sid,
		"reply_id":   aiMsg.ID,
		"answer":     aiMsg.Content,
		"risk_level": risk,
		"actions": []map[string]interface{}{
			{"type": "go_register", "label": "去挂号", "target": "/outpatient/register"},
			{"type": "go_queue", "label": "查看排队", "target": "/outpatient/queue"},
		},
		"created_at": aiMsg.CreatedAt,
	}
	appG.Response(http.StatusOK, "成功", data)
}

func aiChatStream(c *gin.Context, sid, userMessage, risk string) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	c.SSEvent("start", gin.H{"session_id": sid, "risk_level": risk})
	c.Writer.Flush()

	var fullAnswer strings.Builder
	writeDelta := func(delta string) {
		if strings.TrimSpace(delta) == "" {
			return
		}
		fullAnswer.WriteString(delta)
		c.SSEvent("delta", gin.H{"content": delta})
		c.Writer.Flush()
	}

	if risk == "high" {
		msg := "检测到可能高风险症状，请立即前往急诊或拨打120。AI助手不能替代急救与临床诊断。"
		for _, r := range []rune(msg) {
			writeDelta(string(r))
			time.Sleep(15 * time.Millisecond)
		}
	} else {
		if ok := tryModelReplyStream(userMessage, risk, writeDelta); !ok {
			fallback := "我已收到您的问题。建议您描述症状持续时间、是否伴随发热/疼痛、既往病史，以便给出更准确建议。"
			for _, r := range []rune(fallback) {
				writeDelta(string(r))
			}
		}
	}

	answer := strings.TrimSpace(fullAnswer.String())
	if answer == "" {
		answer = "暂时无法回答，请稍后重试。"
	}
	aiMsg := saveChatMessages(sid, userMessage, answer, risk)

	c.SSEvent("done", gin.H{
		"session_id": sid,
		"reply_id":   aiMsg.ID,
		"answer":     aiMsg.Content,
		"risk_level": risk,
		"created_at": aiMsg.CreatedAt,
	})
	c.Writer.Flush()
}

func saveChatMessages(sid, question, answer, risk string) aiMessage {
	now := time.Now().Format(time.RFC3339)
	userMsg := aiMessage{ID: fmt.Sprintf("u_%d", time.Now().UnixNano()), Role: "user", Content: question, CreatedAt: now}
	aiMsg := aiMessage{ID: fmt.Sprintf("ai_%d", time.Now().UnixNano()+1), Role: "assistant", Content: answer, RiskLevel: risk, CreatedAt: now}

	aiStoreMu.Lock()
	s, ok := aiStore[sid]
	if !ok {
		title := question
		if len([]rune(title)) > 12 {
			title = string([]rune(title)[:12]) + "..."
		}
		s = &aiSession{SessionID: sid, Title: title}
		aiStore[sid] = s
	}
	s.Messages = append(s.Messages, userMsg, aiMsg)
	s.LastMessage = aiMsg.Content
	s.UpdatedAt = now
	aiStoreMu.Unlock()

	return aiMsg
}

func AITriage(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(aiTriageRequest)
	if err := c.ShouldBindJSON(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	joined := strings.Join(body.Symptoms, " ") + " " + body.Extra
	risk := detectRisk(joined)

	triagePrompt := triageSystemPrompt + "\n\n用户症状描述：" + strings.TrimSpace(joined) + "\n\n请严格按模板输出。"
	answer, err := callProvider("qwen", triagePrompt)
	if err != nil || strings.TrimSpace(answer) == "" {
		answer, err = callProvider("deepseek", triagePrompt)
	}
	if err != nil || strings.TrimSpace(answer) == "" {
		answer = "- 结论摘要：信息不足，建议完善症状描述后再试。\n- 风险等级：中\n- 建议行动（最多3条）：\n  1) 补充症状持续时间、伴随症状、既往病史\n  2) 若症状加重及时线下就医\n  3) 可先预约全科医学科进一步评估\n- 可选就诊科室：全科医学科\n- 免责声明：本建议仅供健康管理参考，不构成医疗诊断。"
	}

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"risk_level":               risk,
		"triage_output":            answer,
		"is_emergency_recommended": risk == "high",
	})
}

func AIRehabCompanion(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(aiRehabCompanionRequest)
	if err := c.ShouldBindJSON(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if strings.TrimSpace(body.PatientID) == "" || strings.TrimSpace(body.RecordID) == "" {
		appG.Response(http.StatusBadRequest, "失败", "patient_id和record_id不能为空")
		return
	}

	rec, err := queryPrescriptionByRecord(body.PatientID, body.RecordID)
	if err != nil {
		appG.Response(http.StatusBadRequest, "失败", err.Error())
		return
	}

	prompt := buildRehabPrompt(rec, body.UserPrompt)
	answer, err := callProvider("qwen", prompt)
	if err != nil || strings.TrimSpace(answer) == "" {
		answer, err = callProvider("deepseek", prompt)
	}
	if err != nil || strings.TrimSpace(answer) == "" {
		answer = "- 康复阶段判断：请结合临床复诊进一步评估。\n- 用药提醒：\n  1) 按医嘱规律服药\n  2) 不擅自停药或加量\n  3) 若不适及时就医\n- 复诊建议：建议按医生安排按时复诊。\n- 注意事项：\n  1) 规律作息\n  2) 清淡饮食\n  3) 记录症状变化\n- 预警信号（出现即尽快就医）：胸痛、呼吸困难、持续高热等。\n- 免责声明：本建议仅供康复管理参考，不构成诊断或处方调整。"
	}

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"risk_level":   detectRisk(strings.Join(extractSymptomTexts(rec), " ")),
		"rehab_output": answer,
	})
}

func AIReportTranslator(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(aiReportTranslatorRequest)
	if err := c.ShouldBindJSON(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if strings.TrimSpace(body.PatientID) == "" || strings.TrimSpace(body.RecordID) == "" {
		appG.Response(http.StatusBadRequest, "失败", "patient_id和record_id不能为空")
		return
	}

	rec, err := queryPrescriptionByRecord(body.PatientID, body.RecordID)
	if err != nil {
		appG.Response(http.StatusBadRequest, "失败", err.Error())
		return
	}

	prompt := buildTranslatorPrompt(rec, body.UserQuestion)
	answer, err := callProvider("qwen", prompt)
	if err != nil || strings.TrimSpace(answer) == "" {
		answer, err = callProvider("deepseek", prompt)
	}
	if err != nil || strings.TrimSpace(answer) == "" {
		answer = "- 一句话总结：当前报告信息建议结合线下医生解读。\n- 术语翻译（专业词 -> 大白话）：\n  1) 诊断结果：医生对当前病情的专业判断\n  2) 检查结果：化验/影像给出的客观信息\n  3) 治疗方案：接下来建议的处理方式\n- 检查/治疗在说什么：重点是评估病情变化并指导后续治疗。\n- 我现在该做什么（最多3条）：\n  1) 按医嘱规范用药\n  2) 记录症状变化\n  3) 按时复诊\n- 什么时候尽快就医：症状明显加重或出现危险信号时。\n- 免责声明：本解释用于信息理解，不构成诊断或治疗方案。"
	}
	answer = normalizeTranslatorOutput(answer)

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"risk_level":        detectRisk(strings.Join(extractSymptomTexts(rec), " ")),
		"translator_output": answer,
	})
}

func AIGetSessions(c *gin.Context) {
	appG := app.Gin{C: c}
	aiStoreMu.RLock()
	items := make([]map[string]interface{}, 0, len(aiStore))
	for _, s := range aiStore {
		items = append(items, map[string]interface{}{
			"session_id":     s.SessionID,
			"title":          s.Title,
			"last_message":   s.LastMessage,
			"updated_at":     s.UpdatedAt,
			"message_length": len(s.Messages),
		})
	}
	aiStoreMu.RUnlock()
	appG.Response(http.StatusOK, "成功", map[string]interface{}{"items": items, "total": len(items)})
}

func AIGetSessionMessages(c *gin.Context) {
	appG := app.Gin{C: c}
	sid := c.Param("id")
	aiStoreMu.RLock()
	s, ok := aiStore[sid]
	aiStoreMu.RUnlock()
	if !ok {
		appG.Response(http.StatusOK, "成功", map[string]interface{}{"session_id": sid, "messages": []aiMessage{}})
		return
	}
	appG.Response(http.StatusOK, "成功", map[string]interface{}{"session_id": sid, "messages": s.Messages})
}

func detectRisk(content string) string {
	highWords := []string{"胸痛", "呼吸困难", "昏迷", "抽搐", "大出血", "意识障碍"}
	for _, w := range highWords {
		if strings.Contains(content, w) {
			return "high"
		}
	}
	return "medium"
}

func tryModelReply(userMessage, risk string) (string, bool) {
	provider := "deepseek"
	if risk == "high" {
		provider = "qwen"
	}
	reply, err := callProvider(provider, userMessage)
	if err == nil && strings.TrimSpace(reply) != "" {
		return reply, true
	}
	alt := "qwen"
	if provider == "qwen" {
		alt = "deepseek"
	}
	reply, err = callProvider(alt, userMessage)
	if err == nil && strings.TrimSpace(reply) != "" {
		return reply, true
	}
	return "", false
}

func tryModelReplyStream(userMessage, risk string, onDelta func(string)) bool {
	provider := "deepseek"
	if risk == "high" {
		provider = "qwen"
	}
	if err := streamProvider(provider, userMessage, onDelta); err == nil {
		return true
	}
	alt := "qwen"
	if provider == "qwen" {
		alt = "deepseek"
	}
	return streamProvider(alt, userMessage, onDelta) == nil
}

func loadDotEnvOnce() {
	dotenvOnce.Do(func() {
		tryLoadEnvFile(".env")
		tryLoadEnvFile("application/server/.env")
	})
}

func tryLoadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		val = strings.Trim(val, `"'`)
		if key == "" {
			continue
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
}

func callProvider(provider, userMessage string) (string, error) {
	return callProviderWithTimeout(provider, userMessage, 30*time.Second)
}

func callProviderWithTimeout(provider, userMessage string, timeout time.Duration) (string, error) {
	loadDotEnvOnce()
	baseURL, apiKey, model := providerConfig(provider)
	if baseURL == "" || apiKey == "" {
		return "", fmt.Errorf("provider config missing")
	}

	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "你是门诊AI健康助手，提供健康管理建议，不做诊断。"},
			{"role": "user", "content": userMessage},
		},
	}
	bs, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/chat/completions", bytes.NewBuffer(bs))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("status=%d", resp.StatusCode)
	}

	var out map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	choices, ok := out["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	item, _ := choices[0].(map[string]interface{})
	msg, _ := item["message"].(map[string]interface{})
	content, _ := msg["content"].(string)
	return strings.TrimSpace(content), nil
}

func streamProvider(provider, userMessage string, onDelta func(string)) error {
	loadDotEnvOnce()
	baseURL, apiKey, model := providerConfig(provider)
	if baseURL == "" || apiKey == "" {
		return fmt.Errorf("provider config missing")
	}

	payload := map[string]interface{}{
		"model":  model,
		"stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": "你是门诊AI健康助手，提供健康管理建议，不做诊断。"},
			{"role": "user", "content": userMessage},
		},
	}
	bs, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/chat/completions", bytes.NewBuffer(bs))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("status=%d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var item map[string]interface{}
		if err := json.Unmarshal([]byte(data), &item); err != nil {
			continue
		}
		choices, _ := item["choices"].([]interface{})
		if len(choices) == 0 {
			continue
		}
		c0, _ := choices[0].(map[string]interface{})
		delta, _ := c0["delta"].(map[string]interface{})
		content, _ := delta["content"].(string)
		if content != "" {
			onDelta(content)
		}
	}
	return scanner.Err()
}

func queryPrescriptionByRecord(patientID, recordID string) (map[string]interface{}, error) {
	resp, err := bc.ChannelQuery("queryPrescription", [][]byte{[]byte(patientID)})
	if err != nil {
		return nil, err
	}
	var list []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &list); err != nil {
		return nil, err
	}
	for _, item := range list {
		id, _ := item["id"].(string)
		if id == recordID {
			return item, nil
		}
	}
	return nil, fmt.Errorf("未找到指定病历")
}

func extractSymptomTexts(rec map[string]interface{}) []string {
	keys := []string{"chief_complaint", "present_illness", "diagnosis_result", "doctor_diagnosis", "diagnosis"}
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if v, ok := rec[k].(string); ok && strings.TrimSpace(v) != "" {
			out = append(out, v)
		}
	}
	return out
}

func buildRehabPrompt(rec map[string]interface{}, userPrompt string) string {
	return rehabCompanionPrompt + "\n\n[病历字段]\n" +
		"- 诊断结果: " + stringify(rec["diagnosis_result"]) + "\n" +
		"- 治疗方案: " + stringify(rec["treatment_plan"]) + "\n" +
		"- 用药建议: " + stringify(rec["medication_advice"]) + "\n" +
		"- 医嘱: " + stringify(rec["doctor_advice"]) + "\n" +
		"- 主诉: " + stringify(rec["chief_complaint"]) + "\n" +
		"- 现病史: " + stringify(rec["present_illness"]) + "\n" +
		"- 检查: " + stringify(rec["lab_exam"]) + " " + stringify(rec["imaging_exam"]) + "\n\n" +
		"[用户单独提示词]\n" + strings.TrimSpace(userPrompt)
}

func buildTranslatorPrompt(rec map[string]interface{}, userPrompt string) string {
	return reportTranslatorPrompt + "\n\n[病历字段-诊断信息]\n" +
		"- 主诉: " + stringify(rec["chief_complaint"]) + "\n" +
		"- 现病史: " + stringify(rec["present_illness"]) + "\n" +
		"- 既往史: " + stringify(rec["past_history"]) + "\n" +
		"- 过敏史: " + stringify(rec["allergy_history"]) + "\n" +
		"- 家族史: " + stringify(rec["family_history"]) + "\n\n" +
		"[病历字段-检查治疗]\n" +
		"- 体温: " + stringify(rec["temperature"]) + "\n" +
		"- 脉搏: " + stringify(rec["pulse"]) + "\n" +
		"- 血压: " + stringify(rec["blood_pressure"]) + "\n" +
		"- 呼吸: " + stringify(rec["respiration"]) + "\n" +
		"- 体格检查: " + stringify(rec["physical_exam"]) + "\n" +
		"- 实验室检查: " + stringify(rec["lab_exam"]) + "\n" +
		"- 影像学检查: " + stringify(rec["imaging_exam"]) + "\n" +
		"- 诊断结果: " + stringify(rec["diagnosis_result"]) + "\n" +
		"- 治疗方案: " + stringify(rec["treatment_plan"]) + "\n" +
		"- 处方用药: " + stringify(rec["medication_advice"]) + "\n" +
		"- 医嘱: " + stringify(rec["doctor_advice"]) + "\n\n" +
		"[用户单独提示词]\n" + strings.TrimSpace(userPrompt)
}

func normalizeTranslatorOutput(answer string) string {
	required := []string{
		"- 一句话总结：",
		"- 术语翻译（专业词 -> 大白话）：",
		"- 检查/治疗在说什么：",
		"- 我现在该做什么（最多3条）：",
		"- 什么时候尽快就医：",
		"- 免责声明：",
	}

	out := strings.TrimSpace(answer)
	if !strings.Contains(out, "- 一句话总结：") {
		out = "- 一句话总结：当前报告涉及的关键信息建议结合临床医生综合判断。\n" + out
	}
	if !strings.Contains(out, "- 术语翻译（专业词 -> 大白话）：") {
		out += "\n- 术语翻译（专业词 -> 大白话）：\n  1) 医学术语表示医生的专业判断\n  2) 检查结果是客观检测信息\n  3) 治疗方案是后续处理建议"
	}
	if !strings.Contains(out, "- 检查/治疗在说什么：") {
		out += "\n- 检查/治疗在说什么：这些信息主要用于评估病情程度与下一步治疗方向。"
	}
	if !strings.Contains(out, "- 我现在该做什么（最多3条）：") {
		out += "\n- 我现在该做什么（最多3条）：\n  1) 按医嘱执行治疗\n  2) 记录异常症状变化\n  3) 按时复诊复查"
	}
	if !strings.Contains(out, "- 什么时候尽快就医：") {
		out += "\n- 什么时候尽快就医：若症状明显加重、突发胸痛呼吸困难、持续高热等，请尽快就医。"
	}
	if !strings.Contains(out, "- 免责声明：") {
		out += "\n- 免责声明：本解释用于信息理解，不构成诊断或治疗方案。"
	}

	// ensure order
	lines := strings.Split(out, "\n")
	sections := map[string][]string{}
	current := ""
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		matched := false
		for _, k := range required {
			if strings.HasPrefix(trim, k) {
				current = k
				sections[current] = append(sections[current], trim)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		if current != "" && trim != "" {
			sections[current] = append(sections[current], line)
		}
	}

	ordered := make([]string, 0)
	for _, k := range required {
		part := sections[k]
		if len(part) == 0 {
			ordered = append(ordered, k+" ...")
			continue
		}
		ordered = append(ordered, part...)
	}
	return strings.Join(ordered, "\n")
}

func stringify(v interface{}) string {
	s, _ := v.(string)
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func providerConfig(provider string) (baseURL, apiKey, model string) {
	if provider == "qwen" {
		baseURL = strings.TrimSpace(os.Getenv("AI_QWEN_BASE_URL"))
		apiKey = strings.TrimSpace(os.Getenv("AI_QWEN_API_KEY"))
		model = strings.TrimSpace(os.Getenv("AI_QWEN_MODEL"))
		if model == "" {
			model = "qwen-plus"
		}
		return
	}
	baseURL = strings.TrimSpace(os.Getenv("AI_DEEPSEEK_BASE_URL"))
	apiKey = strings.TrimSpace(os.Getenv("AI_DEEPSEEK_API_KEY"))
	model = strings.TrimSpace(os.Getenv("AI_DEEPSEEK_MODEL"))
	if model == "" {
		model = "deepseek-chat"
	}
	return
}

// ==================== AI诊室功能 ====================

// AIClinicStart 开始AI诊室会话
func AIClinicStart(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(aiClinicRequest)
	if err := c.ShouldBindJSON(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}

	sid := strings.TrimSpace(body.SessionID)
	if sid == "" {
		sid = fmt.Sprintf("clinic_%d", time.Now().UnixNano())
	}

	now := time.Now().Format(time.RFC3339)
	session := &aiClinicSession{
		SessionID:     sid,
		Title:         "AI诊室咨询",
		CurrentStep:   1,
		Progress:      0,
		CollectedData: clinicCollectedData{},
		Messages:      []aiMessage{},
		StartedAt:     now,
		UpdatedAt:     now,
	}

	aiStoreMu.Lock()
	clinicStore[sid] = session
	aiStoreMu.Unlock()

	// 获取第1轮问询内容（使用自然友好的问询问题）
	question := clinicNaturalQuestions[1]
	aiMsg := aiMessage{
		ID:        fmt.Sprintf("clinic_ai_%d", time.Now().UnixNano()),
		Role:      "assistant",
		Content:   question,
		RiskLevel: "low",
		CreatedAt: now,
	}

	aiStoreMu.Lock()
	session.Messages = append(session.Messages, aiMsg)
	session.UpdatedAt = now
	aiStoreMu.Unlock()

	data := map[string]interface{}{
		"session_id":   sid,
		"message_id":   aiMsg.ID,
		"current_step": 1,
		"step_name":    clinicStepNames[1],
		"progress":     0,
		"content":      aiMsg.Content,
		"created_at":   now,
	}
	appG.Response(http.StatusOK, "成功", data)
}

// AIClinicChat 处理AI诊室对话
func AIClinicChat(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(aiClinicRequest)
	if err := c.ShouldBindJSON(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}

	sid := strings.TrimSpace(body.SessionID)
	if sid == "" {
		appG.Response(http.StatusBadRequest, "失败", "session_id不能为空")
		return
	}

	aiStoreMu.Lock()
	session, ok := clinicStore[sid]
	aiStoreMu.Unlock()
	if !ok {
		appG.Response(http.StatusBadRequest, "失败", "未找到对应的诊室会话，请重新开始")
		return
	}

	// 保存用户回答
	now := time.Now().Format(time.RFC3339)
	userMsg := aiMessage{
		ID:        fmt.Sprintf("clinic_user_%d", time.Now().UnixNano()),
		Role:      "user",
		Content:   body.UserResponse,
		CreatedAt: now,
	}

	aiStoreMu.Lock()
	session.Messages = append(session.Messages, userMsg)
	// 保存本轮回答到对应字段
	saveClinicData(session, session.CurrentStep, body.UserResponse)
	aiStoreMu.Unlock()

	// 检查是否已完成6轮，触发报告生成
	if session.CurrentStep >= 6 {
		// 生成诊断报告
		report := buildClinicReport(session.CollectedData)
		session.Report = report

		// 提取咨询小结
		summary := extractClinicSummary(report)
		session.Summary = summary

		session.CurrentStep = 7
		session.Progress = 100

		reportMsg := aiMessage{
			ID:        fmt.Sprintf("clinic_report_%d", time.Now().UnixNano()),
			Role:      "assistant",
			Content:   report,
			RiskLevel: detectClinicRisk(session.CollectedData),
			CreatedAt: now,
		}

		aiStoreMu.Lock()
		session.Messages = append(session.Messages, reportMsg)
		session.UpdatedAt = now
		aiStoreMu.Unlock()

		appG.Response(http.StatusOK, "成功", map[string]interface{}{
			"session_id":   sid,
			"message_id":   userMsg.ID,
			"ai_message":   reportMsg.Content,
			"current_step": 7,
			"step_name":    clinicStepNames[7],
			"progress":     100,
			"is_complete":  true,
			"report":       report,
			"summary":      summary,
			"risk_level":   reportMsg.RiskLevel,
			"created_at":   now,
		})
		return
	}

	// 进入下一轮（使用自然友好的问询问题）
	nextStep := session.CurrentStep + 1
	question := clinicNaturalQuestions[nextStep]

	session.CurrentStep = nextStep
	session.Progress = float64((nextStep - 1) * 100 / 6)

	aiMsg := aiMessage{
		ID:        fmt.Sprintf("clinic_ai_%d", time.Now().UnixNano()),
		Role:      "assistant",
		Content:   question,
		RiskLevel: "low",
		CreatedAt: now,
	}

	aiStoreMu.Lock()
	session.Messages = append(session.Messages, aiMsg)
	session.UpdatedAt = now
	aiStoreMu.Unlock()

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"session_id":   sid,
		"message_id":   userMsg.ID,
		"ai_message":   aiMsg.Content,
		"current_step": nextStep,
		"step_name":    clinicStepNames[nextStep],
		"progress":     session.Progress,
		"is_complete":  false,
		"created_at":   now,
	})
}

// AIClinicReport 获取诊室诊断报告
func AIClinicReport(c *gin.Context) {
	appG := app.Gin{C: c}
	sid := c.Param("id")

	aiStoreMu.RLock()
	session, ok := clinicStore[sid]
	aiStoreMu.RUnlock()
	if !ok {
		appG.Response(http.StatusBadRequest, "失败", "未找到对应的诊室会话")
		return
	}

	if session.Report == "" {
		appG.Response(http.StatusOK, "成功", map[string]interface{}{
			"session_id":   sid,
			"is_complete":  false,
			"report":       "",
			"summary":      "",
			"current_step": session.CurrentStep,
			"progress":     session.Progress,
		})
		return
	}

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"session_id":     sid,
		"is_complete":    true,
		"report":         session.Report,
		"summary":        session.Summary,
		"risk_level":     detectClinicRisk(session.CollectedData),
		"current_step":   session.CurrentStep,
		"progress":       session.Progress,
		"collected_data": session.CollectedData,
	})
}

// AIClinicGetSession 获取诊室会话状态
func AIClinicGetSession(c *gin.Context) {
	appG := app.Gin{C: c}
	sid := c.Param("id")

	aiStoreMu.RLock()
	session, ok := clinicStore[sid]
	aiStoreMu.RUnlock()
	if !ok {
		appG.Response(http.StatusOK, "成功", map[string]interface{}{
			"session_id": sid,
			"exists":     false,
			"messages":   []aiMessage{},
		})
		return
	}

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"session_id":   sid,
		"exists":       true,
		"title":        session.Title,
		"current_step": session.CurrentStep,
		"step_name":    clinicStepNames[session.CurrentStep],
		"progress":     session.Progress,
		"messages":     session.Messages,
		"report":       session.Report,
		"summary":      session.Summary,
		"started_at":   session.StartedAt,
		"updated_at":   session.UpdatedAt,
	})
}

// AIClinicReset 重置诊室会话
func AIClinicReset(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(aiClinicRequest)
	if err := c.ShouldBindJSON(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}

	sid := strings.TrimSpace(body.SessionID)
	if sid == "" {
		appG.Response(http.StatusBadRequest, "失败", "session_id不能为空")
		return
	}

	aiStoreMu.Lock()
	delete(clinicStore, sid)
	aiStoreMu.Unlock()

	appG.Response(http.StatusOK, "成功", map[string]interface{}{
		"session_id": sid,
		"reset":      true,
	})
}

// 保存诊室收集的数据
func saveClinicData(session *aiClinicSession, step int, response string) {
	switch step {
	case 1:
		session.CollectedData.ChiefComplaint = response
	case 2:
		session.CollectedData.SymptomDetails = response
	case 3:
		session.CollectedData.PastMedicalHistory = response
	case 4:
		session.CollectedData.FamilyHistory = response
	case 5:
		session.CollectedData.CurrentMedications = response
	case 6:
		session.CollectedData.LifestyleHabits = response
	}
}

// 构建诊断报告
func buildClinicReport(data clinicCollectedData) string {
	prompt := clinicStepPrompts[7] + "\n\n" +
		"【已收集的患者信息】\n\n" +
		"【第1轮-主诉症状】\n" + data.ChiefComplaint + "\n\n" +
		"【第2轮-症状详情】\n" + data.SymptomDetails + "\n\n" +
		"【第3轮-过往病史】\n" + data.PastMedicalHistory + "\n\n" +
		"【第4轮-家族遗传病史】\n" + data.FamilyHistory + "\n\n" +
		"【第5轮-用药情况】\n" + data.CurrentMedications + "\n\n" +
		"【第6轮-个人健康状况】\n" + data.LifestyleHabits + "\n\n" +
		"请根据以上收集的信息，严格按模板格式生成诊断报告。"

	answer, err := callProviderWithTimeout("deepseek", prompt, 60*time.Second)
	if err != nil || strings.TrimSpace(answer) == "" {
		answer, err = callProviderWithTimeout("qwen", prompt, 60*time.Second)
	}
	if err != nil || strings.TrimSpace(answer) == "" {
		// 兜底报告
		return `===诊断结论===
【症状总结】` + data.ChiefComplaint + `
【可能患有的疾病】根据描述，症状可能涉及多个系统，建议进一步检查以明确诊断。
【缓解建议】注意休息，保持良好生活习惯，避免诱因。
【何时需要就医】如症状持续加重、出现高热、胸痛、呼吸困难等请立即就医。
【建议就诊科室】建议就诊内科或全科医学科。
【建议做的检查】血常规、尿常规、心电图、胸片（根据症状选择）。
【建议使用的药物】具体用药需面诊后由医生确定，请勿自行用药。
【注意事项】保持充足睡眠，清淡饮食，如有不适及时就诊。

===AI诊室咨询小结===
【患者基本信息】待补充
【患者主诉】` + data.ChiefComplaint + `
【既往史】` + data.PastMedicalHistory + `
【过敏史】待询问确认
【病情分析】根据收集的信息，需要进一步检查以明确诊断。
【建议科室】内科/全科医学科
【建议检查】常规检查及针对性检查
【建议药物】待医生面诊后确定
【注意事项】如有紧急症状请立即就医，本建议仅供参考。

【重要声明】
1. 本报告仅供参考，不能替代医生面诊
2. 如有高危症状（如胸痛、呼吸困难、意识障碍等），请立即就医或拨打120
3. 具体诊断和用药方案请遵医嘱
4. 本AI系统不对诊疗结果承担责任`
	}
	return answer
}

// 提取咨询小结
func extractClinicSummary(report string) string {
	lines := strings.Split(report, "\n")
	var summaryBuilder strings.Builder
	inSummary := false
	for _, line := range lines {
		if strings.Contains(line, "===AI诊室咨询小结===") || strings.Contains(line, "【患者基本信息】") {
			inSummary = true
		}
		if inSummary {
			summaryBuilder.WriteString(line)
			summaryBuilder.WriteString("\n")
			if strings.Contains(line, "【重要声明】") {
				break
			}
		}
	}
	result := strings.TrimSpace(summaryBuilder.String())
	if result == "" {
		return "已完成AI诊室咨询，详见上方诊断报告。"
	}
	return result
}

// 检测诊室风险等级
func detectClinicRisk(data clinicCollectedData) string {
	content := data.ChiefComplaint + " " + data.SymptomDetails
	highWords := []string{"胸痛", "呼吸困难", "昏迷", "抽搐", "大出血", "意识障碍", "剧烈", "持续高热"}
	for _, w := range highWords {
		if strings.Contains(content, w) {
			return "high"
		}
	}
	return "medium"
}

// 流式输出诊室对话
func clinicChatStream(c *gin.Context, sid string, userResponse string, currentStep int) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	nextStep := currentStep + 1
	progress := float64((nextStep - 1) * 100 / 6)

	c.SSEvent("step", gin.H{
		"session_id":   sid,
		"current_step": nextStep,
		"step_name":    clinicStepNames[nextStep],
		"progress":     progress,
		"is_complete":  nextStep > 6,
	})
	c.Writer.Flush()

	if nextStep > 6 {
		// 生成报告
		aiStoreMu.RLock()
		session := clinicStore[sid]
		aiStoreMu.RUnlock()

		if session != nil {
			report := buildClinicReport(session.CollectedData)
			summary := extractClinicSummary(report)
			risk := detectClinicRisk(session.CollectedData)

			c.SSEvent("report", gin.H{
				"session_id": sid,
				"report":     report,
				"summary":    summary,
				"risk_level": risk,
			})
			c.Writer.Flush()
		}
		return
	}

	question := clinicStepPrompts[nextStep]
	var fullAnswer strings.Builder
	writeDelta := func(delta string) {
		if strings.TrimSpace(delta) == "" {
			return
		}
		fullAnswer.WriteString(delta)
		c.SSEvent("delta", gin.H{"content": delta})
		c.Writer.Flush()
	}

	for _, r := range []rune(question) {
		writeDelta(string(r))
		time.Sleep(10 * time.Millisecond)
	}

	c.SSEvent("done", gin.H{
		"session_id":   sid,
		"ai_content":   fullAnswer.String(),
		"current_step": nextStep,
		"step_name":    clinicStepNames[nextStep],
		"progress":     progress,
	})
	c.Writer.Flush()
}
