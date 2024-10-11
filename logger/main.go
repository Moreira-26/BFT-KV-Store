package logger

import (
	"fmt"
	"log"
	"os"
)

const (
	_BLACK        = "0;30"
	_DARK_GRAY    = "1;30"
	_RED          = "0;31"
	_LIGHT_RED    = "1;31"
	_GREEN        = "0;32"
	_LIGHT_GREEN  = "1;32"
	_BROWN_ORANGE = "0;33"
	_YELLOW       = "1;33"
	_BLUE         = "0;34"
	_LIGHT_BLUE   = "1;34"
	_PURPLE       = "0;35"
	_LIGHT_PURPLE = "1;35"
	_CYAN         = "0;36"
	_LIGHT_CYAN   = "1;36"
	_LIGHT_GRAY   = "0;37"
	_WHITE        = "1;37"
	_NOCOLOR      = "0"
)

func withColor(color string, word string) string {
	return fmt.Sprintf("\033[%sm%s\033[%sm", color, word, _NOCOLOR)
}

func print(prefix string, args ...any) {
	text := make([]byte, 0)
	for _, v := range args {
		text = fmt.Append(text, fmt.Sprint(v))
	}
	log.Println(prefix, string(text))
}

func Info(args ...any) {
	print(withColor(_LIGHT_BLUE, "[INFO]"), args...)
}

func Error(args ...any) {
	print(withColor(_RED, "[ERROR]"), args...)
}

func Log(args ...any) {
	print(withColor(_BLUE, "[LOG]"), args...)
}

func Debug(args ...any) {
	print(withColor(_GREEN, "[DEBUG]"), args...)
}

func Alert(args ...any) {
	print(withColor(_YELLOW, "[ALERT]"), args...)
}

func Fatal(args ...any) {
	print(withColor(_BLACK, "[FATAL]"), args...)
	os.Exit(1)
}
