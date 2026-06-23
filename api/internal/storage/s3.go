package storage

// ⊹ ࣪ ˖ s3-compatible storage

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	client   *s3.Client
	bucket   string
	endpoint string
}

type S3Config struct {
	Endpoint  string
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
}

func NewS3(cfg S3Config) *S3 {
	client := s3.New(s3.Options{
		BaseEndpoint: &cfg.Endpoint,
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		UsePathStyle: true,
	})
	return &S3{
		client:   client,
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
	}
}

func (s *S3) Upload(ctx context.Context, key string, r io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   r,
	})
	if err != nil {
		return fmt.Errorf("s3: upload %s: %w", key, err)
	}
	return nil
}

func (s *S3) Download(ctx context.Context, key string, w io.Writer) error {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("s3: download %s: %w", key, err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("s3: read %s: %w", key, err)
	}
	return nil
}

func (s *S3) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("s3: delete %s: %w", key, err)
	}
	return nil
}

func (s *S3) URL(key string) string {
	return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
}
