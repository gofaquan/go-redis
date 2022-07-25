package tcp

/**
 * echo server 用于测试服务器是否正常工作
 */

import (
	"bufio"
	"context"
	"github.com/gofaquan/go-redis/lib/logger"
	"github.com/gofaquan/go-redis/lib/sync/atomic"
	"github.com/gofaquan/go-redis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

//EchoHandler 将接收到的线路回送到客户端，用于测试
//Goland Ctrl + I 可以快速继承某个接口
type EchoHandler struct {
	activeConn sync.Map       //记录有多少个连接
	closing    atomic.Boolean //代表业务引擎状态
}

// NewHandler 创建 EchoHandler
func NewHandler() *EchoHandler {
	return &EchoHandler{}
}

//EchoClient 作为 EchoHandler 的客户端，用于测试
type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

// Close 用于 close connection
func (c *EchoClient) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second) //设置 TTL，还未关闭就强制关闭
	_ = c.Conn.Close()
	return nil
}

// Handle 处理客户发来的请求
func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		//关闭的过程中，不再接收新的连接请求
		_ = conn.Close()
	}

	client := &EchoClient{
		Conn: conn,
	}
	//只存 key，不存 value，就使用空结构体，由 hashMap 变为 hashSet
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		//可能发生：客户端 EOF、客户端超时、服务器提前关闭
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF { //客户端退出了
				logger.Info("一个客户端退出！")
				h.activeConn.Delete(client)
			} else {
				logger.Warn(err) //无法确定处理超时与提前关闭具体情况，先 Warn
			}
			return
		}

		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

// Close stops echo handler
func (h *EchoHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true) //设置正在退出的状态
	//sync.map 的 range
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*EchoClient) //空接口转为 client 客户端类型
		_ = client.Close()
		return true // 返回 true 就操作下一个对象，false 就不再操作，这里一直 true 保证全部连接 close
	})
	return nil //简化 error 返回
}
