// Package token provides functionality for creating and verifying security tokens.
//
// It defines a Maker interface for token management and provides two
// concrete implementations:
//
// - JWTMaker: Uses JSON Web Tokens (JWT) for session management (unused).
//
// - PasetoMaker: Uses Platform-Agnostic Security Tokens (PASETO), which is
// more secure than JWT by default.
package token
