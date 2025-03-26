package mixincached

func WithFeedbacks(redisDBName string, key string) (ret *CtxStrings) {
	return HashKeyStrs(redisDBName, key+":Feedbacks")
}
func (es *CtxStrings) Feedback(Id string, NewFeedback ...string) []string {
	return es.GetInsert(Id, NewFeedback...)
}
