.PHONY: help up down restart logs run test tidy psql clean

# Varsayılan hedef
help:
	@echo "Kullanılabilir komutlar:"
	@echo "  make up       - PostgreSQL docker konteynerini arka planda başlatır"
	@echo "  make down     - PostgreSQL docker konteynerini durdurur"
	@echo "  make restart  - PostgreSQL konteynerini yeniden başlatır"
	@echo "  make logs     - PostgreSQL konteyner loglarını gösterir"
	@echo "  make run      - Go sunucusunu başlatır (migration'lar otomatik çalışır)"
	@echo "  make psql     - PostgreSQL terminaline bağlanır"
	@echo "  make seed     - CSV dosyasındaki araç verilerini veritabanına aktarır"
	@echo "  make tidy     - Go mod bağımlılıklarını günceller"
	@echo "  make test     - Testleri çalıştırır"

# Docker komutları
up:
	docker compose up -d

down:
	docker compose down

restart: down up

logs:
	docker compose logs -f postgres

psql:
	docker compose exec -it postgres psql -U postgres -d car_auction_db

# Go uygulama komutları
tidy:
	go mod tidy

run:
	go run cmd/server/main.go

seed:
	go run cmd/seed/main.go -limit 100

test:
	go test -v ./...
