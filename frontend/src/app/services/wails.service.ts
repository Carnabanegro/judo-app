/**
 * WailsService — thin bridge to the Go backend via the Wails runtime.
 *
 * In development (served by `ng serve`) the `window.go` object does NOT exist.
 * All methods return resolved promises with safe default values so the Angular
 * app can be developed and tested without the Wails shell.
 */
import { Injectable } from '@angular/core';
import { TournamentDTO, DivisionDTO, AthleteDTO, BracketDTO, MatchRow } from '../models/setup';

// Type stub — the real implementation is injected by the Wails runtime.
declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          // Combat
          StartMatch(matchId: string, labelA: string, labelB: string): Promise<void>;
          Pause(matchId: string): Promise<void>;
          Resume(matchId: string): Promise<void>;
          Ippon(matchId: string, athleteIdx: number): Promise<void>;
          WazaAri(matchId: string, athleteIdx: number): Promise<void>;
          Yuko(matchId: string, athleteIdx: number): Promise<void>;
          Shido(matchId: string, athleteIdx: number): Promise<void>;
          StartOsaekomi(matchId: string): Promise<void>;
          StopOsaekomi(matchId: string, athleteIdx: number): Promise<string>;
          Finish(matchId: string): Promise<void>;
          StartPractice(labelA: string, labelB: string): Promise<string>;
          // Setup
          CreateTournament(name: string, location: string, dateISO: string): Promise<TournamentDTO>;
          ListTournaments(): Promise<TournamentDTO[]>;
          CreateDivision(tournamentId: string, ageGroup: string, gender: string, weightClass: string, format: string): Promise<DivisionDTO>;
          ListDivisions(tournamentId: string): Promise<DivisionDTO[]>;
          RegisterAthlete(divisionId: string, name: string, club: string, weight: number, birthDateISO: string): Promise<AthleteDTO>;
          ListAthletes(divisionId: string): Promise<AthleteDTO[]>;
          GenerateBracket(divisionId: string): Promise<void>;
          GetBracket(divisionId: string): Promise<BracketDTO>;
          // Tatami
          ListMatches(tournamentId: string): Promise<MatchRow[]>;
          ClaimMatch(matchId: string, tatamiId: string, labelA: string, labelB: string): Promise<void>;
          RecordMatchResult(categoryId: string, matchId: string, winnerIdx: number, method: string): Promise<void>;
        };
      };
    };
  }
}

function wailsApp() {
  return window.go?.main?.App;
}

const DEV_TOURNAMENT: TournamentDTO = { id: 'dev-t-1', name: 'Dev Tournament', location: 'Localhost', date: '2026-01-01' };
const DEV_DIVISION: DivisionDTO = { id: 'dev-d-1', tournamentId: 'dev-t-1', ageGroup: 'SENIOR', gender: 'MALE', weightClass: '-66kg', format: 'INDIVIDUAL_IJF' };
const DEV_ATHLETE: AthleteDTO = { id: 'dev-a-1', categoryId: 'dev-c-1', name: 'Dev Athlete', club: 'Dev Club', weight: 65, birthDate: '2000-01-01' };
const DEV_BRACKET: BracketDTO = { categoryId: 'dev-c-1', rounds: [], repechage: [] };

@Injectable({ providedIn: 'root' })
export class WailsService {
  private get app() {
    return wailsApp();
  }

  // ── Combat ─────────────────────────────────────────────────────────────────

  startMatch(matchId: string, labelA: string, labelB: string): Promise<void> {
    return this.app?.StartMatch(matchId, labelA, labelB) ?? Promise.resolve();
  }

  pause(matchId: string): Promise<void> {
    return this.app?.Pause(matchId) ?? Promise.resolve();
  }

  resume(matchId: string): Promise<void> {
    return this.app?.Resume(matchId) ?? Promise.resolve();
  }

