package v2

import (
	bc "application/blockchain"
	"application/model"
	"application/pkg/app"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateOutpatientRegistration(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.CreateOutpatientRegistrationRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if body.PatientID == "" || body.DoctorID == "" || body.DepartmentID == "" || body.SlotID == "" || body.VisitDate == "" {
		appG.Response(http.StatusBadRequest, "失败", "参数存在空值")
		return
	}
	resp, err := bc.ChannelExecute("createOutpatientRegistration", [][]byte{[]byte(body.PatientID), []byte(body.DoctorID), []byte(body.DepartmentID), []byte(body.SlotID), []byte(body.VisitDate)})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	data["tx_id"] = resp.TransactionID
	appG.Response(http.StatusOK, "成功", data)
}

func CancelOutpatientRegistration(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.CancelOutpatientRegistrationRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	resp, err := bc.ChannelExecute("cancelOutpatientRegistration", [][]byte{[]byte(body.RegistrationID), []byte(body.OperatorID)})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	data["tx_id"] = resp.TransactionID
	appG.Response(http.StatusOK, "成功", data)
}

func QueryOutpatientRegistration(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.QueryOutpatientRegistrationRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	resp, err := bc.ChannelQuery("queryOutpatientRegistration", [][]byte{})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var list []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &list); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	filtered := make([]map[string]interface{}, 0)
	for _, item := range list {
		if body.PatientID != "" && fmt.Sprint(item["patient_id"]) != body.PatientID {
			continue
		}
		if body.DoctorID != "" && fmt.Sprint(item["doctor_id"]) != body.DoctorID {
			continue
		}
		if body.Status != "" && fmt.Sprint(item["status"]) != body.Status {
			continue
		}
		filtered = append(filtered, item)
	}
	appG.Response(http.StatusOK, "成功", filtered)
}

func CreateScheduleSlot(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.CreateScheduleSlotRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	resp, err := bc.ChannelExecute("createScheduleSlot", [][]byte{[]byte(body.DoctorID), []byte(body.DepartmentID), []byte(body.VisitDate), []byte(body.StartTime), []byte(body.EndTime), []byte(fmt.Sprintf("%d", body.Capacity))})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	data["tx_id"] = resp.TransactionID
	appG.Response(http.StatusOK, "成功", data)
}

func QueryScheduleSlot(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.QueryScheduleSlotRequestBody)
	_ = c.ShouldBind(body)
	resp, err := bc.ChannelQuery("queryScheduleSlot", [][]byte{})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var list []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &list); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	filtered := make([]map[string]interface{}, 0)
	for _, item := range list {
		if body.DepartmentID != "" && fmt.Sprint(item["department_id"]) != body.DepartmentID {
			continue
		}
		if body.DoctorID != "" && fmt.Sprint(item["doctor_id"]) != body.DoctorID {
			continue
		}
		if body.VisitDate != "" && fmt.Sprint(item["visit_date"]) != body.VisitDate {
			continue
		}
		filtered = append(filtered, item)
	}
	appG.Response(http.StatusOK, "成功", filtered)
}

func QueryOutpatientPayment(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.QueryOutpatientPaymentRequestBody)
	_ = c.ShouldBind(body)
	resp, err := bc.ChannelQuery("queryOutpatientPayment", [][]byte{})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var list []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &list); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	filtered := make([]map[string]interface{}, 0)
	for _, item := range list {
		if body.PatientID != "" && fmt.Sprint(item["patient_id"]) != body.PatientID {
			continue
		}
		if body.Status != "" && fmt.Sprint(item["status"]) != body.Status {
			continue
		}
		filtered = append(filtered, item)
	}
	appG.Response(http.StatusOK, "成功", filtered)
}

func PayOutpatientOrder(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.PayOutpatientOrderRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	resp, err := bc.ChannelExecute("payOutpatientOrder", [][]byte{[]byte(body.PaymentID), []byte(body.PatientID)})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	data["tx_id"] = resp.TransactionID
	appG.Response(http.StatusOK, "成功", data)
}

func QueryOutpatientQueue(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.QueryOutpatientQueueRequestBody)
	if err := c.ShouldBind(body); err != nil || strings.TrimSpace(body.DoctorID) == "" {
		appG.Response(http.StatusBadRequest, "失败", "doctor_id不能为空")
		return
	}
	resp, err := bc.ChannelQuery("queryOutpatientQueue", [][]byte{[]byte(body.DoctorID)})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var list []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &list); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	appG.Response(http.StatusOK, "成功", list)
}

func StartOutpatientVisit(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.StartVisitRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	resp, err := bc.ChannelExecute("startOutpatientVisit", [][]byte{[]byte(body.RegistrationID), []byte(body.DoctorID)})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	data["tx_id"] = resp.TransactionID
	appG.Response(http.StatusOK, "成功", data)
}

func FinishOutpatientVisit(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.FinishVisitRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	resp, err := bc.ChannelExecute("finishOutpatientVisit", [][]byte{[]byte(body.RegistrationID), []byte(body.DoctorID)})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	data["tx_id"] = resp.TransactionID
	appG.Response(http.StatusOK, "成功", data)
}

func QueryOutpatientRecord(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.QueryOutpatientRecordRequestBody)
	_ = c.ShouldBind(body)
	resp, err := bc.ChannelQuery("queryOutpatientRecord", [][]byte{})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var list []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &list); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	filtered := make([]map[string]interface{}, 0)
	for _, item := range list {
		reg, ok := item["registration"].(map[string]interface{})
		if !ok {
			continue
		}
		visitDate := fmt.Sprint(reg["visit_date"])
		if body.PatientID != "" && fmt.Sprint(reg["patient_id"]) != body.PatientID {
			continue
		}
		if body.DoctorID != "" && fmt.Sprint(reg["doctor_id"]) != body.DoctorID {
			continue
		}
		if body.StartDate != "" && visitDate < body.StartDate {
			continue
		}
		if body.EndDate != "" && visitDate > body.EndDate {
			continue
		}
		filtered = append(filtered, item)
	}
	appG.Response(http.StatusOK, "成功", filtered)
}
