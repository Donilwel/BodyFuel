package models

type JWTModel struct {
	JWT string `json:"jwt"`
}

func NewJWTCodeResponse(jwt string) JWTModel {
	return JWTModel{JWT: jwt}
}
