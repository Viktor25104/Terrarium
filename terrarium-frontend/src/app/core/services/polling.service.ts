import { Injectable, inject, signal, OnDestroy } from '@angular/core';
import { Subject, timer, switchMap, retry, takeUntil, tap, catchError, EMPTY } from 'rxjs';
import { ApiService } from './api.service';
import { SensorCurrent, RelayState, SystemStatus } from '../models/api.models';
import { environment } from '../../../environments/environment';

/**
 * Сервис polling'а — периодически запрашивает данные с бэкенда.
 * Использует RxJS timer + switchMap для автоматической отмены предыдущих запросов.
 * Интервал настраивается через environment.pollingIntervalMs.
 */
@Injectable({ providedIn: 'root' })
export class PollingService implements OnDestroy {
    private readonly api = inject(ApiService);
    private readonly destroy$ = new Subject<void>();

    // Реактивные сигналы состояния
    readonly sensorData = signal<SensorCurrent | null>(null);
    readonly relayState = signal<RelayState | null>(null);
    readonly systemStatus = signal<SystemStatus | null>(null);
    readonly isLoading = signal(true);
    readonly lastError = signal<string | null>(null);

    private isStarted = false;

    /** Запуск polling'а (вызывать из главного компонента) */
    start(): void {
        if (this.isStarted) return;
        this.isStarted = true;

        const interval = environment.pollingIntervalMs;

        // Polling текущих датчиков
        timer(0, interval).pipe(
            switchMap(() => this.api.getSensorCurrent().pipe(
                catchError(err => {
                    this.lastError.set('Ошибка получения данных датчиков');
                    return EMPTY;
                })
            )),
            tap(data => {
                this.sensorData.set(data);
                this.isLoading.set(false);
                this.lastError.set(null);
            }),
            takeUntil(this.destroy$)
        ).subscribe();

        // Polling состояния реле
        timer(0, interval).pipe(
            switchMap(() => this.api.getRelays().pipe(
                catchError(() => EMPTY)
            )),
            tap(data => this.relayState.set(data)),
            takeUntil(this.destroy$)
        ).subscribe();

        // Polling системного статуса (реже — раз в 15 секунд)
        timer(0, 15000).pipe(
            switchMap(() => this.api.getSystemStatus().pipe(
                catchError(() => EMPTY)
            )),
            tap(data => this.systemStatus.set(data)),
            takeUntil(this.destroy$)
        ).subscribe();
    }

    /** Остановка polling'а */
    stop(): void {
        this.destroy$.next();
        this.isStarted = false;
    }

    ngOnDestroy(): void {
        this.stop();
        this.destroy$.complete();
    }
}
