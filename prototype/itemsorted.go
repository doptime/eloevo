package prototype

type ItemsSorted struct {
	ItemsRefById []string       `description:"Items sorted. Referenced by Items Id"`
	Extra        map[string]any `msgpack:"-" json:"-" description:"Extra, call parameter of Agent"`
}

type SolutionRefinement_v250319 struct {
	ItemIdsShouldKeptInSolution     []string `description:"Items that should be kept in solution, sorted by importance,best first. Referenced by Items Id"`
	ItemIdsShouldRemoveFromSolution []string `description:"Items that should be remove in solution, sorted by importance, worst first. Referenced by Items Id"`
	NewProposedItemContent          string   `description:"If needs to improve current solution, proposed new Item content here; else keep it empty"`
}

type SolutionRefinement struct {
	ProConsToItems   map[string]int `description:"Pros to Items, here key is item Id, value is int ∈ [-5,5], value above 0 will move item to top, below 0 will move item to bottom, item with negative value will be REMOVED from solution"`
	NewProposedItems []string       `description:"If needs to improve current solution, proposed new Item content here; else keep it empty"`
}
type SolutionRefine struct {
	ProConsToItems   map[string]int    `description:"key: string, item Id ; value: int, ∈ [-5,5]. Votes cnt of Pros or Cons to Items, positive value moving item to top, while minus value moving item to bottom, item with negative votes value will be REMOVED from solution"`
	NewProposedItems map[string]string `description:"key: string, parentNodeId  or \"root\"; value: string, new Item content; Leaving NewProposedItems empty is allowed."`
	Extra            map[string]any    `msgpack:"-" json:"-" description:"-"`
}
