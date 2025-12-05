package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	vortex "github.com/teamvortexsoftware/vortex-go-sdk"
)

var vortexClient *vortex.Client

// VortexConfig holds the configuration for Vortex integration
type VortexConfig struct {
	APIKey string
}

// Initialize Vortex client
func initVortex() {
	apiKey := os.Getenv("VORTEX_API_KEY")
	if apiKey == "" {
		apiKey = "demo-api-key"
	}
	vortexClient = vortex.NewClient(apiKey)
	log.Printf("🔧 Vortex client initialized with API key: %s...", apiKey[:min(len(apiKey), 10)])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Authentication routes
func setupAuthRoutes(r *gin.Engine) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", loginHandler)
		auth.POST("/logout", logoutHandler)
		auth.GET("/me", getMeHandler)
	}
}

// Demo routes
func setupDemoRoutes(r *gin.Engine) {
	demo := r.Group("/api/demo")
	{
		demo.GET("/users", getDemoUsersHandler)
		demo.GET("/protected", requireAuth(), getProtectedHandler)
	}
}

// Vortex API routes
func setupVortexRoutes(r *gin.Engine) {
	vortexGroup := r.Group("/api/vortex")
	{
		vortexGroup.POST("/jwt", requireAuth(), generateJWTHandler)
		vortexGroup.GET("/invitations", requireAuth(), getInvitationsHandler)
		vortexGroup.GET("/invitations/:id", requireAuth(), getInvitationHandler)
		vortexGroup.DELETE("/invitations/:id", requireAuth(), revokeInvitationHandler)
		vortexGroup.POST("/invitations/accept", requireAuth(), acceptInvitationsHandler)
		vortexGroup.GET("/invitations/by-group/:type/:id", requireAuth(), getInvitationsByGroupHandler)
		vortexGroup.DELETE("/invitations/by-group/:type/:id", requireAuth(), deleteInvitationsByGroupHandler)
		vortexGroup.POST("/invitations/:id/reinvite", requireAuth(), reinviteHandler)
	}
}

// Auth handlers
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Email and password required"})
		return
	}

	user := authenticateUser(req.Email, req.Password)
	if user == nil {
		c.JSON(401, gin.H{"error": "Invalid credentials"})
		return
	}

	// Create session JWT and set as cookie
	sessionToken, err := createSessionJWT(*user)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create session token"})
		return
	}

	c.SetCookie("session", sessionToken, 24*60*60, "/", "", false, true)

	c.JSON(200, LoginResponse{
		Success: true,
		User:    *user,
	})
}

func logoutHandler(c *gin.Context) {
	c.SetCookie("session", "", -1, "/", "", false, true)
	c.JSON(200, gin.H{"success": true})
}

func getMeHandler(c *gin.Context) {
	user := getCurrentUser(c)
	if user == nil {
		c.JSON(401, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(200, gin.H{"user": user})
}

// Demo handlers
func getDemoUsersHandler(c *gin.Context) {
	c.JSON(200, gin.H{"users": getDemoUsers()})
}

func getProtectedHandler(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(401, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(200, gin.H{
		"message":   "This is a protected route!",
		"user":      user,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// Vortex handlers
func generateJWTHandler(c *gin.Context) {
	user := getCurrentUser(c)
	if user == nil {
		c.JSON(401, gin.H{"error": "Authentication required"})
		return
	}

	// Build user with admin scopes
	vortexUser := &vortex.User{
		ID:    user.ID,
		Email: user.Email,
	}

	if user.IsAutojoinAdmin {
		vortexUser.AdminScopes = []string{"autojoin"}
	}

	jwt, err := vortexClient.GenerateJWT(vortexUser, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate JWT"})
		return
	}

	c.JSON(200, gin.H{"jwt": jwt})
}

func getInvitationsHandler(c *gin.Context) {
	targetType := c.Query("targetType")
	targetValue := c.Query("targetValue")

	if targetType == "" || targetValue == "" {
		c.JSON(400, gin.H{"error": "targetType and targetValue query parameters required"})
		return
	}

	invitations, err := vortexClient.GetInvitationsByTarget(targetType, targetValue)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get invitations"})
		return
	}

	c.JSON(200, gin.H{"invitations": invitations})
}

func getInvitationHandler(c *gin.Context) {
	id := c.Param("id")

	invitation, err := vortexClient.GetInvitation(id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Invitation not found"})
		return
	}

	c.JSON(200, invitation)
}

func revokeInvitationHandler(c *gin.Context) {
	id := c.Param("id")

	err := vortexClient.RevokeInvitation(id)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to revoke invitation"})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

func acceptInvitationsHandler(c *gin.Context) {
	var req struct {
		InvitationIDs []string             `json:"invitationIds" binding:"required"`
		Target        vortex.InvitationTarget `json:"target" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := vortexClient.AcceptInvitations(req.InvitationIDs, req.Target)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to accept invitations"})
		return
	}

	c.JSON(200, result)
}

func getInvitationsByGroupHandler(c *gin.Context) {
	groupType := c.Param("type")
	groupID := c.Param("id")

	invitations, err := vortexClient.GetInvitationsByGroup(groupType, groupID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get group invitations"})
		return
	}

	c.JSON(200, gin.H{"invitations": invitations})
}

func deleteInvitationsByGroupHandler(c *gin.Context) {
	groupType := c.Param("type")
	groupID := c.Param("id")

	err := vortexClient.DeleteInvitationsByGroup(groupType, groupID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete group invitations"})
		return
	}

	c.JSON(200, gin.H{"success": true})
}

func reinviteHandler(c *gin.Context) {
	id := c.Param("id")

	result, err := vortexClient.Reinvite(id)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to reinvite"})
		return
	}

	c.JSON(200, result)
}

func healthHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"vortex": gin.H{
			"configured": true,
			"routes": []string{
				"/api/vortex/jwt",
				"/api/vortex/invitations",
				"/api/vortex/invitations/:id",
				"/api/vortex/invitations/accept",
				"/api/vortex/invitations/by-group/:type/:id",
				"/api/vortex/invitations/:id/reinvite",
			},
		},
	})
}

func main() {
	// Initialize Vortex
	initVortex()

	// Setup Gin router
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./public")
	r.StaticFile("/", "./public/index.html")

	// Setup routes
	setupAuthRoutes(r)
	setupDemoRoutes(r)
	setupVortexRoutes(r)

	// Health check
	r.GET("/health", healthHandler)

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Convert port to int for validation
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatal("Invalid PORT environment variable")
	}

	log.Printf("🚀 Demo Go server starting on port %s", port)
	log.Printf("📱 Visit http://localhost:%s to try the demo", port)
	log.Printf("🔧 Vortex API routes available at http://localhost:%s/api/vortex", port)
	log.Printf("📊 Health check: http://localhost:%s/health", port)
	log.Println()
	log.Println("Demo users:")
	log.Println("  - admin@example.com / password123 (admin role)")
	log.Println("  - user@example.com / userpass (user role)")

	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}