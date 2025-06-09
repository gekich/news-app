package validation

import (
	"github.com/gekich/news-app/models"
	"github.com/go-playground/validator/v10"
)

// Validator instance
var validate = validator.New()

// PostError stores validation errors for the Post model
type PostError struct {
	Title   string
	Content string
}

// ValidatePost validates a post model and returns any validation errors
func ValidatePost(post models.Post) (PostError, bool) {
	var errors PostError
	valid := true

	err := validate.Struct(struct {
		Title   string `validate:"required,min=3,max=100"`
		Content string `validate:"required,min=10"`
	}{
		Title:   post.Title,
		Content: post.Content,
	})

	if err != nil {
		valid = false
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			switch field {
			case "Title":
				errors.Title = getErrorMessage(err)
			case "Content":
				errors.Content = getErrorMessage(err)
			}
		}
	}

	return errors, valid
}

// getErrorMessage returns a human-readable error message based on the validation error
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "This field must be at least " + err.Param() + " characters long"
	case "max":
		return "This field must be at most " + err.Param() + " characters long"
	default:
		return "Invalid input"
	}
}
