package handler

/*
 * A tcp.RespHandler implements redis protocol
 */

import (
	"context"
	"github.com/gofaquan/go-redis/database"
	databaseface "github.com/gofaquan/go-redis/interface/database"
	"github.com/gofaquan/go-redis/lib/logger"
	"github.com/gofaquan/go-redis/lib/sync/atomic"
	"github.com/gofaquan/go-redis/resp/connection"
	"github.com/gofaquan/go-redis/resp/parser"
	"github.com/gofaquan/go-redis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

// RespHandler implements tcp.Handler and serves as a redis handler
type RespHandler struct {
	activeConn sync.Map              // *client -> placeholder 所有客户端连接
	db         databaseface.Database // 业务层
	closing    atomic.Boolean        // refusing new client and new request
}

// MakeHandler creates a RespHandler instance
func MakeHandler() *RespHandler {
	var db databaseface.Database
	db = database.NewEchoDatabase()
	return &RespHandler{
		db: db,
	}
}

func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

// Handle receives and executes redis commands
func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		// closing handler refuse new connection
		_ = conn.Close()
	}

	client := connection.NewConn(conn)
	h.activeConn.Store(client, 1) // 存储客户端

	ch := parser.ParseStream(conn)
	for payload := range ch {
		//异常处理
		if payload.Err != nil {
			if payload.Err == io.EOF || //挥手，断开连接
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// connection closed
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// protocol err
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		//正常逻辑
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		result := h.db.Exec(client, r.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

// Close stops handler
func (h *RespHandler) Close() error {
	logger.Info("handler shutting down...")
	h.closing.Set(true)
	// TODO: concurrent wait
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	h.db.Close()
	return nil
}
