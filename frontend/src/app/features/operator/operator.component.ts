import {
  Component, OnInit, OnDestroy, ChangeDetectionStrategy, ChangeDetectorRef, inject
} from '@angular/core';
import { filter, Subscription } from 'rxjs';
import { CombatWsService } from '../../services/combat-ws.service';
import { WailsService } from '../../services/wails.service';
import {
  CombatUpdatePayload, TimerTickPayload, CombatFinishPayload,
  ScoreDTO, CombatState
} from '../../models/combat';
import { ScoreBoardComponent } from '../../shared/components/score-board.component';
import { TimerDisplayComponent } from '../../shared/components/timer-display.component';
import { FormsModule } from '@angular/forms';

const EMPTY_SCORE: ScoreDTO = { ippon: 0, wazaAri: 0, yuko: 0, shido: 0, hansoku: false };

@Component({
  selector: 'app-operator',
  standalone: true,
  imports: [ScoreBoardComponent, TimerDisplayComponent, FormsModule],
  changeDetection: ChangeDetectionStrategy.OnPush,
  templateUrl: './operator.component.html',
  styleUrls: ['./operator.component.scss']
})
export class OperatorComponent implements OnInit, OnDestroy {
  private readonly ws = inject(CombatWsService);
  private readonly wails = inject(WailsService);
  private readonly cdr = inject(ChangeDetectorRef);
  private subs = new Subscription();

  // Match setup
  matchId = '';
  labelA = 'Atleta A';
  labelB = 'Atleta B';
  matchStarted = false;

  // Live state
  scoreA: ScoreDTO = { ...EMPTY_SCORE };
  scoreB: ScoreDTO = { ...EMPTY_SCORE };
  state: CombatState = 'PENDING';
  remainingMs = 4 * 60 * 1000;
  goldenScore = false;
  osaekomiMs = 0;
  osaekomiActive = false;
  finishMessage = '';

  // Practice mode
  practiceMode = false;

  ngOnInit(): void {
    this.ws.connect();

    this.subs.add(
      this.ws.events$.pipe(filter(e => e.type === 'combat:update')).subscribe(e => {
        const p = e.payload as CombatUpdatePayload;
        if (p.matchId !== this.matchId) return;
        this.scoreA = p.scoreA;
        this.scoreB = p.scoreB;
        this.state = p.state;
        this.labelA = p.labelA;
        this.labelB = p.labelB;
        this.cdr.markForCheck();
      })
    );

    this.subs.add(
      this.ws.events$.pipe(filter(e => e.type === 'combat:tick')).subscribe(e => {
        const p = e.payload as TimerTickPayload;
        if (p.matchId !== this.matchId) return;
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
        if (p.matchId !== this.matchId) return;
        const winner = p.winnerIdx === 0 ? this.labelA : this.labelB;
        this.finishMessage = `${winner} wins by ${p.method}`;
        this.matchStarted = false;
        this.cdr.markForCheck();
      })
    );
  }

  ngOnDestroy(): void {
    this.subs.unsubscribe();
  }

  async startMatch(): Promise<void> {
    if (this.practiceMode) {
      this.matchId = await this.wails.startPractice(this.labelA, this.labelB);
    } else {
      await this.wails.startMatch(this.matchId, this.labelA, this.labelB);
    }
    this.matchStarted = true;
    this.finishMessage = '';
    this.scoreA = { ...EMPTY_SCORE };
    this.scoreB = { ...EMPTY_SCORE };
    this.remainingMs = 4 * 60 * 1000;
    this.cdr.markForCheck();
  }

  async togglePause(): Promise<void> {
    if (this.state === 'PAUSED') {
      await this.wails.resume(this.matchId);
    } else {
      await this.wails.pause(this.matchId);
    }
  }

  async ippon(idx: number): Promise<void> { await this.wails.ippon(this.matchId, idx); }
  async wazaAri(idx: number): Promise<void> { await this.wails.wazaAri(this.matchId, idx); }
  async yuko(idx: number): Promise<void> { await this.wails.yuko(this.matchId, idx); }
  async shido(idx: number): Promise<void> { await this.wails.shido(this.matchId, idx); }

  async toggleOsaekomi(athleteIdx: number): Promise<void> {
    if (this.osaekomiActive) {
      await this.wails.stopOsaekomi(this.matchId, athleteIdx);
      this.osaekomiActive = false;
    } else {
      await this.wails.startOsaekomi(this.matchId);
      this.osaekomiActive = true;
    }
  }
}
