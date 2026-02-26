import { Component, inject, computed } from '@angular/core';
import { PollingService } from '../../core/services/polling.service';
import { ApiService } from '../../core/services/api.service';
import { ToastService } from '../../core/services/toast.service';
import { RELAY_LABELS, RELAY_ICONS, RelayId } from '../../core/models/api.models';
import { DecimalPipe, DatePipe } from '@angular/common';

@Component({
    selector: 'app-dashboard',
    standalone: true,
    imports: [DecimalPipe, DatePipe],
    template: `
    <div class="page-container">
      <h1 class="page-title">📊 Дашборд</h1>

      <!-- Быстрые действия -->
      <div class="quick-actions">
        <button class="cyber-btn cyber-btn-primary" (click)="setMode('AUTO')"
                [disabled]="polling.systemStatus()?.mode === 'AUTO'">
          🤖 Режим AUTO
        </button>
        <button class="cyber-btn cyber-btn-outline" (click)="setMode('MANUAL')"
                [disabled]="polling.systemStatus()?.mode === 'MANUAL'">
          🎮 Режим MANUAL
        </button>
        <button class="cyber-btn cyber-btn-danger" (click)="allOff()">
          ⛔ Все OFF
        </button>
      </div>

      <!-- Карточки показаний -->
      @if (polling.isLoading()) {
        <div class="sensor-grid">
          @for (i of [1,2,3,4]; track i) {
            <div class="cyber-card sensor-card">
              <div class="skeleton" style="height: 20px; width: 60%; margin-bottom: 12px;"></div>
              <div class="skeleton" style="height: 36px; width: 80%;"></div>
            </div>
          }
        </div>
      } @else {
        <div class="sensor-grid">
          <!-- Тёплая зона: температура -->
          <div class="cyber-card sensor-card">
            <div class="sensor-label">🌡️ Тёплая зона</div>
            <div class="sensor-value glow-green">
              {{ polling.sensorData()?.warm_temp | number:'1.1-1' }}°C
            </div>
            <div class="sensor-sub">
              Влажность: {{ polling.sensorData()?.warm_hum | number:'1.0-0' }}%
            </div>
          </div>

          <!-- Холодная зона: температура -->
          <div class="cyber-card sensor-card">
            <div class="sensor-label">❄️ Холодная зона</div>
            <div class="sensor-value glow-cyan">
              {{ polling.sensorData()?.cold_temp | number:'1.1-1' }}°C
            </div>
            <div class="sensor-sub">
              Влажность: {{ polling.sensorData()?.cold_hum | number:'1.0-0' }}%
            </div>
          </div>

          <!-- Тёплая зона: влажность -->
          <div class="cyber-card sensor-card">
            <div class="sensor-label">💧 Влажность (тёплая)</div>
            <div class="sensor-value" style="color: var(--color-neon-purple);">
              {{ polling.sensorData()?.warm_hum | number:'1.1-1' }}%
            </div>
          </div>

          <!-- Холодная зона: влажность -->
          <div class="cyber-card sensor-card">
            <div class="sensor-label">💧 Влажность (холодная)</div>
            <div class="sensor-value" style="color: var(--color-neon-pink);">
              {{ polling.sensorData()?.cold_hum | number:'1.1-1' }}%
            </div>
          </div>
        </div>
      }

      <!-- Состояние реле -->
      <h2 class="section-title">⚡ Состояние реле</h2>
      <div class="relay-grid">
        @for (relayId of relayIds; track relayId) {
          <div class="cyber-card relay-card">
            <div class="relay-header">
              <span class="relay-icon">{{ getRelayIcon(relayId) }}</span>
              <span class="relay-name">{{ getRelayLabel(relayId) }}</span>
            </div>
            <div class="relay-status">
              <span class="relay-indicator" [class.is-on]="isRelayOn(relayId)" [class.is-off]="!isRelayOn(relayId)"></span>
              <span class="relay-status-text" [style.color]="isRelayOn(relayId) ? 'var(--color-neon-green)' : 'var(--color-text-muted)'">
                {{ isRelayOn(relayId) ? 'ВКЛ' : 'ВЫКЛ' }}
              </span>
            </div>
          </div>
        }
      </div>

      <!-- Последнее обновление -->
      <div class="last-update">
        Последнее обновление: {{ polling.sensorData()?.timestamp | date:'HH:mm:ss' }}
        · Режим: {{ polling.systemStatus()?.mode || '...' }}
      </div>
    </div>
  `,
    styles: [`
    .quick-actions {
      display: flex;
      gap: 12px;
      margin-bottom: 24px;
      flex-wrap: wrap;
    }
    .sensor-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
      gap: 16px;
      margin-bottom: 32px;
    }
    .sensor-card {
      padding: 20px;
    }
    .sensor-label {
      font-size: 14px;
      color: var(--color-text-secondary);
      margin-bottom: 8px;
    }
    .sensor-value {
      font-size: 36px;
      font-weight: 700;
      line-height: 1.2;
    }
    .sensor-sub {
      font-size: 13px;
      color: var(--color-text-muted);
      margin-top: 4px;
    }
    .section-title {
      font-size: 20px;
      font-weight: 600;
      margin-bottom: 16px;
      color: var(--color-text-primary);
    }
    .relay-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 12px;
      margin-bottom: 24px;
    }
    .relay-card {
      padding: 16px;
    }
    .relay-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 12px;
    }
    .relay-icon { font-size: 20px; }
    .relay-name {
      font-size: 15px;
      font-weight: 600;
    }
    .relay-status {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .relay-status-text {
      font-size: 14px;
      font-weight: 600;
      letter-spacing: 0.5px;
    }
    .last-update {
      font-size: 12px;
      color: var(--color-text-muted);
      text-align: center;
      padding-top: 16px;
      border-top: 1px solid var(--color-border);
    }
  `]
})
export class DashboardComponent {
    readonly polling = inject(PollingService);
    private readonly api = inject(ApiService);
    private readonly toast = inject(ToastService);

    readonly relayIds: RelayId[] = ['heat_mat', 'fogger', 'light', 'spare'];

    getRelayLabel(id: RelayId): string {
        return RELAY_LABELS[id];
    }

    getRelayIcon(id: RelayId): string {
        return RELAY_ICONS[id];
    }

    isRelayOn(id: RelayId): boolean {
        const state = this.polling.relayState();
        if (!state) return false;
        return state[id];
    }

    setMode(mode: 'AUTO' | 'MANUAL'): void {
        this.api.setSystemMode(mode).subscribe({
            next: () => this.toast.success(`Режим переключён на ${mode}`),
            error: () => this.toast.error('Ошибка переключения режима'),
        });
    }

    allOff(): void {
        // Сначала переключаем в MANUAL, потом выключаем все реле
        this.api.setSystemMode('MANUAL').subscribe({
            next: () => {
                for (const id of this.relayIds) {
                    this.api.toggleRelay(id, false).subscribe();
                }
                this.toast.info('Все реле выключены (MANUAL mode)');
            },
            error: () => this.toast.error('Ошибка выключения'),
        });
    }
}
