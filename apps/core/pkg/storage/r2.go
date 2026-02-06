package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2Config holds Cloudflare R2 configuration
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
}

// R2Client wraps the S3 client for R2 operations
type R2Client struct {
	client     *s3.Client
	presigner  *s3.PresignClient
	bucketName string
	publicURL  string
}

// NewR2Client creates a new R2 client
func NewR2Client(cfg R2Config) (*R2Client, error) {
	r2Endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	s3Client := s3.New(s3.Options{
		BaseEndpoint: aws.String(r2Endpoint),
		Region:       "auto",
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	})

	return &R2Client{
		client:     s3Client,
		presigner:  s3.NewPresignClient(s3Client),
		bucketName: cfg.BucketName,
		publicURL:  cfg.PublicURL,
	}, nil
}

// PresignedURLResult contains the generated presigned URL and related info
type PresignedURLResult struct {
	UploadURL string
	FileURL   string
	ExpiresAt time.Time
}

// GeneratePresignedURL generates a presigned PUT URL for uploading to R2
func (c *R2Client) GeneratePresignedURL(ctx context.Context, objectKey, contentType string) (*PresignedURLResult, error) {
	expiresIn := 15 * time.Minute

	presignedReq, err := c.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	fileURL := fmt.Sprintf("%s/%s", c.publicURL, objectKey)

	return &PresignedURLResult{
		UploadURL: presignedReq.URL,
		FileURL:   fileURL,
		ExpiresAt: time.Now().Add(expiresIn),
	}, nil
}

// GetPublicURL returns the public URL for an object
func (c *R2Client) GetPublicURL(objectKey string) string {
	return fmt.Sprintf("%s/%s", c.publicURL, objectKey)
}
