package handler

import (
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"mime"
	"net/http"
	"slices"

	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/domain"
)

func decodeFields(c echo.Context, order []string) (map[string]*string, domain.ValidationErrors, error) {
	mediaType, _, err := mime.ParseMediaType(c.Request().Header.Get(echo.HeaderContentType))
	if err != nil || mediaType != echo.MIMEApplicationJSON {
		return nil, nil, echo.NewHTTPError(http.StatusUnsupportedMediaType)
	}
	var raw map[string]jsontext.Value
	if err := json.UnmarshalRead(http.MaxBytesReader(c.Response(), c.Request().Body, 64*1024), &raw); err != nil || raw == nil {
		return nil, nil, echo.NewHTTPError(http.StatusBadRequest)
	}
	values := make(map[string]*string, len(order))
	var fields domain.ValidationErrors
	for _, field := range order {
		data, exists := raw[field]
		if !exists {
			continue
		}
		var value *string
		if data.Kind() == jsontext.KindNull || json.Unmarshal(data, &value) != nil || value == nil {
			fields = append(fields, domain.FieldError{Field: field, Message: "Must be a string."})
			continue
		}
		values[field] = value
	}
	return values, fields, nil
}

func mergeFieldErrors(order []string, typeErrs, businessErrs domain.ValidationErrors) domain.ValidationErrors {
	index := make(map[string]int, len(order))
	for i, name := range order {
		index[name] = i
	}
	fields := slices.Clone(typeErrs)
	for _, validation := range businessErrs {
		if slices.ContainsFunc(fields, func(existing domain.FieldError) bool {
			return existing.Field == validation.Field
		}) {
			continue
		}
		fields = append(fields, validation)
	}
	if len(fields) == 0 {
		return nil
	}
	slices.SortStableFunc(fields, func(a, b domain.FieldError) int {
		return cmp.Compare(index[a.Field], index[b.Field])
	})
	return fields
}
