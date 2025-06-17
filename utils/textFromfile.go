package utils

import (
	"fmt"
	"os"
	"strings"
)

func TextFromFile(filename string, Content ...*string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return ""
	}
	text := string(content)
	text = strings.TrimSpace(text)
	if len(Content) > 0 && Content[0] != nil && text != "" {
		*Content[0] = text
	}
	return text
}
