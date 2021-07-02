package filestore

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type GCS struct {
	client     *storage.Client
	bucketName string
	bucket     *storage.BucketHandle
}

func NewGCS(bucketName string, location string, projectID string, serviceAccountFile string) (*GCS, error) {
	g := &GCS{
		bucketName: bucketName,
	}

	ctx := context.Background()

	var clientOptions option.ClientOption
	if serviceAccountFile != "" {
		jsonKey, err := ioutil.ReadFile(serviceAccountFile)
		if err != nil {
			return nil, err
		}

		conf, err := google.JWTConfigFromJSON(jsonKey, storage.ScopeReadWrite)
		if err != nil {
			return nil, err
		}

		ts := conf.TokenSource(ctx)
		clientOptions = option.WithTokenSource(ts)
	} else {
		// defaultClient, err := google.DefaultClient(ctx, storage.ScopeReadWrite)
		// if err != nil {
		// 	return nil, err
		// }
		clientOptions = option.WithScopes(storage.ScopeReadWrite)
	}

	var err error
	g.client, err = storage.NewClient(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	var bucketExists bool
	it := g.client.Buckets(ctx, projectID)
	for {
		bucketAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		if bucketAttrs.Name == bucketName {
			bucketExists = true
		}
	}

	g.bucket = g.client.Bucket(bucketName)

	if !bucketExists {
		err = g.bucket.Create(ctx, projectID, &storage.BucketAttrs{
			Location: location,
		})
		if err != nil {
			return nil, err
		}
	}

	return g, nil
}

func (g *GCS) Read(name string) ([]byte, error) {

	obj := g.bucket.Object(name)

	reader, err := obj.NewReader(context.Background())
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(reader)
}

func (g *GCS) Write(name string, data io.Reader) (string, error) {

	obj := g.bucket.Object(name)

	w := obj.NewWriter(context.Background())
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.ContentType = mime.TypeByExtension(name)

	_, err := io.Copy(w, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.bucketName, name), w.Close()
}

func (g *GCS) Delete(name string) error {
	return g.bucket.Object(name).Delete(context.Background())
}

func (g *GCS) DeleteBucket() error {
	return g.bucket.Delete(context.Background())
}
