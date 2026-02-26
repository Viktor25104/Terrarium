import { Injectable, signal } from '@angular/core';

// Тип тоста
export interface Toast {
    id: number;
    message: string;
    type: 'success' | 'error' | 'info';
}

/**
 * Сервис тост-уведомлений (глобальные нотификации).
 * Использует Angular Signals для реактивного обновления UI.
 */
@Injectable({ providedIn: 'root' })
export class ToastService {
    private nextId = 0;
    readonly toasts = signal<Toast[]>([]);

    /** Показать уведомление об успехе */
    success(message: string): void {
        this.show(message, 'success');
    }

    /** Показать уведомление об ошибке */
    error(message: string): void {
        this.show(message, 'error');
    }

    /** Показать информационное уведомление */
    info(message: string): void {
        this.show(message, 'info');
    }

    /** Удалить конкретный тост */
    remove(id: number): void {
        this.toasts.update(list => list.filter(t => t.id !== id));
    }

    private show(message: string, type: Toast['type']): void {
        const id = this.nextId++;
        this.toasts.update(list => [...list, { id, message, type }]);

        // Автоудаление через 4 секунды
        setTimeout(() => this.remove(id), 4000);
    }
}
