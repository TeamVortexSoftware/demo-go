package main

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// DemoUser represents a user in our demo system
// Supports both new simplified format (IsAutoJoinAdmin) and legacy format (Role, Groups)
type DemoUser struct {
	ID       string      `json:"id"`
	Email    string      `json:"email"`
	Password string      `json:"-"` // Never include password in JSON

	// New simplified field (preferred)
	IsAutoJoinAdmin bool `json:"isAutoJoinAdmin"`

	// Legacy fields (deprecated but still supported for backward compatibility)
	Role   string      `json:"role"`
	Groups []UserGroup `json:"groups"`
}

// UserGroup represents a group membership
type UserGroup struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success bool     `json:"success"`
	User    DemoUser `json:"user,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// Demo users database (in a real app, this would be in a database)
// Demo users with new simplified format (IsAutoJoinAdmin)
// Legacy fields (Role, Groups) are also included for backward compatibility demo
var demoUsers = []DemoUser{
	{
		ID:              "user-1",
		Email:           "admin@example.com",
		Password:        hashPassword("password123"), // hashed 'password123'
		IsAutoJoinAdmin: true,                        // New simplified field
		Role:            "admin",                     // Legacy field
		Groups: []UserGroup{ // Legacy field
			{Type: "team", ID: "team-1", Name: "Engineering"},
			{Type: "organization", ID: "org-1", Name: "Acme Corp"},
		},
	},
	{
		ID:              "user-2",
		Email:           "user@example.com",
		Password:        hashPassword("userpass"), // hashed 'userpass'
		IsAutoJoinAdmin: false,                    // New simplified field
		Role:            "user",                   // Legacy field
		Groups: []UserGroup{ // Legacy field
			{Type: "team", ID: "team-1", Name: "Engineering"},
		},
	},
}

const jwtSecret = "demo-secret-key"

// Simple password hashing using SHA256 (in production, use bcrypt)
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}

// Verify password against hash
func verifyPassword(password, hash string) bool {
	return hashPassword(password) == hash
}

// Create a session JWT for the demo
func createSessionJWT(user DemoUser) (string, error) {
	claims := jwt.MapClaims{
		"userId":          user.ID,
		"email":           user.Email,
		"isAutoJoinAdmin": user.IsAutoJoinAdmin,
		"role":            user.Role,
		"groups":          user.Groups,
		"exp":             time.Now().Add(24 * time.Hour).Unix(),
		"iat":             time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// Verify session JWT
func verifySessionJWT(tokenString string) (*DemoUser, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Convert groups back to UserGroup slice
		var groups []UserGroup
		if groupsInterface, exists := claims["groups"]; exists {
			if groupsSlice, ok := groupsInterface.([]interface{}); ok {
				for _, g := range groupsSlice {
					if groupMap, ok := g.(map[string]interface{}); ok {
						groups = append(groups, UserGroup{
							Type: groupMap["type"].(string),
							ID:   groupMap["id"].(string),
							Name: groupMap["name"].(string),
						})
					}
				}
			}
		}

		// Get isAutoJoinAdmin with default false
		isAutoJoinAdmin := false
		if isAutoJoinAdminInterface, exists := claims["isAutoJoinAdmin"]; exists {
			if val, ok := isAutoJoinAdminInterface.(bool); ok {
				isAutoJoinAdmin = val
			}
		}

		return &DemoUser{
			ID:              claims["userId"].(string),
			Email:           claims["email"].(string),
			IsAutoJoinAdmin: isAutoJoinAdmin,
			Role:            claims["role"].(string),
			Groups:          groups,
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// Authenticate user by email and password
func authenticateUser(email, password string) *DemoUser {
	for _, user := range demoUsers {
		if user.Email == email && verifyPassword(password, user.Password) {
			return &DemoUser{
				ID:              user.ID,
				Email:           user.Email,
				IsAutoJoinAdmin: user.IsAutoJoinAdmin,
				Role:            user.Role,
				Groups:          user.Groups,
			}
		}
	}
	return nil
}

// Get current user from request (checks cookies for session JWT)
func getCurrentUser(c *gin.Context) *DemoUser {
	token, err := c.Cookie("session")
	if err != nil {
		return nil
	}

	user, err := verifySessionJWT(token)
	if err != nil {
		return nil
	}

	return user
}

// Middleware to require authentication
func requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := getCurrentUser(c)
		if user == nil {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Attach user to context for use in other handlers
		c.Set("user", user)
		c.Next()
	}
}

// Get demo users (for testing) - without passwords
func getDemoUsers() []DemoUser {
	var users []DemoUser
	for _, user := range demoUsers {
		users = append(users, DemoUser{
			ID:              user.ID,
			Email:           user.Email,
			IsAutoJoinAdmin: user.IsAutoJoinAdmin,
			Role:            user.Role,
			Groups:          user.Groups,
		})
	}
	return users
}