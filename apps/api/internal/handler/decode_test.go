package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/domain"
)

func newJSONContext(t *testing.T, body string) echo.Context {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return e.NewContext(req, httptest.NewRecorder())
}

func TestMergeFieldErrorsPreservesContractOrder(t *testing.T) {
	order := []string{"name", "description"}
	got := mergeFieldErrors(order,
		domain.ValidationErrors{{Field: "description", Message: "Must be a string."}},
		domain.ValidationErrors{
			{Field: "name", Message: "Name must contain 1 to 100 characters."},
			{Field: "description", Message: "Description must be at most 2000 characters."},
		},
	)
	if len(got) != 2 || got[0].Field != "name" || got[1].Field != "description" {
		t.Fatalf("order = %v", got)
	}
	if got[1].Message != "Must be a string." {
		t.Fatalf("type error should win for description, got %+v", got[1])
	}
}

func TestSecondResourceOnlyNeedsFieldNamesAndPacking(t *testing.T) {
	type noteInput struct{ Title, Body string }
	order := []string{"title", "body"}
	values, typeErrs, err := decodeFields(newJSONContext(t, `{"title":123,"body":"ok"}`), order)
	if err != nil {
		t.Fatal(err)
	}
	in := noteInput{}
	if v := values["title"]; v != nil {
		in.Title = *v
	}
	if v := values["body"]; v != nil {
		in.Body = *v
	}
	var business domain.ValidationErrors
	if in.Title == "" {
		business = append(business, domain.FieldError{Field: "title", Message: "required"})
	}
	fields := mergeFieldErrors(order, typeErrs, business)
	if len(fields) != 1 || fields[0].Field != "title" || !strings.Contains(fields[0].Message, "string") {
		t.Fatalf("second resource should reuse decodeFields; got %+v in=%+v", fields, in)
	}
	if in.Body != "ok" {
		t.Fatalf("packed body = %q", in.Body)
	}
}

func TestDecodeFieldsDistinguishesNullFromMissing(t *testing.T) {
	e := newJSONContext(t, `{"name":null}`)
	values, fields, err := decodeFields(e, []string{"name", "description"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := values["name"]; ok {
		t.Fatalf("null name should be a type error, values=%v", values)
	}
	if _, ok := values["description"]; ok {
		t.Fatalf("missing description should be absent, values=%v", values)
	}
	if len(fields) != 1 || fields[0].Field != "name" || !strings.Contains(fields[0].Message, "string") {
		t.Fatalf("fields = %+v", fields)
	}
}
