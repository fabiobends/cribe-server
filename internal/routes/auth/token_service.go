package auth

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTObject struct {
	UserID int    `json:"user_id"`
	Exp    int64  `json:"exp"`
	Typ    string `json:"typ"`
}

type TokenService interface {
	GenerateHash(text string) (string, error)
	CompareHashAndPassword(hashedPassword, password string) error
	ValidateToken(token string) (*JWTObject, error)
	GetAccessToken(userID int) (string, error)
	GetRefreshToken(userID int) (string, error)
}

type TokenServiceImpl struct {
	secretKey              []byte
	accessTokenExpiration  time.Duration
	refreshTokenExpiration time.Duration
	currentTime            func() time.Time
}

func NewTokenServiceReady() TokenService {
	secretKey := os.Getenv("JWT_SECRET")
	accessTokenExpiration := os.Getenv("JWT_ACCESS_TOKEN_EXPIRATION_TIME_IN_MINUTES")
	refreshTokenExpiration := os.Getenv("JWT_REFRESH_TOKEN_EXPIRATION_TIME_IN_DAYS")
	accessTokenExpirationInt, err := strconv.Atoi(accessTokenExpiration)
	if err != nil {
		log.Println("Error converting access token expiration to int:", err)
		return nil
	}
	refreshTokenExpirationInt, err := strconv.Atoi(refreshTokenExpiration)
	if err != nil {
		log.Println("Error converting refresh token expiration to int:", err)
		return nil
	}

	return NewTokenService([]byte(secretKey), time.Duration(accessTokenExpirationInt)*time.Minute, time.Duration(refreshTokenExpirationInt)*time.Hour*24, time.Now)
}

func NewTokenService(secretKey []byte, accessTokenExpiration, refreshTokenExpiration time.Duration, currentTime func() time.Time) TokenService {
	return &TokenServiceImpl{
		secretKey:              secretKey,
		accessTokenExpiration:  accessTokenExpiration,
		refreshTokenExpiration: refreshTokenExpiration,
		currentTime:            currentTime,
	}
}

func (s *TokenServiceImpl) GenerateHash(text string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (s *TokenServiceImpl) CompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (s *TokenServiceImpl) ValidateToken(accessToken string) (*JWTObject, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (any, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, errors.New("Failed to parse token")
	}

	if !token.Valid {
		return nil, errors.New("Invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Invalid token")
	}

	return &JWTObject{
		UserID: claims["user_id"].(int),
		Exp:    claims["exp"].(int64),
		Typ:    claims["typ"].(string),
	}, nil
}

func (s *TokenServiceImpl) GetAccessToken(userID int) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     s.currentTime().Add(s.accessTokenExpiration).Unix(),
		"typ":     "access",
	})

	return accessToken.SignedString(s.secretKey)
}

func (s *TokenServiceImpl) GetRefreshToken(userID int) (string, error) {
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     s.currentTime().Add(s.refreshTokenExpiration).Unix(),
		"typ":     "refresh",
	})

	return refreshToken.SignedString(s.secretKey)
}
