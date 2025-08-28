package main

import (
	"log"
	"os"

	"github.com/fflewddur/ltbsky"
)

func main() {
	BasicExample()
	ImageExample()
	ImageAndFacetsExample()
}

func BasicExample() {
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	client, err := ltbsky.NewClient(server, handle, password)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	postBuilder := ltbsky.NewPostBuilder("Hello, world!")
	uri, err := client.Post(postBuilder)
	if err != nil {
		log.Fatalf("Error posting: %v", err)
	}
	log.Printf("Post created with URI: %s", uri)
}

func ImageExample() {
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	client, err := ltbsky.NewClient(server, handle, password)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	postBuilder := ltbsky.NewPostBuilder("Hello with image!")
	postBuilder.AddLang("en")
	postBuilder.AddImageFromPath("./test-data/bsky-go-1.png", "Alt text for image")

	uri, err := client.Post(postBuilder)
	if err != nil {
		log.Fatalf("Error posting: %v", err)
	}
	log.Printf("Post created with URI: %s", uri)
}

func ImageAndFacetsExample() {
	server := os.Getenv("BSKY_SERVER")
	handle := os.Getenv("BSKY_HANDLE")
	password := os.Getenv("BSKY_PASSWORD")
	client, err := ltbsky.NewClient(server, handle, password)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	postBuilder := ltbsky.NewPostBuilder("Hello with image and facets! https://go.dev @golang.org #golang")
	postBuilder.AddLang("en")
	postBuilder.AddImageFromPath("./test-data/bsky-go-1.png", "Alt text for image")

	uri, err := client.Post(postBuilder)
	if err != nil {
		log.Fatalf("Error posting: %v", err)
	}
	log.Printf("Post created with URI: %s", uri)
}
