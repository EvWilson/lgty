package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const prefix = "@cee:"

type ExpectedInput struct {
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
	Time  time.Time `json:"time"`
}

func main() {
	scan := bufio.NewScanner(os.Stdin)
	for scan.Scan() {
		line := scan.Text()
		if after, ok := strings.CutPrefix(line, prefix); ok {
			var input ExpectedInput
			err := json.Unmarshal([]byte(after), &input)
			if err != nil {
				Printf(ColorRed, "Error while reading log message: %v\n", err)
				continue
			}
			time := SafeColorize(ColorGray, input.Time.Format("January 2, 2006 3:04 PM"))
			level, err := colorizeLogLevel(input.Level)
			if err != nil {
				Printf(ColorRed, "Error while colorizing log level: %v\n", err)
				continue
			}
			fmt.Printf("%s %s %s\n", time, level, input.Msg)
		} else {
			fmt.Println(line)
		}
	}
}

func colorizeLogLevel(level string) (string, error) {
	switch strings.ToLower(level) {
	case "debug":
		return SafeColorize(ColorBlue, "DEBU"), nil
	case "info":
		return SafeColorize(ColorGreen, "INFO"), nil
	case "warn", "warning":
		return SafeColorize(ColorYellow, "WARN"), nil
	case "error":
		return SafeColorize(ColorRed, "ERRO"), nil
	default:
		return "", fmt.Errorf("unknown log level: %s", level)
	}
}

// ANSI code color utilities
type Color string

const (
	ColorReset  Color = "\033[0m"
	ColorRed    Color = "\033[31m"
	ColorGreen  Color = "\033[32m"
	ColorYellow Color = "\033[33m"
	ColorBlue   Color = "\033[34m"
	ColorGray   Color = "\033[90m"
)

func Printf(color Color, format string, args ...any) {
	fmt.Printf(SafeColorize(color, format), args...)
}

func IsTerminal() bool {
	fileInfo, _ := os.Stdout.Stat()
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func SafeColorize(color Color, text string) string {
	if IsTerminal() {
		return string(color) + text + string(ColorReset)
	}
	return text
}
