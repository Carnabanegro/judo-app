package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"judo-app/internal/application/ports"
	"judo-app/internal/domain"
)

// MatchStatusRepo manages the mutable match state in the `matches` table.
type MatchStatusRepo struct{ db *sql.DB }

// NewMatchStatusRepo creates a new MatchStatusRepo.
func NewMatchStatusRepo(db *sql.DB) *MatchStatusRepo { return &MatchStatusRepo{db: db} }

// UpsertFromBracket inserts (or updates) all matches from a bracket into the matches table.
func (r *MatchStatusRepo) UpsertFromBracket(ctx context.Context, b *domain.Bracket) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	upsert := func(m *domain.Match) error {
		var aID, bID string
		if m.AthleteA != nil {
			aID = m.AthleteA.ID.String()
		}
		if m.AthleteB != nil {
			bID = m.AthleteB.ID.String()
		}
		isRepechage := 0
		if m.IsRepechage {
			isRepechage = 1
		}
		status := string(m.Status)
		if status == "" {
			status = string(domain.MatchPending)
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO matches (id, category_id, round, position, is_repechage, athlete_a_id, athlete_b_id, status, tatami_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				athlete_a_id = excluded.athlete_a_id,
				athlete_b_id = excluded.athlete_b_id`,
			m.ID.String(), b.CategoryID.String(), m.Round, m.Position,
			isRepechage, aID, bID, status, m.TatamiID,
		)
		return err
	}

	for _, round := range b.Rounds {
		for _, m := range round.Matches {
			if err := upsert(m); err != nil {
				return fmt.Errorf("upsert match %s: %w", m.ID, err)
			}
		}
	}
	for _, pool := range b.Repechage {
		for _, m := range pool.Matches {
			if err := upsert(m); err != nil {
				return fmt.Errorf("upsert repechage match %s: %w", m.ID, err)
			}
		}
	}

	return tx.Commit()
}

// ClaimMatch atomically claims a PENDING match for a tatami.
func (r *MatchStatusRepo) ClaimMatch(ctx context.Context, matchID domain.MatchID, tatamiID string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE matches SET status='IN_PROGRESS', tatami_id=?
		WHERE id=? AND status='PENDING'`,
		tatamiID, matchID.String(),
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrMatchNotAvailable
	}
	return nil
}

// FinishMatch marks a match as FINISHED and stores the result JSON.
func (r *MatchStatusRepo) FinishMatch(ctx context.Context, matchID domain.MatchID, result *domain.MatchResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE matches SET status='FINISHED', result_json=? WHERE id=?`,
		string(data), matchID.String(),
	)
	return err
}

// ListByTournament returns all matches for all categories in a tournament.
func (r *MatchStatusRepo) ListByTournament(ctx context.Context, tournamentID domain.TournamentID) ([]ports.MatchRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.category_id, m.round, m.position, m.is_repechage,
		       COALESCE(m.athlete_a_id,''), COALESCE(m.athlete_b_id,''),
		       m.status, m.tatami_id, COALESCE(m.result_json,''),
		       COALESCE(a1.name,''), COALESCE(a1.club,''),
		       COALESCE(a2.name,''), COALESCE(a2.club,''),
		       c.division_id, d.weight_class, d.gender, d.age_group
		FROM matches m
		LEFT JOIN athletes a1 ON a1.id = m.athlete_a_id
		LEFT JOIN athletes a2 ON a2.id = m.athlete_b_id
		JOIN categories c ON c.id = m.category_id
		JOIN divisions d ON d.id = c.division_id
		WHERE d.tournament_id = ?
		ORDER BY d.weight_class, m.round, m.position`,
		tournamentID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []ports.MatchRow
	for rows.Next() {
		var mr ports.MatchRow
		var isRepechage int
		if err := rows.Scan(
			&mr.ID, &mr.CategoryID, &mr.Round, &mr.Position, &isRepechage,
			&mr.AthleteAID, &mr.AthleteBID, &mr.Status, &mr.TatamiID, &mr.ResultJSON,
			&mr.AthleteAName, &mr.AthleteAClub, &mr.AthleteBName, &mr.AthleteBClub,
			&mr.DivisionID, &mr.WeightClass, &mr.Gender, &mr.AgeGroup,
		); err != nil {
			return nil, err
		}
		mr.IsRepechage = isRepechage == 1
		result = append(result, mr)
	}
	return result, rows.Err()
}

// GetMatch returns a single match row by ID.
func (r *MatchStatusRepo) GetMatch(ctx context.Context, matchID domain.MatchID) (*ports.MatchRow, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT m.id, m.category_id, m.round, m.position, m.is_repechage,
		       COALESCE(m.athlete_a_id,''), COALESCE(m.athlete_b_id,''),
		       m.status, m.tatami_id, COALESCE(m.result_json,''),
		       COALESCE(a1.name,''), COALESCE(a1.club,''),
		       COALESCE(a2.name,''), COALESCE(a2.club,''),
		       c.division_id, d.weight_class, d.gender, d.age_group
		FROM matches m
		LEFT JOIN athletes a1 ON a1.id = m.athlete_a_id
		LEFT JOIN athletes a2 ON a2.id = m.athlete_b_id
		JOIN categories c ON c.id = m.category_id
		JOIN divisions d ON d.id = c.division_id
		WHERE m.id = ?`, matchID.String())

	var mr ports.MatchRow
	var isRepechage int
	if err := row.Scan(
		&mr.ID, &mr.CategoryID, &mr.Round, &mr.Position, &isRepechage,
		&mr.AthleteAID, &mr.AthleteBID, &mr.Status, &mr.TatamiID, &mr.ResultJSON,
		&mr.AthleteAName, &mr.AthleteAClub, &mr.AthleteBName, &mr.AthleteBClub,
		&mr.DivisionID, &mr.WeightClass, &mr.Gender, &mr.AgeGroup,
	); err != nil {
		return nil, fmt.Errorf("get match: %w", err)
	}
	mr.IsRepechage = isRepechage == 1
	return &mr, nil
}

// ErrMatchNotAvailable is returned when a match cannot be claimed.
var ErrMatchNotAvailable = errors.New("match is not available (already claimed or finished)")

// Ensure MatchStatusRepo satisfies the port at compile time.
var _ ports.MatchStatusRepo = (*MatchStatusRepo)(nil)
