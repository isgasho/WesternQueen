package master

import (
	"encoding/json"
	"fmt"
	"github.com/arcosx/WesternQueen/util"
	mapset "github.com/deckarep/golang-set"
	md5simd "github.com/minio/md5-simd"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"
)

// 全局变量
// 全局错误traceID
var WrongTraceSet mapset.Set
var TraceData util.TraceData

// 最终结果
var traceMd5Map = make(map[string]string)

var slave1Finished bool
var slave2Finished bool

// 开启运行

// 初始化 RPC 服务
func init() {
	TraceData = make(util.TraceData)
	WrongTraceSet = mapset.NewSet()
	slave1Finished = false
	slave2Finished = false
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

func Finish(node string) {

	if node == "slave1" {
		slave1Finished = true
	}
	if node == "slave2" {
		slave2Finished = true
	}
	//  两个全部完成开启排序并且发送
	if slave1Finished && slave2Finished {
		fmt.Println("all slave finish ! begin sort and upload result")
		server := md5simd.NewServer()
		defer server.Close()

		md5Hash := server.NewHash()
		defer md5Hash.Close()
		//计算 md5
		for traceId := range TraceData {
			// TODO: 排序能否更快一步
			sortBeginTime := time.Now()
			sort.Sort(TraceData[traceId])
			fmt.Println("sort result use : ", time.Since(sortBeginTime))
			md5Hash.Reset()
			for span := range TraceData[traceId] {
				md5Hash.Write(TraceData[traceId][span])
			}
			digest := md5Hash.Sum([]byte{})
			traceMd5Map[traceId] = fmt.Sprintf("%x", digest)
		}

		// 发送结果
		go sendMd5Result()
	}
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

func sendMd5Result() bool {
	result, _ := json.Marshal(traceMd5Map)
	data := make(url.Values)
	data.Add("result", util.Bytes2str(result))
	resp, err := http.PostForm(fmt.Sprintf("http://localhost:%s/api/finished", util.RESULT_UPLOAD_PORT), data)
	if err == nil {
		defer resp.Body.Close()
	} else {
		log.Fatalln(err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		//log.Println("suc to sendCheckSum, result:" + util.Bytes2str(result))
		return true
	}

	log.Println("fail to sendCheckSum:" + resp.Status)
	return false
}
