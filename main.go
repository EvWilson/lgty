package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

type Config struct {
	WithExtra bool
}

func setupConfig(cfg *Config) {
	flag.BoolVar(&cfg.WithExtra, "extra", false, "Show all JSON fields in log after message")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", path.Base(os.Args[0]))
		fmt.Fprintf(os.Stderr, "\nPrettify log lines according to our output.\n")
		fmt.Fprintf(os.Stderr, "Reads from stdin and writes to stdout.\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()
}

const prefix = "@cee:"

type ExpectedInput struct {
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
	Time  time.Time `json:"time"`
}

func main() {
	var cfg Config
	setupConfig(&cfg)

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
			if !cfg.WithExtra {
				fmt.Printf("%s %s %s\n", time, level, input.Msg)
			} else {
				extra, err := getExtra(after)
				if err != nil {
					Printf(ColorRed, "Error while extracting extra info: %v\n", err)
				}
				fmt.Printf("%s %s %s %s\n", time, level, input.Msg, SafeColorize(ColorGray, extra))
			}
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

// Utility to get "extra" fields in log, that aren't the primary fields we're interested in
func marshalWithSpaces(data any) (string, error) {
	compact, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	result := string(compact)
	result = strings.ReplaceAll(result, ":", ": ")
	result = strings.ReplaceAll(result, ",", ", ")
	return result, nil
}

func getExtra(jsonStr string) (string, error) {
	excluded := []string{"msg", "time", "level"}
	var full map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &full); err != nil {
		return "", err
	}
	for _, field := range excluded {
		delete(full, field)
	}
	res, err := marshalWithSpaces(full)
	if err != nil {
		return "", err
	}
	return res, nil
}
