package models

type PresignAvatarRequest struct {
	ContentType string `json:"content_type" validate:"required"`
}

type PresignAvatarResponse struct {
	UploadURL string `json:"upload_url"`
	ObjectKey string `json:"object_key"`
	AvatarURL string `json:"avatar_url"`
}
