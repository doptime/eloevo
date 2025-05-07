package main

import (
	"time"

	"github.com/doptime/eloevo/devops"
	"github.com/doptime/eloevo/devops/projects"
)

type TestStruct struct {
	Name string `case:"lower" trim:"left"`
	Age  int    `min:"18" max:"60"`
}

func main() {
	// agents.AgentFunctioncallTest.Call(context.Background(), map[string]any{})
	// agents.AgentResponseTest()
	//projects.LoadResultsToRedis()
	projects.BusinessPlansDronebotExploration()
	go projects.BusinessPlansEdTechExploration()
	time.Sleep(1000000 * time.Second)
	go projects.BusinessPlansPWAExploration()
	projects.BusinessClusteringExploration()
	devops.UseCaseExploration()
	//projects.GenBusinessPlanParallel()
	//projects.RationalCognitionFrameworkExploration()
	//projects.BusinessUtilityExploration()
	//projects.PersonalAdminExploration()

}
