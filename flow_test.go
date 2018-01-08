package flow_test

import (
	"encoding/json"
	"testing"
	"time"

	"gitee.com/antlinker/flow"
	"gitee.com/antlinker/flow/service/db"
)

func init() {

	flow.Init(&db.Config{
		DSN:          "root:123456@tcp(192.168.1.36:3306)/flows?charset=utf8",
		Trace:        true,
		MaxIdleConns: 100,
		MaxOpenConns: 100,
		MaxLifetime:  time.Hour * 2,
	})

	err := flow.LoadFile("test_data/leave.bpmn")
	if err != nil {
		panic(err)
	}

	err = flow.LoadFile("test_data/apply_sqltest.bpmn")
	if err != nil {
		panic(err)
	}

	err = flow.LoadFile("test_data/parallel_test.bpmn")
	if err != nil {
		panic(err)
	}
}

func TestLeaveBzrApprovalPass(t *testing.T) {
	var (
		flowCode = "process_leave_test"
		bzr      = "T002"
	)

	input := map[string]interface{}{
		"day": 1,
		"bzr": bzr,
	}

	// 开始流程
	result, err := flow.StartFlow(flowCode, "node_start", "T001", input)
	if err != nil {
		t.Fatal(err.Error())
	}

	if result.NextNodes[0].CandidateIDs[0] != bzr {
		t.Fatalf("无效的下一级流转：%s", result.String())
	}

	// 查询待办
	todos, err := flow.QueryTodoFlows(flowCode, bzr)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(todos) != 1 {
		bts, _ := json.Marshal(todos)
		t.Fatalf("无效的待办数据:%s", string(bts))
	}

	// 处理流程（通过）
	input["action"] = "pass"
	result, err = flow.HandleFlow(todos[0].RecordID, bzr, input)
	if err != nil {
		t.Fatal(err.Error())
	}

	// 流程结束
	if !result.IsEnd {
		t.Fatalf("无效的处理结果：%s", result.String())
	}
}

func TestLeaveBzrApprovalBack(t *testing.T) {
	var (
		flowCode = "process_leave_test"
		launcher = "T001"
		bzr      = "T002"
	)

	input := map[string]interface{}{
		"day": 1,
		"bzr": bzr,
	}

	// 开始流程
	result, err := flow.StartFlow(flowCode, "node_start", launcher, input)
	if err != nil {
		t.Fatal(err.Error())
	}

	if result.NextNodes[0].CandidateIDs[0] != bzr {
		t.Fatalf("无效的下一级流转：%s", result.String())
	}

	// 查询待办
	todos, err := flow.QueryTodoFlows(flowCode, bzr)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// 处理流程（退回）
	input["action"] = "back"
	result, err = flow.HandleFlow(todos[0].RecordID, bzr, input)
	if err != nil {
		t.Fatal(err.Error())
	}

	if result.IsEnd ||
		result.NextNodes[0].CandidateIDs[0] != launcher {
		t.Fatalf("无效的处理结果：%s", result.String())
	}

	// 查询退回流程
	todos, err = flow.QueryTodoFlows(flowCode, launcher)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// 处理退回流程
	delete(input, "action")
	result, err = flow.HandleFlow(todos[0].RecordID, launcher, input)
	if err != nil {
		t.Fatal(err.Error())
	}

	if result.NextNodes[0].CandidateIDs[0] != bzr {
		t.Fatalf("无效的下一级流转：%s", result.String())
	}

	// 查询待办流程
	todos, err = flow.QueryTodoFlows(flowCode, bzr)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// 处理流程（通过）
	input["action"] = "pass"
	result, err = flow.HandleFlow(todos[0].RecordID, bzr, input)
	if err != nil {
		t.Fatal(err.Error())
	}

	// 流程结束
	if !result.IsEnd {
		t.Fatalf("无效的处理结果：%s", result.String())
	}
}

