package users

import (
	"context"
	"errors"
	"simplebank/api/core"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *UserHandler) RenewAccess(ctx context.Context, req *pb.RenewAccessRequest) (*pb.RenewAccessResponse, error) {
	refreshPayload, err := h.Server.TokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		return nil, core.UnauthenticatedError(err)
	}

	session, err := h.Server.Store.GetSession(ctx, refreshPayload.ID)
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

	accessToken, accessPayload, err := h.Server.TokenMaker.CreateToken(refreshPayload.Username, refreshPayload.Role, h.Server.Config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create access token: %v", err)
	}

	response := &pb.RenewAccessResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: timestamppb.New(accessPayload.ExpiredAt),
	}

	return response, nil
}
