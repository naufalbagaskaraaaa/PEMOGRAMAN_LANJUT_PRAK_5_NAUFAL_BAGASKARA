package helper

import (
	"time"

	"latihan-fiber/app/model"

	"github.com/golang-jwt/jwt/v5"
)

const LocalsAuthUser = "auth_user"

type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), ttl: ttl}
}

func (m *JWTManager) Issue(user model.AuthUser) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      user.UserID,
		"username": user.Username,
		"role":     user.Role,
		"iat":      now.Unix(),
		"exp":      now.Add(m.ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *JWTManager) Parse(tokenString string) (model.AuthUser, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return m.secret, nil
	})
	if err != nil {
		return model.AuthUser{}, err
	}
	if !token.Valid {
		return model.AuthUser{}, jwt.ErrTokenInvalidClaims
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return model.AuthUser{}, jwt.ErrTokenInvalidClaims
	}
	userID, ok := claims["sub"].(float64)
	if !ok {
		return model.AuthUser{}, jwt.ErrTokenInvalidClaims
	}
	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)
	return model.AuthUser{UserID: int(userID), Username: username, Role: role}, nil
}
