package slave

import (
	rpc "github.com/arcosx/WesternQueen/rpc"
	"github.com/arcosx/WesternQueen/util"
	"google.golang.org/grpc"
	"log"
	"testing"
)

func init() {
	util.DebugMode = true
	util.Mode = "slave1"

}
func TestStart(t *testing.T) {
	conn, err := grpc.Dial("localhost:8003", grpc.WithInsecure(),grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	RPCClient = rpc.NewWesternQueenClient(conn)
	Start()
}
