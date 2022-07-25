package tcp

/**
 * A tcp server
 */

import (
	"context"
	"fmt"
	"github.com/gofaquan/go-redis/interface/tcp"
	"github.com/gofaquan/go-redis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Config 启动 TCP Server 的配置
type Config struct {
	Address string
}

//ListenAndServeWithSignal 绑定端口并处理请求，直到收到停止信号为止
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	//信号发送
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	//系统收到这些信号就转发到 sigCh
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		//sigCh 被传入系统转发来的信号就向 closeChan 发送信号
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{} //发送空结构体代表发送信号，需要停止服务
		}
	}()
	//绑定端口并且监听
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("绑定地址为: %s, 服务端开始监听...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}

//ListenAndServe 绑定端口并处理请求，直到关闭为止
//closeChan <-chan struct{} 传入空结构体，达到只发信号不传值的作用
func ListenAndServe(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) {
	// 监听信号
	go func() {
		<-closeChan //阻塞等待
		logger.Info("正在关闭...")
		_ = listener.Close() // listener.Accept() 返回 err
		_ = handler.Close()  // 关闭连接
	}()

	// 监听端口
	defer func() {
		// 遇到未知错误关闭
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		conn, err := listener.Accept() //接收新连接
		if err != nil {
			break
		}
		// go func 处理新的客户端连接
		logger.Info("接收到了一个新的客户端连接!")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	//保证所有客户端退出后才退出
	waitDone.Wait()
}
