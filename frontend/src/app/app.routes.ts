import { Routes } from '@angular/router';

export const routes: Routes = [
  { path: '', redirectTo: 'operator', pathMatch: 'full' },
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
  { path: '**', redirectTo: 'operator' }
];
