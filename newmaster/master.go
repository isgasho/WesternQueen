package newmaster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/arcosx/WesternQueen/util"
	mapset "github.com/deckarep/golang-set"
	md5simd "github.com/minio/md5-simd"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SpanSlice []string
type TraceData map[string]SpanSlice

// 最终结果
var traceMd5Map = make(map[string]string)

var WrongTraceSet mapset.Set //存在错误的Trace的TraceID
var FullTraceSet mapset.Set  // 整串 Trace
var slave1FirstStage bool    // 从节点一 首阶段
var slave2FirstStage bool    // 从节点二 首阶段
var slave1SecondStage bool   // 从节点一 二阶段
var slave2SecondStage bool   // 从节点二 二阶段
func init() {
	WrongTraceSet = mapset.NewSet()
	FullTraceSet = mapset.NewSet()
}

func GetWrongTraceList(wrongTraceList []string, node string) {
	for _, v := range wrongTraceList {
		WrongTraceSet.Add(v)
	}
	if node == util.SLAVE_ONE_MODE {
		slave1FirstStage = true
	}
	if node == util.SLAVE_TWO_MODE {
		slave2FirstStage = true
	}
	// 两个都收齐
	if slave1FirstStage && slave2FirstStage {
		go SendShareWrongTraceSet(util.SLAVE_ONE_PORT_8000)
		go SendShareWrongTraceSet(util.SLAVE_TWO_PORT_8001)
	}
}

func GetAllTraceList(traceList []string, node string) {
	for _, v := range traceList {
		FullTraceSet.Add(v)
	}
	if node == util.SLAVE_ONE_MODE {
		slave1SecondStage = true
	}
	if node == util.SLAVE_TWO_MODE {
		slave2SecondStage = true
	}
	if slave1SecondStage && slave2SecondStage {
		LastFinish()
	}
}

// 上传最终结果
func LastFinish() {
	fmt.Println("LastFinish Begin")
	var traceData TraceData
	traceData = make(TraceData)
	// 再次切割整串为Map以便写入  TODO: 优化这里
	it := FullTraceSet.Iterator()
	for elem := range it.C {
		fullStr := elem.(string)
		firstIndex := strings.Index(fullStr, "|")
		if firstIndex == -1 {
			continue
		}
		traceId := fullStr[:firstIndex]
		traceData[traceId] = append(traceData[traceId], fullStr)
	}
	// 开始排序 TODO: 优化这里
	fmt.Println("all slave finish ! begin sort and upload result")
	server := md5simd.NewServer()
	defer server.Close()

	md5Hash := server.NewHash()
	defer md5Hash.Close()
	sortBeginTime := time.Now()
	for traceId := range traceData {

		sort.Sort(traceData[traceId])
		md5Hash.Reset()
		for span := range traceData[traceId] {
			md5Hash.Write([]byte(traceData[traceId][span]))
		}
		digest := md5Hash.Sum([]byte{})
		traceMd5Map[traceId] = fmt.Sprintf("%x", digest)
	}
	fmt.Println("sort result and md5 use : ", time.Since(sortBeginTime))
	sendMd5Result()
}

func SendShareWrongTraceSet(port string) {
	// 发送自己的错误traceID到master

	jsonStr, err := json.Marshal(WrongTraceSet)
	if err != nil {
		fmt.Println("SendWrongTraceSet json.Marshal failed", err)
		return
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost%s/getShare", port), bytes.NewBuffer(jsonStr))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("SendWrongTraceSet post failed", err)
		return
	}
	defer resp.Body.Close()
}

func (s SpanSlice) Len() int {
	return len(s)
}

func (s SpanSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

//根据 span 中 timestamp 排序，用于
func (s SpanSlice) Less(i, j int) bool {
	return getStartTime(s[i]) < getStartTime(s[j])
}

func getStartTime(span string) int64 {
	var low, high int
	low = strings.Index(span, "|") + 1
	if low == 0 {
		return -1
	}

	high = strings.Index(span[low:], "|")
	if high == -1 {
		return -1
	}
	ret, _ := strconv.ParseInt(
		span[low:low+high], 10, 64)
	return ret
}

func sendMd5Result() bool {
	result, _ := json.Marshal(traceMd5Map)
	data := make(url.Values)
	data.Add("result", util.Bytes2str(result))
	url := fmt.Sprintf("http://localhost:%s/api/finished", util.RESULT_UPLOAD_PORT)
	resp, err := http.PostForm(url, data)
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
