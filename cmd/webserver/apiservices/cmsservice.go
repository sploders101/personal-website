package apiservices

import (
	"context"

	"github.com/sploders101/personal-website/cmd/webserver/config"
	"github.com/sploders101/personal-website/cmd/webserver/dbapi"
	"github.com/sploders101/personal-website/internal/authutils"
	cmsv1 "github.com/sploders101/personal-website/internal/gen/proto/com/shaunkeys/cms/v1"
)

type CmsService struct {
	config config.ServerConfig
	db     dbapi.Db
}

func NewCmsService(config config.ServerConfig, db dbapi.Db) CmsService {
	return CmsService{
		config: config,
		db:     db,
	}
}

func (auth CmsService) Ping(
	ctx context.Context,
	req *cmsv1.PingRequest,
) (*cmsv1.PingResponse, error) {
	return cmsv1.PingResponse_builder{
		Message: req.GetMessage() + "\n\n" + authutils.MustGetClaims(ctx).SshKeyFingerprint,
	}.Build(), nil
}