  ippon(matchId: string, athleteIdx: number): Promise<void> {
    return this.app?.Ippon(matchId, athleteIdx) ?? Promise.resolve();
  }

  wazaAri(matchId: string, athleteIdx: number): Promise<void> {
    return this.app?.WazaAri(matchId, athleteIdx) ?? Promise.resolve();
  }

  yuko(matchId: string, athleteIdx: number): Promise<void> {
    return this.app?.Yuko(matchId, athleteIdx) ?? Promise.resolve();
  }

  shido(matchId: string, athleteIdx: number): Promise<void> {
    return this.app?.Shido(matchId, athleteIdx) ?? Promise.resolve();
  }

  startOsaekomi(matchId: string): Promise<void> {
    return this.app?.StartOsaekomi(matchId) ?? Promise.resolve();
  }

  stopOsaekomi(matchId: string, athleteIdx: number): Promise<string> {
    return this.app?.StopOsaekomi(matchId, athleteIdx) ?? Promise.resolve('');
  }

  finish(matchId: string): Promise<void> {
    return this.app?.Finish(matchId) ?? Promise.resolve();
  }

  startPractice(labelA: string, labelB: string): Promise<string> {
    return this.app?.StartPractice(labelA, labelB) ?? Promise.resolve('practice-dev-id');
  }

  // ── Setup ──────────────────────────────────────────────────────────────────

  createTournament(name: string, location: string, dateISO: string): Promise<TournamentDTO> {
    return this.app?.CreateTournament(name, location, dateISO) ?? Promise.resolve({ ...DEV_TOURNAMENT, name, location, date: dateISO });
  }

  listTournaments(): Promise<TournamentDTO[]> {
    return this.app?.ListTournaments() ?? Promise.resolve([DEV_TOURNAMENT]);
  }

  createDivision(tournamentId: string, ageGroup: string, gender: string, weightClass: string, format: string): Promise<DivisionDTO> {
    return this.app?.CreateDivision(tournamentId, ageGroup, gender, weightClass, format)
      ?? Promise.resolve({ ...DEV_DIVISION, tournamentId, ageGroup: ageGroup as DivisionDTO['ageGroup'], gender: gender as DivisionDTO['gender'], weightClass });
  }

  listDivisions(tournamentId: string): Promise<DivisionDTO[]> {
    return this.app?.ListDivisions(tournamentId) ?? Promise.resolve([DEV_DIVISION]);
  }

  registerAthlete(divisionId: string, name: string, club: string, weight: number, birthDateISO: string): Promise<AthleteDTO> {
    return this.app?.RegisterAthlete(divisionId, name, club, weight, birthDateISO)
      ?? Promise.resolve({ ...DEV_ATHLETE, name, club, weight, birthDate: birthDateISO });
  }

  listAthletes(divisionId: string): Promise<AthleteDTO[]> {
    return this.app?.ListAthletes(divisionId) ?? Promise.resolve([DEV_ATHLETE]);
  }

  generateBracket(divisionId: string): Promise<void> {
    return this.app?.GenerateBracket(divisionId) ?? Promise.resolve();
  }

  getBracket(divisionId: string): Promise<BracketDTO> {
    return this.app?.GetBracket(divisionId) ?? Promise.resolve({ ...DEV_BRACKET });
  }

  // ── Tatami ─────────────────────────────────────────────────────────────────

  listMatches(tournamentId: string): Promise<MatchRow[]> {
    return this.app?.ListMatches(tournamentId) ?? Promise.resolve([]);
  }

  claimMatch(matchId: string, tatamiId: string, labelA: string, labelB: string): Promise<void> {
    return this.app?.ClaimMatch(matchId, tatamiId, labelA, labelB) ?? Promise.resolve();
  }

  recordMatchResult(categoryId: string, matchId: string, winnerIdx: number, method: string): Promise<void> {
    return this.app?.RecordMatchResult(categoryId, matchId, winnerIdx, method) ?? Promise.resolve();
  }
}
