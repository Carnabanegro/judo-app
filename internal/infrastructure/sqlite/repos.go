package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"judo-app/internal/domain"
)

// TournamentRepo implements ports.TournamentRepo using SQLite.
type TournamentRepo struct{ db *sql.DB }

// NewTournamentRepo creates a new TournamentRepo.
func NewTournamentRepo(db *sql.DB) *TournamentRepo { return &TournamentRepo{db: db} }

func (r *TournamentRepo) Save(ctx context.Context, t *domain.Tournament) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tournaments (id, name, location, date, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, location=excluded.location, date=excluded.date`,
		t.ID.String(), t.Name, t.Location,
		t.Date.UTC().Format(time.RFC3339),
		t.CreatedAt.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *TournamentRepo) FindByID(ctx context.Context, id domain.TournamentID) (*domain.Tournament, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, location, date, created_at FROM tournaments WHERE id = ?`, id.String())
	return scanTournament(row)
}

func (r *TournamentRepo) List(ctx context.Context) ([]*domain.Tournament, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, location, date, created_at FROM tournaments ORDER BY date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.Tournament
	for rows.Next() {
		t, err := scanTournament(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (r *TournamentRepo) Delete(ctx context.Context, id domain.TournamentID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tournaments WHERE id = ?`, id.String())
	return err
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanTournament(s scanner) (*domain.Tournament, error) {
	var t domain.Tournament
	var idStr, dateStr, createdStr string
	if err := s.Scan(&idStr, &t.Name, &t.Location, &dateStr, &createdStr); err != nil {
		return nil, fmt.Errorf("scan tournament: %w", err)
	}
	if err := t.ID.UnmarshalText([]byte(idStr)); err != nil {
		return nil, err
	}
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil, err
	}
	created, err := time.Parse(time.RFC3339, createdStr)
	if err != nil {
		return nil, err
	}
	t.Date = date
	t.CreatedAt = created
	return &t, nil
}

// DivisionRepo implements ports.DivisionRepo using SQLite.
type DivisionRepo struct{ db *sql.DB }

// NewDivisionRepo creates a new DivisionRepo.
func NewDivisionRepo(db *sql.DB) *DivisionRepo { return &DivisionRepo{db: db} }

func (r *DivisionRepo) Save(ctx context.Context, d *domain.Division) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO divisions (id, tournament_id, age_group, gender, weight_class, format)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET age_group=excluded.age_group, gender=excluded.gender,
			weight_class=excluded.weight_class, format=excluded.format`,
		d.ID.String(), d.TournamentID.String(),
		string(d.AgeGroup), string(d.Gender), d.WeightClass, string(d.Format),
	)
	return err
}

func (r *DivisionRepo) FindByID(ctx context.Context, id domain.DivisionID) (*domain.Division, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, tournament_id, age_group, gender, weight_class, format FROM divisions WHERE id = ?`, id.String())
	return scanDivision(row)
}

func (r *DivisionRepo) ListByTournament(ctx context.Context, tID domain.TournamentID) ([]*domain.Division, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, tournament_id, age_group, gender, weight_class, format FROM divisions WHERE tournament_id = ?`, tID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.Division
	for rows.Next() {
		d, err := scanDivision(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

func (r *DivisionRepo) Delete(ctx context.Context, id domain.DivisionID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM divisions WHERE id = ?`, id.String())
	return err
}

func scanDivision(s scanner) (*domain.Division, error) {
	var d domain.Division
	var idStr, tIDStr, ageGroup, gender, format string
	if err := s.Scan(&idStr, &tIDStr, &ageGroup, &gender, &d.WeightClass, &format); err != nil {
		return nil, fmt.Errorf("scan division: %w", err)
	}
	if err := d.ID.UnmarshalText([]byte(idStr)); err != nil {
		return nil, err
	}
	if err := d.TournamentID.UnmarshalText([]byte(tIDStr)); err != nil {
		return nil, err
	}
	d.AgeGroup = domain.AgeGroup(ageGroup)
	d.Gender = domain.Gender(gender)
	d.Format = domain.Format(format)
	return &d, nil
}

// CategoryRepo implements ports.CategoryRepo using SQLite.
type CategoryRepo struct{ db *sql.DB }

// NewCategoryRepo creates a new CategoryRepo.
func NewCategoryRepo(db *sql.DB) *CategoryRepo { return &CategoryRepo{db: db} }

func (r *CategoryRepo) Save(ctx context.Context, c *domain.Category) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO categories (id, division_id) VALUES (?, ?)
		ON CONFLICT(id) DO NOTHING`,
		c.ID.String(), c.DivisionID.String(),
	)
	return err
}

func (r *CategoryRepo) FindByID(ctx context.Context, id domain.CategoryID) (*domain.Category, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, division_id FROM categories WHERE id = ?`, id.String())
	var c domain.Category
	var idStr, divStr string
	if err := row.Scan(&idStr, &divStr); err != nil {
		return nil, fmt.Errorf("scan category: %w", err)
	}
	if err := c.ID.UnmarshalText([]byte(idStr)); err != nil {
		return nil, err
	}
	if err := c.DivisionID.UnmarshalText([]byte(divStr)); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepo) ListByDivision(ctx context.Context, divID domain.DivisionID) ([]*domain.Category, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, division_id FROM categories WHERE division_id = ?`, divID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.Category
	for rows.Next() {
		var c domain.Category
		var idStr, divStr string
		if err := rows.Scan(&idStr, &divStr); err != nil {
			return nil, err
		}
		_ = c.ID.UnmarshalText([]byte(idStr))
		_ = c.DivisionID.UnmarshalText([]byte(divStr))
		result = append(result, &c)
	}
	return result, rows.Err()
}

// AthleteRepo implements ports.AthleteRepo using SQLite.
type AthleteRepo struct{ db *sql.DB }

// NewAthleteRepo creates a new AthleteRepo.
func NewAthleteRepo(db *sql.DB) *AthleteRepo { return &AthleteRepo{db: db} }

func (r *AthleteRepo) Save(ctx context.Context, a *domain.Athlete) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO athletes (id, category_id, name, club, weight, birth_date)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, club=excluded.club, weight=excluded.weight`,
		a.ID.String(), "", a.Name, a.Club, a.Weight,
		a.BirthDate.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *AthleteRepo) SaveToCategory(ctx context.Context, a *domain.Athlete, categoryID domain.CategoryID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO athletes (id, category_id, name, club, weight, birth_date)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, club=excluded.club, weight=excluded.weight`,
		a.ID.String(), categoryID.String(), a.Name, a.Club, a.Weight,
		a.BirthDate.UTC().Format(time.RFC3339),
	)
	return err
}

func (r *AthleteRepo) FindByID(ctx context.Context, id domain.AthleteID) (*domain.Athlete, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, club, weight, birth_date FROM athletes WHERE id = ?`, id.String())
	return scanAthlete(row)
}

func (r *AthleteRepo) ListByCategory(ctx context.Context, catID domain.CategoryID) ([]*domain.Athlete, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, club, weight, birth_date FROM athletes WHERE category_id = ? ORDER BY name`, catID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.Athlete
	for rows.Next() {
		a, err := scanAthlete(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, rows.Err()
}

func (r *AthleteRepo) Delete(ctx context.Context, id domain.AthleteID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM athletes WHERE id = ?`, id.String())
	return err
}

func scanAthlete(s scanner) (*domain.Athlete, error) {
	var a domain.Athlete
	var idStr, birthStr string
	if err := s.Scan(&idStr, &a.Name, &a.Club, &a.Weight, &birthStr); err != nil {
		return nil, fmt.Errorf("scan athlete: %w", err)
	}
	if err := a.ID.UnmarshalText([]byte(idStr)); err != nil {
		return nil, err
	}
	birth, err := time.Parse(time.RFC3339, birthStr)
	if err != nil {
		return nil, err
	}
	a.BirthDate = birth
	return &a, nil
}
