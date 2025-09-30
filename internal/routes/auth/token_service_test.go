package auth

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenService_GenerateHash(t *testing.T) {
	service := NewTokenService([]byte("test-secret"), time.Hour, time.Hour*24, time.Now)

	t.Run("should generate hash successfully", func(t *testing.T) {
		password := "testpassword123"
		hash, err := service.GenerateHash(password)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if hash == "" {
			t.Error("Expected non-empty hash")
		}

		if hash == password {
			t.Error("Hash should not be the same as the original password")
		}
	})

	t.Run("should generate different hashes for same password", func(t *testing.T) {
		password := "testpassword123"
		hash1, err1 := service.GenerateHash(password)
		hash2, err2 := service.GenerateHash(password)

		if err1 != nil || err2 != nil {
			t.Fatalf("Expected no errors, got %v, %v", err1, err2)
		}

		if hash1 == hash2 {
			t.Error("Expected different hashes for same password due to salt")
		}
	})

	t.Run("should handle empty password", func(t *testing.T) {
		hash, err := service.GenerateHash("")

		if err != nil {
			t.Fatalf("Expected no error for empty password, got %v", err)
		}

		if hash == "" {
			t.Error("Expected non-empty hash even for empty password")
		}
	})
}

func TestTokenService_CompareHashAndPassword(t *testing.T) {
	service := NewTokenService([]byte("test-secret"), time.Hour, time.Hour*24, time.Now)

	t.Run("should return nil for matching password and hash", func(t *testing.T) {
		password := "testpassword123"
		hash, _ := service.GenerateHash(password)

		err := service.CompareHashAndPassword(hash, password)

		if err != nil {
			t.Errorf("Expected nil error for matching password, got %v", err)
		}
	})

	t.Run("should return error for non-matching password", func(t *testing.T) {
		password := "testpassword123"
		wrongPassword := "wrongpassword"
		hash, _ := service.GenerateHash(password)

		err := service.CompareHashAndPassword(hash, wrongPassword)

		if err == nil {
			t.Error("Expected error for non-matching password")
		}
	})

	t.Run("should return error for invalid hash", func(t *testing.T) {
		invalidHash := "invalid-hash"
		password := "testpassword123"

		err := service.CompareHashAndPassword(invalidHash, password)

		if err == nil {
			t.Error("Expected error for invalid hash")
		}
	})

	t.Run("should handle empty password correctly", func(t *testing.T) {
		emptyPassword := ""
		hash, _ := service.GenerateHash(emptyPassword)

		err := service.CompareHashAndPassword(hash, emptyPassword)

		if err != nil {
			t.Errorf("Expected nil error for matching empty password, got %v", err)
		}

		err = service.CompareHashAndPassword(hash, "nonempty")

		if err == nil {
			t.Error("Expected error when comparing empty hash with non-empty password")
		}
	})
}

func TestTokenService_GetAccessToken(t *testing.T) {
	fixedTime := time.Now().Add(-time.Minute) // Use a recent time to avoid expiration issues
	service := NewTokenService([]byte("test-secret"), time.Hour, time.Hour*24, func() time.Time {
		return fixedTime
	})

	t.Run("should generate valid access token", func(t *testing.T) {
		userID := 123
		token, err := service.GetAccessToken(userID)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}

		// Verify token can be parsed
		parsedToken, parseErr := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})

		if parseErr != nil {
			t.Fatalf("Failed to parse generated token: %v", parseErr)
		}

		if !parsedToken.Valid {
			t.Error("Generated token is not valid")
		}

		claims, ok := parsedToken.Claims.(*JWTClaims)
		if !ok {
			t.Fatal("Failed to extract claims from token")
		}

		if claims.UserID != userID {
			t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
		}

		if claims.Typ != "access" {
			t.Errorf("Expected token type 'access', got %s", claims.Typ)
		}

		expectedExp := fixedTime.Add(time.Hour).Unix()
		if claims.ExpiresAt.Unix() != expectedExp {
			t.Errorf("Expected expiration %d, got %d", expectedExp, claims.ExpiresAt.Unix())
		}
	})

	t.Run("should generate different tokens for different users", func(t *testing.T) {
		token1, err1 := service.GetAccessToken(123)
		token2, err2 := service.GetAccessToken(456)

		if err1 != nil || err2 != nil {
			t.Fatalf("Expected no errors, got %v, %v", err1, err2)
		}

		if token1 == token2 {
			t.Error("Expected different tokens for different users")
		}
	})

	t.Run("should generate different tokens for same user at different times", func(t *testing.T) {
		userID := 123
		baseTime := time.Now().Add(-time.Minute)

		// First token at base time
		service1 := NewTokenService([]byte("test-secret"), time.Hour, time.Hour*24, func() time.Time {
			return baseTime
		})
		token1, _ := service1.GetAccessToken(userID)

		// Second token at different time
		service2 := NewTokenService([]byte("test-secret"), time.Hour, time.Hour*24, func() time.Time {
			return baseTime.Add(time.Minute)
		})
		token2, _ := service2.GetAccessToken(userID)

		if token1 == token2 {
			t.Error("Expected different tokens for same user at different times")
		}
	})
}

