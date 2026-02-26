import { Component, inject, OnInit, signal } from '@angular/core';
import { ApiService } from '../../core/services/api.service';
import { PollingService } from '../../core/services/polling.service';
import { RelayLogEntry, RELAY_LABELS, RelayId } from '../../core/models/api.models';
import { DatePipe } from '@angular/common';

@Component({
    selector: 'app-system',
    standalone: true,
    imports: [DatePipe],
    template: `
    <div class="page-container">
      <h1 class="page-title">⚙️ Система</h1>

      <!-- Системная информация -->
      <div class="info-grid">
        <div class="cyber-card info-card">
          <div class="info-label">Режим</div>
          <div class="info-value" [style.color]="polling.systemStatus()?.mode === 'AUTO' ? 'var(--color-neon-green)' : 'var(--color-neon-orange)'">
            {{ polling.systemStatus()?.mode || '...' }}
          </div>
        </div>
        <div class="cyber-card info-card">
          <div class="info-label">Uptime</div>
          <div class="info-value glow-cyan">
            {{ formatUptime(polling.systemStatus()?.uptime || 0) }}
          </div>
        </div>
        <div class="cyber-card info-card">
          <div class="info-label">База данных</div>
          <div class="info-value" [style.color]="polling.systemStatus()?.db_status === 'OK' ? 'var(--color-neon-green)' : 'var(--color-status-danger)'">
            {{ polling.systemStatus()?.db_status || '...' }}
          </div>
        </div>
        <div class="cyber-card info-card">
          <div class="info-label">Telegram</div>
          <div class="info-value" style="color: var(--color-neon-cyan);">
            📱 Подключён
          </div>
        </div>
      </div>

      <!-- Журнал реле -->
      <div class="cyber-card log-section">
        <h2 class="section-header">📋 Журнал переключений реле</h2>
        @if (logsLoading()) {
          <div class="skeleton" style="height: 200px;"></div>
        } @else {
          <div class="log-table-wrapper">
            <table class="log-table">
              <thead>
                <tr>
                  <th>Время</th>
                  <th>Реле</th>
                  <th>Состояние</th>
                  <th>Причина</th>
                </tr>
              </thead>
              <tbody>
                @for (log of logs(); track log.id) {
                  <tr>
                    <td class="log-time">{{ log.recorded_at | date:'dd.MM HH:mm:ss' }}</td>
                    <td>{{ relayLabel(log.relay_id) }}</td>
                    <td>
                      <span class="relay-indicator" [class.is-on]="log.state" [class.is-off]="!log.state" style="vertical-align: middle;"></span>
                      {{ log.state ? 'ВКЛ' : 'ВЫКЛ' }}
                    </td>
                    <td class="log-reason">{{ log.reason }}</td>
                  </tr>
                }
              </tbody>
            </table>
          </div>

          @if (logs().length >= 50) {
            <button class="cyber-btn cyber-btn-outline" style="margin-top: 12px;" (click)="loadMore()">
              Загрузить ещё
            </button>
          }
        }
      </div>
    </div>
  `,
    styles: [`
    .info-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 16px;
      margin-bottom: 24px;
    }
    .info-card { padding: 20px; }
    .info-label {
      font-size: 13px;
      color: var(--color-text-secondary);
      margin-bottom: 8px;
    }
    .info-value {
      font-size: 24px;
      font-weight: 700;
    }
    .log-section { padding: 24px; }
    .section-header {
      font-size: 18px;
      font-weight: 600;
      margin-bottom: 16px;
    }
    .log-table-wrapper {
      overflow-x: auto;
    }
    .log-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 14px;
    }
    .log-table th {
      text-align: left;
      padding: 10px 12px;
      border-bottom: 1px solid var(--color-border);
      color: var(--color-text-secondary);
      font-weight: 600;
      font-size: 12px;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }
    .log-table td {
      padding: 10px 12px;
      border-bottom: 1px solid rgba(42, 48, 80, 0.5);
    }
    .log-time {
      font-family: monospace;
      color: var(--color-text-muted);
      font-size: 13px;
    }
    .log-reason {
      font-family: monospace;
      font-size: 12px;
      color: var(--color-neon-cyan);
    }
  `]
})
export class SystemComponent implements OnInit {
    private readonly api = inject(ApiService);
    readonly polling = inject(PollingService);

    readonly logs = signal<RelayLogEntry[]>([]);
    readonly logsLoading = signal(true);
    private offset = 0;

    ngOnInit(): void {
        this.loadLogs();
    }

    relayLabel(id: string): string {
        return RELAY_LABELS[id as RelayId] || id;
    }

    formatUptime(seconds: number): string {
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        const s = seconds % 60;
        if (h > 0) return `${h}ч ${m}м`;
        if (m > 0) return `${m}м ${s}с`;
        return `${s}с`;
    }

    loadLogs(): void {
        this.api.getRelayLogs(50, this.offset).subscribe({
            next: (data) => {
                this.logs.update(existing => [...existing, ...data]);
                this.logsLoading.set(false);
            },
            error: () => this.logsLoading.set(false),
        });
    }

    loadMore(): void {
        this.offset += 50;
        this.loadLogs();
    }
}
