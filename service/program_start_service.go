package program_service

import (
	pb "agent/grpc/service"
	"agent/logger"
	"encoding/json"
	"go.uber.org/zap"
)

type StartProgramService struct{}

type StartProgramBody struct {
	name string `json:"name"`
}

func (service *StartProgramService) Handler(rq *pb.Payload) {
	body := rq.Body
	// 结构体转为 JSON 字符串
	jsonData, err := json.Marshal(body)
	if err != nil {
		logger.Logger.Error("JSON encoding error.", zap.Error(err))
		return
	}

	// JSON 字符串转为其他结构体
	var startProgramBody StartProgramBody
	err = json.Unmarshal(jsonData, &startProgramBody)
	if err != nil {
		logger.Logger.Error("JSON decoding error.", zap.Error(err))
		return
	}

	name := startProgramBody.name
	if len(name) == 0 {
		logger.Logger.Error("Program name length cannot be 0", zap.Error(err))
	}

	p := ProgramDictionary[name]
	p.Start()
}

func (service *StartProgramService) GetType() string {
	return "startProgram"
}
