package avatar

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Presigner is the minimal interface for creating presigned PUT URLs.
// Satisfied by *s3.PresignClient.
type S3Presigner interface {
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

// S3Uploader is the minimal interface for server-side object uploads.
// Satisfied by *s3.Client.
type S3Uploader interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type Config struct {
	S3         *s3.Client // kept for backwards-compat wiring in app.go
	Presigner  S3Presigner
	Uploader   S3Uploader
	Bucket     string
	PresignTTL time.Duration
	PublicURL  string
}

type Service struct {
	presigner  S3Presigner
	uploader   S3Uploader
	bucket     string
	presignTTL time.Duration
	publicURL  string
}

func NewService(cfg Config) *Service {
	presigner := cfg.Presigner
	uploader := cfg.Uploader

	// Convenience: if raw *s3.Client provided, derive presigner/uploader from it.
	if cfg.S3 != nil {
		if presigner == nil {
			presigner = s3.NewPresignClient(cfg.S3)
		}
		if uploader == nil {
			uploader = cfg.S3
		}
	}

	return &Service{
		presigner:  presigner,
		uploader:   uploader,
		bucket:     cfg.Bucket,
		presignTTL: cfg.PresignTTL,
		publicURL:  cfg.PublicURL,
	}
}

// PresignPutAvatar returns a presigned PUT URL for uploading a user avatar.
func (s *Service) PresignPutAvatar(ctx context.Context, userID string, contentType string) (uploadURL string, objectKey string, err error) {
	key := userID

	req, err := s.presigner.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(s.bucket),
			Key:         aws.String(key),
			ContentType: aws.String(contentType),
		},
		s3.WithPresignExpires(s.presignTTL),
	)
	if err != nil {
		return "", "", fmt.Errorf("presign put avatar: %w", err)
	}

	return req.URL, key, nil
}

// PublicAvatarURL returns the public URL for a stored object key.
func (s *Service) PublicAvatarURL(objectKey string) string {
	if s.publicURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", s.publicURL, objectKey)
}

// UploadFoodPhoto uploads a food photo directly from a reader and returns the public URL.
// The object key is: food-photos/<userID>/<objectName>
func (s *Service) UploadFoodPhoto(ctx context.Context, userID, objectName, contentType string, data io.Reader) (string, error) {
	if s.uploader == nil {
		return "", fmt.Errorf("uploader not configured")
	}

	key := fmt.Sprintf("food-photos/%s/%s", userID, objectName)

	_, err := s.uploader.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        data,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload food photo: %w", err)
	}

	return s.PublicAvatarURL(key), nil
}
