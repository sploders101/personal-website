package apiservices

import (
	"errors"

	"connectrpc.com/connect"
)

var ErrAmbiguousInternal = connect.NewError(connect.CodeInternal, errors.New("internal server error"))
var ErrProtocolViolation = connect.NewError(connect.CodeInvalidArgument, errors.New("protocol violation"))
var ErrTimedOut = connect.NewError(connect.CodeCanceled, errors.New("timed out"))
var ErrAuthenticationFailed = connect.NewError(connect.CodeInvalidArgument, errors.New("authentication failed"))
