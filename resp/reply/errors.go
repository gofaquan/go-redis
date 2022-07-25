package reply

// UnknownErrReply represents UnknownErr
type UnknownErrReply struct{}

var unknownErrBytes = []byte("-未知错误！\r\n")

// ToBytes marshals redis.Reply
func (r *UnknownErrReply) ToBytes() []byte {
	return unknownErrBytes
}

func (r *UnknownErrReply) Error() string {
	return "未知错误！"
}

// ArgNumErrReply represents wrong number of arguments for command
type ArgNumErrReply struct {
	Cmd string
}

// ToBytes marshals redis.Reply
func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-指令 '" + r.Cmd + "' 参数个数错误！\r\n")
}

func (r *ArgNumErrReply) Error() string {
	return "指令 '" + r.Cmd + "' 参数个数错误！"
}

// MakeArgNumErrReply represents wrong number of arguments for command
func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{
		Cmd: cmd,
	}
}

// SyntaxErrReply represents meeting unexpected arguments
type SyntaxErrReply struct{}

var syntaxErrBytes = []byte("-语法错误！\r\n")
var theSyntaxErrReply = &SyntaxErrReply{}

// MakeSyntaxErrReply creates syntax error
func MakeSyntaxErrReply() *SyntaxErrReply {
	return theSyntaxErrReply
}

// ToBytes marshals redis.Reply
func (r *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

func (r *SyntaxErrReply) Error() string {
	return "语法错误！"
}

// WrongTypeErrReply represents operation against a key holding the wrong kind of value
// 类型错误，对持有错误类型值的键的操作！
type WrongTypeErrReply struct{}

var wrongTypeErrBytes = []byte("-对持有错误类型值的键的操作！\r\n")

// ToBytes marshals redis.Reply
func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

func (r *WrongTypeErrReply) Error() string {
	return "对持有错误类型值的键的操作！"
}

// ProtocolErr 协议错误

// ProtocolErrReply represents meeting unexpected byte during parse requests
type ProtocolErrReply struct {
	Msg string
}

// ToBytes marshals redis.Reply
func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-协议错误: '" + r.Msg + "'\r\n")
}

func (r *ProtocolErrReply) Error() string {
	return "协议错误: '" + r.Msg
}
