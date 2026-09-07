package domain

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"
)

var ErrNotFound = errors.New("item not found")

type Item struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ItemInput struct{ Name, Description string }
type FieldError struct{ Field, Message string }
type ValidationErrors []FieldError

func (e ValidationErrors) Error() string { return "some fields are invalid" }

var (
	itemInputNamePattern        = regexp.MustCompile(ItemInputNamePattern)
	itemInputDescriptionPattern = regexp.MustCompile(ItemInputDescriptionPattern)
)

func (in ItemInput) Validate() ValidationErrors {
	var fields ValidationErrors
	if n := utf8.RuneCountInString(in.Name); n < ItemInputNameMinLength || n > ItemInputNameMaxLength {
		fields = append(fields, FieldError{"name", fmt.Sprintf("Name must contain %d to %d characters.", ItemInputNameMinLength, ItemInputNameMaxLength)})
	} else if !itemInputNamePattern.MatchString(in.Name) {
		fields = append(fields, FieldError{"name", "Must not contain the NUL character (U+0000)."})
	}
	if utf8.RuneCountInString(in.Description) > ItemInputDescriptionMaxLength {
		fields = append(fields, FieldError{"description", fmt.Sprintf("Description must be at most %d characters.", ItemInputDescriptionMaxLength)})
	} else if !itemInputDescriptionPattern.MatchString(in.Description) {
		fields = append(fields, FieldError{"description", "Must not contain the NUL character (U+0000)."})
	}
	return fields
}

type Pagination struct{ Page, PageSize int32 }
type ItemPage struct {
	Items []Item
	Total int64
}

type ItemRepository interface {
	Create(context.Context, ItemInput) (Item, error)
	Get(context.Context, string) (Item, error)
	Update(context.Context, string, ItemInput) (Item, error)
	Delete(context.Context, string) error
	List(context.Context, Pagination) (ItemPage, error)
}
