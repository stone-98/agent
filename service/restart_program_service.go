package program_service

import (
	pb "agent/grpc/service"
	"agent/logger"
	"agent/plugin_manager"
	"encoding/json"
	"go.uber.org/zap"
)

type RestartProgramService struct{}

type RestartProgramBody struct {
	name string `json:"name"`
}

func (service *RestartProgramService) Handler(rq *pb.Payload) {
	body := rq.Body
	// 结构体转为 JSON 字符串
	jsonData, err := json.Marshal(body)
	if err != nil {
		logger.Logger.Error("JSON encoding error.", zap.Error(err))
		return
	}

	// JSON 字符串转为其他结构体
	var restartProgramBody RestartProgramBody
	err = json.Unmarshal(jsonData, &restartProgramBody)
	if err != nil {
		logger.Logger.Error("JSON decoding error.", zap.Error(err))
		return
	}

	name := restartProgramBody.name
	if len(name) == 0 {
		logger.Logger.Error("Program name length cannot be 0", zap.Error(err))
	}

	p := plugin_manager.ProgramDictionary[name]
	p.Restart()
}

func (service *RestartProgramService) GetType() string {
	return "startProgram"
}
