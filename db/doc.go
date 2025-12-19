// Package db provides the data access layer for the SimpleBank application.
//
// It encompasses the entire database management system, including:
//
// - Database migrations located in the migration/ directory.
//
// - SQL query definitions in the query/ directory used by sqlc.
//
// - Auto-generated Go code for database interactions in the sqlc/ directory.
//
// - Mock implementations for testing in the mock/ directory.
//
// The package provides database models, SQL query functions, and a Store
// interface for executing database transactions. It supports operations
// for accounts, entries, transfers, users, sessions, and email verification.
package db
