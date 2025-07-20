package bsky

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"
)

type Client struct {
	server      string
	handle      string
	password    string
	accessToken string
	httpClient  *http.Client
}

// NewClient creates a new Client instance with the provided server, handle, and token.
func NewClient(server, handle, password string) (*Client, error) {
	if server == "" {
		return nil, fmt.Errorf("server cannot be empty")
	}
	if handle == "" {
		return nil, fmt.Errorf("handle cannot be empty")
	}
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}
	c := &Client{
		server:     server,
		handle:     handle,
		password:   password,
		httpClient: &http.Client{},
	}
	return c, nil
}

// SessionResponse represents the response from the server after a successful login.
type SessionResponse struct {
	AccessJwt  string `json:"accessJwt"`
	RefreshJwt string `json:"refreshJwt"`
}

// Login logs in to the server using the provided handle and password.
func (c *Client) Login() (bool, error) {
	url := fmt.Sprintf("%s/xrpc/com.atproto.server.createSession", c.server)
	requestBody := map[string]string{
		"identifier": c.handle,
		"password":   c.password,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return false, fmt.Errorf("error marshaling request body: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("error reading response body: %w", err)
	}
	var sessionResponse SessionResponse
	if err := json.Unmarshal(b, &sessionResponse); err != nil {
		return false, fmt.Errorf("error unmarshaling response: %w", err)
	}
	c.accessToken = sessionResponse.AccessJwt
	return (err != nil), err
}

// PostResponse represents the response from the server after a successful post.
type PostResponse struct {
	Uri   string `json:"uri"`
	Cid   string `json:"cid"`
	Error string `json:"error,omitempty"` // Optional field to capture any error messages
}

type PostRequest struct {
	Repo       string  `json:"repo"`
	Collection string  `json:"collection"`
	Record     *Record `json:"record"`
}

type Record struct {
	Type      string   `json:"$type"`
	Text      string   `json:"text"`
	CreatedAt string   `json:"createdAt"`
	Langs     []string `json:"langs,omitempty"`
	Facets    []Facet  `json:"facets,omitempty"` // Optional field to capture facets
}

type Facet struct {
	Index    Index     `json:"index"`
	Features []Feature `json:"features"`
}

type Index struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
}

type Feature struct {
	Type   string `json:"$type"`
	Handle string `json:"handle,omitempty"`
	Did    string `json:"did,omitempty"`
	Uri    string `json:"uri,omitempty"`
}

// Post creates a new post with the given content.
func (c *Client) Post(pb *PostBuilder) (string, error) {
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.createRecord", c.server)
	pr, err := pb.BuildFor(c.server)
	if err != nil {
		return "", fmt.Errorf("error building post request: %w", err)
	}
	pr.Repo = c.handle // Set the repo to the user's handle
	jsonBody, err := json.Marshal(pr)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}
	log.Printf("Posting to %s with content: %s", url, jsonBody)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	b, err := io.ReadAll(resp.Body) // Read the response body to ensure we consume it
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	var postResponse PostResponse
	if err := json.Unmarshal(b, &postResponse); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("post failed with status code: %d (%s) error: %s", resp.StatusCode, resp.Status, postResponse.Error)
	}
	return postResponse.Uri, nil
}

type PostBuilder struct {
	content    string
	langs      []string
	imagePaths []string
	imageBytes [][]byte
	facets     []*Facet
}

// NewPostBuilder creates a new PostBuilder with the initial content.
func NewPostBuilder(content string) *PostBuilder {
	return &PostBuilder{
		content:    content,
		langs:      []string{},
		imagePaths: []string{},
		imageBytes: [][]byte{},
	}
}

// AddContent appends text content to the post.
func (pb *PostBuilder) AddContent(content string) *PostBuilder {
	pb.content += content
	return pb
}

// AddLang adds a language code to the post.
func (pb *PostBuilder) AddLang(lang string) *PostBuilder {
	if lang != "" {
		pb.langs = append(pb.langs, lang)
	}
	return pb
}

// AddImageFromPath adds an image to the post from disk.
func (pb *PostBuilder) AddImageFromPath(path string) *PostBuilder {
	pb.imagePaths = append(pb.imagePaths, path)
	return pb
}

