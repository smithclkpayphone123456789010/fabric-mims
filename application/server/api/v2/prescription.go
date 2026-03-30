package v2

import (
	bc "application/blockchain"
	"application/model"
	"application/pkg/app"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxUploadSize = 200 * 1024 * 1024 // 200MB
)

var (
	allowedExt = map[string]bool{
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
)

func CreatePrescription(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.PrescriptionRequestBody)

	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}

	if body.RecordType == "" {
		body.RecordType = "门诊病历"
	}
	validRecordType := map[string]bool{
		"门诊病历": true,
		"住院病历": true,
		"急诊病历": true,
		"检查报告": true,
	}
	if !validRecordType[body.RecordType] {
		appG.Response(http.StatusBadRequest, "失败", "病历类型不合法")
		return
	}
	if len(body.SymptomDescription) > 500 {
		appG.Response(http.StatusBadRequest, "失败", "症状描述不能超过500字符")
		return
	}
	if len(body.DoctorDiagnosis) > 500 {
		appG.Response(http.StatusBadRequest, "失败", "医生诊断不能超过500字符")
		return
	}
	if len(body.Comment) > 500 {
		appG.Response(http.StatusBadRequest, "失败", "备注不能超过500字符")
		return
	}
	if body.ChiefComplaint == "" || body.PresentIllness == "" || body.DiagnosisResult == "" {
		appG.Response(http.StatusBadRequest, "失败", "主诉、现病史、诊断结果为必填")
		return
	}

	file, fileHeader, err := c.Request.FormFile("record_file")
	if err == nil {
		defer file.Close()
		hashHex, encryptedPath, err := saveEncryptedFile(file, fileHeader)
		if err != nil {
			appG.Response(http.StatusBadRequest, "失败", err.Error())
			return
		}
		body.FileHash = hashHex
		body.FileName = fileHeader.Filename
		body.FilePath = encryptedPath
	}

	if body.Doctor == "" || body.Patient == "" || body.Hospital == "" || body.FileHash == "" {
		appG.Response(http.StatusBadRequest, "失败", "参数存在空值")
		return
	}

	if body.Diagnosis == "" {
		body.Diagnosis = body.DoctorDiagnosis
	}

	var bodyBytes [][]byte
	bodyBytes = append(bodyBytes, []byte(body.Doctor))
	bodyBytes = append(bodyBytes, []byte(body.Patient))
	bodyBytes = append(bodyBytes, []byte(body.RecordType))
	bodyBytes = append(bodyBytes, []byte(body.FileHash))
	bodyBytes = append(bodyBytes, []byte(body.FileName))
	bodyBytes = append(bodyBytes, []byte(body.FilePath))
	bodyBytes = append(bodyBytes, []byte(body.SymptomDescription))
	bodyBytes = append(bodyBytes, []byte(body.DoctorDiagnosis))
	bodyBytes = append(bodyBytes, []byte(body.Diagnosis))
	bodyBytes = append(bodyBytes, []byte(body.DrugName))
	bodyBytes = append(bodyBytes, []byte(body.DrugAmount))
	bodyBytes = append(bodyBytes, []byte(body.Hospital))
	bodyBytes = append(bodyBytes, []byte(body.Comment))
	bodyBytes = append(bodyBytes, []byte(body.PatientName))
	bodyBytes = append(bodyBytes, []byte(body.PatientGender))
	bodyBytes = append(bodyBytes, []byte(body.PatientAge))
	bodyBytes = append(bodyBytes, []byte(body.PatientIDCardNo))
	bodyBytes = append(bodyBytes, []byte(body.PatientPhone))
	bodyBytes = append(bodyBytes, []byte(body.InsuranceCardNo))
	bodyBytes = append(bodyBytes, []byte(body.HospitalName))
	bodyBytes = append(bodyBytes, []byte(body.Department))
	bodyBytes = append(bodyBytes, []byte(body.VisitDoctorName))
	bodyBytes = append(bodyBytes, []byte(body.ChiefComplaint))
	bodyBytes = append(bodyBytes, []byte(body.PresentIllness))
	bodyBytes = append(bodyBytes, []byte(body.PastHistory))
	bodyBytes = append(bodyBytes, []byte(body.AllergyHistory))
	bodyBytes = append(bodyBytes, []byte(body.FamilyHistory))
	bodyBytes = append(bodyBytes, []byte(body.Temperature))
	bodyBytes = append(bodyBytes, []byte(body.Pulse))
	bodyBytes = append(bodyBytes, []byte(body.BloodPressure))
	bodyBytes = append(bodyBytes, []byte(body.Respiration))
	bodyBytes = append(bodyBytes, []byte(body.PhysicalExam))
	bodyBytes = append(bodyBytes, []byte(body.LabExam))
	bodyBytes = append(bodyBytes, []byte(body.ImagingExam))
	bodyBytes = append(bodyBytes, []byte(body.DiagnosisResult))
	bodyBytes = append(bodyBytes, []byte(body.TreatmentPlan))
	bodyBytes = append(bodyBytes, []byte(body.MedicationAdvice))
	bodyBytes = append(bodyBytes, []byte(body.DoctorAdvice))

	resp, err := bc.ChannelExecute("createPrescription", bodyBytes)
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	// 记录审计日志（goroutine中无法使用c，需提前捕获变量）
	txID := string(resp.TransactionID)
	actorID := c.GetString("account_id")
	if actorID == "" {
		actorID = "0feceb66ffc1"
	}
	requestPath := c.Request.URL.Path
	requestMethod := c.Request.Method
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	go recordAuditEvent(actorID, "CREATE_RECORD", "L1", "SUCCESS", body.Patient, "", txID,
		"createPrescription", requestPath, requestMethod, clientIP, userAgent, map[string]interface{}{
			"record_id":   data["id"],
			"doctor":      body.Doctor,
			"patient":     body.Patient,
			"record_type": body.RecordType,
		}, "")

	appG.Response(http.StatusOK, "成功", data)
}