func TestTokenService_GetRefreshToken(t *testing.T) {
	fixedTime := time.Now().Add(-time.Minute) // Use a recent time to avoid expiration issues
	service := NewTokenService([]byte("test-secret"), time.Hour, time.Hour*24, func() time.Time {
		return fixedTime
	})

	t.Run("should generate valid refresh token", func(t *testing.T) {
		userID := 123
		token, err := service.GetRefreshToken(userID)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}

		// Verify token can be parsed
		parsedToken, parseErr := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})

		if parseErr != nil {
			t.Fatalf("Failed to parse generated token: %v", parseErr)
		}

		if !parsedToken.Valid {
			t.Error("Generated token is not valid")
		}

		claims, ok := parsedToken.Claims.(*JWTClaims)
		if !ok {
			t.Fatal("Failed to extract claims from token")
		}

		if claims.UserID != userID {
			t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
		}

		if claims.Typ != "refresh" {
			t.Errorf("Expected token type 'refresh', got %s", claims.Typ)
		}

		expectedExp := fixedTime.Add(time.Hour * 24).Unix()
		if claims.ExpiresAt.Unix() != expectedExp {
			t.Errorf("Expected expiration %d, got %d", expectedExp, claims.ExpiresAt.Unix())
		}
	})

	t.Run("should have longer expiration than access token", func(t *testing.T) {
		userID := 123
		accessToken, _ := service.GetAccessToken(userID)
		refreshToken, _ := service.GetRefreshToken(userID)

		// Parse both tokens
		accessClaims := &JWTClaims{}
		refreshClaims := &JWTClaims{}

		_, _ = jwt.ParseWithClaims(accessToken, accessClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})

		_, _ = jwt.ParseWithClaims(refreshToken, refreshClaims, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})

		if refreshClaims.ExpiresAt.Unix() <= accessClaims.ExpiresAt.Unix() {
			t.Error("Refresh token should have longer expiration than access token")
		}
	})
}

