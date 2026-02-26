import { Component, inject, OnInit } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { SidebarComponent } from './shared/components/sidebar/sidebar.component';
import { ToastContainerComponent } from './shared/components/toast-container/toast-container.component';
import { PollingService } from './core/services/polling.service';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, SidebarComponent, ToastContainerComponent],
  templateUrl: './app.html',
  styleUrl: './app.css',
})
export class App implements OnInit {
  private readonly polling = inject(PollingService);

  ngOnInit(): void {
    // Запуск глобального polling'а данных
    this.polling.start();
  }
}
