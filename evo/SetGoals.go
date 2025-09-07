package evo

import (
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/config"
	"github.com/doptime/eloevo/models"
	"github.com/doptime/eloevo/tool"
	"github.com/doptime/eloevo/utils"
	"github.com/samber/lo"
)

var PromptSetGoals = template.Must(template.New("SetGoals").Parse(`
## 当前系统
{{.ContextFiles}}


<演进系统的目标生成与融合的一般原则>
核心哲学（Core Philosophy）
本指南旨在指导您（LLM）扮演一个战略规划者的角色。您的核心任务是理解并融合两种看似矛盾的力量：

[顶层设计] (Top-Level Design): 代表着一个基于宏观趋势、未来愿景和“大颗粒度数据”的理想化、长远目标。它追求的是开辟全新的“适应性生态位”，可能需要创造当前不存在的技术或市场。它回答的问题是：“我们最终想成为什么？”

[底层驱动] (Bottom-Level Drive): 代表着基于当前用户需求、现有技术能力、市场反馈和“细颗粒度数据”的现实拉力。它追求的是在现有环境中最大化效益和生存能力。它回答的问题是：“我们现在最应该做什么？”

您的最终输出不应是这两者的简单相加，而是通过一种“智能插值”计算出的**[融合目标] (Fusion Goal)**，它既能引领我们走向星辰大海，又能确保我们走出的每一步都坚实有力。

第一步：生成长期目标（源自“顶层设计”）
此步骤旨在让LLM不受当前资源和技术的束缚，进行创造性和前瞻性的思考。

Prompt Template 1: Generate Long-Term Goals

# 指令：生成长期目标 (顶层设计)

作为 [项目/组织名称] 的首席战略家，请基于以下信息，为我们构想一个5-10年的长期愿景目标。

**输入信息:**
* **[愿景描述 (Vision Statement)]**: 我们希望世界因我们而发生怎样的根本性改变？ (例如："实现个性化教育的终极形态，让每个孩子都拥有一个永不疲倦的AI导师。")
* **[宏观趋势分析 (Macro-Trends)]**: 描述未来10年可能影响我们领域的3-5个关键技术、社会或市场趋势。 (例如：通用人工智能、脑机接口、终身学习社会化)
* **[颠覆性机会点 (Disruptive Opportunities)]**: 我们的独特能力可以抓住哪些别人看不到的、能定义下一个时代的机遇？ (例如：“我们拥有最全面的儿童认知模型数据。”)

**输出要求:**
1.  **定义愿景世界**: 清晰描述这个5-10年后的“愿景世界”是什么样的，以及我们在其中的角色。
2.  **生成1-3个核心长期目标**: 这些目标应该是大胆的、方向性的，甚至可能需要依赖尚未完全成熟的技术。
3.  **识别新的动力学机制**: 为了实现这些目标，我们需要建立哪些全新的规则、系统或技术范式？
第二步：生成短期目标（源自“底层驱动”）
此步骤旨在让LLM立足当下，识别出最紧迫、最现实、最有价值的行动。

Prompt Template 2: Generate Short-Term Goals

# 指令：生成短期目标 (底层驱动)

作为 [项目/组织名称] 的产品和工程负责人，请基于以下现实情况，规划未来6-12个月的短期目标。

**输入信息:**
* **[当前用户痛点 (Current User Pain Points)]**: 用户当前反馈最激烈、最影响体验的问题是什么？ (例如："我们的应用加载太慢，内容推荐不精准。")
* **[现有技术栈与资源 (Current Tech & Resources)]**: 我们目前拥有的技术能力、团队规模和预算限制。 (例如：“团队擅长Python和React，但缺乏大规模AI模型训练经验；预算有限。”)
* **[关键业务指标 (Key Business Metrics)]**: 公司当前最关注的KPI是什么？ (例如：日活跃用户数(DAU)、用户留存率、付费转化率)

**输出要求:**
1.  **聚焦核心问题**: 明确当前最需要解决的1-3个核心问题。
2.  **生成3-5个SMART短期目标**: 这些目标必须是具体的（Specific）、可衡量的（Measurable）、可实现的（Achievable）、相关的（Relevant）、有时间限制的（Time-bound）。
3.  **说明与KPI的关联**: 解释每个短期目标如何直接或间接地提升关键业务指标。
第三步：计算冲突最小化的融合目标
这是最关键的一步，它要求LLM对前两步的输出进行辩证思考，找到连接未来的“桥梁”。这对应了您提到的“不同动力学粒度适应性的线性插值或几何插值”思想，我们将其操作化为“冲突识别与桥梁构建”。

Prompt Template 3: Calculate Conflict-Minimized Fusion Goal

# 指令：计算冲突最小化的融合目标

作为 [项目/组织名称] 的CEO，你已经收到了来自战略部门的长期目标和来自产品部门的短期目标。现在，你的任务是融合它们，制定出统一的、冲突最小化的“融合目标”。

**输入信息:**
* **[长期目标列表 (Long-Term Goals)]**: (来自第一步的输出)
* **[短期目标列表 (Short-Term Goals)]**: (来自第二步的输出)

**执行步骤:**
1.  **识别冲突与张力 (Identify Conflicts & Tensions)**:
    * 长期目标A和短期目标B在资源分配、技术方向或用户价值上是否存在直接冲突？
    * 例如：长期目标要求“研发通用AI”，而短期目标要求“优化现有推荐算法”，这两者会争夺顶尖的AI人才。

2.  **寻找连接桥梁 (Find Connecting Bridges)**:
    * 思考：哪些短期任务在完成后，不仅能解决眼前问题，还能为某个长期目标铺路？
    * 这就像是在两个点之间进行“插值”。例如：短期“优化推荐算法”所收集的用户行为数据，能否成为未来“研发通用AI导师”的基础训练数据？

3.  **综合生成融合目标 (Synthesize Fusion Goals)**:
    * 基于上述分析，将短期目标重构(reframe)或升级，使其内嵌长期愿景的基因。
    * **坏的例子 (简单相加)**: "我们既要优化推荐算法，也要开始研发通用AI。"
    * **好的例子 (融合插值)**: "我们的新目标是：构建一个'数据飞轮驱动的推荐系统'。在短期（6个月内），它必须通过更精准的推荐将DAU提升20%；在长期（2年内），它收集和标注的数据必须能支持我们第一个通用AI导师模型的原型开发。"

**输出要求:**
* 一份冲突分析报告。
* 一个包含2-4个“融合目标”的列表，每个目标都应同时阐述其短期价值和长期贡献。
第四步：分配目标的优先权重
基于融合目标，LLM需要给出一个清晰的优先级排序，以便团队能够集中精力。这个权重的计算模型，体现了对愿景和现实的平衡。

Prompt Template 4: Assign Goal Priority Weights

# 指令：分配目标优先权重

基于已确定的“融合目标”列表，请为每个目标分配一个优先权重，并解释理由。

**输入信息:**
* **[融合目标列表 (Fusion Goals List)]**: (来自第三步的输出)

**评估维度与权重计算:**
请对每个融合目标，从以下四个维度进行评分 (1-10分)，并计算最终优先级。

1.  **愿景贡献度 (V-Score, Vision Contribution)**: 该目标对实现最终的“顶层设计”有多大的推动作用？(分值越高越重要)
2.  **现实可行性 (F-Score, Feasibility)**: 以我们当前的技术和资源，实现该目标的确定性有多高？(分值越高越优先)
3.  **短期回报率 (R-Score, Short-term Return)**: 该目标能在12个月内对核心KPI产生多大的积极影响？(分值越高越优先)
4.  **学习与探索价值 (L-Score, Learning Value)**: 即使目标失败，我们能从中学到多少对未来至关重要的知识或数据？(这是对“失败退出机制和适应性增强”的量化，分值越高越值得尝试)

**计算公式 (示例)**:
你可以使用一个加权公式来计算最终的优先级分数 'P'。权重可以根据公司当前所处的阶段（例如，初创期 vs 成熟期）进行调整。

一个示例公式:
$$ P = (w_V \cdot V) + (w_F \cdot F) + (w_R \cdot R) + (w_L \cdot L) $$
其中 'w' 是各维度的权重系数。例如，对于一个需要快速验证市场的初创公司，权重可能是：'w_V=0.2, w_F=0.4, w_R=0.3, w_L=0.1'。

**输出要求:**
1.  为每个融合目标打出 V/F/R/L 四项得分。
2.  设定你认为合理的权重系数 'w'，并解释为什么。
3.  计算每个目标的最终优先级分数 'P'。
4.  根据分数 'P' 从高到低，对所有融合目标进行排序，形成最终的行动优先级列表。
通过这套指南，LLM可以系统性地将一个抽象的哲学框架，转化为一套可操作、可迭代的战略规划流程，从而真正实现“顶层设计”与“底层驱动”的动态融合。

</演进系统的目标生成与融合的一般原则>

## TO Do: 生成演进系统的目标
现在我们需要演进这个系统，请深度思考，并且提出有潜力的长期目标和短期目标。

最后请使用多个调用: SetGoals 来提交Goals变更. 
`))

