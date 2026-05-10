package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"judo-app/internal/domain"
)

// BracketRepo implements ports.BracketRepo using SQLite (JSON blob).
type BracketRepo struct{ db *sql.DB }

// NewBracketRepo creates a new BracketRepo.
func NewBracketRepo(db *sql.DB) *BracketRepo { return &BracketRepo{db: db} }

func (r *BracketRepo) Save(ctx context.Context, b *domain.Bracket) error {
	data, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("bracket marshal: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO brackets (category_id, data) VALUES (?, ?)
		ON CONFLICT(category_id) DO UPDATE SET data=excluded.data`,
		b.CategoryID.String(), string(data),
	)
	return err
}

func (r *BracketRepo) FindByCategory(ctx context.Context, catID domain.CategoryID) (*domain.Bracket, error) {
	row := r.db.QueryRowContext(ctx, `SELECT data FROM brackets WHERE category_id = ?`, catID.String())
	var raw string
	if err := row.Scan(&raw); err != nil {
		return nil, fmt.Errorf("bracket scan: %w", err)
	}
	var b domain.Bracket
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		return nil, fmt.Errorf("bracket unmarshal: %w", err)
	}
	return &b, nil
}

// MatchRepo implements ports.MatchRepo using the brackets JSON blob.
// Individual matches are not stored separately — they live inside the bracket JSON.
type MatchRepo struct{ brackets *BracketRepo }

// NewMatchRepo creates a new MatchRepo backed by BracketRepo.
func NewMatchRepo(br *BracketRepo) *MatchRepo { return &MatchRepo{brackets: br} }

func (r *MatchRepo) Save(ctx context.Context, m *domain.Match) error {
	b, err := r.brackets.FindByCategory(ctx, m.CategoryID)
	if err != nil {
		return err
	}
	updated := false
	for i := range b.Rounds {
		for j := range b.Rounds[i].Matches {
			if b.Rounds[i].Matches[j].ID == m.ID {
				b.Rounds[i].Matches[j] = m
				updated = true
			}
		}
	}
	if !updated {
		return fmt.Errorf("match %s not found in bracket", m.ID)
	}
	return r.brackets.Save(ctx, b)
}

func (r *MatchRepo) FindByID(ctx context.Context, id domain.MatchID) (*domain.Match, error) {
	// We need the category_id — search all brackets. Suitable for V1 scale.
	rows, err := r.brackets.db.QueryContext(ctx, `SELECT data FROM brackets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			continue
		}
		var b domain.Bracket
		if err := json.Unmarshal([]byte(raw), &b); err != nil {
			continue
		}
		for _, round := range b.Rounds {
			for _, m := range round.Matches {
				if m.ID == id {
					return m, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("match %s not found", id)
}

func (r *MatchRepo) ListByCategory(ctx context.Context, catID domain.CategoryID) ([]*domain.Match, error) {
	b, err := r.brackets.FindByCategory(ctx, catID)
	if err != nil {
		return nil, err
	}
	var result []*domain.Match
	for i := range b.Rounds {
		result = append(result, b.Rounds[i].Matches...)
	}
	return result, nil
}
