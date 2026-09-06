package usecase

import (
	"context"

	"github.com/sudabon/monorepo_project_template/apps/api/internal/domain"
)

type Items struct{ repository domain.ItemRepository }

func NewItems(repository domain.ItemRepository) *Items { return &Items{repository: repository} }

func (s *Items) Create(ctx context.Context, input domain.ItemInput) (domain.Item, error) {
	if fields := input.Validate(); len(fields) > 0 {
		return domain.Item{}, fields
	}
	return s.repository.Create(ctx, input)
}
func (s *Items) Get(ctx context.Context, id string) (domain.Item, error) {
	return s.repository.Get(ctx, id)
}
func (s *Items) Update(ctx context.Context, id string, input domain.ItemInput) (domain.Item, error) {
	if fields := input.Validate(); len(fields) > 0 {
		return domain.Item{}, fields
	}
	return s.repository.Update(ctx, id, input)
}
func (s *Items) Delete(ctx context.Context, id string) error { return s.repository.Delete(ctx, id) }
func (s *Items) List(ctx context.Context, pagination domain.Pagination) (domain.ItemPage, error) {
	return s.repository.List(ctx, pagination)
}
