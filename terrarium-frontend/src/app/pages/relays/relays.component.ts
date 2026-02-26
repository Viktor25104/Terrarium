import { Component, inject, signal } from '@angular/core';
import { PollingService } from '../../core/services/polling.service';
import { ApiService } from '../../core/services/api.service';
import { ToastService } from '../../core/services/toast.service';
import { RELAY_LABELS, RELAY_ICONS, RelayId } from '../../core/models/api.models';

@Component({
    selector: 'app-relays',
    standalone: true,
    template: `
    <div class="page-container">
      <h1 class="page-title">⚡ Управление реле</h1>

      @if (polling.systemStatus()?.mode !== 'MANUAL') {
        <div class="mode-warning">
          ⚠️ Ручное управление доступно только в режиме <strong>MANUAL</strong>.
          <button class="cyber-btn cyber-btn-outline" style="margin-left: 12px;" (click)="switchToManual()">
            Переключить в MANUAL
          </button>
        </div>
      }

      <div class="relay-controls">
        @for (relayId of relayIds; track relayId) {
          <div class="cyber-card relay-control-card">
            <div class="relay-info">
              <span class="relay-emoji">{{ getIcon(relayId) }}</span>
              <div>
                <div class="relay-title">{{ getLabel(relayId) }}</div>
                <div class="relay-id">{{ relayId }}</div>
              </div>
            </div>

            <div class="relay-controls-row">
              <span class="relay-indicator" [class.is-on]="isOn(relayId)" [class.is-off]="!isOn(relayId)"></span>
              <span class="relay-state-label" [style.color]="isOn(relayId) ? 'var(--color-neon-green)' : 'var(--color-text-muted)'">
                {{ isOn(relayId) ? 'ВКЛЮЧЕНО' : 'ВЫКЛЮЧЕНО' }}
              </span>

              <div class="relay-buttons">
                <button class="cyber-btn cyber-btn-primary"
                        [disabled]="isOn(relayId) || polling.systemStatus()?.mode !== 'MANUAL' || switching()"
                        (click)="toggle(relayId, true)">
                  ON
                </button>
                <button class="cyber-btn cyber-btn-danger"
                        [disabled]="!isOn(relayId) || polling.systemStatus()?.mode !== 'MANUAL' || switching()"
                        (click)="toggle(relayId, false)">
                  OFF
                </button>
              </div>
            </div>
          </div>
        }
      </div>
    </div>
  `,
    styles: [`
    .mode-warning {
      background: rgba(255, 153, 0, 0.1);
      border: 1px solid var(--color-neon-orange);
      border-radius: 8px;
      padding: 14px 20px;
      margin-bottom: 24px;
      color: var(--color-neon-orange);
      font-size: 14px;
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 8px;
    }
    .relay-controls {
      display: flex;
      flex-direction: column;
      gap: 16px;
    }
    .relay-control-card {
      padding: 20px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      flex-wrap: wrap;
      gap: 16px;
    }
    .relay-info {
      display: flex;
      align-items: center;
      gap: 14px;
    }
    .relay-emoji { font-size: 32px; }
    .relay-title {
      font-size: 18px;
      font-weight: 600;
    }
    .relay-id {
      font-size: 12px;
      color: var(--color-text-muted);
      font-family: monospace;
    }
    .relay-controls-row {
      display: flex;
      align-items: center;
      gap: 16px;
    }
    .relay-state-label {
      font-weight: 600;
      font-size: 14px;
      min-width: 100px;
    }
    .relay-buttons {
      display: flex;
      gap: 8px;
    }
  `]
})
export class RelaysComponent {
    readonly polling = inject(PollingService);
    private readonly api = inject(ApiService);
    private readonly toast = inject(ToastService);
    readonly switching = signal(false);

    readonly relayIds: RelayId[] = ['heat_mat', 'fogger', 'light', 'spare'];

    getLabel(id: RelayId): string { return RELAY_LABELS[id]; }
    getIcon(id: RelayId): string { return RELAY_ICONS[id]; }

    isOn(id: RelayId): boolean {
        const state = this.polling.relayState();
        return state ? state[id] : false;
    }

    switchToManual(): void {
        this.api.setSystemMode('MANUAL').subscribe({
            next: () => this.toast.success('Режим переключён на MANUAL'),
            error: () => this.toast.error('Ошибка переключения'),
        });
    }

    toggle(id: RelayId, state: boolean): void {
        this.switching.set(true);
        this.api.toggleRelay(id, state).subscribe({
            next: () => {
                this.toast.success(`${RELAY_LABELS[id]} ${state ? 'включено' : 'выключено'}`);
                this.switching.set(false);
            },
            error: (err) => {
                this.toast.error(err.error?.message || 'Ошибка переключения реле');
                this.switching.set(false);
            },
        });
    }
}
