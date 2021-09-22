package mocks

type RateLimitStoreMock struct {
	CheckFunc func(subnet string) bool
	ResetFunc func(subnet string)
}

func (r *RateLimitStoreMock) Check(subnet string) bool {
	return r.CheckFunc(subnet)
}

func (r *RateLimitStoreMock) Reset(subnet string) {
	r.ResetFunc(subnet)
}
