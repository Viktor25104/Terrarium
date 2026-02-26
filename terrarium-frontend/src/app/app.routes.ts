import { Routes } from '@angular/router';

// Lazy-loaded маршруты для код-сплитинга
export const routes: Routes = [
    {
        path: '',
        redirectTo: 'dashboard',
        pathMatch: 'full',
    },
    {
        path: 'dashboard',
        loadComponent: () =>
            import('./pages/dashboard/dashboard.component').then(m => m.DashboardComponent),
    },
    {
        path: 'relays',
        loadComponent: () =>
            import('./pages/relays/relays.component').then(m => m.RelaysComponent),
    },
    {
        path: 'automation',
        loadComponent: () =>
            import('./pages/automation/automation.component').then(m => m.AutomationComponent),
    },
    {
        path: 'history',
        loadComponent: () =>
            import('./pages/history/history.component').then(m => m.HistoryComponent),
    },
    {
        path: 'system',
        loadComponent: () =>
            import('./pages/system/system.component').then(m => m.SystemComponent),
    },
    {
        path: '**',
        redirectTo: 'dashboard',
    },
];
