package ltbsky

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	goimage "image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"golang.org/x/image/draw"
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

type PostBuilder struct {
	content string
	langs   []string
	images  []*localImage
	facets  []*facet
}

// NewPostBuilder creates a new PostBuilder with the initial content.
func NewPostBuilder(content string) *PostBuilder {
	return &PostBuilder{
		content: content,
		langs:   []string{},
		images:  make([]*localImage, 0),
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
func (pb *PostBuilder) AddImageFromPath(path string, alt string) *PostBuilder {
	localImg := &localImage{
		Path: path,
		Alt:  alt,
	}
	pb.images = append(pb.images, localImg)
	return pb
}

// AddImageFromBytes adds an image to the post from memory.
func (pb *PostBuilder) AddImageFromBytes(data []byte, alt string) *PostBuilder {
	localImg := &localImage{
		Bytes: data,
		Alt:   alt,
	}
	pb.images = append(pb.images, localImg)
	return pb
}

func (pb *PostBuilder) BuildFor(server string, c *http.Client) (*postRequest, error) {
	createdAt := time.Now().UTC().Format(time.RFC3339)
	record := &record{
		Type:      "app.bsky.feed.post",
		Text:      pb.content,
		CreatedAt: createdAt,
		Langs:     pb.langs,
	}

	// Load images from disk
	for _, img := range pb.images {
		if img.Path != "" && len(img.Bytes) == 0 {
			dat, err := os.ReadFile(img.Path)
			if err != nil {
				log.Printf("Error reading image from path %s: %v", img.Path, err)
				continue
			}
			img.Bytes = dat
		}
	}

	pb.parseLinks()
	pb.parseMentions(server, c)
	if len(pb.facets) > 0 {
		record.Facets = make([]facet, len(pb.facets))
		for i, f := range pb.facets {
			record.Facets[i] = facet{
				Index:    f.Index,
				Features: f.Features,
			}
		}
	}

	return &postRequest{
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
		f := &facet{
			Features: []feature{
				{Type: "app.bsky.richtext.facet#link", Uri: link},
			},
		}
		f.Index.ByteStart = start
		f.Index.ByteEnd = end
		pb.facets = append(pb.facets, f)
	}
}

func (pb *PostBuilder) parseMentions(server string, c *http.Client) {
	// regex based on: https://atproto.com/specs/handle#handle-identifier-syntax
	handle_regex := `[$|\W](?P<handle>@([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)`
	r, err := regexp.Compile(handle_regex)
	if err != nil {
		log.Printf("Error compiling regex: %v", err)
		return
	}
	matches := r.FindAllSubmatchIndex([]byte(pb.content), -1)

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
		f := &facet{
			Features: []feature{
				{Type: "app.bsky.richtext.facet#mention", Did: resolveResponse.Did},
			},
		}
		f.Index.ByteStart = start
		f.Index.ByteEnd = end
		pb.facets = append(pb.facets, f)

	}
}

type postRequest struct {
	Repo       string  `json:"repo"`
	Collection string  `json:"collection"`
	Record     *record `json:"record"`
}

type record struct {
	Type      string   `json:"$type"`
	Text      string   `json:"text"`
	CreatedAt string   `json:"createdAt"`
	Langs     []string `json:"langs,omitempty"`
	Facets    []facet  `json:"facets,omitempty"`
	Embed     *struct {
		Type   string   `json:"$type"`
		Images []*image `json:"images,omitempty"`
	} `json:"embed,omitempty"`
}

type facet struct {
	Index struct {
		ByteStart int `json:"byteStart"`
		ByteEnd   int `json:"byteEnd"`
	} `json:"index"`
	Features []feature `json:"features"`
}

type feature struct {
	Type   string `json:"$type"`
	Handle string `json:"handle,omitempty"`
	Did    string `json:"did,omitempty"`
	Uri    string `json:"uri,omitempty"`
}

type image struct {
	Alt         string      `json:"alt"`
	Image       *imageEmbed `json:"image"`
	AspectRatio *struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"aspectRatio"`
}

type imageEmbed struct {
	Type string `json:"$type"`
	Ref  *struct {
		Link string `json:"$link,omitempty"`
	} `json:"ref"`
	Mimetype string `json:"mimeType,omitempty"`
	Size     int    `json:"size,omitempty"`
}

type localImage struct {
	Path  string
	Bytes []byte
	Alt   string
}

// Post creates a new post with the given content.
func (c *Client) Post(pb *PostBuilder) (string, error) {
	err := c.auth()
	if err != nil {
		return "", fmt.Errorf("error authenticating: %w", err)
	}

	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.createRecord", c.server)
	pr, err := pb.BuildFor(c.server, c.httpClient)
	if err != nil {
		return "", fmt.Errorf("error building post request: %w", err)
	}
	pr.Repo = c.handle // Set the repo to the user's handle

	err = c.embedImagesInPost(pb, pr)
	if err != nil {
		return "", fmt.Errorf("error embedding images in post: %w", err)
	}

	jsonBody, err := json.Marshal(pr)
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
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	b, err := io.ReadAll(resp.Body) // Read the response body to ensure we consume it
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	var postResponse struct {
		Uri   string `json:"uri"`
		Cid   string `json:"cid"`
		Error string `json:"error,omitempty"`
	}
	if err := json.Unmarshal(b, &postResponse); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("post failed with status code: %d (%s) error: %s", resp.StatusCode, resp.Status, postResponse.Error)
	}
	return postResponse.Uri, nil
}

// auth logs in to the server using the provided handle and password.
func (c *Client) auth() error {
	url := fmt.Sprintf("%s/xrpc/com.atproto.server.createSession", c.server)
	requestBody := map[string]string{
		"identifier": c.handle,
		"password":   c.password,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("error marshaling request body: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	// var sessionResponse SessionResponse
	var sessionResponse struct {
		AccessJwt  string `json:"accessJwt"`
		RefreshJwt string `json:"refreshJwt"`
	}
	if err := json.Unmarshal(b, &sessionResponse); err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}
	c.accessToken = sessionResponse.AccessJwt
	return err
}

func (c *Client) embedImagesInPost(pb *PostBuilder, pr *postRequest) error {
	if len(pb.images) == 0 {
		return nil
	}
	uploadUrl := fmt.Sprintf("%s/xrpc/com.atproto.repo.uploadBlob", c.server)

	// First, upload the images and save their references
	embeddedImages := make([]*image, 0, len(pb.images))
	for _, img := range pb.images {
		// If the image file size is too large, scale it until it is under 1MiB
		data := make([]byte, len(img.Bytes))
		copy(data, img.Bytes)
		scaleFactor := 1.0
		var err error
		for len(data) > 1_000_000 {
			scaleFactor *= 0.9 // Reduce size by 10% each iteration
			data, err = scaleImage(img.Bytes, scaleFactor)
			if err != nil {
				return fmt.Errorf("error scaling image: %w", err)
			}
		}

		// Figure out the image type and dimensions
		mimetype := http.DetectContentType(data)
		config, _, err := goimage.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			log.Printf("Error decoding image config: %v", err)
			continue
		}

		// Upload the image to the server
		req, err := http.NewRequest("POST", uploadUrl, bytes.NewBuffer(data))
		if err != nil {
			return fmt.Errorf("error creating upload request: %w", err)
		}
		req.Header.Set("Content-Type", mimetype)
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("error uploading image: %w", err)
		}
		defer func() {
			err = errors.Join(err, resp.Body.Close())
		}()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading upload response body: %w", err)
		}
		blob := struct {
			Blob imageEmbed `json:"blob"`
		}{}
		if err := json.Unmarshal(b, &blob); err != nil {
			return fmt.Errorf("error unmarshaling upload response: %w", err)
		}

		// Create the JSON object for this image
		image := &image{
			Image: &blob.Blob,
			Alt:   img.Alt,
			AspectRatio: &struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			}{
				Width:  config.Width,
				Height: config.Height,
			},
		}
		embeddedImages = append(embeddedImages, image)
	}

	// Then, embed the image references in the post record
	pr.Record.Embed = &struct {
		Type   string   "json:\"$type\""
		Images []*image "json:\"images,omitempty\""
	}{
		Type:   "app.bsky.embed.images",
		Images: embeddedImages,
	}

	return nil
}

func scaleImage(data []byte, scale float64) ([]byte, error) {
	src, format, err := goimage.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}
	dst := goimage.NewRGBA(goimage.Rect(0, 0, int(float64(src.Bounds().Dx())*scale), int(float64(src.Bounds().Dy())*scale)))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	buf := new(bytes.Buffer)
	switch format {
	case "jpeg":
		if err := jpeg.Encode(buf, dst, &jpeg.Options{Quality: 85}); err != nil {
			return nil, fmt.Errorf("error encoding JPEG image: %w", err)
		}
		return buf.Bytes(), nil
	case "gif":
		if err := gif.Encode(buf, dst, nil); err != nil {
			return nil, fmt.Errorf("error encoding GIF image: %w", err)
		}
		return buf.Bytes(), nil
	default:
		if err := png.Encode(buf, dst); err != nil {
			return nil, fmt.Errorf("error encoding PNG image: %w", err)
		}
		return buf.Bytes(), nil
	}
}
