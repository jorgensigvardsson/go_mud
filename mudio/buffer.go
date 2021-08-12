package mudio

import (
	"fmt"
	"strings"
)

type buffer struct {
	builder strings.Builder
}

func (b *buffer) Println(args ...interface{}) {
	b.builder.WriteString(fmt.Sprintln(args...))
}

func (b *buffer) Printlnf(text string, args ...interface{}) {
	b.builder.WriteString(fmt.Sprintln(fmt.Sprintf(text, args...)))
}

func (b *buffer) Printf(text string, args ...interface{}) {
	b.builder.WriteString(fmt.Sprintf(text, args...))
}

func (b *buffer) ToString() string {
	return b.builder.String()
}
