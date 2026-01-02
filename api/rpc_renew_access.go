package api

import (
	"context"
	"errors"
	db "simplebank/db/sqlc"
	"simplebank/pb/users"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) RenewAccess(ctx context.Context, req *users.RenewAccessRequest) (*users.RenewAccessResponse, error) {
	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "session not found")
		}

		return nil, status.Errorf(codes.Internal, "failed to get session")
	}

	if session.IsBlocked {
		return nil, status.Errorf(codes.PermissionDenied, "session is blocked")
	}

	if session.Username != refreshPayload.Username {
		return nil, status.Errorf(codes.PermissionDenied, "invalid session user")
	}

	if session.RefreshToken != req.RefreshToken {
		return nil, status.Errorf(codes.PermissionDenied, "invalid session token")
	}

	if time.Now().After(session.ExpiresAt.Time) {
		return nil, status.Errorf(codes.PermissionDenied, "expired session")
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(refreshPayload.Username, refreshPayload.Role, server.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create access token: %v", err)
	}

	response := &users.RenewAccessResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: timestamppb.New(accessPayload.ExpiredAt),
	}

	return response, nil
}
