package api

import (
	"chaincode/model"
	"chaincode/pkg/utils"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func nowStr() string { return time.Now().Format("2006-01-02 15:04:05") }

func CreateScheduleSlot(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 6 {
		return shim.Error("参数个数不满足")
	}
	doctorID, departmentID, visitDate, startTime, endTime, capacityStr := args[0], args[1], args[2], args[3], args[4], args[5]
	capacity, err := strconv.Atoi(capacityStr)
	if err != nil || capacity <= 0 {
		return shim.Error("capacity不合法")
	}
	if doctorID == "" || departmentID == "" || visitDate == "" || startTime == "" || endTime == "" {
		return shim.Error("参数存在空值")
	}

	slot := &model.OutpatientScheduleSlot{
		ID: stub.GetTxID()[:16], DoctorID: doctorID, DepartmentID: departmentID, VisitDate: visitDate,
		StartTime: startTime, EndTime: endTime, Capacity: capacity, BookedCount: 0, Status: "OPEN",
		CreatedTime: nowStr(), UpdatedTime: nowStr(), TxID: stub.GetTxID(),
	}
	if err := utils.WriteLedger(slot, stub, model.OutpatientSlotKey, []string{slot.ID}); err != nil {
		return shim.Error(err.Error())
	}
	b, _ := json.Marshal(slot)
	return shim.Success(b)
}

func QueryScheduleSlot(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var list []model.OutpatientScheduleSlot
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.OutpatientSlotKey, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	for _, v := range results {
		var item model.OutpatientScheduleSlot
		if err := json.Unmarshal(v, &item); err == nil {
			list = append(list, item)
		}
	}
	b, _ := json.Marshal(list)
	return shim.Success(b)
}

func CreateOutpatientRegistration(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 5 {
		return shim.Error("参数个数不满足")
	}
	patientID, doctorID, departmentID, slotID, visitDate := args[0], args[1], args[2], args[3], args[4]
	if patientID == "" || doctorID == "" || departmentID == "" || slotID == "" || visitDate == "" {
		return shim.Error("参数存在空值")
	}

	key, _ := stub.CreateCompositeKey(model.OutpatientSlotKey, []string{slotID})
	slotBytes, err := stub.GetState(key)
	if err != nil || slotBytes == nil {
		return shim.Error("号源不存在")
	}
	var slot model.OutpatientScheduleSlot
	if err = json.Unmarshal(slotBytes, &slot); err != nil {
		return shim.Error("号源反序列化失败")
	}
	if slot.Status != "OPEN" {
		return shim.Error("号源已关闭")
	}
	if slot.BookedCount >= slot.Capacity {
		return shim.Error("号源已满")
	}

	slot.BookedCount += 1
	slot.UpdatedTime = nowStr()
	if err = utils.WriteLedger(&slot, stub, model.OutpatientSlotKey, []string{slot.ID}); err != nil {
		return shim.Error(err.Error())
	}

	queueNo := fmt.Sprintf("%03d", slot.BookedCount)
	reg := &model.OutpatientRegistration{
		ID: stub.GetTxID()[:16], PatientID: patientID, DoctorID: doctorID, DepartmentID: departmentID, ScheduleSlotID: slotID,
		VisitDate: visitDate, Status: "BOOKED", FeeAmount: "20", FeeStatus: "UNPAID", QueueNo: queueNo,
		CreatedTime: nowStr(), UpdatedTime: nowStr(), TxID: stub.GetTxID(),
	}
	if err = utils.WriteLedger(reg, stub, model.OutpatientRegistrationKey, []string{reg.ID}); err != nil {
		return shim.Error(err.Error())
	}
	_ = utils.WriteLedger(reg, stub, model.OutpatientRegistrationPatientIdxKey, []string{reg.PatientID, reg.ID})
	_ = utils.WriteLedger(reg, stub, model.OutpatientRegistrationDoctorIdxKey, []string{reg.DoctorID, reg.ID})

	pay := &model.OutpatientPayment{ID: "pay-" + reg.ID, OrderType: "REG_FEE", RegistrationID: reg.ID, PatientID: patientID, Amount: reg.FeeAmount, Status: "UNPAID", CreatedTime: nowStr(), TxID: stub.GetTxID()}
	if err = utils.WriteLedger(pay, stub, model.OutpatientPaymentKey, []string{pay.ID}); err != nil {
		return shim.Error(err.Error())
	}

	queue := &model.OutpatientQueueItem{ID: "queue-" + reg.ID, RegistrationID: reg.ID, DoctorID: doctorID, PatientID: patientID, QueueNo: queueNo, Status: "WAITING", CreatedTime: nowStr(), TxID: stub.GetTxID()}
	if err = utils.WriteLedger(queue, stub, model.OutpatientQueueDoctorIdxKey, []string{doctorID, queue.ID}); err != nil {
		return shim.Error(err.Error())
	}

	b, _ := json.Marshal(reg)
	return shim.Success(b)
}

func CancelOutpatientRegistration(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("参数个数不满足")
	}
	regID := args[0]
	key, _ := stub.CreateCompositeKey(model.OutpatientRegistrationKey, []string{regID})
	b, err := stub.GetState(key)
	if err != nil || b == nil {
		return shim.Error("挂号记录不存在")
	}
	var reg model.OutpatientRegistration
	if err = json.Unmarshal(b, &reg); err != nil {
		return shim.Error("反序列化失败")
	}
	if reg.Status != "BOOKED" {
		return shim.Error("仅BOOKED状态可取消")
	}
	reg.Status = "CANCELLED"
	reg.UpdatedTime = nowStr()
	if err = utils.WriteLedger(&reg, stub, model.OutpatientRegistrationKey, []string{reg.ID}); err != nil {
		return shim.Error(err.Error())
	}
	out, _ := json.Marshal(reg)
	return shim.Success(out)
}

