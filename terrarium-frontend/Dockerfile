# Стадия сборки (Build Stage)
FROM node:20-alpine AS builder

WORKDIR /app

# Кэширование node modules
COPY package.json package-lock.json* ./
RUN npm ci || npm install --legacy-peer-deps

COPY . .

# Сборка приложения Angular для production
RUN npm run build -- --configuration production

# Финальная стадия Production (Nginx)
FROM nginx:alpine

# Копирование пользовательской конфигурации Nginx для роутинга Angular
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Копирование собранных файлов из стадии builder
# Angular 19 выводит сборку в dist/<project-name>/browser
COPY --from=builder /app/dist/terrarium-app/browser /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
