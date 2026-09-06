package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/domain"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/generated"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/usecase"
)

type Items struct{ usecase *usecase.Items }

// Every contract operation must be implemented before this package can build.
var _ generated.ServerInterface = (*Items)(nil)

func (h *Items) CreateItem(c echo.Context) error {
	in, err := readInput(c)
	if err != nil {
		return err
	}
	item, err := h.usecase.Create(c.Request().Context(), in)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toItem(item))
}
func (h *Items) GetItem(c echo.Context, id generated.ItemId) error {
	item, err := h.usecase.Get(c.Request().Context(), id.String())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toItem(item))
}
func (h *Items) UpdateItem(c echo.Context, id generated.ItemId) error {
	in, err := readInput(c)
	if err != nil {
		return err
	}
	item, err := h.usecase.Update(c.Request().Context(), id.String(), in)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toItem(item))
}
func (h *Items) DeleteItem(c echo.Context, id generated.ItemId) error {
	if err := h.usecase.Delete(c.Request().Context(), id.String()); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
func (h *Items) ListItems(c echo.Context, params generated.ListItemsParams) error {
	p := domain.Pagination{Page: 1, PageSize: 20}
	if params.Page != nil {
		p.Page = *params.Page
	}
	if params.PageSize != nil {
		p.PageSize = *params.PageSize
	}
	if p.Page < 1 || p.PageSize < 1 || p.PageSize > 100 {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	page, err := h.usecase.List(c.Request().Context(), p)
	if err != nil {
		return err
	}
	items := make([]generated.Item, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, toItem(item))
	}
	return c.JSON(http.StatusOK, generated.ItemPage{Items: items, Page: p.Page, PageSize: p.PageSize, Total: page.Total})
}

func toItem(item domain.Item) generated.Item {
	return generated.Item{Id: uuid.MustParse(item.ID), Name: item.Name, Description: item.Description, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}
}

func readInput(c echo.Context) (domain.ItemInput, error) {
	mediaType, _, err := mime.ParseMediaType(c.Request().Header.Get(echo.HeaderContentType))
	if err != nil || mediaType != echo.MIMEApplicationJSON {
		return domain.ItemInput{}, echo.NewHTTPError(http.StatusUnsupportedMediaType)
	}
	decoder := json.NewDecoder(http.MaxBytesReader(c.Response(), c.Request().Body, 64*1024))
	var raw map[string]json.RawMessage
	if err := decoder.Decode(&raw); err != nil || raw == nil {
		return domain.ItemInput{}, echo.NewHTTPError(http.StatusBadRequest)
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return domain.ItemInput{}, echo.NewHTTPError(http.StatusBadRequest)
	}
	// Generated value types cannot distinguish JSON null from a missing field.
	// Decode the supplied fields explicitly so all type/length errors are returned.
	var input generated.ItemInput
	var fields domain.ValidationErrors
	for _, field := range []string{"name", "description"} {
		data, exists := raw[field]
		if !exists {
			continue
		}
		var value *string
		if err := json.Unmarshal(data, &value); err != nil || value == nil {
			fields = append(fields, domain.FieldError{Field: field, Message: "Must be a string."})
			continue
		}
		if field == "name" {
			input.Name = *value
		} else {
			input.Description = value
		}
	}
	in := domain.ItemInput{Name: input.Name}
	if input.Description != nil {
		in.Description = *input.Description
	}
	// Merge business constraints with type errors, preserving contract field order.
	for _, validation := range in.Validate() {
		duplicate := false
		for _, field := range fields {
			if field.Field == validation.Field {
				duplicate = true
			}
		}
		if !duplicate {
			fields = append(fields, validation)
		}
	}
	if len(fields) > 0 {
		ordered := make(domain.ValidationErrors, 0, len(fields))
		for _, name := range []string{"name", "description"} {
			for _, field := range fields {
				if field.Field == name {
					ordered = append(ordered, field)
				}
			}
		}
		return domain.ItemInput{}, ordered
	}
	return in, nil
}
