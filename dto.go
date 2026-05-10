package main

import (
	"judo-app/internal/domain"
)

// ── DTOs ─────────────────────────────────────────────────────────────────────

// TournamentDTO is the Wails-serialisable representation of a Tournament.
type TournamentDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	Date     string `json:"date"` // YYYY-MM-DD
}

// DivisionDTO is the Wails-serialisable representation of a Division.
type DivisionDTO struct {
	ID           string `json:"id"`
	TournamentID string `json:"tournamentId"`
	AgeGroup     string `json:"ageGroup"`
	Gender       string `json:"gender"`
	WeightClass  string `json:"weightClass"`
	Format       string `json:"format"`
}

// AthleteDTO is the Wails-serialisable representation of an Athlete.
type AthleteDTO struct {
	ID         string  `json:"id"`
	CategoryID string  `json:"categoryId"`
	Name       string  `json:"name"`
	Club       string  `json:"club"`
	Weight     float64 `json:"weight"`
	BirthDate  string  `json:"birthDate"` // YYYY-MM-DD
}

// MatchDTO is the Wails-serialisable representation of a Match (bracket node).
type MatchDTO struct {
	ID          string      `json:"id"`
	Round       int         `json:"round"`
	Position    int         `json:"position"`
	AthleteA    *AthleteDTO `json:"athleteA,omitempty"`
	AthleteB    *AthleteDTO `json:"athleteB,omitempty"`
	Status      string      `json:"status"`
	TatamiID    string      `json:"tatamiId"`
	IsRepechage bool        `json:"isRepechage"`
	WinnerID    string      `json:"winnerId,omitempty"`
	Method      string      `json:"method,omitempty"`
}

// RoundDTO groups matches by round number.
type RoundDTO struct {
	Number  int        `json:"number"`
	Matches []MatchDTO `json:"matches"`
}

// BracketDTO is the Wails-serialisable representation of a Bracket.
type BracketDTO struct {
	CategoryID string     `json:"categoryId"`
	Rounds     []RoundDTO `json:"rounds"`
	Repechage  []RoundDTO `json:"repechage"`
}

// ── Mappers ───────────────────────────────────────────────────────────────────

func tournamentToDTO(t *domain.Tournament) *TournamentDTO {
	return &TournamentDTO{
		ID:       t.ID.String(),
		Name:     t.Name,
		Location: t.Location,
		Date:     t.Date.Format("2006-01-02"),
	}
}

func divisionToDTO(d *domain.Division) *DivisionDTO {
	return &DivisionDTO{
		ID:           d.ID.String(),
		TournamentID: d.TournamentID.String(),
		AgeGroup:     string(d.AgeGroup),
		Gender:       string(d.Gender),
		WeightClass:  d.WeightClass,
		Format:       string(d.Format),
	}
}

func athleteToDTO(a *domain.Athlete, categoryID string) *AthleteDTO {
	return &AthleteDTO{
		ID:         a.ID.String(),
		CategoryID: categoryID,
		Name:       a.Name,
		Club:       a.Club,
		Weight:     a.Weight,
		BirthDate:  a.BirthDate.Format("2006-01-02"),
	}
}

func matchToDTO(m *domain.Match) MatchDTO {
	dto := MatchDTO{
		ID:          m.ID.String(),
		Round:       m.Round,
		Position:    m.Position,
		Status:      string(m.Status),
		TatamiID:    m.TatamiID,
		IsRepechage: m.IsRepechage,
	}
	if m.AthleteA != nil {
		a := athleteToDTO(m.AthleteA, "")
		dto.AthleteA = a
	}
	if m.AthleteB != nil {
		b := athleteToDTO(m.AthleteB, "")
		dto.AthleteB = b
	}
	if m.Result != nil {
		dto.WinnerID = m.Result.WinnerID.String()
		dto.Method = string(m.Result.Method)
	}
	return dto
}

func bracketToDTO(b *domain.Bracket) *BracketDTO {
	dto := &BracketDTO{
		CategoryID: b.CategoryID.String(),
	}
	for _, r := range b.Rounds {
		rd := RoundDTO{Number: r.Number}
		for _, m := range r.Matches {
			rd.Matches = append(rd.Matches, matchToDTO(m))
		}
		dto.Rounds = append(dto.Rounds, rd)
	}
	for i, pool := range b.Repechage {
		rd := RoundDTO{Number: i + 1}
		for _, m := range pool.Matches {
			rd.Matches = append(rd.Matches, matchToDTO(m))
		}
		dto.Repechage = append(dto.Repechage, rd)
	}
	return dto
}

// ── Domain enum helpers ───────────────────────────────────────────────────────

// domain_AgeGroup converts a string to domain.AgeGroup (defaults to SENIOR).
func domain_AgeGroup(s string) domain.AgeGroup {
	switch s {
	case "CADET":
		return domain.AgeGroupCadet
	case "JUNIOR":
		return domain.AgeGroupJunior
	case "U18":
		return domain.AgeGroupU18
	default:
		return domain.AgeGroupSenior
	}
}

// domain_Gender converts a string to domain.Gender (defaults to MIXED).
func domain_Gender(s string) domain.Gender {
	switch s {
	case "MALE":
		return domain.GenderMale
	case "FEMALE":
		return domain.GenderFemale
	default:
		return domain.GenderMixed
	}
}

// domain_Format converts a string to domain.Format (defaults to INDIVIDUAL_IJF).
func domain_Format(s string) domain.Format {
	if s == "TEAMS" {
		return domain.FormatTeams
	}
	return domain.FormatIndividualIJF
}

// domain_FinishMethod converts a string to domain.FinishMethod (defaults to IPPON).
func domain_FinishMethod(s string) domain.FinishMethod {
	switch s {
	case "WAZA_ARI_AWASETE_IPPON":
		return domain.FinishWazaAriAwasete
	case "HANSOKU_MAKE":
		return domain.FinishHansokuMake
	case "KIKEN_GACHI":
		return domain.FinishKikenGachi
	case "FUSEN_GACHI":
		return domain.FinishFusenGachi
	case "GOLDEN_SCORE":
		return domain.FinishGoldenScore
	default:
		return domain.FinishIppon
	}
}
