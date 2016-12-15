package server

import (
	"fmt"
	"io"
)

type StringFormatter struct{}

func (f *StringFormatter) Write(data interface{}, into io.Writer) error {
	s, ok := data.(string)
	if !ok {
		return fmt.Errorf("This formatter only supports writing string data")
	}

	_, err := into.Write([]byte(s))
	return err
}
