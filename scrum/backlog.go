package scrum

import (
	"time"
)

type Backlog struct {
	Info      string
	Reference string
	Sponsor   string
	CreateAt  time.Time
	Expired   bool
}
type ProductGoal struct {
	Title string
}
type Increment struct {
	Id string
}
type Team struct {
	Name        string
	Description string
}
type AdaptionDefinition string
type Sprint struct {
	SprintGoal          string
	Backlog             []Backlog
	Increment           []Increment
	DayBegin            time.Time
	DayEnd              time.Time
	IndependentVariable string
	DependentVariable   string
	DefinitionDone      AdaptionDefinition
}
