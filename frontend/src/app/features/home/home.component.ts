import { Component, inject, signal, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';
import { WailsService } from '../../services/wails.service';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [],
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent implements OnDestroy {
  private router = inject(Router);
  private wails = inject(WailsService);

  readonly active = signal<{id:string,name:string,location:string,date:string} | null>(null);
  readonly loading = signal(false);

  // Evaluated lazily each call so it works even when window.go is injected async by Wails
  get isWails(): boolean {
    return !!(window as any).go?.main?.App;
  }

  private pollInterval: ReturnType<typeof setInterval> | null = null;

  constructor() {
    // Give Wails runtime 300ms to inject window.go before first load
    setTimeout(() => {
      this.loadActive();
      this.pollInterval = setInterval(() => {
        if (!this.active()) this.loadActive();
      }, 3000);
    }, 300);
  }

  ngOnDestroy() {
    if (this.pollInterval !== null) {
      clearInterval(this.pollInterval);
    }
  }

  async loadActive(): Promise<void> {
    this.loading.set(true);
    try {
      if (this.isWails) {
        const t = await this.wails.getActiveTournament();
        this.active.set(t);
      } else {
        const res = await fetch('/api/active-tournament');
        this.active.set(res.ok ? await res.json() : null);
      }
    } catch {
      this.active.set(null);
    } finally {
      this.loading.set(false);
    }
  }

  go(path: string) {
    this.router.navigate([path]);
  }
}
