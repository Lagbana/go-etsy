package etsy

import (
	"bytes"
	"encoding/json"

	//"bytes"
	//"context"
	//"encoding/json"
	"fmt"
	"io"
	"strings"

	//"io"
	//"io/ioutil"
	"net/http"
	"net/url"
	//"strings"

	//"github.com/google/go-querystring/query"
)

// Version is the Etsy API current version.
const (
	Version        = "3.0.0"
	defaultBaseURL = "https://api.etsy.com/v3/application/"
	userAgent      = "go-etsy"
)

type Client struct {
	client  *http.Client
	baseURL *url.URL

	// Etsy API auth options
	opts Options

	// User agent used when communicating with the Etsy API.
	UserAgent string

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for talking to different parts of the Etsy API.
	//Responses *ResponseService
	//Forms     *FormService
}

type service struct {
	client *Client
	opts   Options
}

// Options fields - userID and accessToken are optional but required for requests that require oauth2 scope
// Reference: https://developers.etsy.com/documentation/reference/#section/Authentication
type Options struct {
	apiKey      string
	userID      int
	accessToken string
}

type Option func(*Options) error

// WithApp functional parameter to set API Key
func WithApp(apiKey string) Option {
	return func(o *Options) error {
		o.apiKey = apiKey
		return nil
	}
}

// WithOauth - functional parameter to authenticate.
// Needed for requests that require ouath2 scope
// https://developers.etsy.com/documentation/essentials/requests
func WithOauth(userID int, accessToken string) Option {
	return func(o *Options) error {
		o.accessToken = accessToken
		o.userID = userID
		return nil
	}
}

// NewClient returns a new go-etsy API client.
// If a nil httpClient is provided, the default client http.DefaultClient will be used.
func NewClient(httpClient *http.Client, opts ...Option) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)
	c := &Client{
		client:  httpClient,
		baseURL: baseURL,
		UserAgent: userAgent,
	}

	for _, opt := range opts {
		if err := opt(&c.opts); err != nil {
			return nil, err
		}
	}

	c.common.client = c
	c.common.opts = c.opts


	return c, nil
}

// NewRequest creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.baseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.baseURL)
	}
	u, err := c.baseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set( "x-api-key", c.opts.apiKey)
	req.Header.Set("Host", "openapi.etsy.com")
	req.Header.Set("User-Agent", c.UserAgent)

	return req, nil
}