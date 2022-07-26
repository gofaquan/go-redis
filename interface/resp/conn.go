package resp

// Connection represents a connection with redis client
type Connection interface {
	Write([]byte) error // 返回错误
	GetDBIndex() int    // redis 默认 16 个 DB,下标表示
	SelectDB(int)       // 选择 DB
}
