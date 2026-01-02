package core

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"simplebank/token"
	"strings"
)

const (
	AuthorizationHeader = "authorization"
	BearerPrefix        = "bearer"
)

func AuthorizeUser(tokenMaker token.Maker, ctx context.Context, accessibleRoles []string) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	values := md.Get(AuthorizationHeader)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing authorization header")
	}

	authHeader := values[0]
	fields := strings.Fields(authHeader)
	if len(fields) < 2 {
		return nil, fmt.Errorf("invalid authorization format")
	}

	authType := strings.ToLower(fields[0])
	if authType != BearerPrefix {
		return nil, fmt.Errorf("only bearer authorization is supported")
	}

	accessToken := fields[1]
	payload, err := tokenMaker.VerifyToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("invalid access token: %w", err)
	}

	if !HasPermission(payload.Role, accessibleRoles) {
		return nil, fmt.Errorf("access denied")
	}

	return payload, nil
}

func HasPermission(userRole string, accessibleRoles []string) bool {
	for _, role := range accessibleRoles {
		if userRole == role {
			return true
		}
	}

	return false
}
