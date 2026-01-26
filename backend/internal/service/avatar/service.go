package avatar

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"time"
)

type Config struct {
	S3         *s3.Client
	Bucket     string
	PresignTTL time.Duration
	PublicURL  string
}

type Service struct {
	s3         *s3.Client
	bucket     string
	presignTTL time.Duration
	publicURL  string
}

func NewService(cfg Config) *Service {
	return &Service{
		s3:         cfg.S3,
		bucket:     cfg.Bucket,
		presignTTL: cfg.PresignTTL,
		publicURL:  cfg.PublicURL,
	}
}

func (s *Service) PresignPutAvatar(ctx context.Context, userID string, contentType string) (uploadURL string, objectKey string, err error) {

	key := fmt.Sprintf("%s", userID)

	presignClient := s3.NewPresignClient(s.s3)

	req, err := presignClient.PresignPutObject(
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

func (s *Service) PublicAvatarURL(objectKey string) string {
	if s.publicURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", s.publicURL, objectKey)
}
