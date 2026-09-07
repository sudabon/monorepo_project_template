package handler

import (
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
	out, err := toItem(item)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, out)
}
func (h *Items) GetItem(c echo.Context, id generated.ItemId) error {
	item, err := h.usecase.Get(c.Request().Context(), id.String())
	if err != nil {
		return err
	}
	out, err := toItem(item)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
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
	out, err := toItem(item)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, out)
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
		out, err := toItem(item)
		if err != nil {
			return err
		}
		items = append(items, out)
	}
	return c.JSON(http.StatusOK, generated.ItemPage{Items: items, Page: p.Page, PageSize: p.PageSize, Total: page.Total})
}

func toItem(item domain.Item) (generated.Item, error) {
	id, err := uuid.Parse(item.ID)
	if err != nil {
		return generated.Item{}, err
	}
	return generated.Item{Id: id, Name: item.Name, Description: item.Description, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt}, nil
}

func readInput(c echo.Context) (domain.ItemInput, error) {
	order := []string{"name", "description"}
	values, typeErrs, err := decodeFields(c, order)
	if err != nil {
		return domain.ItemInput{}, err
	}
	in := domain.ItemInput{}
	if v := values["name"]; v != nil {
		in.Name = *v
	}
	if v := values["description"]; v != nil {
		in.Description = *v
	}
	if fields := mergeFieldErrors(order, typeErrs, in.Validate()); len(fields) > 0 {
		return domain.ItemInput{}, fields
	}
	return in, nil
}