type GoalsSetted struct {
	FileName string `description:"-"`
	Action   string `description:"edit, add, delete"`

	GoalNameAsID       string `description:"The file names to reserve in the context"`
	GoalDescription    string `description:"The description of the goal"`
	ParentGoalNameAsID string `description:"optional, the parent goal name id"`

	VisionContributionDescription string  `description:"The vision contribution of the goal"`
	VisionContributionScore       float64 `description:"[1.0-10], the vision contribution of the goal"`

	FeasibilityDescription string  `description:"The feasibility of the goal"`
	FeasibilityScore       float64 `description:"[1.0-10], the feasibility of the goal"`

	ShortTermReturnDescription string  `description:"The short term return of the goal. "`
	ShortTermReturnScore       float64 `description:"[1.0-10], the short term return of the goal"`

	LearningValueDescription string  `description:"The learning value of the goal"`
	LearningValueScore       float64 `description:"[1.0-10], the learning value of the goal"`

	Weights  []float64 `description:"array, the weights for each evaluation dimension. w_V, w_F, w_R, w_L respectively"`
	Priority float64   `description:"-"`

	ProblemToResolve string    `description:"The problems to resolve. What is the current System Pain Points"`
	WhatToAchieve    string    `description:"The goals to achieve. What is the expected System Output"`
	HowToAchieve     string    `description:"Tow to achieve the goals. What is the expected System Behavior"`
	Result           *[]string `description:"-"`
}

