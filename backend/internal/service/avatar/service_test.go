package avatar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ── mocks ──────────────────────────────────────────────────────────────────

type mockPresigner struct{ mock.Mock }

func (m *mockPresigner) PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*v4.PresignedHTTPRequest), args.Error(1)
}

type mockUploader struct{ mock.Mock }

func (m *mockUploader) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

// ── helpers ────────────────────────────────────────────────────────────────

func newService(presigner S3Presigner, uploader S3Uploader) *Service {
	return &Service{
		presigner:  presigner,
		uploader:   uploader,
		bucket:     "test-bucket",
		presignTTL: 15 * time.Minute,
		publicURL:  "https://cdn.example.com/test-bucket",
	}
}

// ── PublicAvatarURL ────────────────────────────────────────────────────────

func TestPublicAvatarURL(t *testing.T) {
	tests := []struct {
		name      string
		publicURL string
		key       string
		want      string
	}{
		{
			name:      "returns full URL",
			publicURL: "https://cdn.example.com/bucket",
			key:       "user-123",
			want:      "https://cdn.example.com/bucket/user-123",
		},
		{
			name:      "empty publicURL returns empty string",
			publicURL: "",
			key:       "user-123",
			want:      "",
		},
		{
			name:      "nested key",
			publicURL: "https://cdn.example.com/bucket",
			key:       "food-photos/user-id/photo.jpg",
			want:      "https://cdn.example.com/bucket/food-photos/user-id/photo.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{publicURL: tt.publicURL}
			assert.Equal(t, tt.want, svc.PublicAvatarURL(tt.key))
		})
	}
}

// ── PresignPutAvatar ───────────────────────────────────────────────────────

func TestPresignPutAvatar_Success(t *testing.T) {
	ctx := context.Background()
	presigner := &mockPresigner{}
	presigner.On("PresignPutObject", mock.Anything, mock.MatchedBy(func(p *s3.PutObjectInput) bool {
		return *p.Bucket == "test-bucket" && *p.Key == "user-42" && *p.ContentType == "image/jpeg"
	})).Return(&v4.PresignedHTTPRequest{URL: "https://s3.example.com/presigned"}, nil)

	svc := newService(presigner, nil)
	url, key, err := svc.PresignPutAvatar(ctx, "user-42", "image/jpeg")

	assert.NoError(t, err)
	assert.Equal(t, "https://s3.example.com/presigned", url)
	assert.Equal(t, "user-42", key)
	presigner.AssertExpectations(t)
}

func TestPresignPutAvatar_Error(t *testing.T) {
	ctx := context.Background()
	presigner := &mockPresigner{}
	presigner.On("PresignPutObject", mock.Anything, mock.Anything).
		Return(nil, errors.New("s3 error"))

	svc := newService(presigner, nil)
	_, _, err := svc.PresignPutAvatar(ctx, "user-42", "image/jpeg")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "presign put avatar")
}

// ── UploadFoodPhoto ────────────────────────────────────────────────────────

func TestUploadFoodPhoto_Success(t *testing.T) {
	ctx := context.Background()
	uploader := &mockUploader{}
	uploader.On("PutObject", mock.Anything, mock.MatchedBy(func(p *s3.PutObjectInput) bool {
		return *p.Bucket == "test-bucket" &&
			strings.HasPrefix(*p.Key, "food-photos/user-99/") &&
			strings.HasSuffix(*p.Key, "meal.jpg") &&
			*p.ContentType == "image/jpeg"
	})).Return(&s3.PutObjectOutput{}, nil)

	svc := newService(nil, uploader)
	photoURL, err := svc.UploadFoodPhoto(ctx, "user-99", "meal.jpg", "image/jpeg", bytes.NewReader([]byte("img")))

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(photoURL, "https://cdn.example.com/test-bucket/food-photos/user-99/"))
	uploader.AssertExpectations(t)
}

func TestUploadFoodPhoto_UploaderError(t *testing.T) {
	ctx := context.Background()
	uploader := &mockUploader{}
	uploader.On("PutObject", mock.Anything, mock.Anything).
		Return(nil, errors.New("s3 write error"))

	svc := newService(nil, uploader)
	_, err := svc.UploadFoodPhoto(ctx, "user-99", "meal.jpg", "image/jpeg", bytes.NewReader([]byte("img")))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload food photo")
}

func TestUploadFoodPhoto_NoUploader(t *testing.T) {
	svc := &Service{}
	_, err := svc.UploadFoodPhoto(context.Background(), "u", "f.jpg", "image/jpeg", bytes.NewReader(nil))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uploader not configured")
}

func TestUploadFoodPhoto_ObjectKeyContainsUserAndFilename(t *testing.T) {
	ctx := context.Background()
	uploader := &mockUploader{}

	var capturedKey string
	uploader.On("PutObject", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			p := args.Get(1).(*s3.PutObjectInput)
			capturedKey = *p.Key
		}).
		Return(&s3.PutObjectOutput{}, nil)

	svc := newService(nil, uploader)
	svc.publicURL = "https://cdn.example.com/bucket"
	_, _ = svc.UploadFoodPhoto(ctx, "abc", "photo.png", "image/png", bytes.NewReader(nil))

	assert.Equal(t, fmt.Sprintf("food-photos/abc/photo.png"), capturedKey)
}

// ── NewService wiring ──────────────────────────────────────────────────────

func TestNewService_NilS3_KeepsNilPresignerUploader(t *testing.T) {
	svc := NewService(Config{
		Bucket:     "b",
		PresignTTL: time.Minute,
		PublicURL:  "https://cdn",
	})
	assert.Nil(t, svc.presigner)
	assert.Nil(t, svc.uploader)
}

func TestNewService_ExplicitPresignerUploader(t *testing.T) {
	p := &mockPresigner{}
	u := &mockUploader{}
	svc := NewService(Config{
		Presigner: p,
		Uploader:  u,
		Bucket:    "b",
	})
	assert.Equal(t, p, svc.presigner)
	assert.Equal(t, u, svc.uploader)
}
