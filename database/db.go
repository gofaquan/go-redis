// Package database is a memory database with redis compatible interface
package database

import (
	"github.com/gofaquan/go-redis/datastruct/dict"
	"github.com/gofaquan/go-redis/interface/database"
	"github.com/gofaquan/go-redis/interface/resp"
	"github.com/gofaquan/go-redis/resp/reply"
	"strings"
)

// DB stores data and execute user's commands
type DB struct {
	index int
	// key -> DataEntity
	data   dict.Dict
	addAof func(CmdLine)
}

// ExecFunc is interface for command executor
// args don't include cmd line
// 指令的实现通过这类 func
type ExecFunc func(db *DB, args [][]byte) resp.Reply

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

// makeDB create DB instance
func makeDB() *DB {
	db := &DB{
		data: dict.MakeSyncDict(),
	}
	return db
}

// Exec executes command within one database
// Connection 输入指令的连接，cmdLine 对应的指令
func (db *DB) Exec(c resp.Connection, cmdLine [][]byte) resp.Reply {
	// ping set setnx get 等
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := cmdTable[cmdName] // 得到 执行函数 与 对应参数数量
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "'")
	}
	if !validateArity(cmd.arity, cmdLine) {
		//  比如 set k
		return reply.MakeArgNumErrReply(cmdName) // 报参数数量错
	}
	fun := cmd.executor
	return fun(db, cmdLine[1:]) // set k v 只要 k v
}

// 校验合法性
// set k v
// exists k1 k2 ...  至少两个参数，这样就 >= -2 代表 参数无上限
func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	if arity >= 0 {
		return argNum == arity
	}
	return argNum >= -arity
}

/* ---- data Access ----- */

// GetEntity returns DataEntity bind to given key
func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := db.data.Get(key) // 得到原始的值
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity) // 类型转换
	return entity, true
}

// PutEntity a DataEntity into DB
func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

// PutIfExists edit an existing DataEntity
func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

// PutIfAbsent insert an DataEntity only if the key not exists
func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

// Remove the given key from db
func (db *DB) Remove(key string) {
	db.data.Remove(key)
}

// Removes the given keys from db
func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.data.Get(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

// Flush clean database
func (db *DB) Flush() {
	db.data.Clear()

}
