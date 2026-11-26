.PHONY: help up down restart logs psql tidy run server worker seed test test-integration

# Docker'daki veritabanı bilgileri (.env ile aynı olmalıdır)
DB_USER ?= postgres
DB_NAME ?= car_auction_db

# Varsayılan hedef
help:
	@echo "Kullanılabilir komutlar:"
	@echo "  make up               - Tüm servisleri (PostgreSQL, web, worker) docker ile arka planda başlatır"
	@echo "  make down             - Docker servislerini durdurur"
	@echo "  make restart          - Docker servislerini yeniden başlatır"
	@echo "  make logs             - Docker servislerinin loglarını gösterir"
	@echo "  make psql             - PostgreSQL terminaline bağlanır"
	@echo "  make run              - Web sunucusunu ve açık artırma worker'ını birlikte yerelde başlatır"
	@echo "  make server           - Yalnızca web sunucusunu başlatır (migration'lar otomatik çalışır)"
	@echo "  make worker           - Yalnızca açık artırma worker'ını başlatır"
	@echo "  make seed             - CSV dosyasındaki araç verilerini veritabanına aktarır"
	@echo "  make tidy             - Go mod bağımlılıklarını günceller"
	@echo "  make test             - Birim testlerini çalıştırır"
	@echo "  make test-integration - Entegrasyon testini çalıştırır (TEST_DATABASE_URL gerekir)"

# Docker komutları
up:
	docker compose up -d

down:
	docker compose down

restart: down up

logs:
	docker compose logs -f

psql:
	docker compose exec -it postgres psql -U $(DB_USER) -d $(DB_NAME)

# Go uygulama komutları
tidy:
	go mod tidy

# İki mikroservisi birlikte çalıştırır; Ctrl+C ikisini de durdurur.
# Önce web sunucusu başlar (migration'ları uygular), ardından worker.
run:
	@trap 'kill 0' INT TERM EXIT; \
	go run ./cmd/server & \
	sleep 3; \
	go run ./cmd/auction-worker & \
	wait

server:
	go run ./cmd/server

worker:
	go run ./cmd/auction-worker

seed:
	go run ./cmd/seed -limit 100

test:
	go test ./...

test-integration:
	go test -tags integration -count=1 ./internal/repository/
