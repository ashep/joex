package pipeline

import (
	"context"
	"fmt"

	"github.com/ashep/joex/pkg/ddbtx"
)

type repo interface {
	NewTx() *ddbtx.Tx
	Create(*ddbtx.Tx, Pipeline) error
	FindByID(ctx context.Context, id string) (Pipeline, error)
}

type Service struct {
	repo repo
}

func NewService(repo repo) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, p Pipeline) error {
	if err := p.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	tx := s.repo.NewTx()

	if err := s.repo.Create(tx, p); err != nil {
		return fmt.Errorf("repo: create: %w", err)
	}

	if _, err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tx: commit: %w", err)
	}

	return nil
}

func (s *Service) FindByID(ctx context.Context, id string) (Pipeline, error) {
	return s.repo.FindByID(ctx, id)
}
