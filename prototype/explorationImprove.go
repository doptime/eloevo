package prototype

type ExporationImprovement struct {
	Id                     string   `description:"required, ID"`
	ImprovementSuggestions []string `description:"required, 改进方案: 提出现有方案的不足; 提出整合第一性原理的创新性解决方案；提出实践可能存在的问题；指出需要如何才能改善动态调整机制。指出现有的方案应对环境的困难。指出现有的理论存在的问题和深入讨论改进的办法。"`
}
