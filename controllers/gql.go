package controllers

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/zaker/anachrome-be/config"
)

type GQL struct {
	conf handler.Config
}

func Init(c config.WebConfig) (*GQL, error) {

	gql := new(GQL)
	// Schema
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return nil, err
	}
	gql.conf = handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: c.IsDebug,
	}
	return gql, nil
}

func (gql *GQL) Handler() *handler.Handler {

	gqlh := handler.New(&gql.conf)

	return gqlh
}
