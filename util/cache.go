package util

// d614959183521b4b -> d614959183521b4b|1587457762873000|d614959183521b4b|0|311601|order|getOrder|192.168.1.3|http.status_code=200

type Span []byte      // just a line
type SpanSlice []Span // some lines
type TraceData map[string]SpanSlice


type TraceCache []string
