package newslave

import (
	"bufio"
	"bytes"
	jsoniter "github.com/json-iterator/go"
	"fmt"
	"github.com/arcosx/WesternQueen/util"
	mapset "github.com/deckarep/golang-set"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var WrongTraceSet mapset.Set
var WrongTraceMap sync.Map
var FullTraceSet mapset.Set // 整串 Trace
func init() {
	WrongTraceSet = mapset.NewSet()
	FullTraceSet = mapset.NewSet()

}

// 第一遍拉取仅过滤错误节点
func Start() {
	dataSourcePath := getDataSourcePath()
	if dataSourcePath == "" {
		fmt.Println("getDataSourcePath failed")
		return
	}
	fmt.Println("data source path:", dataSourcePath)

	var wg sync.WaitGroup
	var beginTime = time.Now()
	res, _ := http.Head(dataSourcePath)
	maps := res.Header
	length, _ := strconv.Atoi(maps["Content-Length"][0])
	limit := util.ParaPull
	len_sub := length / limit
	diff := length % limit
	for i := 0; i < limit; i++ {
		wg.Add(1)

		min := len_sub * i
		max := len_sub * (i + 1)

		if i == limit-1 {
			max += diff
		}

		go func(min int, max int, i int) {
			client := &http.Client{}
			req, _ := http.NewRequest("GET", dataSourcePath, nil)
			range_header := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1)
			req.Header.Add("Range", range_header)
			resp, _ := client.Do(req)
			defer resp.Body.Close()
			bufReader := bufio.NewReader(resp.Body)
			for {
				line, err := bufReader.ReadBytes('\n')
				if err != nil && err != io.EOF {
					log.Println("bufReader.ReadBytes meet unsolved error")
					panic(err)
				}
				if len(line) == 0 && err == io.EOF {
					break
				}
				linsStr := util.Bytes2str(line)
				firstIndex := strings.Index(linsStr, "|")
				if firstIndex == -1 {
					continue
				}
				traceId := line[:firstIndex]

				//获得 tags
				lastIndex := strings.LastIndex(linsStr, "|")
				if lastIndex == -1 {
					continue
				}
				tags := line[lastIndex : len(line)-1]
				if len(tags) > 0 {
					//  判断表达式
					if len(tags) > 8 {
						str := util.Bytes2str(tags)
						if strings.Contains(str, "error=1") ||
							(strings.Contains(str, "http.status_code=") &&
								!strings.Contains(str, "http.status_code=200")) {
							go WrongTraceSet.Add(util.Bytes2str(traceId))
						}
					}
				}
			}
			wg.Done()
		}(min, max, i)
	}
	wg.Wait()
	fmt.Println("finish used time: ", time.Since(beginTime))
	fmt.Println(WrongTraceSet)
	fmt.Println("begin send wrong data")
	SendWrongTraceSet()
}

// 第二遍拉取过滤特定traceID
func again() {
	dataSourcePath := getDataSourcePath()
	if dataSourcePath == "" {
		fmt.Println("getDataSourcePath failed")
		return
	}
	fmt.Println("data source path:", dataSourcePath)

	var wg sync.WaitGroup
	var beginTime = time.Now()
	res, _ := http.Head(dataSourcePath)
	maps := res.Header
	length, _ := strconv.Atoi(maps["Content-Length"][0])
	limit := util.ParaPull
	len_sub := length / limit
	diff := length % limit
	for i := 0; i < limit; i++ {
		wg.Add(1)

		min := len_sub * i
		max := len_sub * (i + 1)

		if i == limit-1 {
			max += diff
		}

		go func(min int, max int, i int) {
			client := &http.Client{}
			req, _ := http.NewRequest("GET", dataSourcePath, nil)
			range_header := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1)
			req.Header.Add("Range", range_header)
			resp, _ := client.Do(req)
			defer resp.Body.Close()
			bufReader := bufio.NewReader(resp.Body)
			for {
				line, err := bufReader.ReadBytes('\n')
				if err != nil && err != io.EOF {
					log.Println("bufReader.ReadBytes meet unsolved error")
					panic(err)
				}
				if len(line) == 0 && err == io.EOF {
					break
				}
				lineStr := util.Bytes2str(line)
				firstIndex := strings.Index(lineStr, "|")
				if firstIndex == -1 {
					continue
				}
				traceId := line[:firstIndex]
				// 在错误集中找到
				_, ok := WrongTraceMap.Load(util.Bytes2str(traceId))
				if ok {
					FullTraceSet.Add(lineStr)
				}
			}
			wg.Done()
		}(min, max, i)
	}
	wg.Wait()
	fmt.Println("finish used time: ", time.Since(beginTime))
	fmt.Println("begin send full trace data ")
	go SendFullTraceSet()
}

// 发送自己的错误traceID到master
func SendWrongTraceSet() {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonStr, err := json.Marshal(WrongTraceSet)
	if err != nil {
		fmt.Println("SendWrongTraceSet json.Marshal failed", err)
		return
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost%s/getWrong?node=%s", util.MASTER_PORT_8002, util.Mode), bytes.NewBuffer(jsonStr))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("SendWrongTraceSet post failed", err)
		return
	}
	defer resp.Body.Close()
}

// 获得共享的错误traceID
func GetShareWrongTraceSet(wrongTraceList []string) {
	// 进行错误traceID的set合并
	for _, v := range wrongTraceList {
		WrongTraceSet.Add(v)
	}
	// 开始第二次拉取
	fmt.Println("finish GetShareWrongTraceSet", WrongTraceSet)
	for traceId := range WrongTraceSet.Iter() {
		WrongTraceMap.Store(traceId.(string), "")
	}
	again()
}

// 发送全量traceID到master
func SendFullTraceSet() {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	jsonStr, err := json.Marshal(FullTraceSet)
	if err != nil {
		fmt.Println("SendFullTraceSet json.Marshal failed", err)
		return
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost%s/all?node=%s", util.MASTER_PORT_8002, util.Mode), bytes.NewBuffer(jsonStr))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("SendFullTraceSet post failed", err)
		return
	}
	defer resp.Body.Close()
}


// 开发或者跑本地测试例时 走本地目录
func getDataSourcePath() string {
	if util.Mode == util.SLAVE_ONE_MODE {
		if util.DebugMode {
			return fmt.Sprintf("http://localhost:%s/trace1.data", util.DEBUG_DATA_SOURCE_PORT)
		}
		return fmt.Sprintf("http://localhost:%s/trace1.data", util.DATA_SOURCE_PORT)
	}
	if util.Mode == util.SLAVE_TWO_MODE {
		if util.DebugMode {
			return fmt.Sprintf("http://localhost:%s/trace2.data", util.DEBUG_DATA_SOURCE_PORT)
		}
		return fmt.Sprintf("http://localhost:%s/trace2.data", util.DATA_SOURCE_PORT)
	}
	return ""
}
