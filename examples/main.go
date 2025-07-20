package main

import (
	"log"
	"os"

	"github.com/fflewddur/bsky"
)

func main() {
	log.Println("bsky test")
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
	postBuilder.AddLang("en")
	pr, err := postBuilder.BuildFor(server)
	if err != nil {
		log.Fatalf("Error building post: %v", err)
	}
	log.Printf("Post request: %+v", pr)
	uri, err := client.Post(postBuilder)
	if err != nil {
		log.Fatalf("Error posting: %v", err)
	}
	log.Printf("Post created with URI: %s", uri)
	log.Println("bsky test completed")
}
