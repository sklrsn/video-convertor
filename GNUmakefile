.DEFAULT_GOAL: all

all: down up

.PHONY: setup
setup:
	mkdir -p storage/
	mkdir -p downloads/
	mkdir -p recordings/

.PHONY: up
up: setup
	docker compose -f docker-compose.yaml up --build

.PHONY: down
down:
	docker compose -f docker-compose.yaml down

.PHONY: clean
clean:
	rm -rf storage/
	rm -rf downloads/
	rm -rf recordings/