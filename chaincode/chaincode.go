package main

import (
	"chaincode/api"
	"chaincode/model"
	"chaincode/pkg/utils"
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

type BlockChainMedicalInfoManageSystem struct {
}

// Init 链码部署到链上并进行初始化时会执行该方法
func (t *BlockChainMedicalInfoManageSystem) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("链码初始化")

	var accountV2Ids = [7]string{"0feceb66ffc1", "1feceb66ffc1", "2b86b273ff31", "34735e3a261e", "4e17408561be", "5b227771d4dd", "6f2d121de37b"}
	var userNameV2s = [7]string{"管理员", "医生", "①号病人", "②号病人", "③号病人", "药店", "保险机构"}
	for i, val := range accountV2Ids {
		role := "patient"
		switch userNameV2s[i] {
		case "管理员":
			role = "admin"
		case "医生":
			role = "doctor"
		case "药店":
			role = "drugstore"
		case "保险机构":
			role = "insurance"
		}
		account := &model.AccountV2{AccountId: val, AccountName: userNameV2s[i], Role: role}
		if err := utils.WriteLedger(account, stub, model.AccountV2Key, []string{val}); err != nil {
			return shim.Error(fmt.Sprintf("%s", err))
		}
	}

	return shim.Success(nil)
}

// Invoke 实现Invoke接口调用智能合约
func (t *BlockChainMedicalInfoManageSystem) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "hello":
		return api.Hello(stub, args)
	case "createAccountV2":
		return api.CreateAccountV2(stub, args)
	case "queryAccountV2List":
		return api.QueryAccountV2List(stub, args)
	case "createPrescription":
		return api.CreatePrescription(stub, args)
	case "queryPrescription":
		return api.QueryPrescription(stub, args)
	case "createInsuranceCover":
		return api.CreateInsuranceCover(stub, args)
	case "queryInsuranceCover":
		return api.QueryInsuranceCover(stub, args)
	case "updateInsuranceCover":
		return api.UpdateInsuranceCover(stub, args)
	case "deleteInsuranceCover":
		return api.DeleteInsuranceCover(stub, args)
	case "createDrugOrder":
		return api.CreateDrugOrder(stub, args)
	case "queryDrugOrder":
		return api.QueryDrugOrder(stub, args)
	case "grantRecordAuthorization":
		return api.GrantRecordAuthorization(stub, args)
	case "renewRecordAuthorization":
		return api.RenewRecordAuthorization(stub, args)
	case "queryRecordAuthorizationsByPatient":
		return api.QueryRecordAuthorizationsByPatient(stub, args)
	case "revokeRecordAuthorization":
		return api.RevokeRecordAuthorization(stub, args)
	case "queryAccessibleRecordsByDoctor":
		return api.QueryAccessibleRecordsByDoctor(stub, args)
	case "checkRecordAccess":
		return api.CheckRecordAccess(stub, args)
	default:
		return shim.Error(fmt.Sprintf("没有该功能: %s", funcName))
	}
}

func main() {
	timeLocal, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	time.Local = timeLocal
	err = shim.Start(new(BlockChainMedicalInfoManageSystem))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
