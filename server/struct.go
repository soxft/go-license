package server

type ConnData struct {
	Status    int32 // license 状态 0: 无效, 1: 有效
	StatusStr string
	Timestamp int64
	Sign      string
}

type MachineInfo struct {
	Mac       string
	Serial    string
	Timestamp int64
	Sign      string
}
