package util

import (
	"strconv"
	"strings"
	"unsafe"
)

// d614959183521b4b -> d614959183521b4b|1587457762873000|d614959183521b4b|0|311601|order|getOrder|192.168.1.3|http.status_code=200

type Span []byte      // just a line
type SpanSlice []Span // some lines
type TraceData map[string]SpanSlice

// 索引标记 最后一次出现的位置
type TraceCache map[string]int64

type TraceDataDim struct {
	TraceId    string
	SpanSlices SpanSlice
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

func getStartTime(span []byte) int64 {
	var low, high int
	low = strings.Index(Bytes2str(span), "|") + 1
	if low == 0 {
		return -1
	}

	high = strings.Index(Bytes2str(span[low:]), "|")
	if high == -1 {
		return -1
	}
	ret, _ := strconv.ParseInt(
		Bytes2str(span[low:low+high]), 10, 64)
	return ret
}


//string convert to bytes, please pay attention! it doesn't use copy
func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

//bytes convert to string,  please pay attention! it doesn't use copy
func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
