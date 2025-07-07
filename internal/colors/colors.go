package colors

import (
	"os"
)

const (
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Bold    = "\033[1m"
	DimColor = "\033[2m"
)

func isColorSupported() bool {
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}
	
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	
	return true
}

func colorize(color, text string) string {
	if !isColorSupported() {
		return text
	}
	return color + text + Reset
}

func Success(text string) string {
	return colorize(Green, text)
}

func Warning(text string) string {
	return colorize(Yellow, text)
}

func Error(text string) string {
	return colorize(Red, text)
}

func Info(text string) string {
	return colorize(Blue, text)
}

func Prompt(text string) string {
	return colorize(Cyan+Bold, text)
}

func Group(text string) string {
	return colorize(Magenta+Bold, text)
}

func Software(text string) string {
	return colorize(Bold, text)
}

func Dim(text string) string {
	return colorize(DimColor, text)
}