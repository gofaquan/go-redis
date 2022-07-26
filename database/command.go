package database

import (
	"strings"
)

var cmdTable = make(map[string]*command) // 记录系统里所有的指令，每个指令对应一个 command 结构体

type command struct {
	executor ExecFunc
	// 参数数量
	arity int // allow number of args, arity < 0 means len(args) >= -arity
}

// RegisterCommand registers a new command
// arity means allowed number of cmdArgs, arity < 0 means len(args) >= -arity.
// for example: the arity of `get` is 2, `mget` is -2
// 注册指令的实现
func RegisterCommand(name string, executor ExecFunc, arity int) {
	name = strings.ToLower(name) // 统一转小写
	cmdTable[name] = &command{
		executor: executor,
		arity:    arity,
	}
}