func TestLeaveFdyApprovalPass(t *testing.T) {
	var (
		flowCode = "process_leave_test"
		launcher = "T001"
		bzr      = "T002"
		fdy      = "T003"
	)

	input := map[string]interface{}{
		"day": 3,
		"bzr": bzr,
		"fdy": fdy,
	}

	// 开始流程
	result, err := flow.StartFlow(flowCode, "node_start", launcher, input)
	if err != nil {
		t.Fatal(err.Error())
	}

	if result.NextNodes[0].CandidateIDs[0] != bzr {
		t.Fatalf("无效的下一级流转：%s", result.String())
	}

	// 查询待办
	todos, err := flow.QueryTodoFlows(flowCode, bzr)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// 处理流程（通过）
	input["action"] = "pass"
	result, err = flow.HandleFlow(todos[0].RecordID, bzr, input)
	if err != nil {
		t.Fatal(err.Error())
	}

	// 查询待办
	todos, err = flow.QueryTodoFlows(flowCode, fdy)
	if err != nil {
		t.Fatalf(err.Error())
	}

	// 处理流程（通过）
	input["action"] = "pass"
	result, err = flow.HandleFlow(todos[0].RecordID, fdy, input)
	if err != nil {
		t.Fatal(err.Error())
	}

	// 流程结束
	if !result.IsEnd {
		t.Fatalf("无效的处理结果：%s", result.String())
	}
}

func TestApplySQLPass(t *testing.T) {
	var (
		flowCode = "process_apply_sqltest"
	)

	input := map[string]interface{}{
		"form": "apply",
	}

	// 开始流程
	result, err := flow.StartFlow(flowCode, "node_start", "A001", input)
	if err != nil {
		t.Fatal(err.Error())
	}

	cIDs := result.NextNodes[0].CandidateIDs
	if len(cIDs) != 2 {
		t.Fatalf("无效的下一级流转：%s", result.String())
	}

	var (
		nodeInstanceID string
		userID         string
	)
	for _, cid := range cIDs {
		// 查询待办
		todos, err := flow.QueryTodoFlows(flowCode, cid)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if len(todos) != 1 {
			bts, _ := json.Marshal(todos)
			t.Fatalf("无效的待办数据:%s", string(bts))
		}

		nodeInstanceID = todos[0].RecordID
		userID = cid
	}

	// 处理流程（通过）
	input["action"] = "pass"
	result, err = flow.HandleFlow(nodeInstanceID, userID, input)
	if err != nil {
		t.Fatal(err.Error())
	}

	// 流程结束
	if !result.IsEnd {
		t.Fatalf("无效的处理结果：%s", result.String())
	}
}

func TestParallel(t *testing.T) {
	var (
		flowCode = "process_parallel_test"
	)

	input := map[string]interface{}{
		"form": "countersign",
	}

	// 开始流程
	result, err := flow.StartFlow(flowCode, "node_start", "H001", input)
	if err != nil {
		t.Fatal(err.Error())
	}

	if len(result.NextNodes) != 3 {
		t.Fatalf("无效的下一级流转：%s", result.String())
	}

	for i, node := range result.NextNodes {
		if len(node.CandidateIDs) != 1 {
			t.Fatalf("无效的节点处理人：%v", node.CandidateIDs)
		}

		todos, err := flow.QueryTodoFlows(flowCode, node.CandidateIDs[0])
		if err != nil {
			t.Fatalf(err.Error())
		} else if len(todos) != 1 {
			bts, _ := json.Marshal(todos)
			t.Fatalf("无效的待办数据:%s", string(bts))
		}

		input["sign"] = node.CandidateIDs[0]
		result, err := flow.HandleFlow(todos[0].RecordID, node.CandidateIDs[0], input)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if i == 2 {
			if !result.IsEnd {
				t.Fatalf("无效的处理结果：%s", result.String())
			}
			break
		}

		if result.IsEnd {
			t.Fatalf("无效的处理结果：%s", result.String())
		}
	}

}
