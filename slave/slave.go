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

// 开启运行
func Start() {
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
		fmt.Println(string(traceId),string(tags))
	}
	fmt.Println("finish used time: ", time.Since(beginTime))
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
