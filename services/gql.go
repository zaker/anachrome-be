package services

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"

	"github.com/zaker/anachrome-be/stores/blog"
)

// GQL graphql setup for anachro.me
type GQL struct {
	conf      handler.Config
	blogStore blog.BlogStore
}

func getBlogMetaType() *graphql.Object {
	var blogPostMetaType *graphql.Object
	blogInterface := graphql.NewInterface(graphql.InterfaceConfig{
		Name: "BlogPostMetaInterface",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The id of the post.",
			},
			"title": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The title of the post.",
			},
			"published": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The date first published.",
			},
			"updated": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The date last updated.",
			},
		},
		ResolveType: func(p graphql.ResolveTypeParams) *graphql.Object {

			return blogPostMetaType
		},
	})
	blogPostMetaType = graphql.NewObject(graphql.ObjectConfig{
		Name:        "BlogPostMeta",
		Description: "A blob with some textual content",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The id of the post.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if meta, ok := p.Source.(blog.BlogPostMeta); ok {
						return meta.ID, nil
					}
					return "", nil
				},
			},
			"title": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The title of the post.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if meta, ok := p.Source.(blog.BlogPostMeta); ok {
						return meta.Title, nil
					}
					return nil, nil
				},
			},
			"published": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The date first published.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if meta, ok := p.Source.(blog.BlogPostMeta); ok {
						return meta.Published, nil
					}
					return nil, nil
				},
			},
			"updated": &graphql.Field{
				Type:        graphql.DateTime,
				Description: "The date last updated.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if meta, ok := p.Source.(blog.BlogPostMeta); ok {
						return meta.Updated, nil
					}
					return nil, nil
				},
			},
		},
		Interfaces: []*graphql.Interface{
			blogInterface,
		},
	})

	return blogPostMetaType
}

// InitGQL initializes components
func InitGQL(isDevMode bool, blogStore blog.BlogStore) (*GQL, error) {

	gql := &GQL{blogStore: blogStore}

	blogMetaType := getBlogMetaType()
	blogType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "BlogPost",
		Description: "A blob with some textual content",
		Fields: graphql.Fields{
			"meta": &graphql.Field{
				Type:        graphql.NewNonNull(blogMetaType),
				Description: "The blog post metadata.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if blog, ok := p.Source.(blog.BlogPost); ok {
						return blog.Meta, nil
					}
					return nil, nil
				},
			},
			"content": &graphql.Field{
				Type:        graphql.String,
				Description: "The content.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if blog, ok := p.Source.(blog.BlogPost); ok {
						return blog.Content, nil
					}
					return nil, nil
				},
			},
		},
	})
	// Schema
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
		"blogs": &graphql.Field{
			Type: graphql.NewList(blogMetaType),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				posts, err := gql.blogStore.GetBlogPostsMeta(context.TODO())
				if err != nil {
					return nil, err
				}
				return posts, nil
			},
		},

		"blog": &graphql.Field{
			Type: blogType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Description: "id of the blog post",
					Type:        graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				post, err := gql.blogStore.GetBlogPost(context.TODO(), p.Args["id"].(string))
				if err != nil {
					return nil, err
				}
				return post, nil
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
		Schema:   &schema,
		Pretty:   isDevMode,
		GraphiQL: isDevMode,
		// Playground: isDevMode,
	}
	return gql, nil
}

func (gql *GQL) Conf() *handler.Config {
	return &gql.conf
}
