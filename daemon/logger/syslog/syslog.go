package syslog

import (
	"log/syslog"
	"os"
	"path"

	"github.com/docker/docker/daemon/logger"
)

type Syslog struct {
	writer	*syslog.Writer
	tag	string	
}

func New(tag string) (logger.Logger, error) {
	log, err := syslog.New(syslog.LOG_USER, path.Base(os.Args[0]))
	if err != nil {
		return nil, err
	}
	return &Syslog{
			writer: log,
			tag: tag,
	}, nil
}

func (s *Syslog) Log(msg *logger.Message) error {
	if msg.Source == "stderr" {
		s.writer.Err(s.tag + " " + string(msg.Line))
	} else { 
		s.writer.Info(s.tag + " " + string(msg.Line))	
	}
	return nil
}

func (s *Syslog) Close() error {
	return s.writer.Close()
}

func (s *Syslog) Name() string {
	return "Syslog"
}
