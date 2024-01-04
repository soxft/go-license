package server

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/imroc/biu"
	"github.com/soxft/go-license/stru"
	"github.com/wenzhenxi/gorsa"
	"gorm.io/gorm"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

var PrivateKey string

func loadPrivateKey() {
	pKeys, err := os.ReadFile("keys/private.pem")
	if err != nil {
		log.Fatalf("读取私钥失败: %s", err.Error())
	}
	PrivateKey = string(pKeys)
}

var method string
var serialNumber string
var dueTime string
var listen string

func Run() {
	flag.StringVar(&method, "m", "server", "指定运行模式, `server`: running server, `set` set license due time")
	flag.StringVar(&serialNumber, "s", "", "指定 serialNumber")
	flag.StringVar(&dueTime, "d", "2004-11-23", "指定过期时间, ./license -m ser -s xxxxxx -d 2023-12-24")
	flag.StringVar(&listen, "l", "127.0.0.1:1111", "指定监听")
	flag.Parse()

	// 处理数据库
	loadMysql()

	if method == "set" {

		date, err := time.Parse("2006-01-02", dueTime)
		if err != nil {
			fmt.Println("无法解析日期:", err)
			return
		}

		if err = SetLicenseDueTime(serialNumber, date.Unix()); err != nil {
			log.Printf("[Err] 修改 license 效期失败: %s", err)
		}

		log.Printf("修改成功 %s -> %s", serialNumber, dueTime)

		return
	}

	// load private key
	loadPrivateKey()

	// 启动 Web 服务
	r := gin.Default()

	initRoute(r)

	log.Println("Server running at " + listen)
	if err := r.Run(listen); err != nil {
		log.Fatalf("运行服务器失败: %s", err.Error())
	}
}

func initRoute(r *gin.Engine) {
	// ping handler
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "pong",
		})
	})

	// handler check License
	r.POST("/check_license", checkLicense)
}

// SetLicenseDueTime 设置 指定 serial 的过期时间
func SetLicenseDueTime(serial string, dueTime int64) error {
	// 先检查是否存在 这个 License
	var lic License
	db := D.Model(License{}).Where(License{Serial: serial}).Take(&lic)
	if errors.Is(db.Error, gorm.ErrRecordNotFound) {
		// 创建
		return D.Model(License{}).Create(&License{
			ID:      0,
			Serial:  serial,
			DueTime: dueTime,
		}).Error
	} else if db.Error != nil {
		// 修改
		log.Println(serial, dueTime, "2")
		return db.Error
	}

	return D.Model(License{}).Where(License{Serial: serial}).Updates(License{DueTime: dueTime}).Error
}

func checkLicense(c *gin.Context) {
	// 获取完整的请求body
	rawBody, err := c.GetRawData()
	if err != nil {
		c.Data(http.StatusOK, "text/plain", GetEnc(stru.ConnData{
			Status:    -1,
			StatusStr: "非法的请求构造, err: body.get.err",
		}))
		return
	}

	signBytes := biu.BinaryStringToBytes(string(rawBody))
	// decode
	jData, err := gorsa.PriKeyDecrypt(string(signBytes), PrivateKey)
	if err != nil {
		c.Data(http.StatusOK, "text/plain", GetEnc(stru.ConnData{
			Status:    -1,
			StatusStr: "非法的请求构造, err: enc.decode.err",
		}))
		return
	}

	var jDataS stru.MachineInfo
	if err := json.Unmarshal([]byte(jData), &jDataS); err != nil {
		c.Data(http.StatusOK, "text/plain", GetEnc(stru.ConnData{
			Status:    -1,
			StatusStr: "非法的请求构造, err: json.decode.err",
		}))
		return
	}

	// 判断服务器客户端时间 是否大于 一分钟
	if math.Abs(float64(jDataS.Timestamp-time.Now().Unix())) >= 60 {
		c.Data(http.StatusOK, "text/plain", GetEnc(stru.ConnData{
			Status:    -1,
			StatusStr: "非法的请求构造, err: time.too_large.err",
		}))
		return
	}

	var lic License
	db := D.Model(License{}).Where(License{Serial: jDataS.Serial}).Take(&lic)
	if errors.Is(db.Error, gorm.ErrRecordNotFound) {
		c.Data(http.StatusOK, "text/plain", GetEnc(stru.ConnData{
			Status:    -1,
			StatusStr: "无效的 License, Serial: " + jDataS.Serial,
		}))
		return
	}

	if time.Now().Unix() >= lic.DueTime {
		c.Data(http.StatusOK, "text/plain", GetEnc(stru.ConnData{
			Status:    -2,
			StatusStr: "License 已过期, Serial: " + jDataS.Serial,
		}))
		return
	}

	// 构造 响应
	dueTime := time.Unix(lic.DueTime, 0)

	c.Data(http.StatusOK, "text/plain", GetEnc(stru.ConnData{
		Status:    -1,
		StatusStr: "有效的License, 有效期至: " + dueTime.Format("2006-01-02"),
	}))
	return

	// 构造 响应
	c.Data(http.StatusOK, "text/plain", GetEnc(stru.ConnData{
		Status:    0,
		StatusStr: "有效的License, 有效期至: " + dueTime.Format("2006-01-02"),
	}))
	return
}

func GetEnc(cData stru.ConnData) []byte {
	cData.Timestamp = time.Now().Unix()

	jcData, _ := json.Marshal(cData)
	encJcData, _ := gorsa.PriKeyEncrypt(string(jcData), PrivateKey)

	binaryString := biu.BytesToBinaryString([]byte(encJcData))
	return []byte(binaryString)
}
