package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post represents a news post in our application
type Post struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// CreatePostInput defines the input for creating a new post
type CreatePostInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// UpdatePostInput defines the input for updating an existing post
type UpdatePostInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
