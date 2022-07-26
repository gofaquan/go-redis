package tcp

import (
	"context"
	"net"
)

// Handler 接口抽象 Handle 和 Close 方法
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() // 用于关闭连接
}
