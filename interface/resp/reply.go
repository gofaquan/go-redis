package resp

// Reply is the interface of redis serialization protocol message
// RESP 格式接口
type Reply interface {
	ToBytes() []byte
}