// recordAuditEvent 记录审计事件
func recordAuditEvent(actorID, eventType, eventLevel, actionResult, targetPatientID, targetRecordID, txID, chaincodeFunc, requestPath, requestMethod, clientIP, userAgent string, detail map[string]interface{}, failReason string) {
	detailJSON, _ := json.Marshal(detail)

	event := map[string]interface{}{
		"event_type":        eventType,
		"event_level":       eventLevel,
		"actor_id":          actorID,
		"target_patient_id": targetPatientID,
		"target_record_id":  targetRecordID,
		"tx_id":             txID,
		"chaincode_func":    chaincodeFunc,
		"request_path":      requestPath,
		"request_method":    requestMethod,
		"client_ip":         clientIP,
		"user_agent":        userAgent,
		"action_result":     actionResult,
		"detail_json":       string(detailJSON),
		"fail_reason":       failReason,
	}

	eventBytes, _ := json.Marshal(event)

	// 调用链码写入审计日志
	resp, err := bc.ChannelExecute("createAuditEvent", [][]byte{eventBytes})
	if err != nil {
		fmt.Printf("[AUDIT ERROR] createAuditEvent failed: %v\n", err)
		return
	}

	fmt.Printf("[AUDIT DEBUG] createAuditEvent success, txID=%s, payload=%s\n", txID, string(resp.Payload))
}

func QueryPrescriptionList(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.PrescriptionQueryRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}

	var bodyBytes [][]byte
	if body.Patient != "" {
		bodyBytes = append(bodyBytes, []byte(body.Patient))
	}
	resp, err := bc.ChannelQuery("queryPrescription", bodyBytes)
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}

	if body.DoctorID != "" {
		allowedMap := make(map[string]bool)
		accessResp, accessErr := bc.ChannelQuery("queryAccessibleRecordsByDoctor", [][]byte{[]byte(body.DoctorID), []byte(""), []byte(""), []byte(""), []byte(""), []byte("")})
		if accessErr != nil {
			appG.Response(http.StatusInternalServerError, "失败", accessErr.Error())
			return
		}
		var accessList []map[string]interface{}
		if err = json.Unmarshal(bytes.NewBuffer(accessResp.Payload).Bytes(), &accessList); err != nil {
			appG.Response(http.StatusInternalServerError, "失败", err.Error())
			return
		}
		for _, item := range accessList {
			recordObj, ok := item["record"].(map[string]interface{})
			if !ok {
				continue
			}
			recordID, _ := recordObj["id"].(string)
			if recordID != "" {
				allowedMap[recordID] = true
			}
		}
		filtered := make([]map[string]interface{}, 0)
		for _, item := range data {
			id, _ := item["id"].(string)
			if allowedMap[id] {
				filtered = append(filtered, item)
			}
		}
		appG.Response(http.StatusOK, "成功", filtered)
		return
	}

	appG.Response(http.StatusOK, "成功", data)
}

