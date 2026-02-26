import { Component, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../core/services/api.service';
import { ToastService } from '../../core/services/toast.service';
import { ConfigPayload, Schedule, ScheduleRequest, RelayId, RELAY_LABELS } from '../../core/models/api.models';

@Component({
    selector: 'app-automation',
    standalone: true,
    imports: [FormsModule],
    template: `
    <div class="page-container">
      <h1 class="page-title">🤖 Настройки автоматизации</h1>

      <!-- Конфигурация порогов -->
      <div class="cyber-card config-section">
        <h2 class="section-header">🌡️ Пороги температуры и влажности</h2>

        @if (configLoading()) {
          <div class="skeleton" style="height: 200px;"></div>
        } @else if (config()) {
          <div class="config-grid">
            <div class="config-field">
              <label>Тёплая зона MIN (°C)</label>
              <input type="number" class="cyber-input" [(ngModel)]="config()!.warm_target_min" step="0.5" min="20" max="40">
            </div>
            <div class="config-field">
              <label>Тёплая зона MAX (°C)</label>
              <input type="number" class="cyber-input" [(ngModel)]="config()!.warm_target_max" step="0.5" min="20" max="40">
            </div>
            <div class="config-field">
              <label>Холодная зона макс (°C)</label>
              <input type="number" class="cyber-input" [(ngModel)]="config()!.cold_max_threshold" step="0.5" min="20" max="35">
            </div>
            <div class="config-field">
              <label>🚨 Аварийный порог (°C)</label>
              <input type="number" class="cyber-input" [(ngModel)]="config()!.emergency_max_threshold" step="0.5" min="30" max="45">
            </div>
            <div class="config-field">
              <label>Влажность MIN (%)</label>
              <input type="number" class="cyber-input" [(ngModel)]="config()!.humidity_min" step="1" min="0" max="100">
            </div>
            <div class="config-field">
              <label>Влажность MAX (%)</label>
              <input type="number" class="cyber-input" [(ngModel)]="config()!.humidity_max" step="1" min="0" max="100">
            </div>
            <div class="config-field">
              <label>Гистерезис темп. (°C)</label>
              <input type="number" class="cyber-input" [(ngModel)]="config()!.hysteresis_temp" step="0.1" min="0.1" max="5">
            </div>
            <div class="config-field">
              <label>Гистерезис влажн. (%)</label>
              <input type="number" class="cyber-input" [(ngModel)]="config()!.hysteresis_hum" step="0.5" min="0.5" max="10">
            </div>
          </div>
          <button class="cyber-btn cyber-btn-primary" style="margin-top: 16px;" (click)="saveConfig()" [disabled]="saving()">
            {{ saving() ? 'Сохранение...' : '💾 Сохранить' }}
          </button>
        }
      </div>

      <!-- Расписания -->
      <div class="cyber-card config-section" style="margin-top: 24px;">
        <h2 class="section-header">📅 Расписания реле</h2>

        @if (schedulesLoading()) {
          <div class="skeleton" style="height: 100px;"></div>
        } @else {
          @if (schedules().length === 0) {
            <p class="empty-text">Расписаний нет. Создайте первое!</p>
          }

          @for (s of schedules(); track s.id) {
            <div class="schedule-item">
              <span class="schedule-relay">{{ relayLabel(s.relay_id) }}</span>
              <span class="schedule-time">{{ s.start_time }} → {{ s.end_time }}</span>
              <span class="schedule-active" [style.color]="s.is_active ? 'var(--color-neon-green)' : 'var(--color-text-muted)'">
                {{ s.is_active ? 'Активно' : 'Неактивно' }}
              </span>
              <button class="cyber-btn cyber-btn-danger" style="padding: 6px 12px; font-size: 12px;" (click)="deleteSchedule(s.id)">
                ✕
              </button>
            </div>
          }

          <!-- Форма нового расписания -->
          <div class="new-schedule-form">
            <select class="cyber-select" [(ngModel)]="newSchedule.relay_id">
              <option value="heat_mat">🔥 Нагрев</option>
              <option value="fogger">💨 Туман</option>
              <option value="light">💡 Свет</option>
              <option value="spare">🔌 Запасной</option>
            </select>
            <input type="time" class="cyber-input" [(ngModel)]="newSchedule.start_time" style="width: auto;">
            <input type="time" class="cyber-input" [(ngModel)]="newSchedule.end_time" style="width: auto;">
            <button class="cyber-btn cyber-btn-primary" (click)="addSchedule()">
              ➕ Добавить
            </button>
          </div>
        }
      </div>
    </div>
  `,
    styles: [`
    .config-section {
      padding: 24px;
    }
    .section-header {
      font-size: 18px;
      font-weight: 600;
      margin-bottom: 20px;
      color: var(--color-text-primary);
    }
    .config-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
      gap: 16px;
    }
    .config-field label {
      display: block;
      font-size: 13px;
      color: var(--color-text-secondary);
      margin-bottom: 6px;
    }
    .schedule-item {
      display: flex;
      align-items: center;
      gap: 16px;
      padding: 12px 0;
      border-bottom: 1px solid var(--color-border);
      flex-wrap: wrap;
    }
    .schedule-relay {
      font-weight: 600;
      min-width: 80px;
    }
    .schedule-time {
      font-family: monospace;
      color: var(--color-neon-cyan);
    }
    .schedule-active {
      font-size: 13px;
      font-weight: 500;
    }
    .new-schedule-form {
      display: flex;
      gap: 12px;
      align-items: center;
      margin-top: 16px;
      flex-wrap: wrap;
    }
    .empty-text {
      color: var(--color-text-muted);
      font-size: 14px;
    }
  `]
})
export class AutomationComponent implements OnInit {
    private readonly api = inject(ApiService);
    private readonly toast = inject(ToastService);

    readonly config = signal<ConfigPayload | null>(null);
    readonly configLoading = signal(true);
    readonly saving = signal(false);
    readonly schedules = signal<Schedule[]>([]);
    readonly schedulesLoading = signal(true);

    newSchedule: ScheduleRequest = {
        relay_id: 'light',
        start_time: '08:00',
        end_time: '20:00',
    };

    ngOnInit(): void {
        this.loadConfig();
        this.loadSchedules();
    }

    relayLabel(id: string): string {
        return RELAY_LABELS[id as RelayId] || id;
    }

    loadConfig(): void {
        this.api.getConfig().subscribe({
            next: (cfg) => { this.config.set(cfg); this.configLoading.set(false); },
            error: () => { this.configLoading.set(false); this.toast.error('Не удалось загрузить конфигурацию'); }
        });
    }

    saveConfig(): void {
        const cfg = this.config();
        if (!cfg) return;
        this.saving.set(true);
        this.api.updateConfig(cfg).subscribe({
            next: () => { this.saving.set(false); this.toast.success('Конфигурация сохранена!'); },
            error: (err) => { this.saving.set(false); this.toast.error(err.error?.message || 'Ошибка сохранения'); }
        });
    }

    loadSchedules(): void {
        this.api.getSchedules().subscribe({
            next: (list) => { this.schedules.set(list); this.schedulesLoading.set(false); },
            error: () => { this.schedulesLoading.set(false); }
        });
    }

    addSchedule(): void {
        this.api.createSchedule(this.newSchedule).subscribe({
            next: (created) => {
                this.schedules.update(list => [created, ...list]);
                this.toast.success('Расписание создано');
            },
            error: () => this.toast.error('Ошибка создания расписания'),
        });
    }

    deleteSchedule(id: string): void {
        this.api.deleteSchedule(id).subscribe({
            next: () => {
                this.schedules.update(list => list.filter(s => s.id !== id));
                this.toast.success('Расписание удалено');
            },
            error: () => this.toast.error('Ошибка удаления'),
        });
    }
}
