package tavily

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// TavilyRequest Payload structure to hit the Tavily API
type TavilyRequest struct {
	ApiKey            string   `json:"api_key"`
	Query             string   `json:"query"`
	SearchDepth       string   `json:"search_depth,omitempty"`
	IncludeImages     bool     `json:"include_images,omitempty"`
	IncludeAnswer     bool     `json:"include_answer,omitempty"`
	IncludeRawContent bool     `json:"include_raw_content,omitempty"`
	MaxResults        uint32   `json:"max_results,omitempty"`
	IncludeDomains    []string `json:"include_domains,omitempty"`
	ExcludeDomains    []string `json:"exclude_domains,omitempty"`
}

// TavilyResponse Responses will take this c
type TavilyResponse struct {
	Answer            string         `json:"answer"`
	Query             string         `json:"query"`
	ResponseTime      float64        `json:"response_time"`
	Images            []string       `json:"images,omitempty"`
	FollowUpQuestions []string       `json:"follow_up_questions"`
	Results           []TavilyResult `json:"results"`
}

type TavilyResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	RawContent string  `json:"raw_content"`
	Score      float64 `json:"score"`
}

// Client A client to work with Tavily
type Client struct {
	APIKey      string
	maxResults  uint32
	searchDepth string
	timeout     time.Duration
}

func (c *Client) SetMaxResults(d uint32) {
	c.maxResults = 1
}

// SetTimeout Set the request timeout in Millisecond (ms)
func (c *Client) SetTimeout(d uint32) {
	c.timeout = time.Millisecond * time.Duration(d)
}

func (c *Client) SetSearchDepth(d string) error {
	switch d {
	case "basic":
		c.searchDepth = d
	case "advanced":
		c.searchDepth = d
	default:
		return ErrUnknownDepth
	}
	return nil
}

// NewClient Produce a new Tavily Client with the given API Key
// If the API key is empty, it will return an error
func NewClient(key string) (*Client, error) {
	if key == "" {
		return nil, errors.New("No API Key was provided")
	}

	c := Client{
		APIKey:      key,
		maxResults:  1,
		searchDepth: "basic",
		timeout:     30_000 * time.Millisecond,
	}

	return &c, nil
}

// Search Use Tavily AI to search the internet. Must provide a query string.
// Optionally, provide a request object to customize the search query.
// If an array of request objects are input, only the first will be used
func (c *Client) Search(q string, params ...TavilyRequest) (*TavilyResponse, error) {
	ctx := context.Background()
	r := c.defaultReq(q)

	for _, cfg := range params {
		r.SearchDepth = cfg.SearchDepth
		r.IncludeImages = cfg.IncludeImages
		r.IncludeAnswer = cfg.IncludeAnswer
		r.IncludeRawContent = cfg.IncludeRawContent
		r.MaxResults = cfg.MaxResults
		r.IncludeDomains = cfg.IncludeDomains
		r.ExcludeDomains = cfg.ExcludeDomains
		break
	}

	return c.search(ctx, r)
}

// QASearch - Question and Answer Search start. Defaults are changed to search depth advanced
// regardless how the client is set unless explicitly added in the params to get the best answer.
// Returns only the answer to the question instead of the entire response
// Params sent accept search depth, max results, and domains to include or exclude. All else is ignored
// Only the first TavilyRequest Struct will be concidered if multiple are sent
func (c *Client) QASearch(q string, params ...TavilyRequest) (string, error) {
	ctx := context.Background()
	r := c.defaultReq(q)
	r.IncludeAnswer = true
	r.SearchDepth = "advanced"

	for _, cfg := range params {
		r.SearchDepth = cfg.SearchDepth
		r.MaxResults = cfg.MaxResults
		r.IncludeDomains = cfg.IncludeDomains
		r.ExcludeDomains = cfg.ExcludeDomains
		break
	}

	resp, err := c.search(ctx, r)
	if err != nil {
		slog.Debug("Tavily: Unable to complete our Q&A", "err", err)
		return "", err
	}

	return resp.Answer, nil
}

