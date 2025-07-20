package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestValidateToken(t *testing.T) {
	durationTime := time.Now().Add(time.Hour)
	duration := durationTime.Unix()
	expiredDuration := durationTime.Add(-2 * time.Hour).Unix()

	secret := []byte("test-secret-123")
	wrongSecret := []byte("fake-rsa-key")

	validClaims := jwt.MapClaims{
		"user_id": 42,
		"exp":     duration,
	}
	expiredClaims := jwt.MapClaims{
		"user_id": 42,
		"exp":     expiredDuration,
	}

	validToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims).SignedString(secret)
	expiredToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString(secret)
	wrongSecretToken, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, validClaims).SignedString(wrongSecret)

	tests := []struct {
		name        string
		token       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:        "Expired token",
			token:       expiredToken,
			wantErr:     true,
			errContains: "token has invalid claims: token is expired",
		},
		{
			name:        "Wrong signing method",
			token:       wrongSecretToken,
			wantErr:     true,
			errContains: "token is malformed: token contains an invalid number of segments",
		},
		{
			name:        "Malformed token",
			token:       "garbage-token",
			wantErr:     true,
			errContains: "token is malformed: token contains an invalid number of segments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token, secret)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, float64(42), (*claims)["user_id"])
		})
	}
}
