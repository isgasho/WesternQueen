package master

import (
	"fmt"
	"github.com/arcosx/WesternQueen/util"
	mapset "github.com/deckarep/golang-set"
	"time"
)

// 全局变量
// 全局错误traceID
var WrongTraceSet mapset.Set
var TraceData util.TraceData

// 开启运行

// 初始化 RPC 服务
func init() {
	TraceData = make(util.TraceData)
	WrongTraceSet = mapset.NewSet()
}

func ReceiveWrongTraceData(wrongTraceId string) {
	//fmt.Println("ReceiveWrongTraceData : ", wrongTraceId)
	WrongTraceSet.Add(wrongTraceId)
}

func ReceiveTraceData(traceId string, spans [][]byte) {
	// 存在就追加
	if _, ok := TraceData[traceId]; ok {
		spanSlice := TraceData[traceId]
		tmpSpanSlice := make(util.SpanSlice, len(spans))
		for k := range spans {
			tmpSpanSlice[k] = spans[k]
		}
		spanSlice = append(tmpSpanSlice)
		TraceData[traceId] = spanSlice
	} else {
		// 不存在就创建
		spanSlice := make(util.SpanSlice, len(spans))
		for k := range spans {
			spanSlice[k] = spans[k]
		}
		TraceData[traceId] = spanSlice
	}
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
		<-t
		fmt.Println("WrongTraceSet", WrongTraceSet.Cardinality())
		fmt.Println("TraceData", len(TraceData))
	}
}
