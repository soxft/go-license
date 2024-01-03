package server

import (
	_ "embed"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/imroc/biu"
	"github.com/wenzhenxi/gorsa"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

var PrivateKey string

func Run() {
	// load private key

	pkeys, err := os.ReadFile("../keys/private.pem")
	if err != nil {
		log.Fatalf("读取私钥失败: %s", err.Error())
	}
	PrivateKey = string(pkeys)

	r := gin.Default()

	initRoute(r)

	log.Println("Server running at 127.0.0.1:8080")
	if err = r.Run("127.0.0.1:8080"); err != nil {
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

func checkLicense(c *gin.Context) {
	// 获取完整的请求body
	rawBody, err := c.GetRawData()
	if err != nil {
		c.Data(http.StatusOK, "text/plain", GetEnc(ConnData{
			Status:    -1,
			StatusStr: "非法的请求构造, err: body.get.err",
		}))
		return
	}

	signBytes := biu.BinaryStringToBytes(string(rawBody))
	// decode
	jData, err := gorsa.PriKeyDecrypt(string(signBytes), PrivateKey)
	if err != nil {
		c.Data(http.StatusOK, "text/plain", GetEnc(ConnData{
			Status:    -1,
			StatusStr: "非法的请求构造, err: enc.decode.err",
		}))
		return
	}

	var jDataS MachineInfo
	if err := json.Unmarshal([]byte(jData), &jDataS); err != nil {
		c.Data(http.StatusOK, "text/plain", GetEnc(ConnData{
			Status:    -1,
			StatusStr: "非法的请求构造, err: json.decode.err",
		}))
		return
	}

	// 判断服务器客户端时间 是否大于 一分钟
	if math.Abs(float64(jDataS.Timestamp-time.Now().Unix())) >= 60 {
		c.Data(http.StatusOK, "text/plain", GetEnc(ConnData{
			Status:    -1,
			StatusStr: "非法的请求构造, err: time.too_large.err",
		}))
		return
	}

	// 构造 响应
	c.Data(http.StatusOK, "text/plain", GetEnc(ConnData{
		Status:    1,
		StatusStr: "有效的License",
	}))
	return
}

func GetEnc(cData ConnData) []byte {
	cData.Timestamp = time.Now().Unix()

	jcData, _ := json.Marshal(cData)
	encJcData, _ := gorsa.PriKeyEncrypt(string(jcData), PrivateKey)

	binaryString := biu.BytesToBinaryString([]byte(encJcData))
	return []byte(binaryString)
}
