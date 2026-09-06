package domain

import (
	"context"
	"errors"
	"strings"
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

func (in ItemInput) Validate() ValidationErrors {
	var fields ValidationErrors
	if n := utf8.RuneCountInString(in.Name); n < 1 || n > 100 {
		fields = append(fields, FieldError{"name", "Name must contain 1 to 100 characters."})
	} else if strings.ContainsRune(in.Name, '\x00') {
		fields = append(fields, FieldError{"name", "Must not contain the NUL character (U+0000)."})
	}
	if utf8.RuneCountInString(in.Description) > 2000 {
		fields = append(fields, FieldError{"description", "Description must be at most 2000 characters."})
	} else if strings.ContainsRune(in.Description, '\x00') {
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