var ToolSetGoals = tool.NewTool("SetGoals", "要最大化这个系统的长期潜力和短期潜力。要设置哪些目标？	", func(commits *GoalsSetted) {
	//calculate priority, with weights normalized
	totalWeight := 0.0
	for _, w := range commits.Weights {
		totalWeight += w
	}
	if totalWeight <= 0.0001 {
		totalWeight = 1
	}
	for i := range commits.Weights {
		commits.Weights[i] /= totalWeight
	}
	commits.Priority = (commits.VisionContributionScore*commits.Weights[0] +
		commits.FeasibilityScore*commits.Weights[1] +
		commits.ShortTermReturnScore*commits.Weights[2] +
		commits.LearningValueScore*commits.Weights[3]) / totalWeight

	utils.TextFromFile(commits.FileName)
	commitsList := []GoalsSetted{}
	toml.DecodeFile(commits.FileName, &commitsList)
	commitsList = append(commitsList, *commits)

	var ioFileToSave *os.File
	ioFileToSave, _ = os.OpenFile(commits.FileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)

	toml.NewEncoder(ioFileToSave).Encode(commitsList)

})

func SetGoals(goalFile string, realms ...string) (NewContextFiles []string) {

	_realm := lo.Filter(lo.Values(config.EvoRealms), func(r *config.EvoRealm, i int) bool {
		return lo.Contains(realms, r.Name)
	})
	files := utils.TextFromEvoRealms(map[string]bool{}, _realm...)
	var ReturnLineKept = &[]string{}
	agent.NewAgent(PromptSetGoals).WithTools(ToolSetGoals).Call(map[string]any{
		"ContextFiles": files,
		"Result":       ReturnLineKept,
		agent.UseModel: models.Qwen3B235Thinking2507,
	})

	return *ReturnLineKept
}
