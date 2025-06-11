//go:build integration

package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gekich/news-app/models"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	mongoDB     *mongo.Database
	repository  *PostRepository
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "4.4",
		Env: []string{
			"MONGO_INITDB_DATABASE=test_db",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		mongoURI := fmt.Sprintf("mongodb://localhost:%s", resource.GetPort("27017/tcp"))

		mongoClient, err = mongo.Connect(
			context.Background(),
			options.Client().ApplyURI(mongoURI),
		)
		if err != nil {
			return err
		}

		return mongoClient.Ping(context.Background(), nil)
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	mongoDB = mongoClient.Database("test_db")
	repository = NewPostRepository(mongoDB)
	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestPostRepository_Create(t *testing.T) {
	// Clean up collection before test
	_, err := repository.collection.DeleteMany(context.Background(), bson.M{})
	require.NoError(t, err)

	post := models.Post{
		Title:   "Test Post",
		Content: "This is a test post content",
	}

	id, err := repository.Create(context.Background(), post)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	var createdPost models.Post
	objectID, err := primitive.ObjectIDFromHex(id)
	require.NoError(t, err)

	err = repository.collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&createdPost)
	require.NoError(t, err)
	assert.Equal(t, post.Title, createdPost.Title)
	assert.Equal(t, post.Content, createdPost.Content)
	assert.False(t, createdPost.CreatedAt.IsZero())
	assert.False(t, createdPost.UpdatedAt.IsZero())
}

func TestPostRepository_FindByID(t *testing.T) {
	_, err := repository.collection.DeleteMany(context.Background(), bson.M{})
	require.NoError(t, err)

	post := models.Post{
		Title:     "Test Post for FindByID",
		Content:   "This is a test post content for FindByID",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := repository.collection.InsertOne(context.Background(), post)
	require.NoError(t, err)
	insertedID := result.InsertedID.(primitive.ObjectID)

	foundPost, err := repository.FindByID(context.Background(), insertedID.Hex())
	require.NoError(t, err)
	assert.Equal(t, post.Title, foundPost.Title)
	assert.Equal(t, post.Content, foundPost.Content)
}

func TestPostRepository_Update(t *testing.T) {
	_, err := repository.collection.DeleteMany(context.Background(), bson.M{})
	require.NoError(t, err)

	post := models.Post{
		Title:     "Original Title",
		Content:   "Original Content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := repository.collection.InsertOne(context.Background(), post)
	require.NoError(t, err)
	insertedID := result.InsertedID.(primitive.ObjectID)

	updatedPost := models.Post{
		Title:   "Updated Title",
		Content: "Updated Content",
	}

	err = repository.Update(context.Background(), insertedID.Hex(), updatedPost)
	require.NoError(t, err)

	var retrievedPost models.Post
	err = repository.collection.FindOne(context.Background(), bson.M{"_id": insertedID}).Decode(&retrievedPost)
	require.NoError(t, err)
	assert.Equal(t, updatedPost.Title, retrievedPost.Title)
	assert.Equal(t, updatedPost.Content, retrievedPost.Content)
	assert.False(t, retrievedPost.UpdatedAt.IsZero())
}

func TestPostRepository_Delete(t *testing.T) {
	_, err := repository.collection.DeleteMany(context.Background(), bson.M{})
	require.NoError(t, err)

	post := models.Post{
		Title:     "Post to Delete",
		Content:   "This post will be deleted",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := repository.collection.InsertOne(context.Background(), post)
	require.NoError(t, err)
	insertedID := result.InsertedID.(primitive.ObjectID)

	err = repository.Delete(context.Background(), insertedID.Hex())
	require.NoError(t, err)

	count, err := repository.collection.CountDocuments(context.Background(), bson.M{"_id": insertedID})
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestPostRepository_FindAll(t *testing.T) {
	_, err := repository.collection.DeleteMany(context.Background(), bson.M{})
	require.NoError(t, err)

	posts := []interface{}{
		models.Post{
			Title:     "Post 1",
			Content:   "Content 1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		models.Post{
			Title:     "Post 2",
			Content:   "Content 2",
			CreatedAt: time.Now().Add(time.Second),
			UpdatedAt: time.Now().Add(time.Second),
		},
		models.Post{
			Title:     "Post 3",
			Content:   "Content 3",
			CreatedAt: time.Now().Add(2 * time.Second),
			UpdatedAt: time.Now().Add(2 * time.Second),
		},
	}

	_, err = repository.collection.InsertMany(context.Background(), posts)
	require.NoError(t, err)

	foundPosts, totalPages, err := repository.FindAll(context.Background(), 0, 0, "")
	require.NoError(t, err)
	assert.Equal(t, 3, len(foundPosts))
	assert.Equal(t, int64(1), totalPages)

	foundPosts, totalPages, err = repository.FindAll(context.Background(), 1, 2, "")
	require.NoError(t, err)
	assert.Equal(t, 2, len(foundPosts))
	assert.Equal(t, int64(2), totalPages)

	foundPosts, totalPages, err = repository.FindAll(context.Background(), 2, 2, "")
	require.NoError(t, err)
	assert.Equal(t, 1, len(foundPosts))
	assert.Equal(t, int64(2), totalPages)

	foundPosts, totalPages, err = repository.FindAll(context.Background(), 1, 10, "Post 2")
	require.NoError(t, err)
	assert.Equal(t, 1, len(foundPosts))
	assert.Equal(t, "Post 2", foundPosts[0].Title)

	foundPosts, totalPages, err = repository.FindAll(context.Background(), 1, 10, "Content 3")
	require.NoError(t, err)
	assert.Equal(t, 1, len(foundPosts))
	assert.Equal(t, "Post 3", foundPosts[0].Title)
}

func TestPostRepository_CreateMany(t *testing.T) {
	_, err := repository.collection.DeleteMany(context.Background(), bson.M{})
	require.NoError(t, err)

	posts := []models.Post{
		{
			Title:   "Batch Post 1",
			Content: "Batch Content 1",
		},
		{
			Title:   "Batch Post 2",
			Content: "Batch Content 2",
		},
	}

	err = repository.CreateMany(context.Background(), posts)
	require.NoError(t, err)

	count, err := repository.collection.CountDocuments(context.Background(), bson.M{})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	var foundPosts []models.Post
	cursor, err := repository.collection.Find(context.Background(), bson.M{})
	require.NoError(t, err)
	err = cursor.All(context.Background(), &foundPosts)
	require.NoError(t, err)

	for _, post := range foundPosts {
		assert.False(t, post.CreatedAt.IsZero())
		assert.False(t, post.UpdatedAt.IsZero())
	}
}
