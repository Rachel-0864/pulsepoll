package votes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"pollingapp/internal/models"
)

type voteRequest struct {
	OptionID string `json:"optionId"`
}

// fingerprint creates a duplicate-vote key.
//
// For logged-in users, the Authorization token is included so that
// different accounts using the same browser/computer can still vote
// independently.
//
// For anonymous users, IP + User-Agent are used as a basic duplicate
// protection mechanism.
func fingerprint(c *gin.Context) string {
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	authHeader := c.GetHeader("Authorization")

	var raw string

	if authHeader != "" {
		// Logged-in users get a fingerprint that is different for
		// different authentication tokens/accounts.
		raw = "authenticated|" + authHeader
	} else {
		// Anonymous users fall back to IP + User-Agent.
		raw = "anonymous|" + ip + "|" + userAgent
	}

	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// Cast validates the vote against the poll's actual options,
// prevents duplicate votes, increments the Redis live counter,
// publishes the new counts through Redis Pub/Sub, and stores
// a durable vote record in MongoDB.
func Cast(db *mongo.Database, rdb *redis.Client) gin.HandlerFunc {
	polls := db.Collection("polls")
	votes := db.Collection("votes")

	return func(c *gin.Context) {
		id := c.Param("id")

		pollID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid poll id",
			})
			return
		}

		var req voteRequest

		if err := c.ShouldBindJSON(&req); err != nil || req.OptionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "optionId is required",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c, 5*time.Second)
		defer cancel()

		// ------------------------------------------------------------
		// Load poll from MongoDB
		// ------------------------------------------------------------

		var poll models.Poll

		if err := polls.FindOne(
			ctx,
			bson.M{"_id": pollID},
		).Decode(&poll); err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "poll not found",
			})
			return
		}

		// ------------------------------------------------------------
		// Check whether poll is still active
		// ------------------------------------------------------------

		if !poll.IsActive {
			c.JSON(http.StatusConflict, gin.H{
				"error": "this poll is no longer accepting votes",
			})
			return
		}

		// ------------------------------------------------------------
		// Validate option
		// ------------------------------------------------------------

		validOption := false

		for _, opt := range poll.Options {
			if opt.ID == req.OptionID {
				validOption = true
				break
			}
		}

		if !validOption {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "optionId does not belong to this poll",
			})
			return
		}

		// ------------------------------------------------------------
		// Duplicate-vote protection
		// ------------------------------------------------------------

		fp := fingerprint(c)

		votedKey := fmt.Sprintf(
			"poll:%s:voted:%s",
			pollID.Hex(),
			fp,
		)

		// SETNX returns false when this fingerprint has already voted.
		firstVote, err := rdb.SetNX(
			ctx,
			votedKey,
			req.OptionID,
			24*time.Hour,
		).Result()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not record vote",
			})
			return
		}

		if !firstVote {
			c.JSON(http.StatusConflict, gin.H{
				"error": "you've already voted on this poll",
			})
			return
		}

		// ------------------------------------------------------------
		// Increment Redis live counter
		// ------------------------------------------------------------

		countsKey := fmt.Sprintf(
			"poll:%s:counts",
			pollID.Hex(),
		)

		if _, err := rdb.HIncrBy(
			ctx,
			countsKey,
			req.OptionID,
			1,
		).Result(); err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not update live counts",
			})
			return
		}

		// ------------------------------------------------------------
		// Read updated counts
		// ------------------------------------------------------------

		counts, err := rdb.HGetAll(
			ctx,
			countsKey,
		).Result()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "could not read live counts",
			})
			return
		}

		// ------------------------------------------------------------
		// Publish live update to WebSocket clients
		// ------------------------------------------------------------

		payload, err := json.Marshal(gin.H{
			"pollId": pollID.Hex(),
			"counts": counts,
		})

		if err != nil {
			// The vote has already been counted, so don't fail it.
			fmt.Printf(
				"warning: could not build websocket payload for poll %s: %v\n",
				pollID.Hex(),
				err,
			)
		} else {
			channel := fmt.Sprintf(
				"poll:%s:updates",
				pollID.Hex(),
			)

			if err := rdb.Publish(
				ctx,
				channel,
				payload,
			).Err(); err != nil {

				// Non-fatal. The vote was still counted.
				fmt.Printf(
					"warning: could not publish update for poll %s: %v\n",
					pollID.Hex(),
					err,
				)
			}
		}

		// ------------------------------------------------------------
		// Durable MongoDB audit record
		// ------------------------------------------------------------

		voteDoc := models.Vote{
			PollID:           pollID,
			OptionID:         req.OptionID,
			VoterFingerprint: fp,
			CreatedAt:        time.Now(),
		}

		if _, err := votes.InsertOne(ctx, voteDoc); err != nil {
			// The Redis/live vote already succeeded.
			// Log the Mongo issue rather than telling the user the
			// vote failed.
			fmt.Printf(
				"warning: could not persist vote audit record: %v\n",
				err,
			)
		}

		// ------------------------------------------------------------
		// Return latest counts to frontend
		// ------------------------------------------------------------

		c.JSON(http.StatusOK, gin.H{
			"counts": counts,
		})
	}
}