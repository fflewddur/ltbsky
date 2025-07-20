package main

import (
	"log"
	"os"

	"github.com/fflewddur/bsky"
)

func main() {
	// BasicExample()
	ImageExample()
}

func BasicExample() {
	log.Println("bsky basic example")
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	client, err := bsky.NewClient(server, handle, password)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	log.Printf("Client created: %+v", client)
	loggedIn, err := client.Login()
	if err != nil {
		log.Fatalf("Error logging in: %v", err)
	}
	if loggedIn {
		log.Println("Successfully logged in")
	} else {
		log.Println("Login failed")
	}
	postBuilder := bsky.NewPostBuilder("Hello, world!")
	uri, err := client.Post(postBuilder)
	if err != nil {
		log.Fatalf("Error posting: %v", err)
	}
	log.Printf("Post created with URI: %s", uri)
	log.Println("bsky test completed")
}

func ImageExample() {
	log.Println("bsky image example")
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	client, err := bsky.NewClient(server, handle, password)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	loggedIn, err := client.Login()
	if err != nil {
		log.Fatalf("Error logging in: %v", err)
	}
	if !loggedIn {
		log.Fatal("Login failed")
	}
	postBuilder := bsky.NewPostBuilder("Hello with image!")
	postBuilder.AddLang("en")
	postBuilder.AddImageFromPath("./test-data/bsky-go-1.png", "Alt text for image")

	uri, err := client.Post(postBuilder)
	if err != nil {
		log.Fatalf("Error posting: %v", err)
	}
	log.Printf("Post created with URI: %s", uri)
	log.Println("bsky image example completed")
}
