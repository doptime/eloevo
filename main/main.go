package main

import (
	"github.com/doptime/eloevo/devops"
	"github.com/doptime/eloevo/devops/projects"
)

type TestStruct struct {
	Name string `case:"lower" trim:"left"`
	Age  int    `min:"18" max:"60"`
}

func main() {
	projects.BusinessPlansExploration()
	projects.BusinessClusteringExploration()
	devops.UseCaseExploration()
	//projects.GenBusinessPlanParallel()
	//projects.RationalCognitionFrameworkExploration()
	//projects.BusinessUtilityExploration()
	//projects.PersonalAdminExploration()

}
