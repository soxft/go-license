# go-license

基于 Golang 开发的一套 在线 License 系统

# client
```
go get -u github.com/soxft/go-license/client
```


```golang
import license "github.com/soxft/go-license/client"

var Pkey = `-----BEGIN RSA PUBLIC KEY-----
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
-----END RSA PUBLIC KEY-----`

func main() {
  license.Run(3, time.Second*30, "http://server:8080", Pkey)

  // main process
  go core()

  // 使用 license.Exit 阻塞运行
  <-license.Exit
}
```


# server
```golang
func main() {
  server.Run()
}
```

```shell
// 设置指定 serial 的到期时间
license_server -m set -s serialNumber -d 2024-11-23
```
