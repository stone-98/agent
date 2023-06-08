package grpc

import (
	pb "agent/grpc/service"
	"agent/logger"
	"encoding/json"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"net"
	"strconv"
)

const healthCheckService = "grpc.health.v1.Health"

const localhost = "localhost:"

const listenProtocol = "tcp"

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type server struct {
	pb.UnimplementedBiRequestStreamServer
}

func (s *server) RequestBiStream(stream pb.BiRequestStream_RequestBiStreamServer) error {
	Mutex.Lock()
	RegisterConnection(stream)
	Mutex.Unlock()
	for {
		rq, err := GetConnection().Recv()
		if err != nil {
			return err
		}
		printRq(rq)
		RequestAcceptor(rq)
	}
}

func printRq(rq *pb.Payload) {
	bytes, err := json.Marshal(rq)
	if err != nil {
		logger.Logger.Error("An error occurred while converting the request payload to a string.", zap.Error(err))
	}
	logger.Logger.Info("Received message.", zap.String("context", string(bytes)))
}

func StartGrpcServer(grpcServerConfig ServerConfig) {
	address := localhost + strconv.Itoa(grpcServerConfig.Port)
	listen, err := net.Listen(listenProtocol, address)
	if err != nil {
		logger.Logger.Panic("An exception occurred on the listening port.", zap.Error(err))
	}
	// 启动grpc服务
	grpcServer := grpc.NewServer()
	// 注册双向流服务
	pb.RegisterBiRequestStreamServer(grpcServer, &server{})
	// 注册健康检查服务
	healthCheckServer := health.NewServer()
	healthCheckServer.SetServingStatus(healthCheckService, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthCheckServer)
	// 支持服务发现和调试
	reflection.Register(grpcServer)
	if err := grpcServer.Serve(listen); err != nil {
		logger.Logger.Info("Crash when starting grpc service.", zap.Error(err))
	}
}
