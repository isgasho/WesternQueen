package western_queen

import (
	"context"
	"fmt"
	"github.com/arcosx/WesternQueen/master"
	"google.golang.org/grpc"
	"net"
)

type WesternQueenService struct {
	UnimplementedWesternQueenServer
}

func NewWesternQueenService(addr string, opts ...grpc.ServerOption) {
	fmt.Println("Create NewWesternQueenService gRPC on", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Create Kafka Service gRPC failed:Failed net Listen", err)
		return
	}
	grpcServer := grpc.NewServer(opts...)
	RegisterWesternQueenServer(grpcServer, &WesternQueenService{})
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Println("Create NewWesternQueenService gRPC failed:Failed to serve", err)
	}
}

func (w *WesternQueenService) SendWrongTraceData(ctx context.Context, request *WrongTraceDataRequest) (*Empty, error) {
	master.ReceiveWrongTraceData(request.TraceId)
	return new(Empty), nil
}

func (w *WesternQueenService) ReadShareWrongTraceData(_ *Empty, server WesternQueen_ReadShareWrongTraceDataServer) error {

	for {
		var ShareWrongTraceDataReturn ShareWrongTraceDataReturn
		ShareWrongTraceDataReturn.WrongTraceDataRequests = master.GetWrongTraceSet()
		server.Send(&ShareWrongTraceDataReturn)
	}

}
