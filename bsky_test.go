package ltbsky

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	server := "https://bsky.social"
	handle := "test.handle"
	password := "test.password"
	client, err := NewClient(server, handle, password)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	if client.server != server {
		t.Errorf("wanted server '%s', got %s", server, client.server)
	}
	if client.handle != handle {
		t.Errorf("wanted handle '%s', got %s", handle, client.handle)
	}
	if client.password != password {
		t.Errorf("wanted password '%s', got %s", password, client.password)
	}
}

func TestAuth(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, err := NewClient(server.URL, "test.handle", "test.password")
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	err = client.auth()
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestPost(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, err := NewClient(server.URL, "test.handle", "test.password")
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}

	content := "Hello, world!"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestPostWithLinks(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, err := NewClient(server.URL, "test.handle", "test.password")
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}

	content := "Link test: https://go.dev https://pkg.go.dev"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestPostWithMentions(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, err := NewClient(server.URL, "test.handle", "test.password")
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}

	content := "Mention test: @itodd.dev @golang.org"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestPostWithTags(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, err := NewClient(server.URL, "test.handle", "test.password")
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	content := "Tag test: #golang #bsky"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestPostWithMentionsAndLinks(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, err := NewClient(server.URL, "test.handle", "test.password")
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}

	content := "Mention and link test: @itodd.dev https://go.dev @golang.org https://pkg.go.dev"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestPostBuilder(t *testing.T) {
	content := "Test content"
	pb := NewPostBuilder(content)
	if pb.content != content {
		t.Errorf("wanted content '%s', got '%s'", content, pb.content)
	}
	if len(pb.images) != 0 {
		t.Errorf("wanted no images, got %d", len(pb.images))
	}
	if len(pb.facets) != 0 {
		t.Errorf("wanted no facets, got %d", len(pb.facets))
	}
	pb.AddContent(" more content")
	if pb.content != "Test content more content" {
		t.Errorf("wanted content 'Test content more content', got '%s'", pb.content)
	}
}

func TestPostBuilderAddLang(t *testing.T) {
	pb := NewPostBuilder("Test content")
	if len(pb.langs) != 0 {
		t.Errorf("wanted no langs, got %d", len(pb.langs))
	}
	pb.AddLang("en")
	if len(pb.langs) != 1 || pb.langs[0] != "en" {
		t.Errorf("wanted langs ['en'], got %v", pb.langs)
	}
	pb.AddLang("")
	if len(pb.langs) != 1 || pb.langs[0] != "en" {
		t.Errorf("wanted langs ['en'], got %v", pb.langs)
	}
	pb.AddLang("fr")
	if len(pb.langs) != 2 || pb.langs[1] != "fr" {
		t.Errorf("wanted langs ['en', 'fr'], got %v", pb.langs)
	}
}

func TestPostBuilderAddImage(t *testing.T) {
	pb := NewPostBuilder("Test content")
	if len(pb.images) != 0 {
		t.Errorf("wanted no images, got %d", len(pb.images))
	}
	path := "./test-data/bsky-go-1.jpg"
	alt := "Test image"
	pb.AddImageFromPath(path, alt)
	if len(pb.images) != 1 || pb.images[0].Path != path {
		t.Errorf("wanted image paths ['%s'], got %v", path, pb.images[0].Path)
	}
	path = "./test-data/bsky-go-1.png"
	alt = "Another test image"
	pb.AddImageFromPath(path, alt)
	if len(pb.images) != 2 || pb.images[1].Path != path {
		t.Errorf("wanted image paths ['%s'], got %v", path, pb.images[1].Path)
	}
	_, err := pb.buildFor("https://bsky.social", &http.Client{})
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	if len(pb.images[0].Bytes) == 0 {
		t.Error("wanted image[0].Bytes to be set, got empty")
	}
	if len(pb.images[1].Bytes) == 0 {
		t.Error("wanted image[0].Bytes to be set, got empty")
	}
}

func TestSimulatedPost(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, err := NewClient(server.URL, "test.handle", "test.password")
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}

	content := "Hello, world!"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestSimulatedPostWithImage(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, err := NewClient(server.URL, "test.handle", "test.password")
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}

	content := "Hello, world!"
	pb := NewPostBuilder(content)
	pb.AddImageFromPath("./test-data/bsky-go-1.png", "test image")
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestSimulatedPostWithMentionsAndLinks(t *testing.T) {
	server := newMockServer()
	defer server.Close()

	client, err := NewClient(server.URL, "test.handle", "test.password")
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}

	content := "Mention and link test: @itodd.dev https://go.dev @golang.org https://pkg.go.dev"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestAddImageFromBytes(t *testing.T) {
	pb := NewPostBuilder("Test content")
	if len(pb.images) != 0 {
		t.Errorf("wanted no images, got %d", len(pb.images))
	}
	// Add an image from bytes
	pb.AddImageFromBytes([]byte("test image data"), "Test image")
	if len(pb.images) != 1 {
		t.Errorf("wanted one image, got %d", len(pb.images))
	}
	if !bytes.Equal(pb.images[0].Bytes, []byte("test image data")) {
		t.Errorf("wanted image bytes 'test image data', got '%s'", pb.images[0].Bytes)
	}
	if pb.images[0].Alt != "Test image" {
		t.Errorf("wanted image alt 'Test image', got '%s'", pb.images[0].Alt)
	}
	// Add a second image
	pb.AddImageFromBytes([]byte("more test image data"), "Test image two")
	if len(pb.images) != 2 {
		t.Errorf("wanted two images, got %d", len(pb.images))
	}
	if !bytes.Equal(pb.images[1].Bytes, []byte("more test image data")) {
		t.Errorf("wanted second image bytes 'more test image data', got '%s'", pb.images[1].Bytes)
	}
	if pb.images[1].Alt != "Test image two" {
		t.Errorf("wanted second image alt 'Test image two', got '%s'", pb.images[1].Alt)
	}
}

func TestParseLinks(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedFacets []struct {
			ByteStart int
			ByteEnd   int
			Uri       string
		}
	}{
		{
			name:    "Two links",
			content: "Hello https://go.dev and https://pkg.go.dev!",
			expectedFacets: []struct {
				ByteStart int
				ByteEnd   int
				Uri       string
			}{
				{ByteStart: 6, ByteEnd: 20, Uri: "https://go.dev"},
				{ByteStart: 25, ByteEnd: 43, Uri: "https://pkg.go.dev"},
			},
		},
		{
			name:    "No links",
			content: "Hello world!",
			expectedFacets: []struct {
				ByteStart int
				ByteEnd   int
				Uri       string
			}{},
		},
		{
			name:    "One link at start",
			content: "https://go.dev Hello world!",
			expectedFacets: []struct {
				ByteStart int
				ByteEnd   int
				Uri       string
			}{
				{ByteStart: 0, ByteEnd: 14, Uri: "https://go.dev"},
			},
		},
		{
			name:    "One link at end",
			content: "Hello world! https://go.dev",
			expectedFacets: []struct {
				ByteStart int
				ByteEnd   int
				Uri       string
			}{
				{ByteStart: 13, ByteEnd: 27, Uri: "https://go.dev"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPostBuilder(tt.content)
			pb.parseLinks()

			if len(pb.facets) != len(tt.expectedFacets) {
				t.Errorf("wanted %d facets, got %d", len(tt.expectedFacets), len(pb.facets))
				return
			}

			for i, expected := range tt.expectedFacets {
				if pb.facets[i].Index.ByteStart != expected.ByteStart || pb.facets[i].Index.ByteEnd != expected.ByteEnd {
					t.Errorf("facet %d: wanted index [%d,%d], got [%d,%d]", i, expected.ByteStart, expected.ByteEnd, pb.facets[i].Index.ByteStart, pb.facets[i].Index.ByteEnd)
				}
				if pb.facets[i].Features[0].Uri != expected.Uri {
					t.Errorf("facet %d: wanted URI '%s', got '%s'", i, expected.Uri, pb.facets[i].Features[0].Uri)
				}
			}
		})
	}
}

type mockHTTPClient struct {
	responses map[string]string
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if req.URL.Path == "/xrpc/com.atproto.identity.resolveHandle" {
		handle := req.URL.Query().Get("handle")
		if did, ok := m.responses[handle]; ok {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{"did": "%s"}`, did))),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       http.NoBody,
			Header:     make(http.Header),
		}, nil
	}
	return nil, nil
}

func TestParseMentions(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		mockResponses  map[string]string
		expectedFacets []struct {
			ByteStart int
			ByteEnd   int
			Did       string
		}
	}{
		{
			name:    "Two mentions",
			content: "Hello @itodd.dev and @golang.org!",
			mockResponses: map[string]string{
				"itodd.dev":  "did:example:itodd",
				"golang.org": "did:example:golang",
			},
			expectedFacets: []struct {
				ByteStart int
				ByteEnd   int
				Did       string
			}{
				{ByteStart: 6, ByteEnd: 16, Did: "did:example:itodd"},
				{ByteStart: 21, ByteEnd: 32, Did: "did:example:golang"},
			},
		},
		{
			name:    "No mentions",
			content: "Hello world!",
			mockResponses: map[string]string{
				"itodd.dev": "did:example:itodd",
			},
			expectedFacets: []struct {
				ByteStart int
				ByteEnd   int
				Did       string
			}{},
		},
		{
			name:    "One mention at start",
			content: "@itodd.dev Hello world!",
			mockResponses: map[string]string{
				"itodd.dev": "did:example:itodd",
			},
			expectedFacets: []struct {
				ByteStart int
				ByteEnd   int
				Did       string
			}{
				{ByteStart: 0, ByteEnd: 10, Did: "did:example:itodd"},
			},
		},
		{
			name:    "One mention at end",
			content: "Hello world! @itodd.dev",
			mockResponses: map[string]string{
				"itodd.dev": "did:example:itodd",
			},
			expectedFacets: []struct {
				ByteStart int
				ByteEnd   int
				Did       string
			}{
				{ByteStart: 13, ByteEnd: 23, Did: "did:example:itodd"},
			},
		},
		{
			name:          "Unresolved mention",
			content:       "Hello @unknown.dev!",
			mockResponses: map[string]string{},
			expectedFacets: []struct {
				ByteStart int
				ByteEnd   int
				Did       string
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &mockHTTPClient{responses: tt.mockResponses}
			pb := NewPostBuilder(tt.content)
			pb.parseMentions("", c)

			if len(pb.facets) != len(tt.expectedFacets) {
				t.Errorf("wanted %d facets, got %d", len(tt.expectedFacets), len(pb.facets))
				return
			}

			for i, expected := range tt.expectedFacets {
				if pb.facets[i].Index.ByteStart != expected.ByteStart || pb.facets[i].Index.ByteEnd != expected.ByteEnd {
					t.Errorf("facet %d: wanted index [%d,%d], got [%d,%d]", i, expected.ByteStart, expected.ByteEnd, pb.facets[i].Index.ByteStart, pb.facets[i].Index.ByteEnd)
				}
				if pb.facets[i].Features[0].Did != expected.Did {
					t.Errorf("facet %d: wanted DID '%s', got '%s'", i, expected.Did, pb.facets[i].Features[0].Did)
				}
			}
		})
	}
}

func TestParseTags(t *testing.T) {
	content := "This is a test post with #golang and #bsky tags."
	pb := NewPostBuilder(content)
	pb.parseTags()
	if len(pb.facets) != 2 {
		t.Errorf("wanted 2 facets, got %d", len(pb.facets))
	}
	if pb.facets[0].Index.ByteStart != 25 || pb.facets[0].Index.ByteEnd != 32 {
		t.Errorf("wanted first facet index [25,32], got [%d,%d]", pb.facets[0].Index.ByteStart, pb.facets[0].Index.ByteEnd)
	}
	if pb.facets[0].Features[0].Tag != "golang" {
		t.Errorf("wanted first facet text 'golang', got '%s'", pb.facets[0].Features[0].Tag)
	}
	if pb.facets[1].Index.ByteStart != 37 || pb.facets[1].Index.ByteEnd != 42 {
		t.Errorf("wanted second facet index [37,42], got [%d,%d]", pb.facets[1].Index.ByteStart, pb.facets[1].Index.ByteEnd)
	}
	if pb.facets[1].Features[0].Tag != "bsky" {
		t.Errorf("wanted second facet text 'bsky', got '%s'", pb.facets[1].Features[0].Tag)
	}
}

func newMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/xrpc/com.atproto.server.createSession":
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"accessJwt": "test.token"}`))
			if err != nil {
				return
			}
		case "/xrpc/com.atproto.repo.createRecord":
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"uri": "test.uri", "cid": "test.cid"}`))
			if err != nil {
				return
			}
		case "/xrpc/com.atproto.repo.uploadBlob":
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"blob": {"$type": "blob", "ref": {"$link": "test.link"}, "mimeType": "image/png", "size": 1234}}`))
			if err != nil {
				return
			}
		case "/xrpc/com.atproto.identity.resolveHandle":
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"did": "test.did"}`))
			if err != nil {
				return
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}
