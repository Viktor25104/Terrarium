import { Component, inject, OnInit, signal, ViewChild, ElementRef, AfterViewInit, OnDestroy } from '@angular/core';
import { ApiService } from '../../core/services/api.service';
import { SensorDataHistory, EnergyReport } from '../../core/models/api.models';
import * as echarts from 'echarts';

@Component({
    selector: 'app-history',
    standalone: true,
    template: `
    <div class="page-container">
      <h1 class="page-title">📈 Графики и отчёты</h1>

      <!-- Селектор диапазона -->
      <div class="range-selector">
        <button class="cyber-btn" [class.cyber-btn-primary]="range() === '1h'" (click)="setRange('1h')">1ч</button>
        <button class="cyber-btn" [class.cyber-btn-primary]="range() === '6h'" (click)="setRange('6h')">6ч</button>
        <button class="cyber-btn" [class.cyber-btn-primary]="range() === '24h'" (click)="setRange('24h')">24ч</button>
        <button class="cyber-btn" [class.cyber-btn-primary]="range() === '7d'" (click)="setRange('7d')">7д</button>
      </div>

      <!-- График температуры -->
      <div class="cyber-card chart-card">
        <h2 class="chart-title">🌡️ Температура (°C)</h2>
        @if (loading()) {
          <div class="skeleton" style="height: 300px;"></div>
        } @else {
          <div #tempChart class="chart-container"></div>
        }
      </div>

      <!-- График влажности -->
      <div class="cyber-card chart-card">
        <h2 class="chart-title">💧 Влажность (%)</h2>
        @if (loading()) {
          <div class="skeleton" style="height: 300px;"></div>
        } @else {
          <div #humChart class="chart-container"></div>
        }
      </div>

      <!-- Энергопотребление -->
      <div class="cyber-card chart-card">
        <h2 class="chart-title">⚡ Энергопотребление (кВт⋅ч)</h2>
        @if (energyLoading()) {
          <div class="skeleton" style="height: 300px;"></div>
        } @else if (energyData().length === 0) {
          <p class="empty-chart-text">Данные об энергопотреблении пока отсутствуют.</p>
        } @else {
          <div #energyChart class="chart-container"></div>
        }
      </div>
    </div>
  `,
    styles: [`
    .range-selector {
      display: flex;
      gap: 8px;
      margin-bottom: 20px;
    }
    .range-selector .cyber-btn:not(.cyber-btn-primary) {
      background: var(--color-bg-card);
      color: var(--color-text-secondary);
      border: 1px solid var(--color-border);
    }
    .chart-card {
      padding: 20px;
      margin-bottom: 20px;
    }
    .chart-title {
      font-size: 16px;
      font-weight: 600;
      margin-bottom: 16px;
    }
    .chart-container {
      width: 100%;
      height: 320px;
    }
    .empty-chart-text {
      color: var(--color-text-muted);
      text-align: center;
      padding: 60px 0;
    }
  `]
})
export class HistoryComponent implements OnInit, OnDestroy {
    private readonly api = inject(ApiService);

    @ViewChild('tempChart') tempChartRef!: ElementRef;
    @ViewChild('humChart') humChartRef!: ElementRef;
    @ViewChild('energyChart') energyChartRef!: ElementRef;

    readonly range = signal<'1h' | '6h' | '24h' | '7d'>('6h');
    readonly loading = signal(true);
    readonly energyLoading = signal(true);
    readonly sensorData = signal<SensorDataHistory[]>([]);
    readonly energyData = signal<EnergyReport[]>([]);

    private tempChartInstance: echarts.ECharts | null = null;
    private humChartInstance: echarts.ECharts | null = null;
    private energyChartInstance: echarts.ECharts | null = null;

    ngOnInit(): void {
        this.loadData();
    }

    ngOnDestroy(): void {
        this.tempChartInstance?.dispose();
        this.humChartInstance?.dispose();
        this.energyChartInstance?.dispose();
    }

    setRange(r: '1h' | '6h' | '24h' | '7d'): void {
        this.range.set(r);
        this.loadData();
    }

    private getFromDate(): string {
        const now = new Date();
        switch (this.range()) {
            case '1h': now.setHours(now.getHours() - 1); break;
            case '6h': now.setHours(now.getHours() - 6); break;
            case '24h': now.setDate(now.getDate() - 1); break;
            case '7d': now.setDate(now.getDate() - 7); break;
        }
        return now.toISOString();
    }

    private loadData(): void {
        this.loading.set(true);
        const from = this.getFromDate();
        const to = new Date().toISOString();

        this.api.getSensorHistory(from, to, 500).subscribe({
            next: (data) => {
                this.sensorData.set(data.reverse()); // Хронологический порядок
                this.loading.set(false);
                setTimeout(() => this.renderCharts(), 50);
            },
            error: () => {
                this.loading.set(false);
                this.sensorData.set([]);
            }
        });

        this.api.getEnergyReports().subscribe({
            next: (data) => {
                this.energyData.set(data);
                this.energyLoading.set(false);
                setTimeout(() => this.renderEnergyChart(), 50);
            },
            error: () => { this.energyLoading.set(false); }
        });
    }