// QASearchWithCtx - With a custom context value,
// Question and Answer Search start. Defaults are changed to search depth advanced
// regardless how the client is set unless explicitly added in the params to get the best answer.
// Returns only the answer to the question instead of the entire response
// Params sent accept search depth, max results, and domains to include or exclude. All else is ignored
// Only the first TavilyRequest Struct will be concidered if multiple are sent
func (c *Client) QASearchWithCtx(ctx context.Context, q string, params ...TavilyRequest) (string, error) {
	r := c.defaultReq(q)
	r.IncludeAnswer = true
	r.SearchDepth = "advanced"

	for _, cfg := range params {
		r.SearchDepth = cfg.SearchDepth
		r.MaxResults = cfg.MaxResults
		r.IncludeDomains = cfg.IncludeDomains
		r.ExcludeDomains = cfg.ExcludeDomains
		break
	}

	resp, err := c.search(ctx, r)
	if err != nil {
		slog.Debug("Tavily: Unable to complete our Q&A", "err", err)
		return "", err
	}

	return resp.Answer, nil
}

func (c *Client) SearchWithDepth(q string, depth string) (*TavilyResponse, error) {
	if depth != "basic" && depth != "advanced" {
		return nil, ErrUnknownDepth
	}

	ctx := context.Background()
	r := TavilyRequest{
		ApiKey:        c.APIKey,
		Query:         q,
		IncludeAnswer: true,
		MaxResults:    c.maxResults,
		SearchDepth:   depth,
	}

	return c.search(ctx, &r)
}

func (c *Client) SearchWithNResults(q string, n uint32) (*TavilyResponse, error) {
	ctx := context.Background()
	r := TavilyRequest{
		ApiKey:        c.APIKey,
		Query:         q,
		IncludeAnswer: true,
		MaxResults:    n,
		SearchDepth:   c.searchDepth,
	}

	return c.search(ctx, &r)
}

func (c *Client) SearchWithDomains(q string, inc []string, exc []string) (*TavilyResponse, error) {
	ctx := context.Background()
	r := c.defaultReq(q)
	r.IncludeDomains = inc
	r.ExcludeDomains = exc

	return c.search(ctx, r)
}

func (c *Client) SearchWithCtx(ctx context.Context, q string, params ...TavilyRequest) (*TavilyResponse, error) {
	r := c.defaultReq(q)

	for _, cfg := range params {
		r.SearchDepth = cfg.SearchDepth
		r.IncludeImages = cfg.IncludeImages
		r.IncludeAnswer = cfg.IncludeAnswer
		r.IncludeRawContent = cfg.IncludeRawContent
		r.MaxResults = cfg.MaxResults
		r.IncludeDomains = cfg.IncludeDomains
		r.ExcludeDomains = cfg.ExcludeDomains
		break
	}
	return c.search(ctx, r)
}

func (c *Client) defaultReq(q string) *TavilyRequest {
	r := TavilyRequest{
		ApiKey:      c.APIKey,
		Query:       q,
		MaxResults:  c.maxResults,
		SearchDepth: c.searchDepth,
	}

	return &r
}

func (c *Client) search(ctx context.Context, r *TavilyRequest) (*TavilyResponse, error) {
	b, err := json.Marshal(r)
	if err != nil {
		slog.Debug("Tavily: Unable to marshal the request into a JSON", "err", err)
		return nil, err
	}

	reader := bytes.NewReader(b)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, TavilyBaseURL, reader)
	if err != nil {
		slog.Debug("Tavily: Could not make a new request object", "err", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Debug("Tavily: Unable to make a request to Tavily")
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return nil, errors.New("Your request is invalid")
	case http.StatusUnauthorized:
		return nil, errors.New("Your API key is wrong")
	case http.StatusForbidden:
		return nil, errors.New("The endpoint requested is hidden for administrators only")
	case http.StatusNotFound:
		return nil, errors.New("The specified endpoint could not be found")
	case http.StatusMethodNotAllowed:
		return nil, errors.New("You tried to access an endpoint with an invalid method")
	case http.StatusTooManyRequests:
		return nil, errors.New("You're requesting too many results; Slow down!")
	case http.StatusInternalServerError:
		return nil, errors.New("We had a problem with our server. Try again later")
	case http.StatusServiceUnavailable:
		return nil, errors.New("We're temporarily offline for maintenance. Please try again later")
	case http.StatusGatewayTimeout:
		return nil, errors.New("We're temporarily offline for maintenance. Please try again later")
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug("Tavily: Unable to read the stream from the body", "err", err)
		return nil, err
	}

	result := TavilyResponse{}
	if err = json.Unmarshal(body, &result); err != nil {
		slog.Debug("Issue unmarshalling response from server into our expected struct")
		return nil, err
	}

	return &result, nil
}

//func (c *Client) getSearchContext(q string, maxTokens uint, ...TavilyRequest) ()
//func (c *Client) getCompanyInfo(q string, maxTokens uint, ...TavilyRequest) ()
