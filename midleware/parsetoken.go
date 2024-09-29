package midleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AuthMiddleware to validate JWT and check for required role
func AuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract the token from the header (remove "Bearer " prefix)
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		fmt.Println("-----------------------", tokenString)

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		// Check for parsing errors or invalid tokens
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		var claim CustomClaims
		fmt.Println("role", claim.Role)

		// Type assert to extract the CustomClaims from the parsed token
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			// Check if the role in the token matches the required role
			if claims.Role != requiredRole {
				c.JSON(http.StatusForbidden, gin.H{"message": "Insufficient privileges"})
				c.Abort()
				return
			}

			// Store claims in context for further use in handlers
			c.Set("claims", claims)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Proceed to the next middleware or handler
		c.Next()
	}
}
