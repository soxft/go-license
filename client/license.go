package license

import (
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/imroc/biu"
	"github.com/soxft/go-license/stru"
	"github.com/soxft/go-machine-code/machine"
	"github.com/wenzhenxi/gorsa"
	"log"
	"os"
	"time"
)

var Exit chan bool

// 本机运行后 获取一个唯一的机器码 ， 将 MAC 地址  / 机器序列号  加密后 注册至 License Server

// License Server 用于设置 是否运行机器继续运行
// 启动时将检查一次, 随后 每隔 30秒检查一次。连续三次失败，则退出程序

var licenseServer string
var PublicKey string

func Run(maxRetries int, interval time.Duration, _licenseServer string, _publicKey string) {
	Exit = make(chan bool)

	licenseServer = _licenseServer
	PublicKey = _publicKey

	result, err := RunCheck()
	if err != nil {
		log.Printf("[LICENSE] license 服务启动失败: %s", err.Error())
		os.Exit(1)
	} else if result.Status != 0 {
		log.Printf("[LICENSE] license 无效: %s", result.StatusStr)
		os.Exit(1)
	}

	go runLicenseClient(maxRetries, interval)
}

func runLicenseClient(maxRetries int, interval time.Duration) {
	var retries = 1

	for {

		result, err := RunCheck()
		if errors.Is(err, stru.ErrLicenseClientRunFailed) {
			log.Printf("[LICENSE] license Client 启动失败: %s", err.Error())
			Exit <- true
		} else if errors.Is(err, stru.ErrLicenseServerTimeout) {
			log.Printf("[LICENSE] 无法连接到 LICENSE 服务器: %s %d/%d", err.Error(), retries, maxRetries)
			retries += 1
		} else if errors.Is(err, stru.ErrLicenseInvalid) {
			log.Printf("[LICENSE] license 无效: %s", err.Error())
			Exit <- true
		}

		// 未通过
		if result.Status == -1 {
			log.Printf("[LICENSE] license 错误: %s", result.StatusStr)
			Exit <- true
		}

		if result.Status == -2 {
			log.Printf("[LICENSE] license 已过期")
			Exit <- true
		}

		if result.Status == -3 {
			log.Printf("[LICENSE] license 无效")
			Exit <- true
		}

		// 正常状态
		if result.Status == 1 {
			retries = 1
		}

		if retries > maxRetries {
			Exit <- true
		}

		time.Sleep(interval)
	}
}

func RunCheck() (stru.ConnData, error) {
	client := resty.New().SetTimeout(time.Second * 5).R()

	machineInfo, err := getMachineInfo()
	if err != nil {
		//log.Printf("[Error] License 服务启动失败, 错误码: rsa.encode.err")
		//Exit <- true
		return stru.ConnData{}, stru.ErrLicenseClientRunFailed
	}

	res, err := client.SetBody(machineInfo.Sign).Post(licenseServer + "/check_license")

	if err != nil {
		//log.Printf("[check_license] > %s 接到授权服务器失败", err.Error())
		return stru.ConnData{}, stru.ErrLicenseServerTimeout
	}

	// 将返回数据解密
	resp := biu.BinaryStringToBytes(string(res.Body()))
	result, err := gorsa.PublicDecrypt(string(resp), PublicKey)
	if err != nil {
		return stru.ConnData{}, stru.ErrLicenseInvalid
	}

	// 解析返回数据
	var LData stru.ConnData
	if err := json.Unmarshal([]byte(result), &LData); err != nil {
		return stru.ConnData{}, stru.ErrLicenseInvalid
	}

	return LData, nil
}

func getMachineInfo() (stru.MachineInfo, error) {
	serialNumber, err := machine.GetSerialNumber()
	if err != nil {
		log.Printf("无法获取机器序列号, 请重试: %s", err.Error())
		return stru.MachineInfo{}, err
	}

	mac, err := machine.GetMACAddress()
	if err != nil {
		log.Printf("无法获取机器Mac地址, 请重试: %s", err.Error())
		return stru.MachineInfo{}, err
	}

	mInfo := stru.MachineInfo{
		Mac:       mac,
		Serial:    serialNumber,
		Timestamp: time.Now().Unix(),
	}

	EncMachineData(&mInfo)

	return mInfo, nil
}

func EncMachineData(mInfo *stru.MachineInfo) {
	machineInfoStr, _ := json.Marshal(mInfo)
	encMInfo, _ := gorsa.PublicEncrypt(string(machineInfoStr), PublicKey)

	mInfo.Sign = biu.BytesToBinaryString([]byte(encMInfo))
}
