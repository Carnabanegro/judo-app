package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// AgeGroup represents the competitor age division.
type AgeGroup string

const (
	AgeGroupCadet  AgeGroup = "CADET"
	AgeGroupJunior AgeGroup = "JUNIOR"
	AgeGroupSenior AgeGroup = "SENIOR"
	AgeGroupU18    AgeGroup = "U18"
)

// Gender represents the gender division.
type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderMixed  Gender = "MIXED"
)

// Format represents the competition format.
type Format string

const (
	FormatIndividualIJF Format = "INDIVIDUAL_IJF"
	// FormatTeams is reserved for future use.
	FormatTeams Format = "TEAMS"
)

// TournamentID is a type-safe tournament identifier.
type TournamentID = uuid.UUID

// DivisionID is a type-safe division identifier.
type DivisionID = uuid.UUID

// CategoryID is a type-safe category identifier.
type CategoryID = uuid.UUID

// AthleteID is a type-safe athlete identifier.
type AthleteID = uuid.UUID

// Tournament represents a judo competition event.
type Tournament struct {
	ID        TournamentID
	Name      string
	Date      time.Time
	Location  string
	CreatedAt time.Time
}

// NewTournament creates a new Tournament with a generated ID.
func NewTournament(name, location string, date time.Time) (*Tournament, error) {
	if name == "" {
		return nil, errors.New("tournament name must not be empty")
	}
	return &Tournament{
		ID:        uuid.New(),
		Name:      name,
		Date:      date,
		Location:  location,
		CreatedAt: time.Now(),
	}, nil
}

// Division represents a competitive division within a tournament.
type Division struct {
	ID           DivisionID
	TournamentID TournamentID
	AgeGroup     AgeGroup
	Gender       Gender
	WeightClass  string // e.g. "-60kg", "+100kg"
	Format       Format
}

// NewDivision creates a new Division with a generated ID.
func NewDivision(tournamentID TournamentID, ageGroup AgeGroup, gender Gender, weightClass string, format Format) (*Division, error) {
	if weightClass == "" {
		return nil, errors.New("weight class must not be empty")
	}
	return &Division{
		ID:           uuid.New(),
		TournamentID: tournamentID,
		AgeGroup:     ageGroup,
		Gender:       gender,
		WeightClass:  weightClass,
		Format:       format,
	}, nil
}

// Category groups athletes within a division for competition.
type Category struct {
	ID         CategoryID
	DivisionID DivisionID
	Athletes   []*Athlete
}

// NewCategory creates a new Category with a generated ID.
func NewCategory(divisionID DivisionID) *Category {
	return &Category{
		ID:         uuid.New(),
		DivisionID: divisionID,
		Athletes:   make([]*Athlete, 0),
	}
}

// AddAthlete registers an athlete to the category.
func (c *Category) AddAthlete(a *Athlete) {
	c.Athletes = append(c.Athletes, a)
}

// Athlete represents a competitor.
type Athlete struct {
	ID        AthleteID
	Name      string
	Club      string
	Weight    float64
	BirthDate time.Time
}

// NewAthlete creates a new Athlete with a generated ID.
func NewAthlete(name, club string, weight float64, birthDate time.Time) (*Athlete, error) {
	if name == "" {
		return nil, errors.New("athlete name must not be empty")
	}
	if weight < 0 {
		return nil, errors.New("athlete weight must not be negative")
	}
	return &Athlete{
		ID:        uuid.New(),
		Name:      name,
		Club:      club,
		Weight:    weight,
		BirthDate: birthDate,
	}, nil
}
