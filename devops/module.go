package devops

import (
	"github.com/spf13/afero"
)

// 模块结构体（简化版）
type Module struct {
	Name     string // 模块名称
	BranchId string // 模块Id

	EloScore  int64   // 模块评分
	Milestone float64 // 1: file/code constructed, 2:file/code tested, 3:hardware constructed, 4:hardware tested, 5:Income generated

	ProblemToSolve []string // 模块所属问题域
	DesignIdeas    []string
	OuterModuleIds []string
	InnerModuleIds []string
}

func (m *Module) SourceCodes() (fileList []string) {
	fs := afero.NewOsFs()
	files, _ := afero.ReadDir(fs, "./"+m.BranchId)
	for _, file := range files {
		content, _ := afero.ReadFile(fs, "./"+m.BranchId+"/"+file.Name())
		fileList = append(fileList, "file-name:\n"+file.Name()+"\ncontent:\n"+string(content))
	}
	return fileList
}

func (m Module) Id() string {
	return m.BranchId
}

func (m Module) Rating(delta int) int {
	return int(m.EloScore) + delta
}
