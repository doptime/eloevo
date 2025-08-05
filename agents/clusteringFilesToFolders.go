package agents

import (
	"context"
	"log"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
)

type ModuleName struct {
	PreviousPath string `json:"previouspath" description:"previous path of the file"`
	Modulename   string `json:"modulename" description:"module name the file belongs to. Represented as path or Multi-level path"`
	Filename     string `json:"filename" description:"name of file. file should describe the content of the file, use printable characters allowed by os, no more than 60 characters"`
}

func toolClustering(param *ModuleName) {
	if _, ok := clustered[param.PreviousPath]; !ok {
		return
	}
	delete(clustered, param.PreviousPath)
	Moduleame := strings.TrimRight(param.Modulename, "/")
	Moduleame = strings.TrimLeft(Moduleame, "/")
	newPath := Moduleame + "/" + param.Filename
	clustered[newPath] = clustered[param.PreviousPath]

	keyProjectsClustered.HSet(newPath, clustered[param.PreviousPath])
	keyProjectsClustered.HDel(param.PreviousPath)
}

var AgentClustering = agent.NewAgent(template.Must(template.New("AgentClustering").Parse(`
## 这是现有的系统文件的目录结构：
{{range .Directories}}
{{.}}
{{end}}

## 这是其中一个文件：
{{.Files}}
EOF

## 任务说明
请为提供的文件，创建聚类/分类描述。目标文件的聚类描述包括路径和文件名称两个部分。
modulename是尝试对所有文件进行多级模块化。意在构建最优的最合理的模块化结构。请确保系统的模块结构能合理地实现文件按模块分类的意图。
文件名称部分应该尝试对文件内容进行有效描述，但不超过60个字。

## 返回格式说明 annotation
name = "FileReClustering"
description = "cluster current file, to new module and new name, 要求采用toml格式"

[parameter]
# previous path of the file
PreviousPath = "previous/path/to/file"
# module name the file belongs to. Represented as path or multi-level path
Modulename = "module/name/path"
# name of file. file should describe the content of the file, use printable characters allowed by OS, no more than 60 characters
Filename = "filename.ext"


`))).WithToolcallParser(func(resp openai.ChatCompletionResponse) (toolCalls []*agent.FunctionCall) {
	messege := resp.Choices[0].Message.Content
	ind := strings.Index(messege, "[parameter]")
	if ind == -1 {
		return nil
	}
	parameter := messege[ind:]
	var toolCall ModuleName

	// 解析 TOML 文件
	if _, err := toml.DecodeFile(parameter, &toolCall); err != nil {
		log.Fatalf("无法解析 TOML 文件: %v", err)
	}
	toolClustering(&toolCall)
	return nil
}).WithModels(models.EloModels.SelectOne("roundrobin"))

var mutex = &sync.Mutex{}

type FileData struct {
	Module  string
	Content string
}

func (f *FileData) String() string {
	return "\n\nModule: " + f.Module + "\nContent: \n" + f.Content + "\nEOF\n"
}
func Clustering() {
	mutex.Lock()
	if len(clustered) == 0 {
		LoadEvoTable()
	}
	mutex.Unlock()

	dirs := lo.Keys(clustered)
	//take random 1000 files from the list
	dirs = lo.Shuffle(dirs)
	dirs = dirs[:300]
	file := dirs[rand.Intn(len(dirs))]
	memory := map[string]interface{}{
		"Directories": dirs,
		"Files":       &FileData{Module: file, Content: clustered[file]},
	}
	AgentClustering.Call(context.Background(), memory)
}

var clustered map[string]string

// func LoadEvoTable() {
// 	var keyProjectsAll = redisdb.HashKey[string, string](redisdb.Opt.Key("NicheMarketOpportunityAll"))
// 	OptunityDescription, _ = keyProjectsAll.HGetAll()
// 	Elos, _ = keyElos.HGetAll()
// 	for k := range OptunityDescription {
// 		if _, ok := Elos[k]; !ok {
// 			Elos[k] = &elo.Elo{Id: k, Rating: 1000, Matches: nil}
// 		}
// 	}
// 	keyElos.HMSet(Elos)
// }

var keyProjectsClustered = redisdb.NewHashKey[string, string](redisdb.Opt.Key("MarketOpportunityClustered"))

func LoadEvoTable() {
	var keyProjectsAll = redisdb.NewHashKey[string, string](redisdb.Opt.Key("NicheMarketOpportunityAll"))
	OptunityDescriptionOriginal, _ := keyProjectsAll.HGetAll()
	var keyProjectsClustered = redisdb.NewHashKey[string, string](redisdb.Opt.Key("MarketOpportunityClustered"))
	clustered, _ = keyProjectsClustered.HGetAll()
	values := lo.Values(clustered)
	ModifiedN := 0
	for k, v := range OptunityDescriptionOriginal {
		if !slices.Contains(values, v) {
			clustered[k] = v
			ModifiedN++
		}
	}
	if ModifiedN > 0 {
		keyProjectsClustered.HMSet(clustered)
	}

}
