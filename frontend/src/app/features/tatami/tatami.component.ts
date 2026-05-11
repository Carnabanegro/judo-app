import { Component, OnInit, OnDestroy, inject, signal, computed } from '@angular/core';
import { Subscription, filter } from 'rxjs';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { TatamiOperatorService } from '../../services/tatami-operator.service';
import { MatchRow } from '../../models/setup';
import { CombatWsService } from '../../services/combat-ws.service';


@Component({
  selector: 'app-tatami',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './tatami.component.html',
  styleUrl: './tatami.component.scss'
})
export class TatamiComponent implements OnInit, OnDestroy {
  private svc = inject(TatamiOperatorService);
  private fb = inject(FormBuilder);
  private ws = inject(CombatWsService);

  // ── State ─────────────────────────────────────────────────────────────────

  readonly tatamiId = signal<string>('');
  readonly tournamentId = signal<string>('');
  readonly matches = signal<MatchRow[]>([]);
  readonly loading = signal(false);
  readonly error = signal<string | null>(null);
  readonly claimingId = signal<string | null>(null);

  readonly pendingMatches = computed(() =>
    this.matches().filter(m => m.status === 'PENDING')
  );
  readonly myMatches = computed(() =>
    this.matches().filter(m => m.tatamiId === this.tatamiId() && m.status === 'IN_PROGRESS')
  );
  readonly finishedMatches = computed(() =>
    this.matches().filter(m => m.status === 'FINISHED')
  );

  // ── Forms ─────────────────────────────────────────────────────────────────

  readonly setupForm = this.fb.group({
    tournamentId: ['', Validators.required],
    tatamiId:     ['', Validators.required],
  });

  readonly resultForm = this.fb.group({
    matchId:    ['', Validators.required],
    categoryId: ['', Validators.required],
    winnerIdx:  [0, Validators.required],
    method:     ['IPPON', Validators.required],
  });

  readonly METHODS = ['IPPON', 'WAZA_ARI_AWASETE_IPPON', 'HANSOKU_MAKE', 'KIKEN_GACHI', 'FUSEN_GACHI', 'GOLDEN_SCORE'];

  private wsSub: Subscription | null = null;

  // ── Lifecycle ─────────────────────────────────────────────────────────────

  ngOnInit(): void {
    // Restore session if stored.
    const stored = sessionStorage.getItem('tatami-config');
    if (stored) {
      const cfg = JSON.parse(stored);
      this.tatamiId.set(cfg.tatamiId ?? '');
      this.tournamentId.set(cfg.tournamentId ?? '');
      if (this.tatamiId() && this.tournamentId()) {
        this.ws.connect();
        this.wsSub?.unsubscribe();
        this.wsSub = this.ws.events$.pipe(filter(e => e.type === 'bracket:update')).subscribe(() => this.loadMatches());
      }

    }
  }

  ngOnDestroy(): void {
    if (this.wsSub) {
      this.wsSub.unsubscribe();
      this.wsSub = null;
    }
    this.ws.disconnect();
  }

  // ── Setup ─────────────────────────────────────────────────────────────────

  async connect(): Promise<void> {
    if (this.setupForm.invalid) return;
    const { tournamentId, tatamiId } = this.setupForm.value;
    this.tournamentId.set(tournamentId!);
    this.tatamiId.set(tatamiId!);
    sessionStorage.setItem('tatami-config', JSON.stringify({ tournamentId, tatamiId }));
    await this.loadMatches();
    this.ws.connect();
    this.wsSub?.unsubscribe();
    this.wsSub = this.ws.events$.pipe(filter(e => e.type === 'bracket:update')).subscribe(() => this.loadMatches());
  }

  disconnect(): void {
    if (this.wsSub) {
      this.wsSub.unsubscribe();
      this.wsSub = null;
    }
    this.ws.disconnect();
    this.tatamiId.set('');
    this.tournamentId.set('');
    this.matches.set([]);
    sessionStorage.removeItem('tatami-config');
    this.setupForm.reset();
  }

  // ── Matches ───────────────────────────────────────────────────────────────

  async loadMatches(): Promise<void> {
    const tId = this.tournamentId();
    if (!tId) return;
    try {
      const rows = await this.svc.listMatches(tId);
      this.matches.set(rows ?? []);
      this.error.set(null);
    } catch (e: unknown) {
      this.error.set(e instanceof Error ? e.message : String(e));
    }
  }

  async claim(m: MatchRow): Promise<void> {
    this.claimingId.set(m.id);
    this.error.set(null);
    try {
      await this.svc.claimMatch(m.id, this.tatamiId(), m.athleteAName || 'A', m.athleteBName || 'B');
      await this.loadMatches();
    } catch (e: unknown) {
      this.error.set(e instanceof Error ? e.message : String(e));
    } finally {
      this.claimingId.set(null);
    }
  }

  async recordResult(): Promise<void> {
    if (this.resultForm.invalid) return;
    const { matchId, categoryId, winnerIdx, method } = this.resultForm.value;
    this.loading.set(true);
    this.error.set(null);
    try {
      await this.svc.recordResult(categoryId!, matchId!, winnerIdx!, method!);
      this.resultForm.reset({ method: 'IPPON', winnerIdx: 0 });
      await this.loadMatches();
    } catch (e: unknown) {
      this.error.set(e instanceof Error ? e.message : String(e));
    } finally {
      this.loading.set(false);
    }
  }

  prefillResult(m: MatchRow): void {
    this.resultForm.patchValue({ matchId: m.id, categoryId: m.categoryId });
  }

  // ── WebSocket updates ─────────────────────────────────────────────────────
  // Uses CombatWsService to receive `bracket:update` events and refresh matches.

  // ── Helpers ───────────────────────────────────────────────────────────────

  readonly isConnected = computed(() => !!this.tatamiId() && !!this.tournamentId());

  athleteName(m: MatchRow, idx: 0 | 1): string {
    return idx === 0 ? (m.athleteAName || 'TBD') : (m.athleteBName || 'TBD');
  }
}
