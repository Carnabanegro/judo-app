import {
  Component, OnInit, OnDestroy, ChangeDetectionStrategy, ChangeDetectorRef, inject
} from '@angular/core';
import { filter, Subscription } from 'rxjs';
import { CombatWsService } from '../../services/combat-ws.service';
import {
  CombatUpdatePayload, TimerTickPayload, CombatFinishPayload,
  ScoreDTO, CombatState
} from '../../models/combat';
import { ScoreBoardComponent } from '../../shared/components/score-board.component';
import { TimerDisplayComponent } from '../../shared/components/timer-display.component';

const EMPTY_SCORE: ScoreDTO = { ippon: 0, wazaAri: 0, yuko: 0, shido: 0, hansoku: false };

@Component({
  selector: 'app-display',
  standalone: true,
  imports: [ScoreBoardComponent, TimerDisplayComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './display.component.html',
  styleUrls: ['./display.component.scss']
})
export class DisplayComponent implements OnInit, OnDestroy {
  private readonly ws = inject(CombatWsService);
  private readonly cdr = inject(ChangeDetectorRef);
  private subs = new Subscription();

  labelA = 'Atleta A';
  labelB = 'Atleta B';
  scoreA: ScoreDTO = { ...EMPTY_SCORE };
  scoreB: ScoreDTO = { ...EMPTY_SCORE };
  state: CombatState = 'PENDING';
  remainingMs = 4 * 60 * 1000;
  goldenScore = false;
  osaekomiMs = 0;
  finishMessage = '';

  ngOnInit(): void {
    this.ws.connect();

    this.subs.add(
      this.ws.events$.pipe(filter(e => e.type === 'combat:update')).subscribe(e => {
        const p = e.payload as CombatUpdatePayload;
        this.scoreA = p.scoreA;
        this.scoreB = p.scoreB;
        this.state = p.state;
        this.labelA = p.labelA;
        this.labelB = p.labelB;
        this.finishMessage = '';
        this.cdr.markForCheck();
      })
    );

    this.subs.add(
      this.ws.events$.pipe(filter(e => e.type === 'combat:tick')).subscribe(e => {
        const p = e.payload as TimerTickPayload;
        this.remainingMs = p.remainingMs;
        this.goldenScore = p.goldenScore;
        this.osaekomiMs = p.osaekoiMs;
        this.state = p.state;
        this.cdr.markForCheck();
      })
    );

    this.subs.add(
      this.ws.events$.pipe(filter(e => e.type === 'combat:finished')).subscribe(e => {
        const p = e.payload as CombatFinishPayload;
        const winner = p.winnerIdx === 0 ? this.labelA : this.labelB;
        this.finishMessage = `${winner} — ${p.method}`;
        this.state = 'FINISHED';
        this.cdr.markForCheck();
      })
    );
  }

  ngOnDestroy(): void {
    this.subs.unsubscribe();
  }
}
