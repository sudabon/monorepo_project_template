package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sudabon/monorepo_project_template/apps/api/internal/domain"
)

type Items struct{ db *sql.DB }

var _ domain.ItemRepository = (*Items)(nil)

func NewItems(db *sql.DB) *Items { return &Items{db: db} }

const columns = "id, name, description, created_at, updated_at"

func scanItem(row interface{ Scan(...any) error }) (domain.Item, error) {
	var item domain.Item
	err := row.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Item{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("scan item: %w", err)
	}
	return item, nil
}

func (r *Items) Create(ctx context.Context, input domain.ItemInput) (domain.Item, error) {
	return scanItem(r.db.QueryRowContext(ctx, `INSERT INTO items(name,description) VALUES ($1,$2) RETURNING `+columns, input.Name, input.Description))
}
func (r *Items) Get(ctx context.Context, id string) (domain.Item, error) {
	return scanItem(r.db.QueryRowContext(ctx, `SELECT `+columns+` FROM items WHERE id=$1`, id))
}
func (r *Items) Update(ctx context.Context, id string, input domain.ItemInput) (domain.Item, error) {
	return scanItem(r.db.QueryRowContext(ctx, `UPDATE items SET name=$2,description=$3,updated_at=clock_timestamp() WHERE id=$1 RETURNING `+columns, id, input.Name, input.Description))
}
func (r *Items) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM items WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("deleted rows: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *Items) List(ctx context.Context, p domain.Pagination) (domain.ItemPage, error) {
	// Count and rows share a snapshot, including pages beyond the last page.
	// Separate HTTP requests use offset pagination as defined by the contract.
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return domain.ItemPage{}, fmt.Errorf("begin list: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	page := domain.ItemPage{Items: make([]domain.Item, 0)}
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM items`).Scan(&page.Total); err != nil {
		return domain.ItemPage{}, fmt.Errorf("count items: %w", err)
	}
	offset := (int64(p.Page) - 1) * int64(p.PageSize)
	rows, err := tx.QueryContext(ctx, `SELECT `+columns+` FROM items ORDER BY created_at DESC,id ASC LIMIT $1 OFFSET $2`, p.PageSize, offset)
	if err != nil {
		return domain.ItemPage{}, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return domain.ItemPage{}, err
		}
		page.Items = append(page.Items, item)
	}
	if err := rows.Err(); err != nil {
		return domain.ItemPage{}, fmt.Errorf("read items: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return domain.ItemPage{}, fmt.Errorf("commit list: %w", err)
	}
	return page, nil
}
