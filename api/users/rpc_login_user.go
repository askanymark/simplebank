package users

import (
	"context"
	"errors"
	"simplebank/api/core"
	db "simplebank/db/sqlc"
	"simplebank/pb"
	"simplebank/util"
	"simplebank/val"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *UserHandler) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	violations := validateLoginUserRequest(req)
	if violations != nil {
		return nil, core.InvalidArgumentError(violations)
	}

	user, err := h.Server.Store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "cannot find user: %v", err)
		}

		return nil, status.Errorf(codes.Internal, "cannot get user: %v", err)
	}

	err = util.CheckPassword(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "unauthenticated")
	}

	accessToken, accessPayload, err := h.Server.TokenMaker.CreateToken(user.Username, user.Role, h.Server.Config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create access token: %v", err)
	}

	refreshToken, refreshPayload, err := h.Server.TokenMaker.CreateToken(user.Username, user.Role, h.Server.Config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create refresh token: %v", err)
	}

	metadata := core.ExtractMetadata(ctx)
	session, err := h.Server.Store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    metadata.UserAgent,
		ClientIp:     metadata.ClientIP,
		IsBlocked:    false,
		ExpiresAt: pgtype.Timestamp{
			Time:  refreshPayload.ExpiredAt,
			Valid: true,
		},
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create session: %v", err)
	}

	response := &pb.LoginUserResponse{
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
		User:                  user.ToResponse(),
	}
	return response, nil
}

func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, core.FieldViolation("username", err))
	}

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, core.FieldViolation("password", err))
	}

	return violations
}
