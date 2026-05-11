import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: 'setup', pathMatch: 'full' },
  {
    path: 'setup',
    loadComponent: () =>
      import('./features/setup/setup.component').then(m => m.SetupComponent)
  },
  {
    path: 'tatami',
    loadComponent: () =>
      import('./features/tatami/tatami.component').then(m => m.TatamiComponent)
  },
  {
    path: 'operator',
    loadComponent: () =>
      import('./features/operator/operator.component').then(m => m.OperatorComponent)
  },
  {
    path: 'display',
    loadComponent: () =>
      import('./features/display/display.component').then(m => m.DisplayComponent)
  },
  { path: '**', redirectTo: 'setup' }
];
