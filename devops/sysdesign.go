package devops

type Model struct {
	Id           string
	PathName     string
	Version      int
	Dependencies map[string]*Model
	Metadata     map[string]any
}

type ModelInterface interface {
	FileContent(revision ...string) string
	DesignIdeas() []string
	TechNotation() []string
	FeedBacks() []string
}
