package prototype

// "相关要素": "问题背景、要素分析、理论原则、效用函数、实际限制、改进建议、动态反馈机制。",
// "实际局限": "现有范式可能无法充分应对高度动态和不确定的环境。",
// "改进方案": "提出动态分层模块化认知框架，整合第一性原理与实践反馈，建立实时动态调整机制。",
// "效用函数": "评估方案的灵活性、适应性、效率和可扩展性。",
// "理论局限": "依赖明确的模块边界和独立性，可能在实际中遇到协调问题。"

// Description string `description:"框架的简要描述 (例如：问题背景 + 要素分析 的文本摘要)"`

// RelatedElements        []string `description:"相关要素: 问题背景、要素分析、理论原则、效用函数、实际限制、改进建议、动态反馈机制。"`
// ActualLimitations      []string `description:"实际局限: 现有范式可能无法充分应对高度动态和不确定的环境。"`
// ImprovementSuggestions []string `description:"改进方案: 提出动态分层模块化认知框架，整合第一性原理与实践反馈，建立实时动态调整机制。"`
// UtilFunctions          []string `description:"效用函数: 评估方案的灵活性、适应性、效率和可扩展性。"`
// TheoreticalPrinciples  []string `description:"理论局限: 依赖明确的模块边界和独立性，可能在实际中遇到协调问题。"`
type ExporationFramwork struct {
	Id string `description:"required, ID"`
	//RelatedElements        []string `description:"required, 相关要素: 问题背景、要素分析、理论原则、效用函数、实际限制、改进建议、动态反馈机制。"`
	Background                        string `description:"required, Basic background of the problem."`
	ElementsAnalysis                  string `description:"required, Elements analysis of the problem."`
	TheoraticalFirstPrincipleAnalysis string `description:"required, First principles analysis of the problem."`
	//ImprovementInnovations            []string `description:"required, 改进方案: 提出动态分层模块化认知框架，整合第一性原理与实践反馈，建立实时动态调整机制。"`
	UtilityFunction string `description:"required, UtilityFunction A quantitative evaluation function"`
}

func (u *ExporationFramwork) GetId() string {
	return u.Id
}

// func (u *ExporationFramwork) ScoreAccessor(delta ...int) int {
// 	return mixincached.WithElo("projects", "ExporationFramwork", 1000).ScoreAccessor(u.Id, float64(append(delta, 0)[0]))
// }

// func (u *ExporationFramwork) Feedbacks(newFeedback ...string) []string {
// 	return mixincached.WithFeedbacks("projects", "ExporationFramwork").Feedback(u.Id, newFeedback...)
// }
