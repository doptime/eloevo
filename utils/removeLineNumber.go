package utils

import (
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