func QueryOutpatientRegistration(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.OutpatientRegistrationKey, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	var list []model.OutpatientRegistration
	for _, v := range results {
		var it model.OutpatientRegistration
		if err := json.Unmarshal(v, &it); err == nil {
			list = append(list, it)
		}
	}
	b, _ := json.Marshal(list)
	return shim.Success(b)
}

func QueryOutpatientPayment(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.OutpatientPaymentKey, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	var list []model.OutpatientPayment
	for _, v := range results {
		var it model.OutpatientPayment
		if err := json.Unmarshal(v, &it); err == nil {
			list = append(list, it)
		}
	}
	b, _ := json.Marshal(list)
	return shim.Success(b)
}

func PayOutpatientOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("参数个数不满足")
	}
	paymentID := args[0]
	key, _ := stub.CreateCompositeKey(model.OutpatientPaymentKey, []string{paymentID})
	b, err := stub.GetState(key)
	if err != nil || b == nil {
		return shim.Error("支付单不存在")
	}
	var p model.OutpatientPayment
	if err = json.Unmarshal(b, &p); err != nil {
		return shim.Error("反序列化失败")
	}
	if p.Status == "PAID" {
		return shim.Error("订单已支付")
	}
	p.Status = "PAID"
	p.PaidTime = nowStr()
	p.TxID = stub.GetTxID()
	if err = utils.WriteLedger(&p, stub, model.OutpatientPaymentKey, []string{p.ID}); err != nil {
		return shim.Error(err.Error())
	}

	regKey, _ := stub.CreateCompositeKey(model.OutpatientRegistrationKey, []string{p.RegistrationID})
	regBytes, _ := stub.GetState(regKey)
	if regBytes != nil {
		var reg model.OutpatientRegistration
		if json.Unmarshal(regBytes, &reg) == nil {
			reg.FeeStatus = "PAID"
			reg.UpdatedTime = nowStr()
			_ = utils.WriteLedger(&reg, stub, model.OutpatientRegistrationKey, []string{reg.ID})
		}
	}
	out, _ := json.Marshal(p)
	return shim.Success(out)
}

func QueryOutpatientQueue(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 1 {
		return shim.Error("doctor_id不能为空")
	}
	doctorID := args[0]
	results, err := utils.GetStateByPartialCompositeKeys2(stub, model.OutpatientQueueDoctorIdxKey, []string{doctorID})
	if err != nil {
		return shim.Error(err.Error())
	}
	var list []model.OutpatientQueueItem
	for _, v := range results {
		var it model.OutpatientQueueItem
		if err := json.Unmarshal(v, &it); err == nil {
			list = append(list, it)
		}
	}
	b, _ := json.Marshal(list)
	return shim.Success(b)
}

