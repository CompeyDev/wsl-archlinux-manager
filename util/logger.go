package logger

import (
	"os"

	"github.com/gookit/color"
)

func Error(log string) {
	color.Red.Println("\r    ❎ ", log)
	os.Exit(1)
}

func Warn(log string) {
	color.Yellow.Println("\r    ❎ ", log)
}

func Info(log string) {
	color.Green.Println("\r    ✅ ", log)
}

func Progress(log string) {
	color.Bold.Println("\n✅ ", log)
}
