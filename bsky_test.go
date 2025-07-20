package bsky

import (
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
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

func TestLogin(t *testing.T) {
	t.Skip("Skipping network login test")
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	client, err := NewClient(server, handle, password)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	loggedIn, err := client.Login()
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	if !loggedIn {
		t.Error("wanted logged in to be true, got false")
	}
}

func TestPost(t *testing.T) {
	t.Skip("Skipping network post test")
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	client, err := NewClient(server, handle, password)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	loggedIn, err := client.Login()
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	if !loggedIn {
		t.Error("wanted logged in to be true, got false")
	}

	content := "Hello, world!"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestPostWithLinks(t *testing.T) {
	t.Skip("Skipping test for posting links")
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	client, err := NewClient(server, handle, password)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	loggedIn, err := client.Login()
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	if !loggedIn {
		t.Error("wanted logged in to be true, got false")
	}

	content := "Link test: https://go.dev https://pkg.go.dev"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestPostWithMentions(t *testing.T) {
	t.Skip("Skipping test for posting mentions")
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	client, err := NewClient(server, handle, password)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	loggedIn, err := client.Login()
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	if !loggedIn {
		t.Error("wanted logged in to be true, got false")
	}

	content := "Mention test: @itodd.dev @golang.org"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}

func TestPostWithMentionsAndLinks(t *testing.T) {
	t.Skip("Skipping test for posting mentions and links")
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	client, err := NewClient(server, handle, password)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	loggedIn, err := client.Login()
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
	if !loggedIn {
		t.Error("wanted logged in to be true, got false")
	}

	content := "Mention and link test: @itodd.dev https://go.dev @golang.org https://pkg.go.dev"
	pb := NewPostBuilder(content)
	_, err = client.Post(pb)
	if err != nil {
		t.Fatalf("wanted no error, got %v", err)
	}
}
