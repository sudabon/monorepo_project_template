package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/sudabon/monorepo_project_template/apps/api/internal/domain"
)

// The external persistence boundary is the only test double.
type itemRepo struct {
	input      domain.ItemInput
	id         string
	pagination domain.Pagination
	called     bool
	err        error
}

func (r *itemRepo) Create(_ context.Context, in domain.ItemInput) (domain.Item, error) {
	r.called = true
	r.input = in
	return domain.Item{ID: "created", Name: in.Name, Description: in.Description}, r.err
}
func (r *itemRepo) Get(_ context.Context, id string) (domain.Item, error) {
	r.id = id
	return domain.Item{ID: id, Name: "stored"}, r.err
}
func (r *itemRepo) Update(_ context.Context, id string, in domain.ItemInput) (domain.Item, error) {
	r.called = true
	r.id = id
	r.input = in
	return domain.Item{ID: id, Name: in.Name}, r.err
}
func (r *itemRepo) Delete(_ context.Context, id string) error { r.id = id; return r.err }
func (r *itemRepo) List(_ context.Context, p domain.Pagination) (domain.ItemPage, error) {
	r.pagination = p
	return domain.ItemPage{Items: []domain.Item{{ID: "listed"}}, Total: 3}, r.err
}

func TestItemsOperations(t *testing.T) {
	ctx := context.Background()
	repo := &itemRepo{}
	svc := NewItems(repo)
	created, err := svc.Create(ctx, domain.ItemInput{Name: "new", Description: "details"})
	if err != nil || created.ID != "created" || created.Description != "details" {
		t.Fatalf("create = %+v, %v", created, err)
	}
	got, err := svc.Get(ctx, "read-id")
	if err != nil || got.Name != "stored" || repo.id != "read-id" {
		t.Fatalf("get = %+v, %v", got, err)
	}
	updated, err := svc.Update(ctx, "update-id", domain.ItemInput{Name: "updated"})
	if err != nil || updated.Name != "updated" || repo.id != "update-id" || repo.input.Description != "" {
		t.Fatalf("update = %+v, %v", updated, err)
	}
	if err := svc.Delete(ctx, "delete-id"); err != nil || repo.id != "delete-id" {
		t.Fatalf("delete = %v", err)
	}
	page, err := svc.List(ctx, domain.Pagination{Page: 2, PageSize: 1})
	if err != nil || page.Total != 3 || page.Items[0].ID != "listed" || repo.pagination.Page != 2 {
		t.Fatalf("list = %+v, %v", page, err)
	}
}

func TestInvalidWritesNeverReachRepository(t *testing.T) {
	for _, update := range []bool{false, true} {
		repo := &itemRepo{}
		svc := NewItems(repo)
		var err error
		if update {
			_, err = svc.Update(context.Background(), "id", domain.ItemInput{})
		} else {
			_, err = svc.Create(context.Background(), domain.ItemInput{})
		}
		var validation domain.ValidationErrors
		if !errors.As(err, &validation) || repo.called {
			t.Fatalf("validation=%v, persisted=%v", err, repo.called)
		}
	}
}

func TestRepositoryErrorsRemainIdentifiable(t *testing.T) {
	for _, failure := range []error{domain.ErrNotFound, errors.New("database unavailable")} {
		svc := NewItems(&itemRepo{err: failure})
		ctx := context.Background()
		_, create := svc.Create(ctx, domain.ItemInput{Name: "valid"})
		_, get := svc.Get(ctx, "id")
		_, update := svc.Update(ctx, "id", domain.ItemInput{Name: "valid"})
		remove := svc.Delete(ctx, "id")
		_, list := svc.List(ctx, domain.Pagination{Page: 1, PageSize: 20})
		for _, err := range []error{create, get, update, remove, list} {
			if !errors.Is(err, failure) {
				t.Fatalf("error = %v, want %v", err, failure)
			}
		}
	}
}
