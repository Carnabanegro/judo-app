/**
 * TatamiOperatorService — bridge between the /tatami UI and the Go backend.
 *
 * When running inside the Wails desktop app: uses window.go bindings.
 * When running in a remote browser (operator at http://<host>:8080): uses the
 * REST API exposed by the display server at the same origin.
 */
import { Injectable, inject } from '@angular/core';
import { WailsService } from './wails.service';
import { MatchRow } from '../models/setup';

@Injectable({ providedIn: 'root' })
export class TatamiOperatorService {
  private wails = inject(WailsService);

  /** True when running inside the Wails shell (desktop app). */
  private get isWails(): boolean {
    return !!window.go?.main?.App;
  }

  // ── Matches ───────────────────────────────────────────────────────────────

  async listMatches(tournamentId: string): Promise<MatchRow[]> {
    if (this.isWails) {
      return this.wails.listMatches(tournamentId);
    }
    const url = tournamentId ? `/api/matches?tournamentId=${encodeURIComponent(tournamentId)}` : `/api/matches`;
    const res = await fetch(url);
    if (!res.ok) throw new Error(`listMatches: ${res.statusText}`);
    return res.json();
  }

  async getActiveTournament(): Promise<{id:string,name:string,location:string,date:string} | null> {
    if (this.isWails) {
      // Wails shell: no binding yet for active tournament; backend may still expose API
      return null;
    }
    const res = await fetch('/api/active-tournament');
    if (res.status === 404) return null;
    if (!res.ok) throw new Error(`getActiveTournament: ${res.statusText}`);
    return res.json();
  }

  async claimMatch(matchId: string, tatamiId: string, labelA: string, labelB: string): Promise<void> {
    if (this.isWails) {
      return this.wails.claimMatch(matchId, tatamiId, labelA, labelB);
    }
    const res = await fetch(`/api/matches/${encodeURIComponent(matchId)}/claim`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tatamiId, labelA, labelB }),
    });
    if (!res.ok) throw new Error(`claimMatch: ${res.statusText}`);
  }

  async recordResult(categoryId: string, matchId: string, winnerIdx: number, method: string): Promise<void> {
    if (this.isWails) {
      return this.wails.recordMatchResult(categoryId, matchId, winnerIdx, method);
    }
    const res = await fetch(`/api/matches/${encodeURIComponent(matchId)}/result`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ categoryId, winnerIdx, method }),
    });
    if (!res.ok) throw new Error(`recordResult: ${res.statusText}`);
  }
}
