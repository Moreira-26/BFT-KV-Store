package logger

import (
	"fmt"
	"log"
	"os"
)

const (
	BLACK        = "0;30"
	DARK_GRAY    = "1;30"
	RED          = "0;31"
	LIGHT_RED    = "1;31"
	GREEN        = "0;32"
	LIGHT_GREEN  = "1;32"
	BROWN_ORANGE = "0;33"
	YELLOW       = "1;33"
	BLUE         = "0;34"
	LIGHT_BLUE   = "1;34"
	PURPLE       = "0;35"
	LIGHT_PURPLE = "1;35"
	CYAN         = "0;36"
	LIGHT_CYAN   = "1;36"
	LIGHT_GRAY   = "0;37"
	WHITE        = "1;37"
	NOCOLOR      = "0"
)

func withColor(color string, word string) string {
	return fmt.Sprintf("\033[%sm%s\033[%sm", color, word, NOCOLOR)
}

func print(prefix string, args ...string) {
	text := make([]byte, 0)
	for _, v := range args {
		text = fmt.Append(text, v)
	}
	log.Println(prefix, string(text))
}

func Info(args ...string) {
	print(withColor(LIGHT_BLUE, "[INFO]"), args...)
}

func Error(args ...string) {
	print(withColor(RED, "[ERROR]"), args...)
}

func Log(args ...string) {
	print(withColor(BLUE, "[LOG]"), args...)
}

func Debug(args ...string) {
	print(withColor(GREEN, "[DEBUG]"), args...)
}

func Alert(args ...string) {
	print(withColor(YELLOW, "[ALERT]"), args...)
}

func Fatal(args ...string) {
	print(withColor(BLACK, "[FATAL]"), args...)
	os.Exit(1)
}