    private renderCharts(): void {
        if (!this.tempChartRef?.nativeElement) return;

        const data = this.sensorData();
        const timestamps = data.map(d => new Date(d.timestamp).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' }));

        // Тёмная тема ECharts
        const baseOption = {
            backgroundColor: 'transparent',
            grid: { left: 50, right: 20, top: 20, bottom: 40 },
            tooltip: {
                trigger: 'axis' as const,
                backgroundColor: '#1a1f35',
                borderColor: '#2a3050',
                textStyle: { color: '#e2e8f0' }
            },
            xAxis: {
                type: 'category' as const,
                data: timestamps,
                axisLine: { lineStyle: { color: '#2a3050' } },
                axisLabel: { color: '#94a3b8', fontSize: 11 },
            },
            dataZoom: [{ type: 'inside' as const }],
        };

        // График температуры
        this.tempChartInstance?.dispose();
        this.tempChartInstance = echarts.init(this.tempChartRef.nativeElement);
        this.tempChartInstance.setOption({
            ...baseOption,
            yAxis: {
                type: 'value' as const,
                axisLabel: { color: '#94a3b8', formatter: '{value}°C' },
                splitLine: { lineStyle: { color: '#2a3050' } },
            },
            legend: {
                data: ['Тёплая зона', 'Холодная зона'],
                textStyle: { color: '#94a3b8' },
                top: 0,
            },
            series: [
                {
                    name: 'Тёплая зона',
                    type: 'line',
                    data: data.map(d => d.warm_temp),
                    smooth: true,
                    lineStyle: { color: '#39ff14', width: 2 },
                    itemStyle: { color: '#39ff14' },
                    areaStyle: { color: 'rgba(57, 255, 20, 0.05)' },
                },
                {
                    name: 'Холодная зона',
                    type: 'line',
                    data: data.map(d => d.cold_temp),
                    smooth: true,
                    lineStyle: { color: '#00f5ff', width: 2 },
                    itemStyle: { color: '#00f5ff' },
                    areaStyle: { color: 'rgba(0, 245, 255, 0.05)' },
                },
            ],
        });

        // График влажности
        this.humChartInstance?.dispose();
        this.humChartInstance = echarts.init(this.humChartRef.nativeElement);
        this.humChartInstance.setOption({
            ...baseOption,
            yAxis: {
                type: 'value' as const,
                axisLabel: { color: '#94a3b8', formatter: '{value}%' },
                splitLine: { lineStyle: { color: '#2a3050' } },
            },
            legend: {
                data: ['Тёплая зона', 'Холодная зона'],
                textStyle: { color: '#94a3b8' },
                top: 0,
            },
            series: [
                {
                    name: 'Тёплая зона',
                    type: 'line',
                    data: data.map(d => d.warm_hum),
                    smooth: true,
                    lineStyle: { color: '#bf00ff', width: 2 },
                    itemStyle: { color: '#bf00ff' },
                    areaStyle: { color: 'rgba(191, 0, 255, 0.05)' },
                },
                {
                    name: 'Холодная зона',
                    type: 'line',
                    data: data.map(d => d.cold_hum),
                    smooth: true,
                    lineStyle: { color: '#ff006e', width: 2 },
                    itemStyle: { color: '#ff006e' },
                    areaStyle: { color: 'rgba(255, 0, 110, 0.05)' },
                },
            ],
        });
    }

    private renderEnergyChart(): void {
        if (!this.energyChartRef?.nativeElement || this.energyData().length === 0) return;

        const data = this.energyData();
        this.energyChartInstance?.dispose();
        this.energyChartInstance = echarts.init(this.energyChartRef.nativeElement);
        this.energyChartInstance.setOption({
            backgroundColor: 'transparent',
            grid: { left: 50, right: 20, top: 40, bottom: 40 },
            tooltip: {
                trigger: 'axis',
                backgroundColor: '#1a1f35',
                borderColor: '#2a3050',
                textStyle: { color: '#e2e8f0' }
            },
            legend: {
                data: ['Нагрев', 'Свет', 'Туман', 'Итого'],
                textStyle: { color: '#94a3b8' },
            },
            xAxis: {
                type: 'category',
                data: data.map(d => d.report_date),
                axisLine: { lineStyle: { color: '#2a3050' } },
                axisLabel: { color: '#94a3b8' },
            },
            yAxis: {
                type: 'value',
                axisLabel: { color: '#94a3b8', formatter: '{value} кВт⋅ч' },
                splitLine: { lineStyle: { color: '#2a3050' } },
            },
            series: [
                { name: 'Нагрев', type: 'bar', stack: 'energy', data: data.map(d => d.heat_mat_kwh), itemStyle: { color: '#ff6600' } },
                { name: 'Свет', type: 'bar', stack: 'energy', data: data.map(d => d.light_kwh), itemStyle: { color: '#e6ff00' } },
                { name: 'Туман', type: 'bar', stack: 'energy', data: data.map(d => d.fogger_kwh), itemStyle: { color: '#bf00ff' } },
                { name: 'Итого', type: 'line', data: data.map(d => d.total_kwh), lineStyle: { color: '#39ff14', width: 2 }, itemStyle: { color: '#39ff14' } },
            ],
        });
    }
}