func TestTokenService_ValidateToken(t *testing.T) {
	fixedTime := time.Now().Add(-time.Minute) // Use a recent time to avoid expiration issues
	service := NewTokenService([]byte("test-secret"), time.Hour, time.Hour*24, func() time.Time {
		return fixedTime
	})

	t.Run("should validate valid access token", func(t *testing.T) {
		userID := 123
		token, _ := service.GetAccessToken(userID)

		jwtObject, err := service.ValidateToken(token)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if jwtObject == nil {
			t.Fatal("Expected non-nil JWT object")
		}

		if jwtObject.UserID != userID {
			t.Errorf("Expected UserID %d, got %d", userID, jwtObject.UserID)
		}

		if jwtObject.Typ != "access" {
			t.Errorf("Expected token type 'access', got %s", jwtObject.Typ)
		}

		expectedExp := fixedTime.Add(time.Hour).Unix()
		if jwtObject.Exp != expectedExp {
			t.Errorf("Expected expiration %d, got %d", expectedExp, jwtObject.Exp)
		}
	})

	t.Run("should validate valid refresh token", func(t *testing.T) {
		userID := 456
		token, _ := service.GetRefreshToken(userID)

		jwtObject, err := service.ValidateToken(token)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if jwtObject.UserID != userID {
			t.Errorf("Expected UserID %d, got %d", userID, jwtObject.UserID)
		}

		if jwtObject.Typ != "refresh" {
			t.Errorf("Expected token type 'refresh', got %s", jwtObject.Typ)
		}
	})

	t.Run("should reject token with invalid signature", func(t *testing.T) {
		// Create token with different secret
		wrongService := NewTokenService([]byte("wrong-secret"), time.Hour, time.Hour*24, func() time.Time {
			return fixedTime
		})
		token, _ := wrongService.GetAccessToken(123)

		// Try to validate with correct service
		jwtObject, err := service.ValidateToken(token)

		if err == nil {
			t.Error("Expected error for token with invalid signature")
		}

		if jwtObject != nil {
			t.Error("Expected nil JWT object for invalid token")
		}
	})

	t.Run("should reject malformed token", func(t *testing.T) {
		malformedToken := "invalid.token.string"

		jwtObject, err := service.ValidateToken(malformedToken)

		if err == nil {
			t.Error("Expected error for malformed token")
		}

		if jwtObject != nil {
			t.Error("Expected nil JWT object for malformed token")
		}
	})

	t.Run("should reject empty token", func(t *testing.T) {
		jwtObject, err := service.ValidateToken("")

		if err == nil {
			t.Error("Expected error for empty token")
		}

		if jwtObject != nil {
			t.Error("Expected nil JWT object for empty token")
		}
	})

	t.Run("should reject expired token", func(t *testing.T) {
		pastTime := time.Now().Add(-2 * time.Hour) // Create token in the past

		// Create a service with very short expiration
		shortService := NewTokenService([]byte("test-secret"), time.Nanosecond, time.Nanosecond, func() time.Time {
			return pastTime
		})
		token, _ := shortService.GetAccessToken(123)

		// Token should be expired by now when validated
		jwtObject, err := service.ValidateToken(token)

		if err == nil {
			t.Error("Expected error for expired token")
		}

		if jwtObject != nil {
			t.Error("Expected nil JWT object for expired token")
		}
	})

	t.Run("should reject token with wrong signing method", func(t *testing.T) {
		// Create a manually crafted invalid token that claims to use RS256 algorithm
		// This token claims to use RS256 algorithm but has an invalid signature
		invalidToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjMsInR5cCI6ImFjY2VzcyIsImV4cCI6MTcwNDEwODAwMCwiaWF0IjoxNzA0MTA0NDAwfQ.invalid_signature"

		jwtObject, err := service.ValidateToken(invalidToken)

		if err == nil {
			t.Error("Expected error for token with wrong signing method")
		}

		if jwtObject != nil {
			t.Error("Expected nil JWT object for token with wrong signing method")
		}
	})
}

