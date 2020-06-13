package main

import (
	"flag"
	"fmt"
	"github.com/arcosx/WesternQueen/master"
	"github.com/arcosx/WesternQueen/slave"
	"github.com/arcosx/WesternQueen/util"
	"github.com/gin-gonic/gin"
)

// 暴漏HTTP端口
var finalRunPort string

func main() {
	// 读取命令行参数判断是 slave 还是 master
	flag.StringVar(&util.Mode, "mode", "SLAVE", "Run Mode")
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
		if util.IsSlave() {
			go slave.Start()
		}
		if util.IsMaster() {
			go master.Start()
		}
		c.JSON(200, gin.H{
			"message": "setParameter success",
		})
	})
	// only master
	if util.IsMaster() {
		r.GET("/finish", func(c *gin.Context) {

		})
	}

	// 根据模式选择端口
	switch util.Mode {
	case util.MASTER_MODE:
		finalRunPort = util.MASTER_PORT_8002
	case util.SLAVE_ONE_MODE:
		finalRunPort = util.SLAVE_ONE_PORT_8001
	case util.SLAVE_TWO_MODE:
		finalRunPort = util.SLAVE_TWO_PORT_8002

	}
	_ = r.Run(finalRunPort) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
