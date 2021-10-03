package context

import (
	"context"

	"lenslocked.com/models"
)

const (
	userKey privateKey = "user"
)

// create a new type call privateKey that takes in a string
// the
type privateKey string

// Set a user onto a context by setting the context to a privateKey type
// The context is of type 'privateKey'

func WithUser(cxt context.Context, user *models.User) context.Context {
	return context.WithValue(cxt, userKey, user)
}

// Return a user context
// The context value must match the TYPE 'privateKey' and VALUE
// i.e. constant userKey, thus providing type safety as privateKey type is accessible
// outside of context.go

func User(cxt context.Context) *models.User {
	if temp := cxt.Value(userKey); temp != nil {
		if user, ok := temp.(*models.User); ok {
			return user
		}
	}
	return nil
}