func StartOutpatientVisit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("参数个数不满足")
	}
	registrationID, doctorID := args[0], args[1]
	qid := "queue-" + registrationID
	qkey, _ := stub.CreateCompositeKey(model.OutpatientQueueDoctorIdxKey, []string{doctorID, qid})
	qb, err := stub.GetState(qkey)
	if err != nil || qb == nil {
		return shim.Error("队列项不存在")
	}
	var q model.OutpatientQueueItem
	if err = json.Unmarshal(qb, &q); err != nil {
		return shim.Error("反序列化失败")
	}
	if q.Status != "WAITING" {
		return shim.Error("仅WAITING可开始就诊")
	}
	q.Status = "IN_PROGRESS"
	q.CalledTime = nowStr()
	q.TxID = stub.GetTxID()
	if err = utils.WriteLedger(&q, stub, model.OutpatientQueueDoctorIdxKey, []string{doctorID, qid}); err != nil {
		return shim.Error(err.Error())
	}
	b, _ := json.Marshal(q)
	return shim.Success(b)
}

func FinishOutpatientVisit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("参数个数不满足")
	}
	registrationID, doctorID := args[0], args[1]
	qid := "queue-" + registrationID
	qkey, _ := stub.CreateCompositeKey(model.OutpatientQueueDoctorIdxKey, []string{doctorID, qid})
	qb, err := stub.GetState(qkey)
	if err != nil || qb == nil {
		return shim.Error("队列项不存在")
	}
	var q model.OutpatientQueueItem
	if err = json.Unmarshal(qb, &q); err != nil {
		return shim.Error("反序列化失败")
	}
	if q.Status != "IN_PROGRESS" {
		return shim.Error("仅IN_PROGRESS可完成")
	}
	q.Status = "DONE"
	q.FinishedTime = nowStr()
	q.TxID = stub.GetTxID()
	if err = utils.WriteLedger(&q, stub, model.OutpatientQueueDoctorIdxKey, []string{doctorID, qid}); err != nil {
		return shim.Error(err.Error())
	}

	regKey, _ := stub.CreateCompositeKey(model.OutpatientRegistrationKey, []string{registrationID})
	regBytes, _ := stub.GetState(regKey)
	if regBytes != nil {
		var reg model.OutpatientRegistration
		if json.Unmarshal(regBytes, &reg) == nil {
			reg.Status = "VISITED"
			reg.UpdatedTime = nowStr()
			_ = utils.WriteLedger(&reg, stub, model.OutpatientRegistrationKey, []string{reg.ID})
		}
	}
	b, _ := json.Marshal(q)
	return shim.Success(b)
}

func QueryOutpatientRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	regs, err := utils.GetStateByPartialCompositeKeys2(stub, model.OutpatientRegistrationKey, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	pays, err := utils.GetStateByPartialCompositeKeys2(stub, model.OutpatientPaymentKey, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}
	queueAll, err := utils.GetStateByPartialCompositeKeys2(stub, model.OutpatientQueueDoctorIdxKey, []string{})
	if err != nil {
		return shim.Error(err.Error())
	}

	payMap := map[string]model.OutpatientPayment{}
	for _, b := range pays {
		var p model.OutpatientPayment
		if json.Unmarshal(b, &p) == nil {
			payMap[p.RegistrationID] = p
		}
	}
	queueMap := map[string]model.OutpatientQueueItem{}
	for _, b := range queueAll {
		var q model.OutpatientQueueItem
		if json.Unmarshal(b, &q) == nil {
			queueMap[q.RegistrationID] = q
		}
	}
	var out []map[string]interface{}
	for _, b := range regs {
		var r model.OutpatientRegistration
		if json.Unmarshal(b, &r) != nil {
			continue
		}
		out = append(out, map[string]interface{}{
			"registration": r,
			"payment":      payMap[r.ID],
			"queue":        queueMap[r.ID],
		})
	}
	resp, _ := json.Marshal(out)
	return shim.Success(resp)
}
