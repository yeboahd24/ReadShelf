package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrDuplicateAnnotation = errors.New("duplicate annotation")
	ErrConflictAnnotation  = errors.New("conflicting annotation")
	ErrBookLimitReached    = errors.New("book limit reached for current plan")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrEmailTaken          = errors.New("email already taken")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidFileType     = errors.New("only PDF files are allowed")
	ErrFileTooLarge        = errors.New("file exceeds size limit")
)
