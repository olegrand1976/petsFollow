SHELL := /bin/bash
COMPOSE_FILE := deploy/docker-compose.yml
ENV_FILE     := .env
COMPOSE      := docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE)

.DEFAULT_GOAL := help

.PHONY: help env brand-sync up up-infra down migrate seed api-dev nuxtjs-dev test test-go test-flutter test-nuxt test-auth smoke gcp-setup gcp-github gcp-setup-media gcp-setup-stripe gcp-deploy gcp-domain gcp-smoke firebase-flutter-setup firebase-android-dist

help:
	@echo "petsFollow — commandes"
	@echo "  make env            .env.example → .env"
	@echo "  make brand-sync     tokens CSS + Dart"
	@echo "  make up             stack Docker complète"
	@echo "  make up-infra       db + redis + mailhog"
	@echo "  make migrate        migrations API"
	@echo "  make seed           seed demo"
	@echo "  make api-dev        API Go (bloque le terminal, port 8291)"
	@echo "  make nuxtjs-dev     Web Pro Nuxt (autre terminal, port 3002)"
	@echo "  make test-go        tests Go"
	@echo "  make test-auth      garde-fou login/forgot (Go + Vitest)"
	@echo "  make smoke          smoke API MVP"
	@echo "  make gcp-deploy     Cloud Build staging"
	@echo "  make firebase-flutter-setup  apps Firebase Android/iOS"
	@echo "  make firebase-android-dist   APK → Firebase App Distribution (groupe petsfollow-testers)"
	@echo "  make gcp-setup-stripe        secrets Stripe GCP (placeholders + instructions)"
	@echo ""
	@echo "Dev local — 2 terminaux :"
	@echo "  T1: make up-infra && make migrate && make seed && make api-dev"
	@echo "  T2: make nuxtjs-dev  →  http://localhost:3002"

env:
	@test -f $(ENV_FILE) || cp .env.example $(ENV_FILE)

brand-sync:
	@bash scripts/brand-sync.sh

up: env brand-sync
	$(COMPOSE) up --build -d

up-infra: env
	$(COMPOSE) up -d db redis mailhog

down:
	$(COMPOSE) down

migrate: env
	@set -a && source $(ENV_FILE) && set +a && cd go && GOTOOLCHAIN=local go run ./cmd/petsfollow-api migrate

seed: env
	@set -a && source $(ENV_FILE) && set +a && cd go && GOTOOLCHAIN=local go run ./cmd/petsfollow-api seed

api-dev: env
	@set -a && source $(ENV_FILE) && set +a && cd go && GOTOOLCHAIN=local MIGRATE_ON_BOOT=true DEV_SEED_ENABLED=true go run ./cmd/petsfollow-api

nuxtjs-dev: env
	@set -a && source $(ENV_FILE) && set +a && cd nuxtjs && npm install && npx nuxt dev --port $${PETSFOLLOW_NUXTJS_PORT:-3002} --host 0.0.0.0

test-go:
	cd go && GOTOOLCHAIN=local go test ./...

test-flutter:
	cd flutter && flutter pub get && flutter test

test-nuxt:
	cd nuxtjs && npm install && npm test

# Garde-fou connexion / reset — à lancer avant deploy (DB seedée pour le volet Go).
test-auth:
	cd go && GOTOOLCHAIN=local go test ./internal/handlers/ -run 'TestAuth' -count=1
	cd nuxtjs && npm test -- tests/unit/useAuth.spec.ts

test: test-go

smoke:
	@bash scripts/smoke-test.sh

smoke-staging:
	PETSFOLLOW_API_URL=https://api.petsfollow.ll-it-sc.be bash scripts/smoke-test.sh

gcp-github:
	bash infra/gcp/setup-github-deploy.sh

gcp-setup:
	bash infra/gcp/setup-gcp.sh

gcp-setup-media:
	bash infra/gcp/setup-gcs-media.sh

gcp-setup-stripe:
	bash infra/gcp/setup-stripe-secrets.sh

gcp-deploy:
	gcloud builds submit --config=infra/gcp/cloudbuild.yaml .

gcp-domain:
	bash infra/gcp/setup-custom-domain.sh

firebase-flutter-setup:
	bash infra/firebase/setup-flutter-firebase.sh

firebase-android-dist:
	bash infra/firebase/distribute-android.sh

gcp-smoke: smoke-staging
