package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sudabon/monorepo_project_template/apps/api/internal/generated"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/handler"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/repository"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/testdb"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/usecase"
)

func request(t *testing.T, api http.Handler, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	api.ServeHTTP(rec, req)
	if rec.Code != status {
		t.Fatalf("%s %s = %d %s, want %d", method, path, rec.Code, rec.Body, status)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request ID")
	}
	return rec
}

func TestHTTPPersistentCRUD(t *testing.T) {
	db := testdb.Open(t)
	newAPI := func() http.Handler { return handler.New(usecase.NewItems(repository.NewItems(db)), db.PingContext) }
	api := newAPI()
	created := request(t, api, "POST", "/api/items", `{"name":"original","description":"details"}`, 201)
	var item generated.Item
	if err := json.Unmarshal(created.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	if item.Name != "original" || item.Description != "details" || item.CreatedAt.IsZero() {
		t.Fatalf("item = %+v", item)
	}
	path := "/api/items/" + item.Id.String()
	api = newAPI()
	got := request(t, api, "GET", path, "", 200)
	if !bytes.Equal(created.Body.Bytes(), got.Body.Bytes()) {
		t.Fatalf("persisted item changed: %s", got.Body)
	}
	request(t, api, "PUT", path, `{"name":"updated"}`, 200)
	got = request(t, api, "GET", path, "", 200)
	if err := json.Unmarshal(got.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	if item.Name != "updated" || item.Description != "" {
		t.Fatalf("updated = %+v", item)
	}
	deleted := request(t, api, "DELETE", path, "", 204)
	if deleted.Body.Len() != 0 {
		t.Fatal("204 must be empty")
	}
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		rec := request(t, api, method, path, `{"name":"missing"}`, 404)
		var e generated.Error
		_ = json.Unmarshal(rec.Body.Bytes(), &e)
		if e.Code != "not_found" {
			t.Fatalf("error = %+v", e)
		}
	}
}

func TestHTTPValidationDoesNotPersist(t *testing.T) {
	db := testdb.Open(t)
	api := handler.New(usecase.NewItems(repository.NewItems(db)), db.PingContext)
	tooLong := strings.Repeat("文", 2001)
	for _, tc := range []struct {
		body   string
		fields []string
	}{
		{`{"name":""}`, []string{"name"}},
		{`{"name":"","description":"` + tooLong + `"}`, []string{"name", "description"}},
		{`{}`, []string{"name"}},
		{`{"name":null,"description":null}`, []string{"name", "description"}},
		{`{"name":123,"description":true}`, []string{"name", "description"}},
	} {
		rec := request(t, api, "POST", "/api/items", tc.body, 422)
		var e generated.ValidationError
		if err := json.Unmarshal(rec.Body.Bytes(), &e); err != nil {
			t.Fatal(err)
		}
		if e.Code != "validation_error" || len(e.Errors) != len(tc.fields) {
			t.Fatalf("validation = %+v", e)
		}
		for i, field := range tc.fields {
			if e.Errors[i].Field != field || e.Errors[i].Message == "" {
				t.Fatalf("field = %+v", e.Errors[i])
			}
		}
	}
	for _, body := range []string{`{`, `null`, `[]`, `{"name":"ok"} {}`} {
		request(t, api, "POST", "/api/items", body, 400)
	}
	empty := request(t, api, "GET", "/api/items", "", 200)
	var page generated.ItemPage
	if err := json.Unmarshal(empty.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Items == nil || len(page.Items) != 0 || page.Total != 0 || page.Page != 1 || page.PageSize != 20 {
		t.Fatalf("empty = %+v", page)
	}
	valid := `{"name":"` + strings.Repeat("名", 100) + `","description":"` + strings.Repeat("文", 2000) + `"}`
	request(t, api, "POST", "/api/items", valid, 201)
}

func TestHTTPPaginationAndMalformedParameters(t *testing.T) {
	db := testdb.Open(t)
	api := handler.New(usecase.NewItems(repository.NewItems(db)), db.PingContext)
	for _, name := range []string{"a", "b", "c"} {
		request(t, api, "POST", "/api/items", `{"name":"`+name+`"}`, 201)
	}
	seen := map[string]bool{}
	for _, query := range []string{"page=1&pageSize=2", "page=2&pageSize=2"} {
		rec := request(t, api, "GET", "/api/items?"+query, "", 200)
		var page generated.ItemPage
		if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
			t.Fatal(err)
		}
		if page.Total != 3 || page.PageSize != 2 {
			t.Fatalf("page = %+v", page)
		}
		for _, item := range page.Items {
			if seen[item.Id.String()] {
				t.Fatal("duplicate across pages")
			}
			seen[item.Id.String()] = true
		}
	}
	if len(seen) != 3 {
		t.Fatal("missing item")
	}
	request(t, api, "GET", "/api/items?page=2147483647&pageSize=100", "", 200)
	for _, path := range []string{"/api/items?page=0", "/api/items?pageSize=101", "/api/items?page=-1", "/api/items?pageSize=0", "/api/items?page=abc", "/api/items?page=2147483648", "/api/items/not-a-uuid"} {
		rec := request(t, api, "GET", path, "", 400)
		var e generated.Error
		_ = json.Unmarshal(rec.Body.Bytes(), &e)
		if e.Code != "bad_request" || e.Message == "" {
			t.Fatalf("error = %+v", e)
		}
	}
}

func TestHTTPRejectsNULWithoutCreatingOrUpdating(t *testing.T) {
	db := testdb.Open(t)
	api := handler.New(usecase.NewItems(repository.NewItems(db)), db.PingContext)
	created := request(t, api, "POST", "/api/items", `{"name":"original","description":"keep"}`, 201)
	var item generated.Item
	if err := json.Unmarshal(created.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	path := "/api/items/" + item.Id.String()
	for _, method := range []string{"POST", "PUT"} {
		for _, tc := range []struct {
			name, body string
			fields     []string
		}{
			{"name", `{"name":"a\u0000b"}`, []string{"name"}},
			{"description", `{"name":"valid","description":"\u0000"}`, []string{"description"}},
			{"both", `{"name":"\u0000name","description":"details\u0000"}`, []string{"name", "description"}},
			{"mixed constraints", `{"name":"","description":"\u0000"}`, []string{"name", "description"}},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				target := "/api/items"
				if method == "PUT" {
					target = path
				}
				rec := request(t, api, method, target, tc.body, 422)
				var validation generated.ValidationError
				if err := json.Unmarshal(rec.Body.Bytes(), &validation); err != nil {
					t.Fatal(err)
				}
				if validation.Code != "validation_error" || len(validation.Errors) != len(tc.fields) {
					t.Fatalf("validation = %+v", validation)
				}
				for i, field := range tc.fields {
					if validation.Errors[i].Field != field || validation.Errors[i].Message == "" {
						t.Fatalf("field error = %+v", validation.Errors[i])
					}
				}
				got := request(t, api, "GET", path, "", 200)
				if !bytes.Equal(got.Body.Bytes(), created.Body.Bytes()) {
					t.Fatalf("invalid write changed existing item: %s", got.Body)
				}
				list := request(t, api, "GET", "/api/items", "", 200)
				var page generated.ItemPage
				if err := json.Unmarshal(list.Body.Bytes(), &page); err != nil {
					t.Fatal(err)
				}
				if page.Total != 1 {
					t.Fatalf("invalid write persisted: total = %d", page.Total)
				}
			})
		}
	}
	// A literal backslash escape and newlines are not the NUL character.
	valid := request(t, api, "POST", "/api/items", `{"name":"\\u0000","description":"first\nsecond"}`, 201)
	var literal generated.Item
	if err := json.Unmarshal(valid.Body.Bytes(), &literal); err != nil {
		t.Fatal(err)
	}
	if literal.Name != `\u0000` || literal.Description != "first\nsecond" {
		t.Fatalf("valid text changed: %+v", literal)
	}
}
