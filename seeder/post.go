package seeder

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gekich/news-app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GenerateSamplePosts creates a specified number of sample posts
func GenerateSamplePosts(count int) []models.Post {
	posts := make([]models.Post, count)
	titles := []string{
		"Breaking News: Major Tech Breakthrough Announced",
		"Local Community Celebrates Annual Festival",
		"New Study Reveals Surprising Health Benefits",
		"Global Economic Trends Point to Recovery",
		"Environmental Initiative Gains International Support",
		"Sports Team Clinches Championship in Dramatic Fashion",
		"Education Reform Bill Passes After Long Debate",
		"Cultural Event Draws Record Attendance",
		"Scientific Discovery Changes Understanding of Universe",
		"Political Leaders Announce Historic Agreement",
	}

	contents := []string{
		"Researchers have announced a breakthrough in quantum computing technology that could revolutionize data processing. The new approach, developed by a team of international scientists, demonstrates quantum advantage in solving complex problems that would take traditional computers thousands of years to complete.",
		"The annual harvest festival brought together thousands of community members this weekend, featuring local cuisine, artisan crafts, and live performances. Now in its 25th year, the celebration has become a cornerstone of cultural heritage in the region.",
		"A comprehensive 10-year study published in the Journal of Medical Science indicates that moderate daily exercise can significantly reduce the risk of cognitive decline in older adults. The findings suggest even light activity provides measurable benefits.",
		"Leading economists are pointing to several indicators suggesting a strong economic recovery is underway. Decreased unemployment rates, increased consumer spending, and stability in global markets all contribute to the positive outlook for the coming quarters.",
		"The international Climate Action Coalition has secured commitments from 45 countries to reduce carbon emissions by 50% before 2030. This landmark agreement represents the most ambitious global environmental initiative to date.",
		"In a thrilling final match that went into double overtime, the hometown team secured their first championship title in 15 years. The victory caps a remarkable season that few analysts predicted at the outset.",
		"After months of negotiations, legislators have passed a comprehensive education reform bill that will increase teacher salaries, reduce class sizes, and expand access to early childhood education programs across the state.",
		"The international film festival concluded yesterday with record-breaking attendance figures. Over 150,000 cinema enthusiasts gathered to view premieres from acclaimed directors and emerging filmmakers from 35 countries.",
		"Astronomers using the newly deployed space telescope have discovered evidence of water vapor in the atmosphere of an exoplanet located just 40 light years from Earth, raising new possibilities in the search for habitable worlds.",
		"Representatives from previously conflicting nations signed a historic peace accord today, ending decades of tension. The agreement includes provisions for economic cooperation, cultural exchange programs, and joint environmental protection efforts.",
	}

	for i := 0; i < count; i++ {
		titleIndex := i % len(titles)
		contentIndex := i % len(contents)

		// Add some randomness to make each post unique
		title := titles[titleIndex]
		if i >= len(titles) {
			title = fmt.Sprintf("%s - Part %d", title, (i/len(titles))+1)
		}

		content := contents[contentIndex]
		if i >= len(contents) {
			content = fmt.Sprintf("%s\n\nThis is additional information for part %d of this series.", content, (i/len(contents))+1)
		}

		// Create post with random timestamp within the last month
		posts[i] = models.Post{
			ID:        primitive.NewObjectID(),
			Title:     title,
			Content:   content,
			CreatedAt: time.Now().AddDate(0, 0, -rand.Intn(30)),
			UpdatedAt: time.Now().AddDate(0, 0, -rand.Intn(15)),
		}
	}

	return posts
}
