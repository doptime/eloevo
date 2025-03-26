package tools

import (
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
)

type FileNameString struct {
	Filename string `json:"filename,omitempty" description:"Filename of output"`
	Content  string `json:"content,omitempty" description:"The content string to save"`
}

func SaveToFile(param *FileNameString) {
	if param.Filename == "" || param.Content == "" {
		return
	}
	utils.StringToFile(param.Filename, param.Content)

}

var SaveStringToFile = tool.NewTool("SaveToFile", "Save string to localfile", SaveToFile)
