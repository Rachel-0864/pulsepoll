package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"pollingapp/internal/auth"
	"pollingapp/internal/config"
	"pollingapp/internal/db"
	"pollingapp/internal/polls"
	"pollingapp/internal/votes"
	"pollingapp/internal/ws"
)

func main() {
	// ------------------------------------------------------------
	// Load environment variables
	// ------------------------------------------------------------

	if err := godotenv.Load(); err != nil {
		log.Println(
			"warning: .env file not found, using system environment variables",
		)
	}

	// ------------------------------------------------------------
	// Load application configuration
	// ------------------------------------------------------------

	cfg := config.Load()

	// ------------------------------------------------------------
	// Connect to MongoDB
	// ------------------------------------------------------------

	mongoDB := db.ConnectMongo(
		cfg.MongoURI,
		cfg.MongoDBName,
	)

	// ------------------------------------------------------------
	// Connect to Redis
	// ------------------------------------------------------------

	redisClient := db.ConnectRedis(
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisTLS,
	)

	// ------------------------------------------------------------
	// Create WebSocket hub
	// ------------------------------------------------------------

	hub := ws.NewHub(redisClient)

	// ------------------------------------------------------------
	// Create Gin router
	// ------------------------------------------------------------

	router := gin.Default()

	// ------------------------------------------------------------
	// CORS
	// ------------------------------------------------------------

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			cfg.AllowedOrigin,
		},

		AllowMethods: []string{
			"GET",
			"POST",
			"OPTIONS",
		},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},

		AllowCredentials: true,
	}))

	// ------------------------------------------------------------
	// API routes
	// ------------------------------------------------------------

	api := router.Group("/api")

	{
		// ========================================================
		// Authentication
		// ========================================================

		api.POST(
			"/auth/signup",
			auth.Signup(
				mongoDB,
				cfg.JWTSecret,
			),
		)

		api.POST(
			"/auth/login",
			auth.Login(
				mongoDB,
				cfg.JWTSecret,
			),
		)

		// ========================================================
		// Public poll routes
		// ========================================================

		api.GET(
			"/polls/:id",
			polls.GetPoll(
				mongoDB,
				redisClient,
			),
		)

		// Voting remains PUBLIC.
		//
		// OptionalAuth identifies logged-in users when a valid
		// JWT is present, but does not prevent anonymous voting.
		voteRoute := api.Group("")
		voteRoute.Use(
			auth.OptionalAuth(cfg.JWTSecret),
		)

		voteRoute.POST(
			"/polls/:id/vote",
			votes.Cast(
				mongoDB,
				redisClient,
			),
		)

		// ========================================================
		// Authenticated routes
		// ========================================================

		authed := api.Group("")

		authed.Use(
			auth.RequireAuth(cfg.JWTSecret),
		)

		{
			// Create poll
			authed.POST(
				"/polls",
				polls.CreatePoll(
					mongoDB,
					redisClient,
				),
			)

			// Get current user's polls
			authed.GET(
				"/polls",
				polls.ListMyPolls(
					mongoDB,
				),
			)
		}
	}

	// ------------------------------------------------------------
	// WebSocket route
	// ------------------------------------------------------------

	router.GET(
		"/ws/poll/:id",
		hub.ServeWs,
	)

	// ------------------------------------------------------------
	// Health check
	// ------------------------------------------------------------

	router.GET(
		"/health",
		func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		},
	)

	// ------------------------------------------------------------
	// Start server
	// ------------------------------------------------------------

	log.Printf(
		"server listening on :%s",
		cfg.Port,
	)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf(
			"server error: %v",
			err,
		)
	}
}