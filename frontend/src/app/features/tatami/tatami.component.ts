import { Component, OnInit, OnDestroy, inject, signal, computed } from '@angular/core';
import { Subscription, filter } from 'rxjs';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { TatamiOperatorService } from '../../services/tatami-operator.service';
import { CombatWsService } from '../../services/combat-ws.service';
import { MatchRow } from '../../models/setup';

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

  // Removed full setup form — active tournament used and simple tatami selector only

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
      if (this.tatamiId()) {
        this.connectWs();
      }
    }

    // Try to auto-resolve active tournament (browser mode)
    (async () => {
      try {
        const active = await this.svc.getActiveTournament();
        if (active) {
          this.tournamentId.set(active.id);
          sessionStorage.setItem('tatami-config', JSON.stringify({ tournamentId: active.id, tatamiId: this.tatamiId() }));
          if (this.tatamiId()) this.connectWs();
        }
      } catch (_) {
        // ignore
      }
    })();
  }

  ngOnDestroy(): void {
    this.disconnectWs();
  }

  // ── Setup ─────────────────────────────────────────────────────────────────

  // Connect with simple tatami number — tournament is auto-selected via active tournament
  async connect(): Promise<void> {
    if (!this.tatamiId()) return;
    sessionStorage.setItem('tatami-config', JSON.stringify({ tournamentId: this.tournamentId(), tatamiId: this.tatamiId() }));
    await this.loadMatches();
    this.connectWs();
  }

  disconnect(): void {
    this.disconnectWs();
    this.tatamiId.set('');
    this.tournamentId.set('');
    this.matches.set([]);
    sessionStorage.removeItem('tatami-config');
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

  // ── WebSocket ─────────────────────────────────────────────────────────────

  private connectWs(): void {
    this.ws.connect();
    this.wsSub?.unsubscribe();
    this.wsSub = this.ws.events$
      .pipe(filter(e => e.type === 'bracket:update'))
      .subscribe(() => this.loadMatches());
  }

  private disconnectWs(): void {
    this.wsSub?.unsubscribe();
    this.wsSub = null;
    this.ws.disconnect();
  }

  // ── Helpers ───────────────────────────────────────────────────────────────

  readonly isConnected = computed(() => !!this.tatamiId() && !!this.tournamentId());

  athleteName(m: MatchRow, idx: 0 | 1): string {
    return idx === 0 ? (m.athleteAName || 'TBD') : (m.athleteBName || 'TBD');
  }
}
