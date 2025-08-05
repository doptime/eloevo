package projects

type TestBuilder struct {
	Id string `description:"ID of the test scenario, unique"`

	Catalogue string `description:"Catalogue of the test scenario"`

	Name string `description:"Name of the test scenario"`

	Objective string `description:"Main objective of the test scenario"`

	ComplexityLevel string `description:"Complexity level of the scenario"`

	EnvironmentConditions string `description:"Conditions under which the scenario is tested"`

	EvaluationCriteria string `description:"Criteria for evaluating the success of the solution"`

	Description string `description:"Description of the test scenario"`
}
