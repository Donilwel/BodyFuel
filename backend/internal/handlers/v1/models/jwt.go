package models

import "backend/internal/service/auth"

type JWTModel struct {
	JWT string `json:"jwt"`
}

func NewJWTCodeResponse(jwt string) JWTModel {
	return JWTModel{JWT: jwt}
}

type TokenPairModel struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewTokenPairModel(pair auth.TokenPair) TokenPairModel {
	return TokenPairModel{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}
}
