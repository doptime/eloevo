package evotests

import (
	"fmt"
)

func (t *TestScenario) String() string {
	return fmt.Sprintf(
		"Test Scenario:\n"+
			"  ID: %s\n"+
			"  Catalogue: %s\n"+
			"  Name: %s\n"+
			"  Objective: %s\n"+
			"  ComplexityLevel: %s\n"+
			"  EnvironmentConditions: %s\n"+
			"  EvaluationCriteria: %s\n"+
			"  Description: %s",
		t.Id, t.Catalogue, t.Name, t.Objective, t.ComplexityLevel, t.EnvironmentConditions, t.EvaluationCriteria, t.Description,
	)
}
func (m *Model) String() string {
	return fmt.Sprintf(
		"Model:\n"+
			"  Name: %s\n"+
			"  ModelId: %s\n"+
			"  Milestone: %.2f\n"+
			"  ProblemToSolve: %v\n"+
			"  DesignIdeas: %v\n"+
			"  Dependencies: %v\n"+
			"  DevelopFeedbacks: %v\n"+
			"  CompileFeedbacks: %v",
		m.Name, m.ModelId, m.Milestone, m.ProblemToSolve, m.DesignIdeas, m.Dependencies, m.DevelopFeedbacks, m.CompileFeedbacks,
	)
}
