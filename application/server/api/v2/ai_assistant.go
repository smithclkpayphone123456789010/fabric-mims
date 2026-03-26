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
	aiStoreMu  sync.RWMutex
	aiStore    = map[string]*aiSession{}
	dotenvOnce sync.Once
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

	client := &http.Client{Timeout: 30 * time.Second}
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
