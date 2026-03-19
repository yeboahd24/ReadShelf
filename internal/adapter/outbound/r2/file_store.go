package r2

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dominic/readshelf/internal/core/port/outbound"
)

type fileStore struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucketName string
}

func NewFileStore(accountID, accessKeyID, secretAccessKey, bucketName string) outbound.FileStore {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String(endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
	})

	return &fileStore{
		client:     client,
		presigner:  s3.NewPresignClient(client),
		bucketName: bucketName,
	}
}

func (f *fileStore) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	_, err := f.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(f.bucketName),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	return err
}

func (f *fileStore) SignedURL(ctx context.Context, key string) (string, error) {
	resp, err := f.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(f.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", err
	}
	return resp.URL, nil
}

func (f *fileStore) Delete(ctx context.Context, key string) error {
	_, err := f.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(f.bucketName),
		Key:    aws.String(key),
	})
	return err
}
