package stru

import "errors"

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

var (
	ErrLicenseInvalid         = errors.New("license 无效")
	ErrLicenseClientRunFailed = errors.New("license Client 启动失败")
	ErrLicenseServerTimeout   = errors.New("无法连接到 license 服务器")
)
