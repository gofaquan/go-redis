package aof

import (
	"github.com/gofaquan/go-redis/config"
	databaseface "github.com/gofaquan/go-redis/interface/database"
	"github.com/gofaquan/go-redis/lib/logger"
	"github.com/gofaquan/go-redis/lib/utils"
	"github.com/gofaquan/go-redis/resp/connection"
	"github.com/gofaquan/go-redis/resp/parser"
	"github.com/gofaquan/go-redis/resp/reply"
	"io"
	"os"
	"strconv"
)

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

const (
	aofQueueSize = 1 << 16
)

type payload struct {
	cmdLine CmdLine
	dbIndex int
}

// AofHandler receive msgs from channel and write to AOF file
type AofHandler struct {
	db          databaseface.Database
	aofChan     chan *payload
	aofFile     *os.File
	aofFilename string
	currentDB   int
}

// NewAOFHandler creates a new aof.AofHandler
func NewAOFHandler(db databaseface.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.db = db
	handler.LoadAof()
	// 有就打开没有就创建
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payload, aofQueueSize)
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

// AddAof send command to aof goroutine through channel
// 数据落盘
func (handler *AofHandler) AddAof(dbIndex int, cmdLine CmdLine) {
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}

// handleAof listen aof channel and write into file
// 从 channel 取数据
func (handler *AofHandler) handleAof() {
	// serialized execution
	handler.currentDB = 0 //设定当前 DB
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB { // DB 发生切换
			// select db
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data) // 写入 说明 DB 发生切换
			if err != nil {
				logger.Warn(err)
				continue // skip this command
			}
			handler.currentDB = p.dbIndex
		}
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
		}
	}
}

// LoadAof read aof file
func (handler *AofHandler) LoadAof() {

	file, err := os.Open(handler.aofFilename)
	if err != nil {
		logger.Warn(err)
		return
	}
	defer file.Close()
	ch := parser.ParseStream(file)
	fakeConn := &connection.Connection{} // only used for save dbIndex
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Error("parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		ret := handler.db.Exec(fakeConn, r.Args)
		if reply.IsErrorReply(ret) {
			logger.Error("exec err", err)
		}
	}
}
