package spider

type ParseResult struct {
	Requesrts []*Request
	Items     []interface{}
}

type RuleTree struct {
	Root  func() ([]*Request, error)
	Trunk map[string]*Rule
}

type Rule struct {
	ItemFields []string
	ParseFunc  func(*Context) (ParseResult, error)
}
