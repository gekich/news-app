package repository

import (
	"context"
	"time"

	"github.com/gekich/news-app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const postCollection = "posts"

// PostRepository handles database operations for posts
type PostRepository struct {
	collection *mongo.Collection
}

// NewPostRepository creates a new PostRepository
func NewPostRepository(db *mongo.Database) *PostRepository {
	return &PostRepository{
		collection: db.Collection(postCollection),
	}
}

// FindAll retrieves all posts with optional pagination and search
func (r *PostRepository) FindAll(ctx context.Context, page, limit int64, search string) ([]models.Post, int64, error) {
	opts := options.Find()
	opts.SetSort(bson.D{{"created_at", -1}})

	if limit > 0 {
		opts.SetSkip((page - 1) * limit)
		opts.SetLimit(limit)
	}

	// Create filter based on search parameter
	filter := bson.M{}
	if search != "" {
		// Search in both title and content fields
		filter = bson.M{
			"$or": []bson.M{
				{"title": bson.M{"$regex": search, "$options": "i"}},
				{"content": bson.M{"$regex": search, "$options": "i"}},
			},
		}
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var posts []models.Post
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, 0, err
	}

	totalCount, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return posts, 0, err
	}

	totalPages := int64(1)
	if limit > 0 && totalCount > 0 {
		totalPages = (totalCount + limit - 1) / limit
	}

	return posts, totalPages, nil
}

// FindByID retrieves a post by its ID
func (r *PostRepository) FindByID(ctx context.Context, id string) (models.Post, error) {
	var post models.Post

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return post, err
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&post)
	return post, err
}

// Create inserts a new post
func (r *PostRepository) Create(ctx context.Context, post models.Post) (string, error) {
	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	result, err := r.collection.InsertOne(ctx, post)
	if err != nil {
		return "", err
	}

	id := result.InsertedID.(primitive.ObjectID).Hex()
	return id, nil
}

// Update modifies an existing post
func (r *PostRepository) Update(ctx context.Context, id string, post models.Post) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	post.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"title":      post.Title,
			"content":    post.Content,
			"updated_at": post.UpdatedAt,
		},
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

// Delete removes a post from the repository by its ID
func (r *PostRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

// CreateMany inserts multiple posts into the repository
func (r *PostRepository) CreateMany(ctx context.Context, posts []models.Post) error {
	if len(posts) == 0 {
		return nil
	}

	now := time.Now()
	documents := make([]interface{}, len(posts))

	for i := range posts {
		posts[i].CreatedAt = now
		posts[i].UpdatedAt = now
		documents[i] = posts[i]
	}

	_, err := r.collection.InsertMany(ctx, documents)
	return err
}
