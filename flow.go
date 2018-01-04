package flow

import (
	"gitee.com/antlinker/flow/schema"
	"gitee.com/antlinker/flow/service/db"
)

var (
	engine *Engine
)

// Init 初始化流程配置
func Init(dbConfig *db.Config) {
	db, err := db.NewDB(dbConfig)
	if err != nil {
		panic(err)
	}
	engine = new(Engine).Init(db, NewXMLParser(), NewQLangExecer())
}

// SetParser 设定解析器
func SetParser(parser Parser) {
	engine.SetParser(parser)
}

// SetExecer 设定表达式执行器
func SetExecer(execer Execer) {
	engine.SetExecer(execer)
}

// LoadFile 加载流程文件数据
func LoadFile(name string) error {
	return engine.LoadFile(name)
}

// StartFlow 启动流程
// flowCode 流程编号
// nodeCode 开始节点编号
// userID 发起人
// inputData 输入数据
func StartFlow(flowCode, nodeCode, userID string, inputData []byte) (*HandleResult, error) {
	return engine.StartFlow(flowCode, nodeCode, userID, inputData)
}

// HandleFlow 处理流程节点
// nodeInstanceID 节点实例内码
// userID 处理人
// inputData 输入数据
func HandleFlow(nodeInstanceID, userID string, inputData []byte) (*HandleResult, error) {
	return engine.HandleFlow(nodeInstanceID, userID, inputData)
}

// QueryTodoFlows 查询流程待办数据
// flowCode 流程编号
// userID 待办人
func QueryTodoFlows(flowCode, userID string) ([]*schema.NodeInstances, error) {
	return engine.QueryTodoFlows(flowCode, userID)
}