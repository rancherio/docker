package dockerlog

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	"github.com/docker/docker/daemon/logger"
	"github.com/docker/docker/pkg/jsonlog"
)

// DockerLogger is Logger implementation for default docker logging:
// JSON objects to file
type DockerLogger struct {
	buf      *bytes.Buffer
	f        *os.File // store for closing
	capacity int64
	n        int
}

// New creates new DockerLogger which writes to filename
func New(filename string) (logger.Logger, error) {
	log, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return &DockerLogger{
		f:        log,
		buf:      bytes.NewBuffer(nil),
		capacity: -1,
		n:        1,
	}, nil
}

func NewWithCap(filename string, capacity string, n int) (logger.Logger, error) {
	var multiplier int64
	capString := capacity[:len(capacity)-1]
	switch capacity[len(capacity)-1:] {
	case "K":
		multiplier = 1024
	case "k":
		multiplier = 1024
	case "M":
		multiplier = 1024 * 1024
	case "m":
		multiplier = 1024 * 1024
	case "G":
		multiplier = 1024 * 1024 * 1024
	case "g":
		multiplier = 1024 * 1024 * 1024
	default:
		multiplier = 1
		capString = capacity
	}

	capVal, err := strconv.ParseInt(capString, 10, 64)
	if err != nil {
		return nil, err
	}
	capVal *= multiplier
	log, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}
	return &DockerLogger{
		f:        log,
		buf:      bytes.NewBuffer(nil),
		capacity: capVal,
		n:        n,
	}, nil
}

// Log converts logger.Message to jsonlog.JSONLog and serializes it to file
func (l *DockerLogger) Log(msg *logger.Message) error {
	if err := (&jsonlog.JSONLog{Log: string(msg.Line) + "\n", Stream: msg.Source, Created: msg.Timestamp}).MarshalJSONBuf(l.buf); err != nil {
		return err
	}
	l.buf.WriteByte('\n')
	_, err := writeLog(l)
	return err
}

func writeLog(l *DockerLogger) (int64, error) {
	if l.capacity == -1 {
		return l.buf.WriteTo(l.f)
	}
	meta, err := l.f.Stat()
	if err != nil {
		fmt.Println(os.IsNotExist(err))
		fmt.Println(meta)
		return -1, err
	}
	if meta.Size() >= l.capacity/int64(l.n) {
		l.f.Sync()
		name := l.f.Name()
		if err := l.f.Close(); err != nil {
			fmt.Println("error while closing file")
			return -1, err
		}
		if err := rotate(name, l.n); err != nil {
			return -1, err
		}
		os.Remove(name)
		file, err := os.OpenFile(name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil {
			return -1, err
		}
		l.f = file
	}
	return l.buf.WriteTo(l.f)
}

func rotate(name string, n int) error {
	if n < 2 {
		return nil
	}
	for i := 1; i < n-1; i++ {
		oldFile := name + "-" + strconv.Itoa(i-1)
		replacingFile := name + "-" + strconv.Itoa(i)
		if err := backup(oldFile, replacingFile); err != nil {
			return err
		}
	}
	if err := backup(name+"-"+strconv.Itoa(n-2), name); err != nil {
		return err
	}
	return nil
}

func backup(old, curr string) error {
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		err := os.Remove(old)
		if err != nil {
			return err
		}
	}
	if _, err := os.Stat(curr); os.IsNotExist(err) {
		if f, err := os.Create(curr); err != nil {
			return err
		} else {
			f.Close()
		}
	}
	return os.Rename(curr, old)
}

// Close closes underlying file
func (l *DockerLogger) Close() error {
	return l.f.Close()
}

// Name returns name of this logger
func (l *DockerLogger) Name() string {
	return "Default"
}
