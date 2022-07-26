package database

import (
	"github.com/gofaquan/go-redis/interface/resp"
)

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

// Database is the interface for redis style storage engine
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	AfterClientClose(c resp.Connection) // 连接关闭后的处理
	Close()
}

// DataEntity stores data bound to a key, including a string, list, hash, set and so on
type DataEntity struct { //指代 redis 数据结构
	Data interface{}
}
