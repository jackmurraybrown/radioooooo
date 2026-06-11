package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenDuration = 15 * time.Minute

// IssueAccessToken returns a signed JWT valid for 15 minutes.
func IssueAccessToken(secret, userID, stationID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        userID,
		"station_id": stationID,
		"exp":        time.Now().Add(accessTokenDuration).Unix(),
	})
	return token.SignedString([]byte(secret))
}

// ParseAccessToken validates a signed JWT and returns its claims.
func ParseAccessToken(secret, raw string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}
