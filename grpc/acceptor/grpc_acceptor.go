package acceptor

import (
	pb "agent/grpc/service"
	program_service "agent/service"
	"log"
)

type RequestBiStreamAcceptor interface {
	Handler(rq *pb.Payload)
	GetType() string
}

func init() {
	registerRequestBiStreamAcceptor(&program_service.PluginService{})
}

var requestBiStreamAcceptors []RequestBiStreamAcceptor

func registerRequestBiStreamAcceptor(handler RequestBiStreamAcceptor) {
	requestBiStreamAcceptors = append(requestBiStreamAcceptors, handler)
}

func RequestAcceptor(rq *pb.Payload) {
	for _, acceptor := range requestBiStreamAcceptors {
		if acceptor.GetType() == rq.Metadata.Type {
			acceptor.Handler(rq)
		}
		log.Println("The request does not know the specific acceptor.")
	}
}
