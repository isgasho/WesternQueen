package util

// 运行模式

var Mode string
var DebugMode bool

const (
	SLAVE_ONE_MODE string = "slave1"
	SLAVE_TWO_MODE string = "slave2"
	MASTER_MODE    string = "master"
)

// 端口定义
const MASTER_PORT_8002 = ":8002"
const MASTER_PORT_8003 = ":8003"
const MASTER_PORT_8004 = ":8004"

const SLAVE_ONE_PORT_8000 = ":8000"
const SLAVE_TWO_PORT_8001 = ":8001"

// 程序运行参数

// 批次处理大小
const ProcessBatchSize = 25000

// 同 traceID 出现位置
const MaxSpanSplitSize = 20000

// 同时拉取的进程数
const ParaPull = 20
// 外部传递

var DEBUG_DATA_SOURCE_PORT = "9971"
var DATA_SOURCE_PORT string   //拉取数据端口
var RESULT_UPLOAD_PORT string // 上传checksum结果端口 理论来说和拉取数据端口一致
