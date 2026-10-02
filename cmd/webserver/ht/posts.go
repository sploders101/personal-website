package ht

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/sploders101/personal-website/cmd/webserver/config"
	queries "github.com/sploders101/personal-website/cmd/webserver/dbapi/gen"
	"github.com/sploders101/personal-website/internal/markdown"
)

type postTemplateVars struct {
	baseTemplateVars
	Slug         string
	Title        string
	Description  string
	PublishedAt  time.Time
	PostContents template.HTML
}

func getPostTemplateConfig(
	cfg config.ServerConfig,
	req *http.Request,
	db *queries.Queries,
) (postTemplateVars, error) {
	ctx := req.Context()
	baseVars, err := getBasePageConfig(cfg, req)
	if err != nil {
		return postTemplateVars{}, err
	}

	slug := req.PathValue("slug")
	article, err := db.GetPublishedRevisionBySlug(ctx, slug)
	if err != nil {
		return postTemplateVars{}, err
	}

	_, htmlContents, err := markdown.Render(
		[]byte(article.ArticlesRevision.Body),
		markdown.RenderOptions{
			EnableXHTML: false,
			Highlight: markdown.HighlightOptions{
				InlineStyles: true,
			},
		},
	)
	if err != nil {
		return postTemplateVars{}, fmt.Errorf("failed to render markdown: %w", err)
	}

	return postTemplateVars{
		baseTemplateVars: baseVars,
		Slug:             article.Article.Slug,
		Title:            article.ArticlesRevision.Title,
		Description:      article.ArticlesRevision.Description,
		PublishedAt:      article.ArticlesRevision.PublishedAt.Time,
		PostContents:     template.HTML(htmlContents),
	}, nil
}

type postFeedTemplateVars struct {
	baseTemplateVars
	Posts []postFeedPost
}

type postFeedPost struct {
	Url         string
	Slug        string
	Title       string
	Description string
	CreatedAt   time.Time
}

func getPostFeedTemplateConfig(
	cfg config.ServerConfig,
	req *http.Request,
	db *queries.Queries,
) (postFeedTemplateVars, error) {
	ctx := req.Context()
	baseVars, err := getBasePageConfig(cfg, req)
	if err != nil {
		return postFeedTemplateVars{}, err
	}

	dbPosts, err := db.GetArticleFeed(ctx, queries.GetArticleFeedParams{
		// TODO: Pagination & filtering
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		return postFeedTemplateVars{}, err
	}
	posts := make([]postFeedPost, len(dbPosts))
	for i, dbPost := range dbPosts {
		url, err := url.JoinPath(cfg.BaseUrl, "posts", dbPost.Article.Slug, "/")
		if err != nil {
			return postFeedTemplateVars{}, err
		}
		posts[i] = postFeedPost{
			Url:         url,
			Slug:        dbPost.Article.Slug,
			Title:       dbPost.ArticlesRevision.Title,
			Description: dbPost.ArticlesRevision.Description,
			CreatedAt:   dbPost.ArticlesRevision.CreatedAt,
		}
	}

	return postFeedTemplateVars{
		baseTemplateVars: baseVars,
		Posts:            posts,
	}, nil
}
