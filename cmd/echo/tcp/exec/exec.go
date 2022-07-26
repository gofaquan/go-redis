package main

import (
	"fmt"
	echo "github.com/gofaquan/go-redis/cmd/echo/tcp"
	"github.com/gofaquan/go-redis/conf"
	"github.com/gofaquan/go-redis/log"
	"github.com/gofaquan/go-redis/tcp"
	"go.uber.org/zap"
)

func main() {
	if err := conf.Init(); err != nil {
		fmt.Printf("load config failed, err:%v\n", err)
		return
	}
	var host = conf.Conf.RedisConfig.Host
	var port = conf.Conf.RedisConfig.Port
	var logConf = conf.Conf.LogConfig

	if err := logger.Init(logConf, conf.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}
	zap.L().Info("server is starting , will arrive in " + host + ":" + port)
	tcp.ListenAndServeWithSignal(
		echo.NewEchoHandler(), &tcp.Addr{
			Host:    host,
			Port:    port,
			NetType: "tcp",
		})
}