func PreviewPrescriptionFile(c *gin.Context) {
	filePath := strings.TrimSpace(c.Query("file_path"))
	fileName := strings.TrimSpace(c.Query("file_name"))
	doctorID := strings.TrimSpace(c.Query("doctor_id"))
	recordID := strings.TrimSpace(c.Query("record_id"))
	if filePath == "" {
		c.String(http.StatusBadRequest, "file_path不能为空")
		return
	}

	if doctorID != "" {
		if recordID == "" {
			c.String(http.StatusBadRequest, "record_id不能为空")
			return
		}
		accessResp, err := bc.ChannelQuery("checkRecordAccess", [][]byte{[]byte(doctorID), []byte(recordID)})
		if err != nil {
			c.String(http.StatusInternalServerError, "权限校验失败")
			return
		}
		var accessData map[string]interface{}
		if err = json.Unmarshal(bytes.NewBuffer(accessResp.Payload).Bytes(), &accessData); err != nil {
			c.String(http.StatusInternalServerError, "权限校验失败")
			return
		}
		if allowed, ok := accessData["allowed"].(bool); !ok || !allowed {
			c.String(http.StatusForbidden, "无病历访问授权")
			return
		}
	}

	cleanPath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanPath, filepath.Join("uploads", "records")) {
		c.String(http.StatusBadRequest, "非法文件路径")
		return
	}

	cipherData, err := os.ReadFile(cleanPath)
	if err != nil {
		c.String(http.StatusNotFound, "文件不存在")
		return
	}

	plainData, err := decryptAES256GCM(cipherData)
	if err != nil {
		c.String(http.StatusInternalServerError, "文件解密失败")
		return
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		ext = strings.ToLower(filepath.Ext(cleanPath))
	}

	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".pdf":
		contentType = "application/pdf"
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, contentType, plainData)
}

func saveEncryptedFile(file multipart.File, fileHeader *multipart.FileHeader) (string, string, error) {
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedExt[ext] {
		return "", "", fmt.Errorf("不支持的文件格式")
	}
	if fileHeader.Size <= 0 || fileHeader.Size > maxUploadSize {
		return "", "", fmt.Errorf("文件大小需在0-200MB之间")
	}

	plainData, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("读取文件失败: %s", err.Error())
	}

	hashBytes := sha256.Sum256(plainData)
	hashHex := hex.EncodeToString(hashBytes[:])

	cipherData, err := encryptAES256GCM(plainData)
	if err != nil {
		return "", "", fmt.Errorf("文件加密失败: %s", err.Error())
	}

	storeDir := filepath.Join("uploads", "records", time.Now().Format("20060102"))
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return "", "", fmt.Errorf("创建文件目录失败: %s", err.Error())
	}

	storedFileName := fmt.Sprintf("%d_%s.enc", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(fileHeader.Filename), ext))
	storedPath := filepath.Join(storeDir, storedFileName)
	if err := os.WriteFile(storedPath, cipherData, 0o644); err != nil {
		return "", "", fmt.Errorf("保存文件失败: %s", err.Error())
	}

	return hashHex, storedPath, nil
}

func encryptAES256GCM(plainData []byte) ([]byte, error) {
	key := sha256.Sum256([]byte("fabric-mims-file-encryption-key-2026"))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	cipherText := gcm.Seal(nil, nonce, plainData, nil)
	result := make([]byte, 0, len(nonce)+len(cipherText))
	result = append(result, nonce...)
	result = append(result, cipherText...)
	return result, nil
}

func decryptAES256GCM(cipherData []byte) ([]byte, error) {
	key := sha256.Sum256([]byte("fabric-mims-file-encryption-key-2026"))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(cipherData) < nonceSize {
		return nil, fmt.Errorf("密文长度非法")
	}
	nonce := cipherData[:nonceSize]
	enc := cipherData[nonceSize:]
	return gcm.Open(nil, nonce, enc, nil)
}
