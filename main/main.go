package main

import (
	"time"

	"github.com/doptime/doptime/httpserve"
	"github.com/doptime/eloevo/devops/projects"
	"github.com/doptime/eloevo/utils"
	//_ "github.com/doptime/doptime/httpserve"
)

type TestStruct struct {
	Name string `case:"lower" trim:"left"`
	Age  int    `min:"18" max:"60"`
}

func main() {
	projects.EvoLearningSolution()
	httpserve.Debug()
	time.Sleep(1000000 * time.Second)
	a, e := utils.GetEmbedding("hello world")
	println("embedding result", a, e.Error())
	projects.AgentSelectAndExecute()
	// agents.AgentFunctioncallTest.Call(context.Background(), map[string]any{})
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
