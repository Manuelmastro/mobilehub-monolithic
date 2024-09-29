package midleware

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type CustomClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	ID    uint   `json:"id"`
	jwt.StandardClaims
}

var jwtSecret = []byte("your-secret-key")

// Function to generate a JWT token
func GenerateJWT(role string, email string, id uint) (string, error) {
	// Set custom claims
	claims := &CustomClaims{
		Email: email,
		Role:  role,
		ID:    id,
		StandardClaims: jwt.StandardClaims{
			// Token expiration time
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			// Token issued at time
			IssuedAt: time.Now().Unix(),
			Issuer:   "mobilehub",
		},
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	return token.SignedString(jwtSecret)
}
