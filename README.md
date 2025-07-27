# ltbsky

ltbsky is a lightweight library for posting to Bluesky.

## Features

- Create a post
- Embed images in a post
- Specify the language(s) of a post
- Automatically parse web links and Bluesky mentions from a post
- Automatically reduce image size to fit within Bluesky's 1MB limit

## Examples

### Create a post

First, we create a client using `ltbsky.NewClient(server, handle, password)`.
The `server` is your Bluesky host URL (e.g., `"https://bsky.social"`), `handle`
is your username (e.g., "golang.org"), and `password` is an app password from
your Bluesky account settings. You can reuse the same client for multiple
posts.

Then we create a post using `NewPostBuilder(content)`, where `content` is the
text of the post. Web links and Bluesky mentions will automatically appear as
links in the post (e.g., `"Visit https://go.dev to learn more"` will show
"go.dev" as a link when viewed on Bluesky).

Finally, we call `client.Post(postBuilder)` to publish the post. The `Post`
method will automatically authenticate with your Bluesky server using your
provided credentials.

```go
server := os.Getenv("BSKY_SERVER") // e.g., "https://bsky.social"
handle := os.Getenv("BSKY_HANDLE") // username, e.g. "golang.org"
password := os.Getenv("BSKY_PASSWORD") // from Bluesky -> Settings -> Privacy and security -> App passwords

client, err := ltbsky.NewClient(server, handle, password)
if err != nil {
    log.Fatalf("Error creating client: %v", err)
}
postBuilder := ltbsky.NewPostBuilder("Hello, world! How's it going? https://go.dev")
uri, err := client.Post(postBuilder)
if err != nil {
    log.Fatalf("Error posting: %v", err)
}
log.Printf("Post created with URI: %s", uri)
```

### Create a post with an image

To embed an image in a post, we add a call to
`PostBuilder.AddImageFromPath(path, altText)` or
`PostBuilder.AddImageFromBytes(bytes, altText)`:

```go
// [continued from above]

postBuilder = ltbsky.NewPostBuilder("There's an image in this post. Isn't that right, @golang.org?")
postBuilder.AddImageFromPath("./test-data/bsky-go-1.png", "A screenshot of the Go installation process")
postBuilder.AddImageFromPath("./test-data/bsky-go-1.jpg", "A second screenshot of the Go installation process")
uri, err = client.Post(postBuilder)
if err != nil {
    log.Fatalf("Error posting: %v", err)
}
log.Printf("Post created with URI: %s", uri)
```

### Specify a post's languages

To specify each language used in a post, we add a call to
`PostBuilder.AddLang(langCode)`:

```go
// [continued from above]

postBuilder = ltbsky.NewPostBuilder("Hello, world! Hola, mundo!")
postBuilder.AddLang("en") // Add English language
postBuilder.AddLang("es") // Add Spanish language
uri, err = client.Post(postBuilder)
if err != nil {
    log.Fatalf("Error posting: %v", err)
}
log.Printf("Post created with URI: %s", uri)
```

## Contributing

Contributions are welcome! Please [open an
issue](https://github.com/fflewddur/ltbsky/issues) or submit a pull request on
GitHub. If your idea involves significant effort or major changes, please open
an issue first to discuss it.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE)
file for details.
