package main

import (
	"time"

	"github.com/doptime/doptime/httpserve"
	"github.com/doptime/eloevo/evo"
	"github.com/doptime/eloevo/evobymeasured"
	"github.com/doptime/eloevo/projects"
	"github.com/doptime/eloevo/utils"
	//_ "github.com/doptime/doptime/httpserve"
)

func main() {
	evo.SetGoals("/Users/yang/doptime/redisdb/intentions_and_goals/goals.toml", "RedisDB")
	evobymeasured.EvoAGoal("RedisDB", "CreateUnifiedCookbook")
	//learnbychoose.EvoLearnByChooseSolution()
	httpserve.Debug()
	time.Sleep(1000000 * time.Second)
	a, e := utils.GetEmbedding("hello world")
	println("embedding result", a, e.Error())
	projects.AgentSelectAndExecute()
	// agents.AgentFunctioncallTest.Call( map[string]any{})
	// agents.AgentResponseTest()
	//projects.LoadResultsToRedis()
	// scrum.AgentSessionArrangementCall()
	// go projects.BusinessPlansEdTechExploration()
	// time.Sleep(1000000 * time.Second)
	// go projects.BusinessPlansPWAExploration()
	// projects.BusinessClusteringExploration()

	//devops.UseCaseExploration()
	//projects.GenBusinessPlanParallel()
	//projects.RationalCognitionFrameworkExploration()
	//projects.BusinessUtilityExploration()
	//projects.PersonalAdminExploration()

}
