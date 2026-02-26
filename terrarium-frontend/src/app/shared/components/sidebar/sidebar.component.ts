import { Component, inject } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { PollingService } from '../../../core/services/polling.service';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [RouterLink, RouterLinkActive],
  template: `
    <nav class="sidebar">
      <div class="sidebar-logo">
        <svg class="logo-snake" viewBox="0 0 40 40" width="36" height="36" fill="none" xmlns="http://www.w3.org/2000/svg">
          <defs>
            <filter id="snakeGlow">
              <feGaussianBlur stdDeviation="1.5" result="blur"/>
              <feMerge><feMergeNode in="blur"/><feMergeNode in="SourceGraphic"/></feMerge>
            </filter>
          </defs>
          <!-- Тело змеи — S-образная кривая -->
          <path class="snake-body" d="M8 6 C14 6, 14 14, 20 14 S26 22, 20 22 S14 30, 20 30 C24 30, 28 28, 32 34"
                stroke="url(#snakeGrad)" stroke-width="3.5" stroke-linecap="round" filter="url(#snakeGlow)"/>
          <!-- Голова -->
          <circle cx="8" cy="6" r="3.5" fill="#39ff14" filter="url(#snakeGlow)"/>
          <!-- Глаз  -->
          <circle cx="7.2" cy="5.2" r="1" fill="#0a0e17"/>
          <!-- Язык -->
          <path class="snake-tongue" d="M5 5.5 L2 3.5 M5 5.5 L2 7" stroke="#ff006e" stroke-width="0.8" stroke-linecap="round"/>
          <!-- Градиент от зелёного к циану -->
          <linearGradient id="snakeGrad" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#39ff14"/>
            <stop offset="100%" stop-color="#00f5ff"/>
          </linearGradient>
        </svg>
        <span class="logo-text">Terrarium</span>
      </div>

      <!-- Индикатор режима -->
      <div class="mode-badge" [class.mode-auto]="polling.systemStatus()?.mode === 'AUTO'"
           [class.mode-manual]="polling.systemStatus()?.mode === 'MANUAL'">
        {{ polling.systemStatus()?.mode || '...' }}
      </div>

      <ul class="nav-list">
        <li>
          <a routerLink="/dashboard" routerLinkActive="active">
            <span class="nav-icon">📊</span>
            <span class="nav-label">Дашборд</span>
          </a>
        </li>
        <li>
          <a routerLink="/relays" routerLinkActive="active">
            <span class="nav-icon">⚡</span>
            <span class="nav-label">Реле</span>
          </a>
        </li>
        <li>
          <a routerLink="/automation" routerLinkActive="active">
            <span class="nav-icon">🤖</span>
            <span class="nav-label">Автоматика</span>
          </a>
        </li>
        <li>
          <a routerLink="/history" routerLinkActive="active">
            <span class="nav-icon">📈</span>
            <span class="nav-label">Графики</span>
          </a>
        </li>
        <li>
          <a routerLink="/system" routerLinkActive="active">
            <span class="nav-icon">⚙️</span>
            <span class="nav-label">Система</span>
          </a>
        </li>
      </ul>

      <div class="sidebar-footer">
        <div class="connection-dot" [class.connected]="!polling.lastError()"></div>
        <span class="connection-label">{{ polling.lastError() ? 'Offline' : 'Online' }}</span>
      </div>
    </nav>
  `,
  styles: [`
    .sidebar {
      position: fixed;
      left: 0;
      top: 0;
      bottom: 0;
      width: 240px;
      background: linear-gradient(180deg, #0d1117 0%, #111827 100%);
      border-right: 1px solid var(--color-border);
      display: flex;
      flex-direction: column;
      z-index: 100;
      padding: 20px 0;
    }
    .sidebar-logo {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 0 20px 20px;
      border-bottom: 1px solid var(--color-border);
    }
    .logo-snake {
      flex-shrink: 0;
    }
    .snake-body {
      stroke-dasharray: 8 4;
      animation: snake-slither 1.5s linear infinite;
    }
    @keyframes snake-slither {
      to { stroke-dashoffset: -24; }
    }
    .snake-tongue {
      animation: tongue-flick 2s ease-in-out infinite;
      transform-origin: 5px 5.5px;
    }
    @keyframes tongue-flick {
      0%, 70%, 100% { opacity: 0; }
      75%, 85% { opacity: 1; }
    }
    .logo-text {
      font-size: 20px;
      font-weight: 700;
      background: linear-gradient(135deg, var(--color-neon-green), var(--color-neon-cyan));
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
    }
    .mode-badge {
      margin: 16px 20px 8px;
      padding: 6px 12px;
      border-radius: 6px;
      font-size: 12px;
      font-weight: 700;
      text-align: center;
      letter-spacing: 1px;
    }
    .mode-auto {
      background: rgba(57, 255, 20, 0.1);
      border: 1px solid var(--color-neon-green);
      color: var(--color-neon-green);
    }
    .mode-manual {
      background: rgba(255, 153, 0, 0.1);
      border: 1px solid var(--color-neon-orange);
      color: var(--color-neon-orange);
    }
    .nav-list {
      list-style: none;
      padding: 12px 0;
      margin: 0;
      flex: 1;
    }
    .nav-list li a {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 12px 20px;
      color: var(--color-text-secondary);
      text-decoration: none;
      font-size: 15px;
      font-weight: 500;
      transition: all 0.2s ease;
      border-left: 3px solid transparent;
    }
    .nav-list li a:hover {
      color: var(--color-text-primary);
      background: rgba(255, 255, 255, 0.03);
    }
    .nav-list li a.active {
      color: var(--color-neon-cyan);
      background: rgba(0, 245, 255, 0.05);
      border-left-color: var(--color-neon-cyan);
    }
    .nav-icon { font-size: 18px; }
    .sidebar-footer {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 16px 20px;
      border-top: 1px solid var(--color-border);
    }
    .connection-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: var(--color-status-danger);
    }
    .connection-dot.connected {
      background: var(--color-neon-green);
      box-shadow: 0 0 6px var(--color-neon-green);
    }
    .connection-label {
      font-size: 12px;
      color: var(--color-text-muted);
    }

    @media (max-width: 768px) {
      .sidebar { width: 60px; }
      .logo-text, .nav-label, .connection-label, .mode-badge { display: none; }
      .sidebar-logo { justify-content: center; padding: 0 0 16px; }
      .nav-list li a { justify-content: center; padding: 14px 0; border-left: none; }
    }
  `]
})
export class SidebarComponent {
  readonly polling = inject(PollingService);
}
