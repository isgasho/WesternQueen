package master

import (
	"fmt"
	mapset "github.com/deckarep/golang-set"
	"time"
)

// 全局变量
// 全局错误traceID
var WrongTraceSet mapset.Set

// 开启运行

// 初始化 RPC 服务
func init() {
	WrongTraceSet = mapset.NewSet()
}

func ReceiveWrongTraceData(wrongTraceId string) {
	//fmt.Println("ReceiveWrongTraceData : ", wrongTraceId)
	WrongTraceSet.Add(wrongTraceId)
}

func GetWrongTraceSet() []string {
	var result []string
	for traceId := range WrongTraceSet.Iter() {
		result = append(result, traceId.(string))
	}
	return result
}

func Start() {
	fmt.Println("master start!")
	d := time.Second * 5

	t := time.Tick(d)
	for {
		<- t
		go fmt.Println(WrongTraceSet.Cardinality())
	}
}
