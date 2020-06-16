package slave

import (
	"bufio"
	"fmt"
	rpc "github.com/arcosx/WesternQueen/rpc"
	"github.com/arcosx/WesternQueen/util"
	"google.golang.org/grpc"
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
var wg sync.WaitGroup

func init() {
	util.DebugMode = true
	util.Mode = "slave1"

}

// 单机调试使用此方法
func TestStart(t *testing.T) {
	conn, err := grpc.Dial("localhost:8003", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	RPCClient = rpc.NewWesternQueenClient(conn)
	Start()
}

func TestGetFile(t *testing.T)  {
	runtime.GOMAXPROCS(4)
	var beginTime = time.Now()
	res, _ := http.Head("http://localhost:9971/trace1.data1"); // 187 MB file of random numbers per line
	maps := res.Header
	length, _ := strconv.Atoi(maps["Content-Length"][0]) // Get the content length from the header request
	limit := 100 // 10 Go-routines for the process so each downloads 18.7MB
	len_sub := length / limit // Bytes for each Go-routine
	diff := length % limit // Get the remaining for the last request
	for i := 0; i < limit ; i++ {
		wg.Add(1)

		min := len_sub * i // Min range
		max := len_sub * (i + 1) // Max range

		if (i == limit - 1) {
			max += diff // Add the remaining bytes in the last request
		}

		go func(min int, max int, i int) {
			client := &http.Client {}
			req, _ := http.NewRequest("GET", "http://localhost:9971/trace1.data1", nil)
			range_header := "bytes=" + strconv.Itoa(min) +"-" + strconv.Itoa(max-1) // Add the data for the Range header of the form "bytes=0-100"
			req.Header.Add("Range", range_header)
			resp,_ := client.Do(req)
			defer resp.Body.Close()
			bufReader := bufio.NewReader(resp.Body)
			for  {
				line, err := bufReader.ReadBytes('\n')
				if err != nil && err != io.EOF {
					log.Println("bufReader.ReadBytes meet unsolved error")
					panic(err)
				}
				if len(line) == 0 && err == io.EOF {
					break
				}
				firstIndex := strings.Index(string(line), "|")
				if firstIndex == -1 {
					continue
				}
				_ = line[:firstIndex]

				//获得 tags
				lastIndex := strings.LastIndex(string(line), "|")
				if lastIndex == -1 {
					continue
				}
				tags := line[lastIndex : len(line)-1]
				if len(tags) > 0 {
					//  判断表达式
					if len(tags) > 8 {
						str := string(tags)
						if strings.Contains(str, "error=1") ||
							(strings.Contains(str, "http.status_code=") &&
								!strings.Contains(str, "http.status_code=200")) {
							//fmt.Println(traceId)
						}
					}
				}
			}
			wg.Done()
			//          ioutil.WriteFile("new_oct.png", []byte(string(body)), 0x777)
		}(min, max, i)
	}
	wg.Wait()
	fmt.Println("finish used time: ", time.Since(beginTime))
}
