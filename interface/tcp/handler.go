package tcp

import (
	"context"
	"net"
)

// Handler 抽象业务逻辑
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
