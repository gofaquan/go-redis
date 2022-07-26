package parser

// 解析器
import (
	"bufio"
	"errors"
	"github.com/gofaquan/go-redis/interface/resp"
	"github.com/gofaquan/go-redis/lib/logger"
	"github.com/gofaquan/go-redis/resp/reply"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

// Payload stores redis.Reply or error
type Payload struct {
	Data resp.Reply
	Err  error
}

// ParseStream reads data from io.Reader and send payloads through channel
// 解析字节流，返回 channel 异步进行
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

type readState struct {
	readingMultiLine  bool
	expectedArgsCount int //应该解析参数
	msgType           byte
	args              [][]byte // 二维数组，已经解析的参数
	bulkLen           int64
}

func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() { //防止 panic
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for {
		// read line
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr { // encounter io err, stop read
				ch <- &Payload{
					Err: err,
				}
				close(ch)
				return
			}
			// protocol err, reset read state
			ch <- &Payload{
				Err: err,
			}
			state = readState{}
			continue
		}

		// parse line
		if !state.readingMultiLine {
			// receive new response
			if msg[0] == '*' {
				// multi bulk reply  *3\r\n
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{} // reset state
					continue
				}
				if state.expectedArgsCount == 0 { //往 channel 输入 空 reply
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
					}
					state = readState{} // reset state
					continue
				}
			} else if msg[0] == '$' { // bulk reply   $4\r\nPING\r\n
				err = parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{} // reset state
					continue
				}
				if state.bulkLen == -1 { // null bulk reply  $-1\r\n
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{} // reset state
					continue
				}
			} else {
				// single line reply
				result, err := parseSingleLineReply(msg) // +OK\r\n  -err\r\n
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{} // reset state
				continue
			}
		} else {
			// receive following bulk reply
			err = readBody(msg, &state)
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{} // reset state
				continue
			}
			// if sending finished
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
			}
		}
	}
}

func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error
	if state.bulkLen == 0 { // read normal line
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else { // read bulk line (binary safe)
		msg = make([]byte, state.bulkLen+2)
		_, err = io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 ||
			msg[len(msg)-2] != '\r' ||
			msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, false, nil
}

// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	// *300\r\n => 300
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 { // *0\r\n => 0 无意义
		state.expectedArgsCount = 0 // 0个参数
		return nil
	} else if expectedLine > 0 {
		// first line of multi bulk reply
		state.msgType = msg[0]                       // * 代表是数组 , $ 字符串
		state.readingMultiLine = true                // 读取多行
		state.expectedArgsCount = int(expectedLine)  // 3个参数
		state.args = make([][]byte, 0, expectedLine) // 初始化已读取
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// $4\r\nPING\r\n
func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // null bulk
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// +OK\r\n  -err\r\n
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n") // 取出 \r\n 前面的
	var result resp.Reply
	switch msg[0] {
	case '+': // status reply
		result = reply.MakeStatusReply(str[1:])
	case '-': // err reply
		result = reply.MakeErrReply(str[1:])
	case ':': // int reply
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// read the non-first lines of multi bulk reply or bulk reply
// $3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// PONG\r\n
func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' { // $3
		// bulk reply
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // null bulk in multi bulks, $0\r\n
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
