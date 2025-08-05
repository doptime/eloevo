package learnbychoose

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/scrum"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

//不使用超边节点

type FileRefine struct {
	Filename          string `description:"string, Ascii filename of current node。using bullet name to denodes node's modualized intention. extension name such as .md ... is needed"`
	BulletDescription string `description:"string, Required when create. BulletDescription 是文件内容的摘要。描述和文件的模块化的意图。规定实现的细节."`
	Delete            bool   `msgpack:"-" description:"bool, Whether this node is deleted. If true, the node will be removed from the system."`
	FileContent       string `description:"string, The contents of the file stored on disk"`

	AllItems    map[string]*FileRefine               `msgpack:"-" description:"-"`
	Backlogs    []*scrum.Backlog                     `msgpack:"-" description:"-"`
	ProductGoal string                               `msgpack:"-" description:"-"`
	HashKey     redisdb.HashKey[string, *FileRefine] `msgpack:"-" description:"-"`
	ThisAgent   *agent.Agent                         `msgpack:"-" description:"-"`
}

func (a *FileRefine) FileSize() string {
	if a == nil || a.FileContent == "" {
		return " (size: 0 B)"
	}
	size := len(a.FileContent)
	if size > 1024*1024 {
		return fmt.Sprintf(" (size: %.2f MB)", float64(size)/1024/1024)
	} else if size > 1024 {
		return fmt.Sprintf(" (size: %.2f KB)", float64(size/1024))
	} else {
		return fmt.Sprintf(" (size: %d B)", size)
	}
}

type FileRefineList []*FileRefine

func (a FileRefineList) Uniq() FileRefineList { return lo.Uniq(a) }
func (a FileRefineList) FullView() string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Filename, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		//description := "\nBulletDescription: " + v.BulletDescription

		s := fmt.Sprint(indence, "\n Pathname", v.Filename, v.FileSize(), "\nFileContent: ", v.FileContent, "\n\n\n\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a FileRefineList) View(FullViewPaths ...string) string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Filename, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		fileContent := lo.Ternary(slices.Contains(FullViewPaths, v.Filename), "\nFileContent: "+v.FileContent, "")

		s := fmt.Sprint(indence, "\n Pathname", v.Filename, v.FileSize(), "\nBulletDescription: ", v.BulletDescription, fileContent, "\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a FileRefineList) PathnameSorted() FileRefineList {
	slices.SortFunc(a, func(a, b *FileRefine) int {
		return strings.Compare(a.Filename, b.Filename)
	})
	return a
}

func LoadExtraPathToMapFileRefineMap(RootPath, ExtraPath string, solution map[string]*FileRefine) {
	extraPath := filepath.Join(RootPath, ExtraPath)
	ExtraPathFiles, _ := os.ReadDir(extraPath)
	for _, file := range ExtraPathFiles {
		fn := filepath.Join(extraPath, file.Name())
		filename := ExtraPath + "/" + file.Name()
		//hidden file skip
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}
		FileContent := utils.TextFromFile(fn)
		if strings.Contains(FileContent, "\x00") {
			continue // skip file with null character
		}
		filename = strings.TrimPrefix(filename, "./")
		solution[filename] = &FileRefine{
			Filename:    filename,
			FileContent: FileContent,
		}
	}
}
func (node *FileRefine) SaveContentToPath(RootPath string) {
	//save to root path
	filename := filepath.Join(RootPath, node.Filename)
	err := os.WriteFile(filename, []byte(node.FileContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file %s: %v\n", filename, err)
	}
}
