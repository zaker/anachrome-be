package services

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"

	"github.com/zaker/anachrome-be/stores"
)

// GQL graphql setup for anachro.me
type GQL struct {
	conf      handler.Config
	blobStore *stores.BlobStore
}

func getBlogType() *graphql.Object {
	var blogType *graphql.Object
	blogInterface := graphql.NewInterface(graphql.InterfaceConfig{
		Name: "Blogpost",
		Fields: graphql.Fields{
			"title": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The title of the post.",
			},
			"content": &graphql.Field{
				Type:        graphql.String,
				Description: "The content of the post.",
			},
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {

			return blogType
		},
	})
	blogType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "BlogPost",
		Description: "A blob with some textual content",
		Fields: graphql.Fields{
			"title": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The title of the post.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if blog, ok := p.Source.(stores.BlogPost); ok {
						return blog.Title, nil
					}
					return nil, nil
				},
			},
			"content": &graphql.Field{
				Type:        graphql.String,
				Description: "The content.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if blog, ok := p.Source.(stores.BlogPost); ok {
						return blog.Content, nil
					}
					return nil, nil
				},
			},
		},
		Interfaces: []*graphql.Interface{
			blogInterface,
		},
	})

	return blogType
}

// InitGQL initializes components
func InitGQL(isDevMode bool) (*GQL, error) {

	gql := new(GQL)

	mockPosts := make([]stores.BlogPost, 0)
	mockPosts = append(mockPosts, stores.BlogPost{Title: "foo", Content: "blabla"})
	mockPosts = append(mockPosts, stores.BlogPost{Title: "bar", Content: "bliblo"})
	// Schema
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
		"blogs": &graphql.Field{
			Type: graphql.NewList(getBlogType()),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {

				return mockPosts, nil
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
		Pretty:     isDevMode,
		GraphiQL:   false,
		Playground: isDevMode,
	}
	return gql, nil
}

func (gql *GQL) Conf() *handler.Config {
	return &gql.conf
}
