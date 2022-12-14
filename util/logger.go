package logger

import (
	"os"

	"github.com/gookit/color"
)

func Error(log string) {
	color.Red.Println("\r    ā ", log)
	os.Exit(1)
}

func Warn(log string) {
	color.Yellow.Println("\r    ā ", log)
}

func Info(log string) {
	color.Green.Println("\r    ā ", log)
}

func Progress(log string) {
	color.Bold.Println("\nā ", log)
}
