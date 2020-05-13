package logger

import "testing"

func TestInfo(t *testing.T) {
	Info("hello", "郭斌")
	Debug("hello", "郭斌")
	Error("hello", "郭斌")
	Warn("hello", "郭斌")
}
