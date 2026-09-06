package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sudabon/monorepo_project_template/apps/api/internal/domain"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/repository"
	"github.com/sudabon/monorepo_project_template/apps/api/internal/testdb"
)

func TestPersistentCRUDAndPagination(t *testing.T) {
	db := testdb.Open(t)
	repo := repository.NewItems(db)
	ctx := context.Background()
	empty, err := repo.List(ctx, domain.Pagination{Page: 1, PageSize: 2})
	if err != nil || empty.Total != 0 || empty.Items == nil || len(empty.Items) != 0 {
		t.Fatalf("empty = %+v, %v", empty, err)
	}
	ids := map[string]bool{}
	for _, name := range []string{"first", "second", "third"} {
		item, err := repo.Create(ctx, domain.ItemInput{Name: name, Description: "stored"})
		if err != nil || item.ID == "" || item.CreatedAt.IsZero() {
			t.Fatalf("create = %+v, %v", item, err)
		}
		ids[item.ID] = true
		// A fresh repository reads persisted state, not process-local memory.
		got, err := repository.NewItems(db).Get(ctx, item.ID)
		if err != nil || got != item {
			t.Fatalf("get = %+v, %v", got, err)
		}
	}
	seen := map[string]bool{}
	for _, pageNo := range []int32{1, 2, 3} {
		page, err := repo.List(ctx, domain.Pagination{Page: pageNo, PageSize: 2})
		if err != nil || page.Total != 3 {
			t.Fatalf("page = %+v, %v", page, err)
		}
		want := map[int32]int{1: 2, 2: 1, 3: 0}[pageNo]
		if len(page.Items) != want {
			t.Fatalf("page %d length = %d", pageNo, len(page.Items))
		}
		for _, item := range page.Items {
			if !ids[item.ID] || seen[item.ID] {
				t.Fatalf("unexpected/duplicate ID %s", item.ID)
			}
			seen[item.ID] = true
		}
	}
	if len(seen) != 3 {
		t.Fatal("pagination lost items")
	}
	for id := range ids {
		updated, err := repo.Update(ctx, id, domain.ItemInput{Name: "updated"})
		if err != nil || updated.Description != "" || updated.UpdatedAt.Before(updated.CreatedAt) {
			t.Fatalf("update = %+v, %v", updated, err)
		}
		got, err := repo.Get(ctx, id)
		if err != nil || got.Name != "updated" {
			t.Fatalf("get update = %+v, %v", got, err)
		}
		if err := repo.Delete(ctx, id); err != nil {
			t.Fatal(err)
		}
		_, getErr := repo.Get(ctx, id)
		_, updateErr := repo.Update(ctx, id, domain.ItemInput{Name: "missing"})
		deleteErr := repo.Delete(ctx, id)
		for _, err := range []error{getErr, updateErr, deleteErr} {
			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("missing = %v", err)
			}
		}
	}
}

func TestListOrdersTimestampTiesByID(t *testing.T) {
	db := testdb.Open(t)
	_, err := db.Exec(`INSERT INTO items (id,name,created_at,updated_at) VALUES
	('00000000-0000-0000-0000-000000000002','second','2026-01-01','2026-01-01'),
	('00000000-0000-0000-0000-000000000001','first','2026-01-01','2026-01-01'),
	('00000000-0000-0000-0000-000000000003','newest','2026-01-02','2026-01-02')`)
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewItems(db)
	for i, want := range []string{"newest", "first", "second"} {
		page, err := repo.List(context.Background(), domain.Pagination{Page: int32(i + 1), PageSize: 1})
		if err != nil || len(page.Items) != 1 || page.Items[0].Name != want {
			t.Fatalf("page = %+v, %v, want %s", page, err, want)
		}
	}
}
