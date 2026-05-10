import { Component, Input, ChangeDetectionStrategy } from '@angular/core';
import { ScoreDTO } from '../../models/combat';
import { NgClass } from '@angular/common';

@Component({
  selector: 'app-score-board',
  standalone: true,
  imports: [NgClass],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="score-board" [ngClass]="{'score-board--compact': compact}">
      <div class="athlete-label">{{ label }}</div>
      <div class="scores">
        <span class="score score--ippon" [class.score--active]="score.ippon > 0">
          {{ score.ippon }}
        </span>
        <span class="score score--waza" [class.score--active]="score.wazaAri > 0">
          {{ score.wazaAri }}
        </span>
        <span class="score score--yuko" [class.score--active]="score.yuko > 0">
          {{ score.yuko }}
        </span>
      </div>
      <div class="penalties">
        @for (i of shidoArray; track i) {
          <span class="shido" [class.shido--active]="i < score.shido">●</span>
        }
        @if (score.hansoku) {
          <span class="hansoku">X</span>
        }
      </div>
    </div>
  `,
  styles: [`
    .score-board {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 0.5rem;
      padding: 1rem;
      background: #1a1a2e;
      border-radius: 8px;
      min-width: 200px;
    }
    .athlete-label {
      font-size: 1.2rem;
      font-weight: bold;
      color: #eee;
      text-align: center;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }
    .scores {
      display: flex;
      gap: 0.75rem;
    }
    .score {
      font-size: 2.5rem;
      font-weight: 900;
      width: 2.5rem;
      text-align: center;
      color: #555;
      transition: color 0.2s;
    }
    .score--ippon { color: #c0392b; }
    .score--waza.score--active { color: #e67e22; }
    .score--yuko.score--active { color: #f1c40f; }
    .score--active { color: inherit; }
    .penalties {
      display: flex;
      gap: 0.25rem;
    }
    .shido {
      color: #555;
      font-size: 1.2rem;
    }
    .shido--active { color: #e74c3c; }
    .hansoku {
      color: #e74c3c;
      font-weight: 900;
      font-size: 1.4rem;
    }
    .score-board--compact .score { font-size: 1.5rem; }
  `]
})
export class ScoreBoardComponent {
  @Input({ required: true }) score!: ScoreDTO;
  @Input({ required: true }) label!: string;
  @Input() compact = false;

  readonly shidoArray = [0, 1, 2];
}
