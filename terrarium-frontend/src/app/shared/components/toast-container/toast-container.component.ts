import { Component, inject } from '@angular/core';
import { ToastService } from '../../../core/services/toast.service';

@Component({
    selector: 'app-toast-container',
    standalone: true,
    template: `
    <div class="toast-container">
      @for (toast of toastService.toasts(); track toast.id) {
        <div class="toast" [class]="'toast-' + toast.type" (click)="toastService.remove(toast.id)">
          {{ toast.message }}
        </div>
      }
    </div>
  `,
})
export class ToastContainerComponent {
    readonly toastService = inject(ToastService);
}
