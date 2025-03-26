package tools

type ModuleName struct {
	PreviousPath string `json:"previouspath" description:"previous path of the file"`
	Modulename   string `json:"modulename" description:"module name the file belongs to. Represented as path or Multi-level path"`
	Filename     string `json:"filename" description:"name of file. file should describe the content of the file, use printable characters allowed by os, no more than 60 characters"`
}
