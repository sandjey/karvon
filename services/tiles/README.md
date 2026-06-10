# Своя карта (self-hosted OpenStreetMap tiles)

Сервис `tiles` в `docker-compose.yml` поднимает собственный сервер тайлов
(tileserver-gl) — чтобы карта на платформе была своя, без зависимости от
внешних публичных тайл-серверов OSM.

## Как запустить

1. Скачайте регион-экстракт в формате `.mbtiles` (вектор/растр). Источники:
   - https://data.maptiler.com/downloads/ (OpenMapTiles, по странам)
   - https://openmaptiles.org/ / https://download.geofabrik.de/ (raw OSM → собрать в mbtiles)
   Для нашего рынка достаточно: Узбекистан + Центральная Азия + СНГ.

2. Положите файл сюда как `tiles.mbtiles`:
   ```
   services/tiles/tiles.mbtiles
   ```

3. Запустите сервис:
   ```
   docker compose up -d tiles
   ```

4. Карта будет доступна на `http://<host>:8083`. Настройте reverse-proxy
   (например `maps.fan.sarbon.me`) и пропишите фронту `MAP_TILES_URL`.

## Фронтенд (Next.js + Leaflet)

Leaflet указывает на свой сервер тайлов:
```js
L.tileLayer('https://maps.fan.sarbon.me/styles/basic/{z}/{x}/{y}.png')
```

Бэкенд отдаёт адрес карты на `GET /api/v1/config` (поле `map_tiles_url`,
берётся из переменной окружения `MAP_TILES_URL`).
