# ltbsky

ltbsky is a lightweight library for posting to Bluesky.

# Features

- Create a post
- Embed images in a post
- Automatically parse web links and Bluesky mentions from a post

# Example

```go
server := os.Getenv("BSKY_SERVER") // e.g., "https://bsky.social"
handle := os.Getenv("BSKY_HANDLE") // username, e.g. "golang.org"
password := os.Getenv("BSKY_PASSWORD") // from Bluesky -> Settings -> Privacy and security -> App passwords

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
```
