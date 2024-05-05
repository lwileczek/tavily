package tavily

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBasicSearch(t *testing.T) {
	sh := SearchHandler{}
	ts := httptest.NewServer(http.HandlerFunc(sh.ServeHTTP))
	defer ts.Close()

	c := Client{
		APIKey:      "A fake API Key",
		maxResults:  1,
		searchDepth: "basic",
		timeout:     30_000 * time.Millisecond,
		tavilyURL:   ts.URL,
	}

	testCases := []struct {
		//a message which describes the test
		msg string
		//A payload to use in the test
		req TavilyRequest
	}{
		{
			msg: "Happy Path Simple Query",
			req: TavilyRequest{
				Query: "A Pretend Query",
			},
		},
	}

	for _, tc := range testCases {
		if _, err := c.Search(tc.req.Query); err != nil {
			t.Errorf("There was an error testing our search!Test: %s\nError:%s\n", tc.msg, err)
		}
	}
}

func TestNewClientSearch(t *testing.T) {
	if _, err := NewClient(""); err == nil {
		t.Error("Passed in an empty string and should have gotten an error!")
	}

	//TODO: Check that the API key has the correct format at least and isn't something silly
	if _, err := NewClient("ladafafdfd"); err != nil {
		t.Errorf("Should have gotten a proper client with no errors but did not!\nError: %s", err)
	}
}
