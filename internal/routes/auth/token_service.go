package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"cribeapp.com/cribe-server/internal/core/logger"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTObject struct {
	UserID int    `json:"user_id"`
	Exp    int64  `json:"exp"`
	Typ    string `json:"typ"`
}

type JWTClaims struct {
	UserID int    `json:"user_id"`
	Typ    string `json:"typ"`
	jwt.RegisteredClaims
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
	log := logger.NewServiceLogger("TokenService")
	secretKey := os.Getenv("JWT_SECRET")
	accessTokenExpiration := os.Getenv("JWT_ACCESS_TOKEN_EXPIRATION_TIME_IN_MINUTES")
	refreshTokenExpiration := os.Getenv("JWT_REFRESH_TOKEN_EXPIRATION_TIME_IN_DAYS")
	accessTokenExpirationInt, err := strconv.Atoi(accessTokenExpiration)
	if err != nil {
		log.Error("Error converting access token expiration to int", map[string]any{
			"error": err.Error(),
			"value": accessTokenExpiration,
		})
		return nil
	}
	refreshTokenExpirationInt, err := strconv.Atoi(refreshTokenExpiration)
	if err != nil {
		log.Error("Error converting refresh token expiration to int", map[string]any{
			"error": err.Error(),
			"value": refreshTokenExpiration,
		})
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
	token, err := jwt.ParseWithClaims(accessToken, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, errors.New("failed to parse token")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return &JWTObject{
		UserID: claims.UserID,
		Exp:    claims.ExpiresAt.Unix(),
		Typ:    claims.Typ,
	}, nil
}

func (s *TokenServiceImpl) GetAccessToken(userID int) (string, error) {
	now := s.currentTime()
	claims := &JWTClaims{
		UserID: userID,
		Typ:    "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *TokenServiceImpl) GetRefreshToken(userID int) (string, error) {
	now := s.currentTime()
	claims := &JWTClaims{
		UserID: userID,
		Typ:    "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}
