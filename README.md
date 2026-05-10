# judo-app

Aplicación de gestión de torneos de judo con soporte multi-tatami.

## Stack

- **Backend**: Go + [Wails v2](https://wails.io) + SQLite (pure Go, sin CGO)
- **Frontend**: Angular 21 + SCSS
- **Arquitectura**: PC central corre el backend; operadores remotos usan browser en la LAN

## Desarrollo

```bash
# Frontend (hot reload)
cd frontend && npm install && npm start

# App completa con Wails
wails dev
```

## Build

```bash
wails build
```
