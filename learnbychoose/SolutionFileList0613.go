package learnbychoose

import (
	"fmt"
	"slices"
	"strings"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/scrum"
	"github.com/doptime/redisdb"
	"github.com/samber/lo"
)

//不使用超边节点

type SolutionFileNode struct {
	Filename          string `description:"string, Ascii filename of current node。using bullet name to denodes node's modualized intention. extension name such as .md ... is needed"`
	BulletDescription string `description:"string, Required when create. BulletDescription 是文件内容的摘要。描述和文件的模块化的意图。规定实现的细节."`
	Delete            bool   `msgpack:"-" description:"bool, Whether this node is deleted. If true, the node will be removed from the system."`
	FileContent       string `description:"string, The contents of the file stored on disk"`

	WhatTodoInFollowingIter string `description:"What is key focus to do in following iter, string"`

	NextIterFeedback string `description:"string, 当前开发迭代的反馈和可操作性洞察，例如已识别的缺失依赖项、约束冲突或潜在的冗余实现。此信息将传递给 EndGoalDrivenPlanner，以用于后续迭代中的战略考量和任务分配。"`

	BackgroundNodesToPass []string `description:"background solution nodes Filename should pass to iterate this file in next iteration, array of string"`

	Fullfillment int64 `description:"integer, Fullfillment is the degree of completion of the solution file. 0 means not started, 100 means completed."`

	AllItems    map[string]*SolutionFileNode               `msgpack:"-" description:"-"`
	Backlogs    []*scrum.Backlog                           `msgpack:"-" description:"-"`
	ProductGoal string                                     `msgpack:"-" description:"-"`
	HashKey     redisdb.HashKey[string, *SolutionFileNode] `msgpack:"-" description:"-"`
	ThisAgent   *agent.Agent                               `msgpack:"-" description:"-"`
}

func (a *SolutionFileNode) FileSize() string {
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

type SolutionFileNodeList []*SolutionFileNode

func (a SolutionFileNodeList) Uniq() SolutionFileNodeList { return lo.Uniq(a) }
func (a SolutionFileNodeList) FullView() string {
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
func (a SolutionFileNodeList) View(FullViewPaths ...string) string {
	var sb strings.Builder
	for _, v := range a {
		numlayter := len(strings.Split(v.Filename, "/")) - 1
		indence := strings.Repeat("\t", numlayter)
		NextIterFeedback := lo.Ternary(v.NextIterFeedback != "", " NextIterFeedback: "+v.NextIterFeedback, "")
		fileContent := lo.Ternary(slices.Contains(FullViewPaths, v.Filename), "\nFileContent: "+v.FileContent, "")

		s := fmt.Sprint(indence, "\n Pathname", v.Filename, v.FileSize(), "\nBulletDescription: ", v.BulletDescription, fileContent, NextIterFeedback, "\n")
		sb.WriteString(s)
	}
	return sb.String()
}
func (a SolutionFileNodeList) PathnameSorted() SolutionFileNodeList {
	slices.SortFunc(a, func(a, b *SolutionFileNode) int {
		return strings.Compare(a.Filename, b.Filename)
	})
	return a
}
func (a SolutionFileNodeList) FullfillmentSorted() SolutionFileNodeList {
	slices.SortFunc(a, func(a, b *SolutionFileNode) int {
		return int(a.Fullfillment - b.Fullfillment)
	})
	return a
}
