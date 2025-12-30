package gapi

import (
	"context"
	"simplebank/pb"
	"simplebank/util"
)

func (server *Server) RenewAccess(ctx context.Context, req *pb.RenewAccessRequest) (*pb.RenewAccessResponse, error) {
	authPayload, err := server.authorizeUser(ctx, []string{util.DepositorRole, util.BankerRole})
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	// TODO implement

	return nil, nil
}
