package main

import (
	"errors"
	"fmt"
)

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// NotFoundError представляет ошибку "не найдено"
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %v not found", e.Resource, e.ID)
}

// ConflictError представляет ошибку конфликта
type ConflictError struct {
	Resource string
	Message  string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("conflict in %s: %s", e.Resource, e.Message)
}

// BusinessLogicError представляет бизнес-логику ошибку
type BusinessLogicError struct {
	Operation string
	Message   string
}

func (e BusinessLogicError) Error() string {
	return fmt.Sprintf("business logic error in %s: %s", e.Operation, e.Message)
}

// IsValidationError проверяет, является ли ошибка ValidationError
func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.Is(err, validationErr) || errors.As(err, &validationErr)
}

// IsNotFoundError проверяет, является ли ошибка NotFoundError
func IsNotFoundError(err error) bool {
	var notFoundErr NotFoundError
	return errors.Is(err, notFoundErr) || errors.As(err, &notFoundErr)
}

// IsConflictError проверяет, является ли ошибка ConflictError
func IsConflictError(err error) bool {
	var conflictErr ConflictError
	return errors.Is(err, conflictErr) || errors.As(err, &conflictErr)
}

// IsBusinessLogicError проверяет, является ли ошибка BusinessLogicError
func IsBusinessLogicError(err error) bool {
	var businessErr BusinessLogicError
	return errors.Is(err, businessErr) || errors.As(err, &businessErr)
}