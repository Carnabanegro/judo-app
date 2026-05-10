/**
 * WailsService — thin bridge to the Go backend via the Wails runtime.
 *
 * In development (served by `ng serve`) the `window.go` object does NOT exist.
 * All methods return resolved promises with safe default values so the Angular
 * app can be developed and tested without the Wails shell.
 */
import { Injectable } from '@angular/core';

// Type stub — the real implementation is injected by the Wails runtime.
declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
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
        };
      };
    };
  }
}

function wailsApp() {
  return window.go?.main?.App;
}

@Injectable({ providedIn: 'root' })
export class WailsService {
  private get app() {
    return wailsApp();
  }

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
}
