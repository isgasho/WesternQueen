package newslave

import (
	"bufio"
	"fmt"
	"github.com/arcosx/WesternQueen/util"
	cmap "github.com/orcaman/concurrent-map"
	"io"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func init() {
	util.DebugMode = true
	util.Mode = "slave1"

}
func init() {
	util.DebugMode = true
	util.Mode = "slave1"

}

func Test_start(t *testing.T) {
	Start()
}

func Test_Down(t *testing.T) {
	m := cmap.New()
	runtime.GOMAXPROCS(2)
	url := "http://localhost:9971/trace1.data1"
	var beginTime = time.Now()
	res, _ := http.Head(url) // 187 MB file of random numbers per line
	maps := res.Header
	length, _ := strconv.Atoi(maps["Content-Length"][0]) // Get the content length from the header request
	limit := 20                                          // 10 Go-routines for the process so each downloads 18.7MB
	len_sub := length / limit                            // Bytes for each Go-routine
	diff := length % limit                               // Get the remaining for the last request
	var wg sync.WaitGroup

	for i := 0; i < limit; i++ {
		wg.Add(1)

		min := len_sub * i       // Min range
		max := len_sub * (i + 1) // Max range

		if i == limit-1 {
			max += diff // Add the remaining bytes in the last request
		}

		go func(min int, max int, i int) {
			client := &http.Client{}
			req, _ := http.NewRequest("GET", url, nil)
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
					str := util.Bytes2str(tags)
					if strings.Contains(str, "error=1") ||
						(strings.Contains(str, "http.status_code=") &&
							!strings.Contains(str, "http.status_code=200")) {
						//WrongTraceSet.Add(util.Bytes2str(traceId))
						m.Set(util.Bytes2str(traceId),"")
					}
				}
			}
			wg.Done()
		}(min, max, i)
	}
	wg.Wait()
	fmt.Println("finish used time: ", time.Since(beginTime))
}