// AddImageFromBytes adds an image to the post from memory.
func (pb *PostBuilder) AddImageFromBytes(data []byte) *PostBuilder {
	pb.imageBytes = append(pb.imageBytes, data)
	return pb
}

func (pb *PostBuilder) BuildFor(server string) (*PostRequest, error) {
	createdAt := time.Now().UTC().Format(time.RFC3339)
	record := &Record{
		Type:      "app.bsky.feed.post",
		Text:      pb.content,
		CreatedAt: createdAt,
		Langs:     pb.langs,
	}

	if len(pb.imagePaths) > 0 {
		for _, path := range pb.imagePaths {
			log.Printf("Adding image from path: %s", path)
			// TODO: Load image bytes from path
			// TODO: Calculate width, height, and format of the image
		}
	}
	if len(pb.imageBytes) > 0 {
		for _, data := range pb.imageBytes {
			log.Printf("Adding image from bytes of size: %d", len(data))
			// TODO: Save a reference to the image bytes
			// TODO: Calculate width, height, and format of the image
		}
	}
	pb.parseLinks()
	pb.parseMentions(server)
	if len(pb.facets) > 0 {
		record.Facets = make([]Facet, len(pb.facets))
		for i, f := range pb.facets {
			record.Facets[i] = Facet{
				Index:    f.Index,
				Features: f.Features,
			}
		}
	}

	return &PostRequest{
		Collection: "app.bsky.feed.post",
		Record:     record,
	}, nil
}

func (pb *PostBuilder) parseLinks() {
	// regex based on: https://docs.bsky.app/docs/advanced-guides/posts#mentions-and-links
	url_regex := `[$|\W](?P<url>https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*[-a-zA-Z0-9@%_\+~#//=])?)`
	r, err := regexp.Compile(url_regex)
	if err != nil {
		log.Printf("Error compiling regex: %v", err)
		return
	}
	matches := r.FindAllSubmatchIndex([]byte(pb.content), -1)
	for _, match := range matches {
		start := match[2] // start position of the 'url' group
		end := match[3]
		link := pb.content[start:end]
		f := &Facet{
			Index: Index{ByteStart: start, ByteEnd: end},
			Features: []Feature{
				{Type: "app.bsky.richtext.facet#link", Uri: link},
			},
		}
		pb.facets = append(pb.facets, f)
	}
}

func (pb *PostBuilder) parseMentions(server string) {
	// regex based on: https://atproto.com/specs/handle#handle-identifier-syntax
	handle_regex := `[$|\W](?P<handle>@([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)`
	r, err := regexp.Compile(handle_regex)
	if err != nil {
		log.Printf("Error compiling regex: %v", err)
		return
	}
	matches := r.FindAllSubmatchIndex([]byte(pb.content), -1)
	if matches != nil {
		c := &http.Client{}
		for _, match := range matches {
			start := match[2] // start position of the 'handle' group
			end := match[3]
			handle := pb.content[start+1 : end] // +1 to skip the '@' character
			url := fmt.Sprintf("%s/xrpc/com.atproto.identity.resolveHandle?handle=%s", server, handle)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Printf("Error creating request for handle %s: %v", handle, err)
				continue
			}
			resp, err := c.Do(req)
			if err != nil {
				log.Printf("Error making request for handle %s: %v", handle, err)
				continue
			}
			defer func() {
				err = errors.Join(err, resp.Body.Close())
			}()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Failed to resolve handle %s with status code: %d", handle, resp.StatusCode)
				continue
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Error reading response body for handle %s: %v", handle, err)
				continue
			}
			var resolveResponse struct {
				Did string `json:"did"`
			}
			if err := json.Unmarshal(b, &resolveResponse); err != nil {
				log.Printf("Error unmarshaling response for handle %s: %v", handle, err)
				continue
			}
			f := &Facet{
				Index: Index{ByteStart: start, ByteEnd: end},
				Features: []Feature{
					{Type: "app.bsky.richtext.facet#mention", Did: resolveResponse.Did},
				},
			}
			pb.facets = append(pb.facets, f)
		}
	}
}
