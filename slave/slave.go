package slave

import (
	"bufio"
	"context"
	"fmt"
	"github.com/arcosx/WesternQueen/rpc"
	"github.com/arcosx/WesternQueen/util"
	mapset "github.com/deckarep/golang-set"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// 全局缓存区
var TraceCache util.TraceCache
var TraceData util.TraceData
var RPCClient western_queen.WesternQueenClient

// 发送chan
var SendTraceDataCh chan util.TraceDataDim

// 完成标志符
var finishFlag bool

// 全局变量
// 全局错误traceID
var WrongTraceSet mapset.Set

func init() {
	TraceData = make(util.TraceData)
	TraceCache = make(util.TraceCache)
	SendTraceDataCh = make(chan util.TraceDataDim)
	WrongTraceSet = mapset.NewSet()
	finishFlag = false
}

// 开启运行
func Start() {
	// 同步错误traceID 队列
	go ReadShareWrongTraceData()
	// 发送 trace 数据队列
	go SendTraceData()
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
	var lineCount int64
	var beginTime = time.Now()

	for {
		// 按行读取 按行处理
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
			TraceCache[string(traceId)] = lineCount
			//  判断表达式
			if len(tags) > 8 {
				if strings.Contains(string(tags), "error=1") ||
					(strings.Contains(string(tags), "http.status_code=") &&
						!strings.Contains(string(tags), "http.status_code=200")) {
					go SendWrongTraceData(traceId)
				}
			}
		}
		// 达到开始批处理
		if lineCount%util.ProcessBatchSize == 0 {
			for traceId, index := range TraceCache {
				pos := lineCount - index
				// 位置超越界限
				if pos > util.MaxSpanSplitSize {
					if findInWrongTraceSet(traceId) {
						var traceDataDim util.TraceDataDim
						traceDataDim.TraceId = traceId
						traceDataDim.SpanSlices = TraceData[traceId]
						SendTraceDataCh <- traceDataDim
						// TODO:后续如果再发现同 traceID 有误呢 ?
						delete(TraceCache, traceId)
						delete(TraceData, traceId)
					} else {
						delete(TraceCache, traceId)
						delete(TraceData, traceId)
					}
				}
			}
		}
	}
	finishFlag = true
	close(SendTraceDataCh)
	fmt.Println("finish used time: ", time.Since(beginTime))
}

func findInWrongTraceSet(traceId string) bool {
	if WrongTraceSet.Contains(traceId) {
		return true
	}
	return false
}

// 写入错误信息
func SendWrongTraceData(traceId []byte) {
	// 先写本地 再写主节点
	WrongTraceSet.Add(string(traceId))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := RPCClient.SendWrongTraceData(ctx, &western_queen.WrongTraceDataRequest{
		TraceId: string(traceId),
	})
	if err != nil {
		fmt.Println("RPCClient SendWrongTraceDataStream error", err)
	}
}

// 流写入出现错误的调用链
func SendTraceData() {
	stream, err := RPCClient.SendTraceDataStream(context.Background())
	if err != nil {
		fmt.Println("SendTraceData error", err)
	}
	for {
		if !finishFlag {
			tmp := <-SendTraceDataCh
			sliceBytes := make([][]byte, len(tmp.SpanSlices))
			for k, v := range tmp.SpanSlices {
				sliceBytes[k] = v
			}
			stream.Send(&western_queen.TraceData{
				TraceId: tmp.TraceId,
				Spans:   sliceBytes,
			})
		} else {
			stream.CloseSend()
		}

	}
}

// 流读取错误信息
func ReadShareWrongTraceData() {
	d := time.Microsecond * 10

	t := time.Tick(d)

	fmt.Println("Read share wrong trace data from master !")
	stream, err := RPCClient.ReadShareWrongTraceData(context.Background(), &western_queen.Empty{
	})
	if err != nil {
		fmt.Println("ReadShareWrongTraceData error", err)
	}
	for {
		<-t
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("ReadShareWrongTraceData stream Recv", err)
		}
		// 合并本地错误列表
		// shit code !
		tmp := make([]interface{}, len(resp.WrongTraceDataRequests))
		for i, v := range resp.WrongTraceDataRequests {
			tmp[i] = v
		}
		shareWrongTraceData := mapset.NewSetFromSlice(tmp)
		WrongTraceSet.Union(shareWrongTraceData) // ? 这里一定需要 union 吗？
	}
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
