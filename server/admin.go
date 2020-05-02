package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/valyala/fasthttp"
	"math"
	"net/url"
	"os"
	"runtime"
	"strings"
)

// AdminWrapper wraps requests to ensure you're admin.
func AdminWrapper(WrappedFunc func(ctx *fasthttp.RequestCtx)) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		unauthorized := func() {
			ctx.Response.SetStatusCode(403)
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBody([]byte("\"Unauthorized.\""))
		}
		TokenAuth := string(ctx.Request.Header.Peek("Token-Auth"))
		if TokenAuth == "" || TokenAuth != os.Getenv("ADMIN_TOKEN") {
			unauthorized()
			return
		}
		WrappedFunc(ctx)
	}
}

// Stats is used to get stats about the current deployment.
func Stats(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Set("Content-Type", "application/json")
	b, err := json.Marshal(&map[string]interface{}{
		"online": RedisClient.SCard("census").Val(),
		"installs": RedisClient.SCard("installs").Val(),
	})
	if err != nil {
		panic(err)
	}
	ctx.Response.SetBody(b)
}

func chunkUpdate(b []byte) [][]byte {
	// Defines the size of a chunk.
	Size := 15 * 1024 * 1024

	// Create the chunks array.
	chunks := make([][]byte, 0, int(math.Ceil(float64(len(b)) / float64(Size))))

	// Create the chunks.
	chunk := make([]byte, 0, Size)
	for _, v := range b {
		chunk = append(chunk, v)
		if len(chunk) >= Size {
			chunks = append(chunks, chunk)
			chunk = make([]byte, 0, Size)
		}
	}
	if len(chunk) != 0 {
		chunks = append(chunks, chunk)
	}

	// Return the chunks.
	return chunks
}

// UploadChunk is used to upload a update chunk.
func UploadChunk(UpdateHash, ChunkHash string, Chunk []byte) *string {
	// Create the URL item.
	u, _ := url.Parse(os.Getenv("UPDATE_CDN_URL"))

	// Defines the S3 key.
	S3Key := UpdateHash+"/"+ChunkHash

	// Set the URL path.
	u.Path = "/"+S3Key

	// Get the URL as a string.
	ChunkURL := u.String()

	// Initialise S3.
	StaticCredential := credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")
	e := os.Getenv("S3_ENDPOINT")
	r := os.Getenv("S3_REGION")
	s3sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    &e,
		Credentials: StaticCredential,
		Region:      &r,
	}))
	svc := s3.New(s3sess)

	// Upload the chunk to S3.
	MimeType := "application/octet-stream"
	ContentLength := int64(len(Chunk))
	Bucket := os.Getenv("S3_BUCKET")
	FileReader := bytes.NewReader(Chunk)
	UploadParams := &s3.PutObjectInput{
		Bucket:             &Bucket,
		Key:                &S3Key,
		ContentType:        &MimeType,
		Body:               FileReader,
		ACL:                aws.String("public-read"),
		ContentLength:      &ContentLength,
		ContentDisposition: aws.String("attachment"),
	}
	_, err := svc.PutObject(UploadParams)
	if err != nil {
		return nil
	}

	// Return the chunk URL.
	return &ChunkURL
}

// PushRelease is used to push a release to the database/chunk it for S3.
func PushRelease(ctx *fasthttp.RequestCtx)  {
	// !! POSSIBLE OPTIMISATION !!
	// Right now, Ironfist currently loads the whole release into RAM. Can we read it in chunks?

	// Get all other attributes from the headers.
	Channel := string(ctx.Request.Header.Peek("Ironfist-Update-Channel"))
	Version := string(ctx.Request.Header.Peek("Ironfist-Update-Version"))
	Changelogs := string(ctx.Request.Header.Peek("Ironfist-Update-Changelogs"))
	Name := string(ctx.Request.Header.Peek("Ironfist-Update-Name"))
	if Name == "" || Changelogs == "" || Version == "" || Channel == "" {
		ctx.Response.SetStatusCode(400)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.Response.SetBody([]byte("\"Missing required metadata.\""))
		return
	}

	// Get the body.
	b := ctx.Request.Body()

	// Create the hash for the update.
	bytearr := sha1.Sum(b)
	byteslice := make([]byte, 20)
	for i, v := range bytearr {
		byteslice[i] = v
	}
	UpdateHash := base64.StdEncoding.EncodeToString(byteslice)
	UpdateHash = strings.ReplaceAll(UpdateHash, "/", "_")
	UpdateHash = strings.ReplaceAll(UpdateHash, "+", "-")

	// Turn the update into chunks.
	Chunks := chunkUpdate(b)

	// Run the GC now to reduce the affects.
	runtime.GC()

	// Defines the array of uploaded chunks.
	UploadedChunks := make([]UpdateChunk, len(Chunks))
	for i, v := range Chunks {
		w := bytes.Buffer{}
		gzipw := gzip.NewWriter(&w)
		_, err := gzipw.Write(v)
		if err != nil {
			panic(err)
		}
		v = w.Bytes()
		bytearr := sha1.Sum(v)
		byteslice := make([]byte, 20)
		for i, v := range bytearr {
			byteslice[i] = v
		}
		B64Encoded := base64.StdEncoding.EncodeToString(byteslice)
		B64Encoded = strings.ReplaceAll(B64Encoded, "/", "_")
		B64Encoded = strings.ReplaceAll(B64Encoded, "+", "-")
		r := UploadChunk(UpdateHash, B64Encoded, v)
		if r == nil {
			ctx.Response.SetStatusCode(500)
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBody([]byte("\"Upload error.\""))
			return
		}
		UploadedChunks[i] = UpdateChunk{
			URL:  *r,
			Hash: B64Encoded,
		}
	}

	// Push the update to the updates API.
	u := Update{
		UpdateHash: UpdateHash,
		Channel:    Channel,
		Chunks:     UploadedChunks,
		Version:    Version,
		Changelogs: Changelogs,
		Name:       Name,
	}
	err := u.PushUpdate()
	if err != nil {
		ctx.Response.SetStatusCode(500)
		ctx.Response.Header.Set("Content-Type", "application/json")
		b, _ = json.Marshal(err.Error())
		ctx.Response.SetBody(b)
		return
	}

	// Return the JSON.
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Set("Content-Type", "application/json")
	b, _ = json.Marshal(&u)
	ctx.Response.SetBody(b)
}

// Initialises the admin routes.
func init() {
	Router.GET("/v1/admin/stats", AdminWrapper(Stats))
	Router.POST("/v1/admin/push", AdminWrapper(PushRelease))
}
