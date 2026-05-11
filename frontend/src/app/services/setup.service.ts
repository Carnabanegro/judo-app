/**
 * SetupService — manages the state of the tournament setup wizard.
 *
 * The wizard has 4 steps:
 *   1. Select or create a tournament
 *   2. Add divisions
 *   3. Register athletes per division
 *   4. Generate brackets
 */
import { Injectable, signal, computed } from '@angular/core';
import { WailsService } from './wails.service';
import { TournamentDTO, DivisionDTO, AthleteDTO, BracketDTO } from '../models/setup';

export type SetupStep = 'tournament' | 'divisions' | 'athletes' | 'bracket';

@Injectable({ providedIn: 'root' })
export class SetupService {
  readonly tournaments = signal<TournamentDTO[]>([]);
  readonly activeTournament = signal<TournamentDTO | null>(null);

  readonly divisions = signal<DivisionDTO[]>([]);
  readonly activeDivision = signal<DivisionDTO | null>(null);

  readonly athletes = signal<AthleteDTO[]>([]);
  readonly bracket = signal<BracketDTO | null>(null);

  readonly step = signal<SetupStep>('tournament');

  readonly canProceedToDivisions = computed(() => this.activeTournament() !== null);
  readonly canProceedToAthletes = computed(() => this.activeDivision() !== null);
  readonly canGenerateBracket = computed(() => this.athletes().length >= 2);

  constructor(private wails: WailsService) {}

  // ── Step 1: Tournaments ───────────────────────────────────────────────────

  async loadTournaments(): Promise<void> {
    const list = await this.wails.listTournaments();
    this.tournaments.set(list ?? []);
  }

  async createTournament(name: string, location: string, date: string): Promise<void> {
    const t = await this.wails.createTournament(name, location, date);
    this.tournaments.update(ts => [...ts, t]);
    this.selectTournament(t);
  }

  selectTournament(t: TournamentDTO): void {
    this.activeTournament.set(t);
    this.divisions.set([]);
    this.activeDivision.set(null);
    this.athletes.set([]);
    this.bracket.set(null);
  }

  // ── Step 2: Divisions ─────────────────────────────────────────────────────

  async loadDivisions(): Promise<void> {
    const t = this.activeTournament();
    if (!t) return;
    const list = await this.wails.listDivisions(t.id);
    this.divisions.set(list ?? []);
  }

  async createDivision(ageGroup: string, gender: string, weightClass: string, format: string): Promise<void> {
    const t = this.activeTournament();
    if (!t) return;
    const d = await this.wails.createDivision(t.id, ageGroup, gender, weightClass, format);
    this.divisions.update(ds => [...ds, d]);
    this.selectDivision(d);
  }

  selectDivision(d: DivisionDTO): void {
    this.activeDivision.set(d);
    this.athletes.set([]);
    this.bracket.set(null);
  }

  // ── Step 3: Athletes ──────────────────────────────────────────────────────

  async loadAthletes(): Promise<void> {
    const d = this.activeDivision();
    if (!d) return;
    const list = await this.wails.listAthletes(d.id);
    this.athletes.set(list ?? []);
  }

  async registerAthlete(name: string, club: string, weight: number, birthDate: string): Promise<void> {
    const d = this.activeDivision();
    if (!d) return;
    const a = await this.wails.registerAthlete(d.id, name, club, weight, birthDate);
    this.athletes.update(as => [...as, a]);
  }

  // ── Step 4: Bracket ───────────────────────────────────────────────────────

  async generateBracket(): Promise<void> {
    const d = this.activeDivision();
    if (!d) return;
    await this.wails.generateBracket(d.id);
    const b = await this.wails.getBracket(d.id);
    this.bracket.set(b);
  }

  // ── Navigation ────────────────────────────────────────────────────────────

  goTo(step: SetupStep): void {
    this.step.set(step);
  }

  // Set the active tournament by ID (UI helper).
  setActiveTournament(id: string): void {
    const t = this.tournaments().find(x => x.id === id) ?? null;
    this.activeTournament.set(t);
  }
}
