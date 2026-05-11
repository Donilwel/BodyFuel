package models

type SendFeedbackRequest struct {
	Message string `json:"message" binding:"required,min=10,max=2000"`
	Email   string `json:"email" binding:"omitempty,email"`
}

type FeedbackResponse struct {
	Message string `json:"message"`
}
