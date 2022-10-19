package logrus

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

type DefaultFormatter struct {
	GetTraceId func() string
}

func (d *DefaultFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	levelText, err := entry.Level.MarshalText()
	if err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	b.WriteString(entry.Time.Format("2006-01-02 15:04:05.000"))
	b.WriteString("|")
	b.WriteString(strings.ToUpper(string(levelText)))
	b.WriteString("|")
	if entry.HasCaller() {
		b.WriteString(entry.Caller.File)
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(entry.Caller.Line))
		b.WriteString("|")
	}
	if d.GetTraceId != nil {
		b.WriteString(d.GetTraceId())
		b.WriteString("|")
	}
	b.WriteString(entry.Message)
	b.WriteString("\n")
	return b.Bytes(), nil
}
