package apiservices

import (
	"errors"

	"connectrpc.com/connect"
)

var ErrAmbiguousInternal = connect.NewError(connect.CodeInternal, errors.New("internal server error"))
var ErrProtocolViolation = connect.NewError(connect.CodeInvalidArgument, errors.New("protocol violation"))
var ErrTimedOut = connect.NewError(connect.CodeCanceled, errors.New("timed out"))
var ErrAuthenticationFailed = connect.NewError(connect.CodeInvalidArgument, errors.New("authentication failed"))
var ErrPermissionDenied = connect.NewError(connect.CodePermissionDenied, errors.New("permission denied"))
var ErrHashMismatch = connect.NewError(connect.CodeFailedPrecondition, errors.New("hash mismatch"))
var ErrSizeMismatch = connect.NewError(connect.CodeFailedPrecondition, errors.New("actual size does not match reported size"))
var ErrArticleNotFound = connect.NewError(connect.CodeNotFound, errors.New("the requested article was not found"))
