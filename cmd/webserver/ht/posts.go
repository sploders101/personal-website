package ht

import (
	"fmt"
	"html/template"
	"net/http"
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
