// Конфигурация для среды разработки (development)
// API-запросы проксируются через Angular dev server → Go backend (proxy.conf.json)
export const environment = {
    production: false,
    apiUrl: '/api/v1',
    pollingIntervalMs: 5000, // Интервал polling'а датчиков (5 секунд)
};
