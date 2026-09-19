package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is a poll creator/manager. Voters never need an account.
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
}

// Option is one choice on a poll. ID is a short stable string (opt1, opt2, ...)
// used both as the MongoDB sub-document key and the Redis hash field, so the
// same identifier flows through validation, storage, and live counts.
type Option struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

// Poll is the durable record of what was asked and by whom.
// Live vote counts do NOT live here — they live in Redis and are only
// the fields that change on every single vote, kept out of Mongo's write path.
type Poll struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Question  string             `bson:"question" json:"question"`
	Options   []Option           `bson:"options" json:"options"`
	CreatorID primitive.ObjectID `bson:"creatorId" json:"creatorId"`
	IsActive  bool               `bson:"isActive" json:"isActive"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

// Vote is the durable audit trail of individual votes. Redis holds the fast
// running totals; this collection is what lets you rebuild/verify those totals,
// export results, or do analysis later.
type Vote struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID           primitive.ObjectID `bson:"pollId" json:"pollId"`
	OptionID         string             `bson:"optionId" json:"optionId"`
	VoterFingerprint string             `bson:"voterFingerprint" json:"-"`
	CreatedAt        time.Time          `bson:"createdAt" json:"createdAt"`
}
