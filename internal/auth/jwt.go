package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned by JWTManager.Verify for any malformed, expired
// or otherwise untrustworthy token.
var ErrInvalidToken = errors.New("invalid or expired token")

// Claims is the JWT payload taskflow-go issues and verifies.
type Claims struct {
	UserID int64  `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTManager issues and verifies HMAC-signed JWTs.
type JWTManager struct {
	secret []byte
	ttl    time.Duration
}

// NewJWTManager builds a JWTManager with the given signing secret and token lifetime.
func NewJWTManager(secret string, ttl time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), ttl: ttl}
}

// Generate issues a signed token for the given user.
func (m *JWTManager) Generate(userID int64, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Verify parses and validates a token string, returning its claims.
func (m *JWTManager) Verify(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