func TestNewTokenService(t *testing.T) {
	t.Run("should create token service with provided parameters", func(t *testing.T) {
		secretKey := []byte("test-secret")
		accessExp := time.Hour
		refreshExp := time.Hour * 24

		service := NewTokenService(secretKey, accessExp, refreshExp, time.Now)

		if service == nil {
			t.Fatal("Expected non-nil token service")
		}

		// Test that the service works
		token, err := service.GetAccessToken(123)
		if err != nil {
			t.Errorf("Expected service to work, got error: %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}
	})
}

// Test error cases that might occur in real scenarios
func TestTokenService_ErrorCases(t *testing.T) {
	service := NewTokenService([]byte("test-secret"), time.Hour, time.Hour*24, time.Now)

	t.Run("should handle bcrypt errors gracefully", func(t *testing.T) {
		// Test with extremely long password that might cause bcrypt to fail
		longPassword := string(make([]byte, 100)) // bcrypt has a limit of 72 bytes
		hash, err := service.GenerateHash(longPassword)

		// bcrypt should handle this, but let's verify it doesn't panic
		if err != nil {
			t.Logf("bcrypt returned error for long password (expected): %v", err)
		} else if hash == "" {
			t.Error("Expected non-empty hash even for long password")
		}
	})

	t.Run("should handle token validation edge cases", func(t *testing.T) {
		testCases := []struct {
			name  string
			token string
		}{
			{"random string", "random-string"},
			{"partial JWT", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
			{"JWT with missing signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxMjN9"},
			{"JWT with invalid base64", "invalid.base64!.signature"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				jwtObject, err := service.ValidateToken(tc.token)

				if err == nil {
					t.Errorf("Expected error for %s", tc.name)
				}

				if jwtObject != nil {
					t.Errorf("Expected nil JWT object for %s", tc.name)
				}
			})
		}
	})
}

func TestNewTokenServiceReady(t *testing.T) {
	t.Run("should create token service with valid environment variables", func(t *testing.T) {
		// Set up environment variables
		_ = os.Setenv("JWT_SECRET", "test-secret-key")
		_ = os.Setenv("JWT_ACCESS_TOKEN_EXPIRATION_TIME_IN_MINUTES", "60")
		_ = os.Setenv("JWT_REFRESH_TOKEN_EXPIRATION_TIME_IN_DAYS", "7")

		// Clean up after test
		defer func() {
			_ = os.Unsetenv("JWT_SECRET")
			_ = os.Unsetenv("JWT_ACCESS_TOKEN_EXPIRATION_TIME_IN_MINUTES")
			_ = os.Unsetenv("JWT_REFRESH_TOKEN_EXPIRATION_TIME_IN_DAYS")
		}()

		service := NewTokenServiceReady()

		if service == nil {
			t.Fatal("Expected non-nil token service")
		}

		// Test that the service works by generating a token
		token, err := service.GetAccessToken(123)
		if err != nil {
			t.Errorf("Expected service to work, got error: %v", err)
		}

		if token == "" {
			t.Error("Expected non-empty token")
		}
	})

	t.Run("should return nil when access token expiration is invalid", func(t *testing.T) {
		// Set up environment variables with invalid access token expiration
		_ = os.Setenv("JWT_SECRET", "test-secret-key")
		_ = os.Setenv("JWT_ACCESS_TOKEN_EXPIRATION_TIME_IN_MINUTES", "invalid-number")
		_ = os.Setenv("JWT_REFRESH_TOKEN_EXPIRATION_TIME_IN_DAYS", "7")

		// Clean up after test
		defer func() {
			_ = os.Unsetenv("JWT_SECRET")
			_ = os.Unsetenv("JWT_ACCESS_TOKEN_EXPIRATION_TIME_IN_MINUTES")
			_ = os.Unsetenv("JWT_REFRESH_TOKEN_EXPIRATION_TIME_IN_DAYS")
		}()

		service := NewTokenServiceReady()

		if service != nil {
			t.Error("Expected nil service when access token expiration is invalid")
		}
	})

	t.Run("should return nil when refresh token expiration is invalid", func(t *testing.T) {
		// Set up environment variables with invalid refresh token expiration
		_ = os.Setenv("JWT_SECRET", "test-secret-key")
		_ = os.Setenv("JWT_ACCESS_TOKEN_EXPIRATION_TIME_IN_MINUTES", "60")
		_ = os.Setenv("JWT_REFRESH_TOKEN_EXPIRATION_TIME_IN_DAYS", "invalid-number")

		// Clean up after test
		defer func() {
			_ = os.Unsetenv("JWT_SECRET")
			_ = os.Unsetenv("JWT_ACCESS_TOKEN_EXPIRATION_TIME_IN_MINUTES")
			_ = os.Unsetenv("JWT_REFRESH_TOKEN_EXPIRATION_TIME_IN_DAYS")
		}()

		service := NewTokenServiceReady()

		if service != nil {
			t.Error("Expected nil service when refresh token expiration is invalid")
		}
	})
}
