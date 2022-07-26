package database

import (
	"github.com/gofaquan/go-redis/interface/resp"
	"github.com/gofaquan/go-redis/lib/wildcard"
	"github.com/gofaquan/go-redis/resp/reply"
)

// execDel removes a key from db
// 删除key，支持多个
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v) // v 是 []byte，要转字符串
	}

	deleted := db.Removes(keys...)
	return reply.MakeIntReply(int64(deleted)) // 返回结果是删除 key 的个数
}

// execExists checks if a is existed in db
// key 是否存在
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// execFlushDB removes all data in current db
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	return &reply.OkReply{}
}

// execType returns the type of entity, including: string, list, hash, set and zset
// 返回类型
// type k1
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])              // 取到 k1
	entity, exists := db.GetEntity(key) // 取 k1 对应的值
	if !exists {
		return reply.MakeStatusReply("none")
	}
	// 类型断言
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
	}
	return &reply.UnknownErrReply{}
}

// execRename a key
// rename k1 k2
func execRename(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}
	src := string(args[0])  // k1
	dest := string(args[1]) // k2

	entity, ok := db.GetEntity(src) // 取出值
	if !ok {
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity) // put
	db.Remove(src)             // remove
	return &reply.OkReply{}
}

// execRenameNx a key, only if the new key does not exist
// Renamenx 命令用于在新的 key 不存在时修改 key 的名称
func execRenameNx(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])

	_, ok := db.GetEntity(dest)
	if ok {
		return reply.MakeIntReply(0) // 存在了就不改
	}
	// 原 key 不存在
	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeErrReply("no such key")
	}
	// 原 key 存在，新 key 不存在 这种情况就能改
	db.Removes(src, dest) // clean src and dest with their ttl
	db.PutEntity(dest, entity)
	return reply.MakeIntReply(1)
}

// execKeys returns all keys matching the given pattern
// 列出所有 key
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0)
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

func init() {
	RegisterCommand("Del", execDel, -2)
	RegisterCommand("Exists", execExists, -2)
	RegisterCommand("Keys", execKeys, 2)
	RegisterCommand("FlushDB", execFlushDB, -1)
	RegisterCommand("Type", execType, 2)
	RegisterCommand("Rename", execRename, 3)
	RegisterCommand("RenameNx", execRenameNx, 3)
}
