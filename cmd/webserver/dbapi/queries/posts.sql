-- name: CreateArticle :one
INSERT INTO articles (
    author,
    slug
) VALUES ($1, $2)
ON CONFLICT (slug) DO NOTHING
RETURNING *;

-- name: StageArticleRevision :one
INSERT INTO articles__revisions(
    article_id,
    title,
    description,
    body
) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetArticleRevision :one
SELECT
    sqlc.embed(articles),
    sqlc.embed(articles__revisions)
FROM articles__revisions
INNER JOIN articles ON articles.id = articles__revisions.article_id
WHERE public_id = $1;

-- name: PublishArticleRevision :exec
UPDATE articles__revisions
SET
    published_at = COALESCE($2, now())
WHERE
    public_id = $1;

-- name: GetArticleBySlug :one
SELECT *
FROM articles
WHERE slug = $1;

-- name: RedactArticle :exec
UPDATE articles__revisions
SET published_at = NULL
WHERE article_id = $1;

-- name: CreateAsset :exec
INSERT INTO assets(
    sha512_hash,
    content_type,
    content_length
) VALUES ($1, $2, $3);

-- name: LinkArticleAsset :exec
INSERT INTO articles__revisions__assets(
    revision_id,
    sha512_hash,
    file_name
) VALUES ($1, $2, $3);

-- name: GetMissingArticleAssets :many
SELECT *
FROM articles__revisions__assets ara
WHERE
    revision_id = $1
    AND NOT EXISTS (
        SELECT 1
        FROM assets
        WHERE assets.sha512_hash = ara.sha512_hash
    );
