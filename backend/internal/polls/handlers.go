package polls

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"pollingapp/internal/models"
)

type createPollRequest struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

// CreatePoll validates the question/options server-side, writes the poll to
// MongoDB, and seeds Redis with zero counts for every option.
func CreatePoll(db *mongo.Database, rdb *redis.Client) gin.HandlerFunc {
	polls := db.Collection("polls")

	return func(c *gin.Context) {
		userID := c.GetString("userID")

		var req createPollRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body",
			})
			return
		}

		req.Question = strings.TrimSpace(req.Question)

		if req.Question == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "question is required",
			})
			return
		}

		var options []models.Option
		seen := make(map[string]bool)

		for i, raw := range req.Options {
			text := strings.TrimSpace(raw)

			if text == "" {
				continue
			}

			key := strings.ToLower(text)

			if seen[key] {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "duplicate option text is not allowed",
				})
				return
			}

			seen[key] = true

			options = append(options, models.Option{
				ID:   fmt.Sprintf("opt%d", i+1),
				Text: text,
			})
		}

		if len(options) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "a poll needs at least 2 non-empty options",
			})
			return
		}

		if len(options) > 10 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "a poll can have at most 10 options",
			})
			return
		}

		creatorID, err := primitive.ObjectIDFromHex(userID)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user",
			})
			return
		}

		poll := models.Poll{
			Question:  req.Question,
			Options:   options,
			CreatorID: creatorID,
			IsActive:  true,
			CreatedAt: time.Now(),
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		// ---------------------------------------------------------
		// Save poll to MongoDB
		// ---------------------------------------------------------

		res, err := polls.InsertOne(ctx, poll)

		if err != nil {
			log.Printf(
				"CreatePoll MongoDB error: %v",
				err,
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not create poll",
			})
			return
		}

		poll.ID = res.InsertedID.(primitive.ObjectID)

		// ---------------------------------------------------------
		// Seed Redis counts
		// ---------------------------------------------------------

		countsKey := fmt.Sprintf(
			"poll:%s:counts",
			poll.ID.Hex(),
		)

		fields := make(map[string]interface{})

		for _, opt := range options {
			fields[opt.ID] = 0
		}

		if err := rdb.HSet(
			ctx,
			countsKey,
			fields,
		).Err(); err != nil {

			log.Printf(
				"CreatePoll Redis seed error for poll %s: %v",
				poll.ID.Hex(),
				err,
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "poll created but live counts could not be initialized",
			})
			return
		}

		c.JSON(http.StatusCreated, poll)
	}
}

// GetPoll returns the poll from MongoDB together with its live Redis counts.
func GetPoll(db *mongo.Database, rdb *redis.Client) gin.HandlerFunc {
	polls := db.Collection("polls")

	return func(c *gin.Context) {
		id := c.Param("id")

		pollID, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid poll id",
			})
			return
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		// ---------------------------------------------------------
		// Load poll from MongoDB
		// ---------------------------------------------------------

		var poll models.Poll

		err = polls.FindOne(
			ctx,
			bson.M{"_id": pollID},
		).Decode(&poll)

		if err != nil {

			// IMPORTANT:
			// A missing document is a 404.
			if errors.Is(err, mongo.ErrNoDocuments) {
				log.Printf(
					"GetPoll: poll %s not found",
					pollID.Hex(),
				)

				c.JSON(http.StatusNotFound, gin.H{
					"error": "poll not found",
				})
				return
			}

			// A timeout/network/database error is NOT a 404.
			log.Printf(
				"GetPoll MongoDB error for poll %s: %v",
				pollID.Hex(),
				err,
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not load poll from database",
			})
			return
		}

		// ---------------------------------------------------------
		// Load live counts from Redis
		// ---------------------------------------------------------

		countsKey := fmt.Sprintf(
			"poll:%s:counts",
			poll.ID.Hex(),
		)

		counts, err := rdb.HGetAll(
			ctx,
			countsKey,
		).Result()

		if err != nil {
			log.Printf(
				"GetPoll Redis error for poll %s: %v",
				poll.ID.Hex(),
				err,
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not load live counts",
			})
			return
		}

		// ---------------------------------------------------------
		// Make sure every poll option has a count.
		// This protects the frontend if a Redis field is missing.
		// ---------------------------------------------------------

		for _, option := range poll.Options {
			if _, exists := counts[option.ID]; !exists {
				counts[option.ID] = "0"
			}
		}

		// ---------------------------------------------------------
		// Return poll + live counts
		// ---------------------------------------------------------

		c.JSON(http.StatusOK, gin.H{
			"poll":   poll,
			"counts": counts,
		})
	}
}

// ListMyPolls returns all polls created by the authenticated user.
func ListMyPolls(db *mongo.Database) gin.HandlerFunc {
	polls := db.Collection("polls")

	return func(c *gin.Context) {
		userID := c.GetString("userID")

		creatorID, err := primitive.ObjectIDFromHex(userID)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user",
			})
			return
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		cursor, err := polls.Find(
			ctx,
			bson.M{"creatorId": creatorID},
		)

		if err != nil {
			log.Printf(
				"ListMyPolls MongoDB error: %v",
				err,
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not fetch polls",
			})
			return
		}

		defer cursor.Close(ctx)

		var results []models.Poll

		if err := cursor.All(ctx, &results); err != nil {
			log.Printf(
				"ListMyPolls cursor error: %v",
				err,
			)

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not read polls",
			})
			return
		}

		// Return an empty array instead of null when there are no polls.
		if results == nil {
			results = []models.Poll{}
		}

		c.JSON(http.StatusOK, results)
	}
}