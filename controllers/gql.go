package controllers

import (
	"github.com/graphql-go/handler"
	"github.com/zaker/anachrome-be/services"
)

// GQLHandler returns a gql query handler
func GQLHandler(gql *services.GQL) func() *handler.Handler {
	gqlh := handler.New(gql.Conf())
	return func() *handler.Handler {

		return gqlh
	}
}
