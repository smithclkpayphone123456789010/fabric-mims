package v2

import (
	bc "application/blockchain"
	"application/model"
	"application/pkg/app"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func QueryAccountV2List(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.AccountRequestBody)
	//解析Body参数
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	var bodyBytes [][]byte
	for _, val := range body.Args {
		bodyBytes = append(bodyBytes, []byte(val.AccountId))
	}
	//调用智能合约
	resp, err := bc.ChannelQuery("queryAccountV2List", bodyBytes)
	if err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	// 反序列化json
	var data []map[string]interface{}
	if err = json.Unmarshal(bytes.NewBuffer(resp.Payload).Bytes(), &data); err != nil {
		appG.Response(http.StatusInternalServerError, "失败", err.Error())
		return
	}
	appG.Response(http.StatusOK, "成功", data)
}

func CreateAccountV2(c *gin.Context) {
	appG := app.Gin{C: c}
	body := new(model.CreateAccountBody)

	//解析Body参数
	if err := c.ShouldBind(body); err != nil {
		appG.Response(http.StatusBadRequest, "失败", fmt.Sprintf("参数出错%s", err.Error()))
		return
	}
	if body.AccountName == "" || body.Operator == "" || body.Role == "" {
		appG.Response(http.StatusBadRequest, "失败", "参数存在空值")
		return
	}

	switch body.Role {
	case "doctor":
		if body.HospitalName == "" || body.Department == "" || body.Title == "" || body.Gender == "" || body.EmployeeNo == "" {
			appG.Response(http.StatusBadRequest, "失败", "医生信息不完整")
			return
		}
	case "patient":
		if body.IDCardNo == "" || body.InsuranceCardNo == "" || body.Gender == "" || body.Age == "" || body.BirthDate == "" || body.Phone == "" {
			appG.Response(http.StatusBadRequest, "失败", "患者信息不完整")
			return
		}
	case "drugstore":
		if body.HospitalName == "" {
			appG.Response(http.StatusBadRequest, "失败", "药店所属医院不能为空")
			return
		}
	case "insurance":
		// 保险机构仅校验基础字段
	default:
		appG.Response(http.StatusBadRequest, "失败", "角色类型不合法")
		return
	}

	var bodyBytes [][]byte
	bodyBytes = append(bodyBytes, []byte(body.AccountName))
	bodyBytes = append(bodyBytes, []byte(body.Role))
	bodyBytes = append(bodyBytes, []byte(body.Operator))
	bodyBytes = append(bodyBytes, []byte(body.HospitalID))
	bodyBytes = append(bodyBytes, []byte(body.HospitalName))
	bodyBytes = append(bodyBytes, []byte(body.Department))
	bodyBytes = append(bodyBytes, []byte(body.Title))
	bodyBytes = append(bodyBytes, []byte(body.Gender))
	bodyBytes = append(bodyBytes, []byte(body.EmployeeNo))
	bodyBytes = append(bodyBytes, []byte(body.IDCardNo))
	bodyBytes = append(bodyBytes, []byte(body.InsuranceCardNo))
	bodyBytes = append(bodyBytes, []byte(body.Age))
	bodyBytes = append(bodyBytes, []byte(body.BirthDate))
	bodyBytes = append(bodyBytes, []byte(body.Phone))

	// 调用智能合约
	resp, err := bc.ChannelExecute("createAccountV2", bodyBytes)
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
