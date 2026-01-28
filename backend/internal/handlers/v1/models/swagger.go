package models

type ErrorResponse struct {
	Error string `json:"{route} error" example:"{method}: {crud}: error message"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"Successfully {action}"`
}
