import { Injectable, inject } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import {
    SensorCurrent,
    SensorDataHistory,
    RelayState,
    RelayToggleRequest,
    ConfigPayload,
    ModeRequest,
    SystemStatus,
    Schedule,
    ScheduleRequest,
    EnergyReport,
    RelayLogEntry,
    RelayId,
} from '../models/api.models';

/**
 * Типизированный API-клиент для взаимодействия с Go-бэкендом.
 * Все HTTP-запросы проходят через этот сервис.
 */
@Injectable({ providedIn: 'root' })
export class ApiService {
    private readonly http = inject(HttpClient);
    private readonly baseUrl = environment.apiUrl;

    // ==========================================
    // ДАТЧИКИ
    // ==========================================

    /** Текущие показания датчиков (из кэша Engine) */
    getSensorCurrent(): Observable<SensorCurrent> {
        return this.http.get<SensorCurrent>(`${this.baseUrl}/sensors/current`);
    }

    /** Историческая выборка показаний с фильтрацией */
    getSensorHistory(from?: string, to?: string, limit?: number): Observable<SensorDataHistory[]> {
        let params = new HttpParams();
        if (from) params = params.set('from', from);
        if (to) params = params.set('to', to);
        if (limit) params = params.set('limit', limit.toString());
        return this.http.get<SensorDataHistory[]>(`${this.baseUrl}/metrics/sensors`, { params });
    }

    // ==========================================
    // РЕЛЕ
    // ==========================================

    /** Состояние всех 4 реле */
    getRelays(): Observable<RelayState> {
        return this.http.get<RelayState>(`${this.baseUrl}/relays`);
    }

    /** Переключить конкретное реле (только в MANUAL) */
    toggleRelay(id: RelayId, state: boolean): Observable<any> {
        const body: RelayToggleRequest = { state };
        return this.http.post(`${this.baseUrl}/relays/${id}/toggle`, body);
    }

    // ==========================================
    // КОНФИГУРАЦИЯ
    // ==========================================

    /** Получить текущую конфигурацию автоматизации */
    getConfig(): Observable<ConfigPayload> {
        return this.http.get<ConfigPayload>(`${this.baseUrl}/config`);
    }

    /** Обновить конфигурацию автоматизации */
    updateConfig(config: ConfigPayload): Observable<ConfigPayload> {
        return this.http.put<ConfigPayload>(`${this.baseUrl}/config`, config);
    }

    // ==========================================
    // СИСТЕМА
    // ==========================================

    /** Системный статус (uptime, mode, db_status) */
    getSystemStatus(): Observable<SystemStatus> {
        return this.http.get<SystemStatus>(`${this.baseUrl}/system/status`);
    }

    /** Переключить режим системы AUTO/MANUAL */
    setSystemMode(mode: 'AUTO' | 'MANUAL'): Observable<ModeRequest> {
        return this.http.post<ModeRequest>(`${this.baseUrl}/system/mode`, { mode });
    }

    // ==========================================
    // РАСПИСАНИЯ
    // ==========================================

    /** Список всех расписаний реле */
    getSchedules(): Observable<Schedule[]> {
        return this.http.get<Schedule[]>(`${this.baseUrl}/schedules`);
    }

    /** Создать расписание */
    createSchedule(req: ScheduleRequest): Observable<Schedule> {
        return this.http.post<Schedule>(`${this.baseUrl}/schedules`, req);
    }

    /** Обновить расписание */
    updateSchedule(id: string, req: ScheduleRequest): Observable<any> {
        return this.http.put(`${this.baseUrl}/schedules/${id}`, req);
    }

    /** Удалить расписание */
    deleteSchedule(id: string): Observable<any> {
        return this.http.delete(`${this.baseUrl}/schedules/${id}`);
    }

    // ==========================================
    // ЭНЕРГОПОТРЕБЛЕНИЕ
    // ==========================================

    /** Отчёты энергопотребления */
    getEnergyReports(from?: string, to?: string): Observable<EnergyReport[]> {
        let params = new HttpParams();
        if (from) params = params.set('from', from);
        if (to) params = params.set('to', to);
        return this.http.get<EnergyReport[]>(`${this.baseUrl}/metrics/energy`, { params });
    }

    // ==========================================
    // ЛОГИ РЕЛЕ
    // ==========================================

    /** Журнал переключений реле */
    getRelayLogs(limit?: number, offset?: number): Observable<RelayLogEntry[]> {
        let params = new HttpParams();
        if (limit) params = params.set('limit', limit.toString());
        if (offset) params = params.set('offset', offset.toString());
        return this.http.get<RelayLogEntry[]>(`${this.baseUrl}/relay-logs`, { params });
    }
}
