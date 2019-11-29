package controllers

import (
	"github.com/graphql-go/handler"
	"github.com/zaker/anachrome-be/services"
)

// Handler returns handler
func GQLHandler(gql *services.GQL) func() *handler.Handler {
	return func() *handler.Handler {

		gqlh := handler.New(gql.Conf())

		return gqlh
	}
}
