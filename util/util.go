package util

// 我是从节点
func IsSlave() bool {
	if Mode == SLAVE_ONE_MODE {
		return true
	}
	if Mode == SLAVE_TWO_MODE {
		return true
	}
	return false
}

// 我是主节点
func IsMaster() bool {
	if Mode == MASTER_MODE {
		return true
	}
	return false
}

