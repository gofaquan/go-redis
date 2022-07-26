package tcp

import (
	"context"
	"github.com/gofaquan/go-redis/Interface/tcp"
	"go.uber.org/zap"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const TcpType = "tcp"

type Addr struct {
	Host    string
	Port    string
	NetType string
}

func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan os.Signal) {
	var ctx = context.Background() // 顶层 context
	var waitDone sync.WaitGroup    // 处理连接断开问题，全部断开程序再终止

	// 监听退出信号，下方是 for 循环，所以写在上面
	go func() {
		<-closeChan
		zap.L().Info("Shutting down....")
		listener.Close() // 负责监听的 socket 关闭连接
		handler.Close()  // 服务端 socket 关闭连接
	}()

	// 新连接处理逻辑
	for {
		conn, err := listener.Accept() // 监听新连接的加入
		if err != nil {
			break
		}
		// 处理新连接
		zap.L().Info("Accept a new link")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn) // 此处出现异常或者断开，会 defer 调用 Done 来 -1
		}()
	}

	waitDone.Wait() // 为 0 就全部结束
}

func ListenAndServeWithSignal(handler tcp.Handler, addr *Addr) {
	//closeChan := make(chan struct{}) // 空结构体作值，只发信号，节约空间
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	address := addr.Host + ":" + addr.Port
	listener, err := net.Listen(addr.NetType, address)
	if err != nil {
		zap.L().Error("ListenAndServeWithSignal:net.Listen ", zap.Error(err))
	}
	ListenAndServe(listener, handler, sigCh)
}
