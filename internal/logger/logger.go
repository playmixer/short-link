package logger

import "fmt"

type Logger struct{}

func New() *Logger {
	return &Logger{}
}

func (l *Logger) INFO(t ...any) {
	fmt.Println(t...)
}

func (l *Logger) WARN(t ...any) {
	fmt.Println(t...)
}

func (l *Logger) ERROR(t ...any) {
	fmt.Println(t...)
}

func (l *Logger) DEBUG(t ...any) {
	fmt.Println(t...)
}
