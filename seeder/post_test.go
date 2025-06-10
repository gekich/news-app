//go:build unit

package seeder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSamplePosts(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{
			name:  "generate zero posts",
			count: 0,
		},
		{
			name:  "generate one post",
			count: 1,
		},
		{
			name:  "generate multiple posts",
			count: 5,
		},
		{
			name:  "generate more posts than templates",
			count: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posts := GenerateSamplePosts(tt.count)
			assert.Equal(t, tt.count, len(posts))

			for i, post := range posts {
				assert.NotEmpty(t, post.ID)
				assert.NotEmpty(t, post.Title)
				assert.NotEmpty(t, post.Content)
				assert.False(t, post.CreatedAt.IsZero())
				assert.False(t, post.UpdatedAt.IsZero())

				if i >= 10 {
					assert.Contains(t, post.Title, "Part")
					assert.Contains(t, post.Content, "additional information")
				}
			}
		})
	}
}
