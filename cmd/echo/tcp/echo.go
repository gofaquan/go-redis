package tcp

import (
	"bufio"
	"context"
	"github.com/gofaquan/go-redis/pkg/sync/atomic"
	"github.com/gofaquan/go-redis/pkg/sync/wait"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
	"time"
)

/**
 * tcp echo 服务端，测试 tcp 包下功能是否正常
 */

type Echo struct {
	activeConn sync.Map       // 保存连接数
	closing    atomic.Boolean // 是否关闭，使用自己封装的原子操作
}

func (e *Echo) Handle(ctx context.Context, conn net.Conn) {
	if e.closing.Get() {
		conn.Close()
	}

	client := &Client{
		Conn:    conn,
		Waiting: wait.Wait{},
	}
	e.activeConn.Store(client, struct{}{}) // 空结构体节省空间，只记录 client

	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('e') // 等待 conn 收到字节流信息
		if err != nil {
			if err != io.EOF {
				zap.L().Info("a connection close")
				e.activeConn.Delete(client)
			} else {
				zap.L().Warn("Echo Handle unknown error!")
			}
			// 以上逻辑只是错误不同吗，最后都会从这里 return 退出
			return
		}
		// Add 的作用是 在下方 Done 调用前不会被 client 强制关闭，写回 reMsg 在退出
		client.Waiting.Add(1)
		reMsg := []byte(msg)
		conn.Write(reMsg)
		client.Waiting.Done()
	}
}

func (e *Echo) Close() {
	zap.L().Info("echo handler is shutting down...")
	e.closing.Set(true) // 设置状态为正在关闭
	e.activeConn.Range(func(key, value any) bool {
		client := key.(*Client)
		client.Close() // 关闭连接
		return true
	})
}

type Client struct {
	Conn    net.Conn // 服务端连接
	Waiting wait.Wait
}

func (c *Client) Close() {
	c.Waiting.WaitWithTimeout(10 * time.Second) // 10s TTL,关闭前给一段时间，等待已有业务处理完毕
	c.Conn.Close()
}

func NewEchoHandler() *Echo {
	return &Echo{}
}
