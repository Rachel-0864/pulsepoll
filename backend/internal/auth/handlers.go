package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"pollingapp/internal/models"
)

type signupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Signup validates the payload server-side (never trust the client), checks
// for an existing account, hashes the password with bcrypt, and returns a JWT
// so the new user is immediately logged in.
func Signup(db *mongo.Database, jwtSecret string) gin.HandlerFunc {
	users := db.Collection("users")

	return func(c *gin.Context) {
		var req signupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		req.Name = strings.TrimSpace(req.Name)

		if req.Name == "" || req.Email == "" || len(req.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name, email, and a password of at least 8 characters are required"})
			return
		}

		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		existing := users.FindOne(ctx, bson.M{"email": req.Email})
		if existing.Err() == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "an account with this email already exists"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process password"})
			return
		}

		user := models.User{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: string(hash),
			CreatedAt:    time.Now(),
		}

		res, err := users.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create account"})
			return
		}

		userID := res.InsertedID.(interface{ Hex() string })
		token, err := GenerateToken(userID.Hex(), jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"token": token, "name": user.Name, "email": user.Email})
	}
}

// Login checks credentials against the stored bcrypt hash and, on success,
// issues a fresh JWT. Deliberately vague error messages (not "wrong password"
// vs "no such user") so we don't leak which emails are registered.
func Login(db *mongo.Database, jwtSecret string) gin.HandlerFunc {
	users := db.Collection("users")

	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		req.Email = strings.ToLower(strings.TrimSpace(req.Email))

		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		var user models.User
		if err := users.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		token, err := GenerateToken(user.ID.Hex(), jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token, "name": user.Name, "email": user.Email})
	}
}
