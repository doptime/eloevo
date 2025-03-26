package tools

import (
	"github.com/doptime/eloevo/memory"
	"github.com/doptime/eloevo/tool"
)

type FileNames struct {
	Category string   `json:"category,omitempty" description:"Category of files"`
	Files    []string `json:"filenames,omitempty" description:"Filenames"`
}

func WithFiles(param *FileNames) {
	if len(param.Files) == 0 || param.Category == "" {
		return
	}
	memory.SharedMemory[param.Category] = param.Files

}

var StoreFilenamesToMemory = tool.NewTool("PickFileNames", "Save the picked out files", WithFiles)
