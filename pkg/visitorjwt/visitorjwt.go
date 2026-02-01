package visitorjwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type VisitorClaims struct {
	VisitorID   string `json:"visitor_id"`
	OwnerUserID uint   `json:"owner_user_id"` // CRM account
	Domain      string `json:"domain"`
	Fingerprint string `json:"fingerprint,omitempty"`
	jwt.RegisteredClaims
}

func GenerateVisitorToken(visitorID string, ownerUserID uint, domain, jwtSecret string) (string, error) {
	claims := &VisitorClaims{
		VisitorID:   visitorID,
		OwnerUserID: ownerUserID,
		Domain:      domain,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // 30 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func ValidateVisitorToken(tokenString, jwtSecret string) (*VisitorClaims, error) {
	claims := &VisitorClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}
