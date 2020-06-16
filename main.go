package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/arcosx/WesternQueen/newmaster"
	"github.com/arcosx/WesternQueen/newslave"
	rpc "github.com/arcosx/WesternQueen/rpc"
	"github.com/arcosx/WesternQueen/slave"
	"github.com/arcosx/WesternQueen/util"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log"
	"os"
	"time"
)

// 暴漏HTTP端口
var finalRunPort string

func main() {
	// 读取环境变量判断是 slave 还是 master
	envPort := os.Getenv("SERVER_PORT")
	if envPort == "8000" {
		util.Mode = util.SLAVE_ONE_MODE
	} else if envPort == "8001" {
		util.Mode = util.SLAVE_TWO_MODE
	} else {
		util.Mode = util.MASTER_MODE
	}
	flag.BoolVar(&util.DebugMode, "debug", false, "Debug Mode")
	flag.Parse()
	fmt.Println("Run Mode is : ", util.Mode)
	fmt.Println("Debug Mode : ", util.DebugMode)

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ready",
		})
	})
	r.GET("/start", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "start",
		})
	})
	r.GET("/setParameter", func(c *gin.Context) {
		util.DATA_SOURCE_PORT = c.Query("port")
		util.RESULT_UPLOAD_PORT = util.DATA_SOURCE_PORT
		fmt.Println(fmt.Sprintf("%s receive parameters start running!", util.Mode))
		c.JSON(200, gin.H{
			"message": "setParameter success",
		})
		// 从节点开始！
		if util.IsSlave() {
			go newslave.Start()
		}
	})

	if util.IsMaster() {
		// master 获取 slave 传递的错误ID
		r.POST("/getWrong", func(c *gin.Context) {
			bytes, _ := c.GetRawData()
			var WrongTraceList []string
			err := json.Unmarshal(bytes, &WrongTraceList)
			if err != nil {
				fmt.Println("json parse error", err)
				c.JSON(500, gin.H{
					"message": "data error",
				})
			}
			go newmaster.GetWrongTraceList(WrongTraceList, c.Query("node"))
			c.JSON(200, gin.H{
				"message": "fuck you slave",
			})
		})
		// master 获取 slave 传递的全量数据
		r.POST("/all", func(c *gin.Context) {
			bytes, _ := c.GetRawData()
			var traceList []string
			err := json.Unmarshal(bytes, &traceList)
			if err != nil {
				fmt.Println("json parse error", err)
				c.JSON(500, gin.H{
					"message": "data error",
				})
			}
			go newmaster.GetAllTraceList(traceList, c.Query("node"))
			c.JSON(200, gin.H{
				"message": "fuck you slave",
			})
		})

	}
	if util.IsSlave() {
		// slave 获取 master 发送的共享错误traceID
		r.POST("/getShare", func(c *gin.Context) {
			bytes, _ := c.GetRawData()
			var WrongTraceList []string
			err := json.Unmarshal(bytes, &WrongTraceList)
			if err != nil {
				fmt.Println("json parse error", err)
				c.JSON(500, gin.H{
					"message": "data error",
				})
			}
			go newslave.GetShareWrongTraceSet(WrongTraceList)
			c.JSON(200, gin.H{
				"message": "fuck you master",
			})
		})
	}
	// 根据模式选择端口
	switch util.Mode {
	case util.MASTER_MODE:
		finalRunPort = util.MASTER_PORT_8002
	case util.SLAVE_ONE_MODE:
		finalRunPort = util.SLAVE_ONE_PORT_8000
	case util.SLAVE_TWO_MODE:
		finalRunPort = util.SLAVE_TWO_PORT_8001

	}
	_ = r.Run(finalRunPort) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

// 初始化RPC服务
func RPCService() {
	fmt.Println("init RPC Service")
	var kaep = keepalive.EnforcementPolicy{
		MinTime:             1 * time.Minute, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}
	var kasp = keepalive.ServerParameters{
		//MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
		//MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 5 * time.Minute, // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  5 * time.Second, // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout:               1 * time.Second, // Wait 1 second for the ping ack before assuming the connection is dead
	}
	rpc.NewWesternQueenService(util.MASTER_PORT_8003, grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))
	fmt.Println("init RPC Finish")
}

func RPCClient() {

	var kacp = keepalive.ClientParameters{
		Time:                2 * time.Minute, // send pings every 10 seconds if there is no activity
		PermitWithoutStream: true,            // send pings even without active streams
	}
	fmt.Println("init RPCClient!")
	conn, err := grpc.Dial(fmt.Sprintf("localhost%s", util.MASTER_PORT_8003), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	slave.RPCClient = rpc.NewWesternQueenClient(conn)
	fmt.Println("init RPCClient finish!")
}
