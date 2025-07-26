package ltbsky

import (
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
	_, err := pb.BuildFor("https://bsky.social", &http.Client{})
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
