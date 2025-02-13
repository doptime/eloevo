package agents

import (
	"context"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/doptime/eloevo/agent"
	"github.com/doptime/eloevo/utils"
	"github.com/doptime/redisdb"
)

var AgentGenNicheMarketOpportunity = agent.NewAgent(template.Must(template.New("question").Parse(`
请作为一位创新思维专家,生成一个在AGI时代可能出现的利基应用领域及其盈利机会.

{{.Now}}

约束条件:
	说明:{{.RestrictionsAnnotation}}

1. 高度随机与多样化:
- 在生成任务的首段，展开与任务无直接关联的启发性思路
- 这些思路可以来源于不同的工作角色需求、跨行业的上下游需求、或新兴技术的发展趋势
- 确保这些思路具有高度随机性和不可预测性，以引导生成多样化的机会
- 每个生成的方案必须采用不同的结构和内容组织方式

2. 差异化与独特性:
- 由于任务将被上百万次地抽样运行，生成的每个机会必须彼此独特，避免重复
- 通过随机组合不同的行业、技术和市场需求来实现
- 每个方案应包含不同的分析维度和内容重点

3. 实现与测试的简易性:
- 市场机会应以软件产品的形式呈现,且具备微型化特征
- 若涉及编码,代码量应控制在30K行以内,以确保易于实现和测试

4. 具体与可操作:
- 每次仅生成一个具体的市场机会
- 确保内容详实且具备实际操作性,便于后续的评估和实施
- 内容结构由模型自行决定，但必须包含NicheMarketOpportunityName字段

5. 创新性:
- 鼓励采用非常规的内容组织方式
- 每个方案应包含独特的分析视角和表达方式
- 避免使用固定的模板结构

NicheMarketOpportunityName: 为该市场机会命名（必须包含）

其他内容: 由模型自行决定需要包含哪些分析维度、充分深度评估该商业机会的潜力和可行性。
   `))).WithContent2RedisHash("NicheMarketOpportunityHumanProsperity", func(content string) (field string) {
	return utils.ExtractTagValue(content, "NicheMarketOpportunityName")
})

func GenNicheMarketOpportunity() (err error) {
	var inputsParams = map[string]any{
		"Now": "当前时间" + time.Now().Format("2006-01-02 15:04:05") + " 随机思路Id：" + redisdb.NanoId(8),
		//"RestrictionsAnnotation": "要求生成的市场机会是基于即将出现的重要的技术突破。",
		//"RestrictionsAnnotation": "要求生成的市场机会是基于设想现在的工作得到了AGI的增强，而出现的重要基于。",
		"RestrictionsAnnotation": "要求生成的市场机会是有助于AGI的背景下，目标商业机服务于人类社会的长期稳定繁荣。",
	}
	fmt.Println("GenNicheMarketOpportunity...")
	//return AgentGenNicheMarketOpportunity.WithModel(models.EloModels.SelectOne("roundrobin")).		Call(context.Background(), inputsParams)
	return AgentGenNicheMarketOpportunity.Call(context.Background(), inputsParams)
}

// GenNicheMarketOpportunityParallel calls GenNicheMarketOpportunity 1000 times in 16 parallel threads.
func GenNicheMarketOpportunityParallel() {
	const numThreads = 4
	const numCallsPerThread = 40

	var wg sync.WaitGroup
	wg.Add(numThreads)

	for i := 0; i < numThreads; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numCallsPerThread; j++ {
				GenNicheMarketOpportunity()
			}
		}()
	}
	wg.Wait()
}
