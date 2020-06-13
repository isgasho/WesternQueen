package slave

import (
	"bufio"
	"fmt"
	"github.com/arcosx/WesternQueen/util"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// 全局缓存区
var TraceCache util.TraceCache
var TraceData util.TraceData

func init() {
	TraceData = make(util.TraceData)
}

// 开启运行
func Start() {
	go ReadShareWrongTraceData()

	dataSourcePath := getDataSourcePath()
	if dataSourcePath == "" {
		fmt.Println("getDataSourcePath failed")
		return
	}
	fmt.Println("data source path:", dataSourcePath)

	// TODO: Range 并发拉取 -> 线程隔离
	resp, err := http.Get(dataSourcePath)
	if err == nil {
		defer resp.Body.Close()
	} else {
		log.Fatalln(err)
	}
	bufReader := bufio.NewReader(resp.Body)
	// 开始拉取
	var lineCount = 0
	var beginTime = time.Now()

	for {
		line, err := bufReader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Println("bufReader.ReadBytes meet unsolved error")
			panic(err)
		}
		if len(line) == 0 && err == io.EOF {
			break
		}
		lineCount++
		firstIndex := strings.Index(string(line), "|")
		if firstIndex == -1 {
			continue
		}
		traceId := line[:firstIndex]

		//获得 tags
		lastIndex := strings.LastIndex(string(line), "|")
		if lastIndex == -1 {
			continue
		}
		tags := line[lastIndex : len(line)-1]
		if len(tags) > 0 {

			TraceData[string(traceId)] = append(TraceData[string(traceId)], line)

			//  判断表达式
			if len(tags) > 8 {
				if strings.Contains(string(tags), "error=1") ||
					(strings.Contains(string(tags), "http.status_code=") &&
						!strings.Contains(string(tags), "http.status_code=200")) {
					go SendWrongTraceData(traceId, line)
				}
			}
		}
		// 达到开始批处理
		if lineCount%util.ProcessBatchSize == 0 {
			// 判断是否存在于本地的错误表信息内
			// 如果存在
			SendTraceData()
			// 否则 
			fmt.Println("get ProcessBatchSize", lineCount)
		}
	}

	fmt.Println("finish used time: ", time.Since(beginTime))
}

// 处理程序
func ProcessData() {

}

// 流写入错误信息
func SendWrongTraceData(traceId []byte, line []byte) {

}

// 流写入出现错误的调用链
func SendTraceData(wrongTraceData util.TraceData) {

}

// 流读取错误信息
func ReadShareWrongTraceData() {

}

func getDataSourcePath() string {
	if util.DebugMode {
		return fmt.Sprintf("http://localhost:%s/trace1.data", util.DEBUG_DATA_SOURCE_PORT)
	}
	if util.Mode == util.SLAVE_ONE_MODE {
		return fmt.Sprintf("http://localhost:%s/trace1.data", util.DATA_SOURCE_PORT)
	}
	if util.Mode == util.SLAVE_TWO_MODE {
		return fmt.Sprintf("http://localhost:%s/trace2.data", util.DATA_SOURCE_PORT)
	}
	return ""
}
