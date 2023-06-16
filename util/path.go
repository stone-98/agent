package util

import (
	"os"
	"strings"
)

func GenCompleteCommand(path string, command string) string {
	p := AppendPathSeparator(path)
	return AppendCommand(p, command)
}

func AppendCommand(path string, command string) string {
	return path + command
}

func AppendPathSeparator(path string) string {
	separator := string(os.PathSeparator)
	if !strings.HasSuffix(path, separator) {
		path += separator
	}
	return path
}
