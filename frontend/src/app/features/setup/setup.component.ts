import { Component, OnInit, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { SetupService } from '../../services/setup.service';
import { TournamentDTO, DivisionDTO } from '../../models/setup';
import { WailsService } from '../../services/wails.service';

@Component({
  selector: 'app-setup',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './setup.component.html',
  styleUrl: './setup.component.scss'
})
export class SetupComponent implements OnInit {
  readonly setup = inject(SetupService);
  private wails = inject(WailsService);
  private fb = inject(FormBuilder);

  readonly loading = signal(false);
  readonly error = signal<string | null>(null);

  // ── Forms ─────────────────────────────────────────────────────────────────

  readonly tournamentForm = this.fb.group({
    name:     ['', Validators.required],
    location: ['', Validators.required],
    date:     ['', Validators.required],
  });

  readonly divisionForm = this.fb.group({
    ageGroup:    ['SENIOR', Validators.required],
    gender:      ['MALE', Validators.required],
    weightClass: ['', Validators.required],
    format:      ['INDIVIDUAL_IJF', Validators.required],
  });

  readonly athleteForm = this.fb.group({
    name:      ['', Validators.required],
    club:      [''],
    weight:    [null as number | null, [Validators.required, Validators.min(0)]],
    birthDate: ['', Validators.required],
  });

  readonly AGE_GROUPS = ['CADET', 'JUNIOR', 'U18', 'SENIOR'];
  readonly GENDERS    = ['MALE', 'FEMALE', 'MIXED'];
  readonly FORMATS    = ['INDIVIDUAL_IJF', 'TEAMS'];

  // ── Lifecycle ─────────────────────────────────────────────────────────────

  async ngOnInit(): Promise<void> {
    await this.run(async () => {
      await this.setup.loadTournaments();
      // Load active tournament to display indicator
      try {
        const res = await fetch('/api/active-tournament');
        if (res.ok) {
          const active = await res.json();
          this.setup.setActiveTournament && this.setup.setActiveTournament(active.id);
        }
      } catch {
        // ignore
      }
    });
  }

  async activateTournament(t: TournamentDTO): Promise<void> {
    try {
      await this.wails.activateTournament(t.id);
      // reflect in UI
      this.setup.setActiveTournament(t.id);
    } catch (e) {
      // ignore for now
    }
  }

  // ── Step 1: Tournament ────────────────────────────────────────────────────

  async createTournament(): Promise<void> {
    if (this.tournamentForm.invalid) return;
    const { name, location, date } = this.tournamentForm.value;
    await this.run(() => this.setup.createTournament(name!, location!, date!));
    this.tournamentForm.reset();
    this.setup.goTo('divisions');
    await this.run(() => this.setup.loadDivisions());
  }

  async selectAndProceed(t: TournamentDTO): Promise<void> {
    this.setup.selectTournament(t);
    this.setup.goTo('divisions');
    await this.run(() => this.setup.loadDivisions());
  }

  // ── Step 2: Divisions ─────────────────────────────────────────────────────

  async createDivision(): Promise<void> {
    if (this.divisionForm.invalid) return;
    const { ageGroup, gender, weightClass, format } = this.divisionForm.value;
    await this.run(() => this.setup.createDivision(ageGroup!, gender!, weightClass!, format!));
    this.divisionForm.patchValue({ weightClass: '' });
  }

  async selectDivisionAndProceed(d: DivisionDTO): Promise<void> {
    this.setup.selectDivision(d);
    this.setup.goTo('athletes');
    await this.run(() => this.setup.loadAthletes());
  }

  // ── Step 3: Athletes ──────────────────────────────────────────────────────

  async registerAthlete(): Promise<void> {
    if (this.athleteForm.invalid) return;
    const { name, club, weight, birthDate } = this.athleteForm.value;
    await this.run(() => this.setup.registerAthlete(name!, club ?? '', weight!, birthDate!));
    this.athleteForm.reset();
  }

  // ── Step 4: Bracket ───────────────────────────────────────────────────────

  async generateBracket(): Promise<void> {
    this.setup.goTo('bracket');
    await this.run(() => this.setup.generateBracket());
  }

  // ── Helpers ───────────────────────────────────────────────────────────────

  private async run(fn: () => Promise<void>): Promise<void> {
    this.loading.set(true);
    this.error.set(null);
    try {
      await fn();
    } catch (e: unknown) {
      this.error.set(e instanceof Error ? e.message : String(e));
    } finally {
      this.loading.set(false);
    }
  }
}
