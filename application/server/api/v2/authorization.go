package v2

import (
	bc "application/blockchain"
	"application/model"
	"application/pkg/app"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GrantRecordAuthorization(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.GrantRecordAuthorizationRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if body.PatientID == "" || body.RecordID == "" || body.DoctorID == "" || body.EndTime == "" {
		appG.Response(http.StatusBadRequest, "失败", "参数存在空值")
		return
	}
	bodyBytes := [][]byte{[]byte(body.PatientID), []byte(body.RecordID), []byte(body.DoctorID), []byte(body.HospitalName), []byte(body.Department), []byte(body.EndTime), []byte(body.Remark)}
	resp, err := bc.ChannelExecute("grantRecordAuthorization", bodyBytes)
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	appG.Response(http.StatusOK, "成功", data)
}

func RenewRecordAuthorization(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.RenewRecordAuthorizationRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if body.PatientID == "" || body.AuthID == "" || body.EndTime == "" {
		appG.Response(http.StatusBadRequest, "失败", "参数存在空值")
		return
	}
	bodyBytes := [][]byte{[]byte(body.PatientID), []byte(body.AuthID), []byte(body.EndTime)}
	resp, err := bc.ChannelExecute("renewRecordAuthorization", bodyBytes)
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	appG.Response(http.StatusOK, "成功", data)
}

func QueryMyAuthorizations(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.QueryMyAuthorizationsRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if body.PatientID == "" {
		appG.Response(http.StatusBadRequest, "失败", "参数存在空值")
		return
	}
	resp, err := bc.ChannelQuery("queryRecordAuthorizationsByPatient", [][]byte{[]byte(body.PatientID)})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	appG.Response(http.StatusOK, "成功", data)
}

func RevokeRecordAuthorization(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.RevokeRecordAuthorizationRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if body.PatientID == "" || body.AuthID == "" {
		appG.Response(http.StatusBadRequest, "失败", "参数存在空值")
		return
	}
	resp, err := bc.ChannelExecute("revokeRecordAuthorization", [][]byte{[]byte(body.PatientID), []byte(body.AuthID)})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	appG.Response(http.StatusOK, "成功", data)
}

func QueryAccessibleRecordsByDoctor(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.QueryAccessibleRecordsByDoctorRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if body.DoctorID == "" {
		appG.Response(http.StatusBadRequest, "失败", "参数存在空值")
		return
	}
	bodyBytes := [][]byte{[]byte(body.DoctorID), []byte(body.PatientNameKeyword), []byte(body.IdCardKeyword), []byte(body.RecordTypeKeyword), []byte(body.CreatedStart), []byte(body.CreatedEnd)}
	resp, err := bc.ChannelQuery("queryAccessibleRecordsByDoctor", bodyBytes)
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	appG.Response(http.StatusOK, "成功", data)
}

func CheckRecordAccess(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.CheckRecordAccessRequestBody)
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if body.DoctorID == "" || body.RecordID == "" {
		appG.Response(http.StatusBadRequest, "失败", "参数存在空值")
		return
	}
	resp, err := bc.ChannelQuery("checkRecordAccess", [][]byte{[]byte(body.DoctorID), []byte(body.RecordID)})
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	var data map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	if allowed, ok := data["allowed"].(bool); ok && !allowed {
		appG.Response(http.StatusForbidden, "失败", "无病历访问授权")
		return
	}
	appG.Response(http.StatusOK, "成功", data)
}
