export interface ScoreDTO {
  ippon: number;
  wazaAri: number;
  yuko: number;
  shido: number;
  hansoku: boolean;
}

export interface CombatUpdatePayload {
  matchId: string;
  scoreA: ScoreDTO;
  scoreB: ScoreDTO;
  state: CombatState;
  labelA: string;
  labelB: string;
}

export interface TimerTickPayload {
  matchId: string;
  remainingMs: number;
  osaekoiMs: number;
  state: CombatState;
  goldenScore: boolean;
}

export interface CombatFinishPayload {
  matchId: string;
  winnerIdx: number;
  method: string;
}

export type CombatState =
  | 'PENDING'
  | 'ACTIVE'
  | 'PAUSED'
  | 'GOLDEN_SCORE'
  | 'FINISHED';

export interface BracketUpdatePayload {
  tournamentId: string;
}

export interface DisplayEvent<T = unknown> {
  type: 'combat:update' | 'combat:tick' | 'combat:finished' | 'bracket:update';
  payload: T;
}
