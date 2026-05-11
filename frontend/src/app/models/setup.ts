// Models matching the Go DTOs returned by the Wails backend.

export interface TournamentDTO {
  id: string;
  name: string;
  location: string;
  date: string; // YYYY-MM-DD
}

export interface DivisionDTO {
  id: string;
  tournamentId: string;
  ageGroup: 'CADET' | 'JUNIOR' | 'SENIOR' | 'U18';
  gender: 'MALE' | 'FEMALE' | 'MIXED';
  weightClass: string;
  format: 'INDIVIDUAL_IJF' | 'TEAMS';
}

export interface AthleteDTO {
  id: string;
  categoryId: string;
  name: string;
  club: string;
  weight: number;
  birthDate: string; // YYYY-MM-DD
}

export interface MatchDTO {
  id: string;
  round: number;
  position: number;
  athleteA?: AthleteDTO;
  athleteB?: AthleteDTO;
  status: 'PENDING' | 'IN_PROGRESS' | 'FINISHED';
  tatamiId: string;
  isRepechage: boolean;
  winnerId?: string;
  method?: string;
}

export interface RoundDTO {
  number: number;
  matches: MatchDTO[];
}

export interface BracketDTO {
  categoryId: string;
  rounds: RoundDTO[];
  repechage: RoundDTO[];
}

export interface MatchRow {
  id: string;
  categoryId: string;
  round: number;
  position: number;
  isRepechage: boolean;
  athleteAId: string;
  athleteBId: string;
  status: 'PENDING' | 'IN_PROGRESS' | 'FINISHED';
  tatamiId: string;
  resultJson: string;
  athleteAName: string;
  athleteAClub: string;
  athleteBName: string;
  athleteBClub: string;
  divisionId: string;
  weightClass: string;
  gender: string;
  ageGroup: string;
}
