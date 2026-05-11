import { Injectable } from '@angular/core';
import { Observable, Subject, share } from 'rxjs';
import { DisplayEvent } from '../models/combat';

const WS_URL = 'ws://localhost:8080/ws';
const RECONNECT_DELAY_MS = 2000;

@Injectable({ providedIn: 'root' })
export class CombatWsService {
  private ws: WebSocket | null = null;
  private readonly _events$ = new Subject<DisplayEvent>();

  /** Stream of all display events from the Go server. */
  readonly events$: Observable<DisplayEvent> = this._events$.pipe(share());

  connect(): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) return;

    this.ws = new WebSocket(WS_URL);

    this.ws.onmessage = (ev) => {
      try {
        const event: DisplayEvent = JSON.parse(ev.data as string);
        this._events$.next(event);
      } catch {
        // ignore malformed frames
      }
    };

    this.ws.onclose = () => {
      setTimeout(() => this.connect(), RECONNECT_DELAY_MS);
    };

    this.ws.onerror = () => {
      this.ws?.close();
    };
  }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }
}
