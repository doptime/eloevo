package utils

import (
	"fmt"
	"strconv"
	"strings"
)

func RemoveLineNumber(s string) string {
	newLinesNoLineNum := strings.Split(s, "\n")
	preNum := -1
	for i, line := range newLinesNoLineNum {
		if items := strings.SplitN(line, ":", 1); len(items) > 1 {
			if num, err := strconv.Atoi(items[0]); err == nil {
				if preNum == -1 || preNum == num-1 {
					newLinesNoLineNum[i] = items[1]
				}
				preNum = num
			}
		}
	}
	return strings.Join(newLinesNoLineNum, "\n")
}
func LineNumberedFileContent(s string, lineNumberStart int) string {
	s = NormalizeFileContent(s)
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = fmt.Sprintf("%d: %s", i+lineNumberStart, lines[i])
	}
	return strings.Join(lines, "\n")
}
func LineNumberedMap(s string, lineNumberStart int) map[int]string {
	s = NormalizeFileContent(s)
	lines := strings.Split(s, "\n")
	linesMap := map[int]string{}
	for i, t := range lines {
		linesMap[i+lineNumberStart] = t
	}
	return linesMap
}
