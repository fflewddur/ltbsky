package bsky

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	log.Printf("Logging in to %s with handle %s", url, c.handle)
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body) // Read the response body to ensure we consume it
	if err != nil {
		return false, fmt.Errorf("error reading response body: %w", err)
	}
	// log.Printf("Login response: %s", string(b))
	var sessionResponse SessionResponse
	if err := json.Unmarshal(b, &sessionResponse); err != nil {
		return false, fmt.Errorf("error unmarshaling response: %w", err)
	}
	c.accessToken = sessionResponse.AccessJwt
	// log.Printf("Access JWT: %s\nRefresh JWT: %s", sessionResponse.AccessJwt, sessionResponse.RefreshJwt)
	return true, nil // Simulate a successful login for now
}

// PostResponse represents the response from the server after a successful post.
type PostResponse struct {
	Uri   string `json:"uri"`
	Cid   string `json:"cid"`
	Error string `json:"error,omitempty"` // Optional field to capture any error messages
}

type PostRequest struct {
	Repo       string `json:"repo"`
	Collection string `json:"collection"`
	Record     Record `json:"record"`
}

type Record struct {
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}

func (c *Client) Post(content string) (string, error) {
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.createRecord", c.server)
	createdAt := time.Now().UTC().Format(time.RFC3339)
	requestBody := PostRequest{
		Repo:       c.handle,
		Collection: "app.bsky.feed.post",
		Record: Record{
			Text:      content,
			CreatedAt: createdAt,
		},
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

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
	defer resp.Body.Close()

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
