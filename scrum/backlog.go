package scrum

import (
	"fmt"
	"time"
)

type Backlog struct {
	Id        string
	Info      string
	Reference string
	Sponsor   string
	CreatedAt time.Time
	UpdatedAt time.Time
	Expired   bool
	Done      bool
}

func (b *Backlog) String() string {
	return fmt.Sprintf("Backlog[Id: %s, Info: %s, Reference: %s, Sponsor: %s, CreatedAt: %s, UpdatedAt: %s, Expired: %t, Done: %t]",
		b.Id, b.Info, b.Reference, b.Sponsor, b.CreatedAt.Format(time.RFC3339), b.UpdatedAt.Format(time.RFC3339), b.Expired, b.Done)
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
