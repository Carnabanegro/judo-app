import { Component, inject, signal, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [],
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent implements OnDestroy {
  private router = inject(Router);

  readonly active = signal<{id:string,name:string,location:string,date:string} | null>(null);
  readonly loading = signal(false);
  readonly isWails = !!(window as any).go?.main?.App;

  private pollInterval: ReturnType<typeof setInterval> | null = null;

  constructor() {
    this.loadActive();
    // Browser mode: poll until a tournament is activated by the desktop app
    if (!this.isWails) {
      this.pollInterval = setInterval(() => {
        if (!this.active()) this.loadActive();
      }, 3000);
    }
  }

  ngOnDestroy() {
    if (this.pollInterval !== null) {
      clearInterval(this.pollInterval);
    }
  }

  async loadActive(): Promise<void> {
    this.loading.set(true);
    try {
      const res = await fetch('/api/active-tournament');
      this.active.set(res.ok ? await res.json() : null);
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
