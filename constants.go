package tavily

import (
	"errors"
)

const (
	TavilyBaseURL        = "https://api.tavily.com/search"
	DefaultModelEncoding = "gpt-3.5-turbo"
	DefaultMaxTokens     = 4000
)

var (
	ErrUnknownDepth = errors.New("Unknown search depth, use 'basic' or 'advanced'")
)
