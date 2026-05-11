import { Component, Input, ChangeDetectionStrategy } from '@angular/core';
import { CombatState } from '../../models/combat';
import { NgClass } from '@angular/common';

@Component({
  selector: 'app-timer-display',
  standalone: true,
  imports: [NgClass],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="timer" [ngClass]="{
      'timer--golden': goldenScore,
      'timer--paused': state === 'PAUSED',
      'timer--finished': state === 'FINISHED'
    }">
      <div class="timer__time">{{ formatted }}</div>
      @if (goldenScore) {
        <div class="timer__label">GOLDEN SCORE</div>
      }
      @if (osaekomiMs > 0) {
        <div class="timer__osaekomi">OSA {{ osaeFormatted }}</div>
      }
    </div>
  `,
  styles: [`
    .timer {
      text-align: center;
      padding: 0.5rem 1rem;
    }
    .timer__time {
      font-size: 4rem;
      font-weight: 900;
      font-variant-numeric: tabular-nums;
      color: #ecf0f1;
      letter-spacing: 0.05em;
      transition: color 0.2s;
    }
    .timer--golden .timer__time { color: #f39c12; }
    .timer--paused .timer__time { color: #7f8c8d; }
    .timer--finished .timer__time { color: #e74c3c; }
    .timer__label {
      font-size: 0.9rem;
      color: #f39c12;
      font-weight: bold;
      letter-spacing: 0.1em;
    }
    .timer__osaekomi {
      font-size: 1.1rem;
      color: #3498db;
      font-weight: bold;
      margin-top: 0.25rem;
    }
  `]
})
export class TimerDisplayComponent {
  @Input({ required: true }) remainingMs!: number;
  @Input({ required: true }) state!: CombatState;
  @Input() goldenScore = false;
  @Input() osaekomiMs = 0;

  get formatted(): string {
    return formatMs(this.remainingMs);
  }

  get osaeFormatted(): string {
    return formatMs(this.osaekomiMs);
  }
}

function formatMs(ms: number): string {
  const totalSec = Math.floor(Math.abs(ms) / 1000);
  const min = Math.floor(totalSec / 60);
  const sec = totalSec % 60;
  return `${min}:${sec.toString().padStart(2, '0')}`;
}
