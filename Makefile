SHELL := /bin/bash
COMPOSE_FILE := deploy/docker-compose.yml
ENV_FILE     := .env
COMPOSE      := docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE)

.DEFAULT_GOAL := help

.PHONY: help env brand-sync usecases-sync usecases-check up up-infra down migrate seed seed-mass api-dev api-billit-live nuxtjs-dev flutter-dev test test-go test-flutter test-flutter-smoke test-nuxt test-auth test-e2e test-e2e-p0 smoke billit-sandbox-smoke gcp-setup gcp-github gcp-setup-media gcp-setup-stripe gcp-retention-scheduler gcp-saas-invoices-scheduler gcp-sales-branches-scheduler gcp-seed-scheduler gcp-delete-seed-scheduler gcp-deploy gcp-domain gcp-smoke firebase-flutter-setup firebase-google-signin-android firebase-android-dist play-android-bundle import-cnk

help:
	@echo "petsFollow — commandes"
	@echo "  make env            .env.example → .env"
	@echo "  make brand-sync     tokens CSS + Dart"
	@echo "  make up             stack Docker complète"
	@echo "  make up-infra       db + redis + mailhog"
	@echo "  make migrate        migrations API"
	@echo "  make seed           seed demo"
	@echo "  make seed-mass      densifie la DB (après seed) — volume démo prod-like"
	@echo "  make api-dev        API Go (bloque le terminal, port 8291)"
	@echo "  make api-billit-live  API Go Billit live (MOCK=false — secrets requis)"
	@echo "  make billit-sandbox-smoke  gates webhook+routes live (doc 34)"
	@echo "  make nuxtjs-dev     Web Pro Nuxt (autre terminal, port 3002)"
	@echo "  make flutter-dev    Flutter pets staging (émulateur) + Google Sign-In"
	@echo "  make test-go        tests Go"
	@echo "  make test-e2e-p0    Playwright @p0 (API+Nuxt locaux)"
	@echo "  make test-e2e       Playwright suite complète"
	@echo "  make test-auth      garde-fou login/forgot (Go + Vitest)"
	@echo "  make smoke          smoke API MVP"
	@echo "  make gcp-deploy     Cloud Build staging"
	@echo "  make firebase-flutter-setup  apps Firebase Android/iOS (staging + prod)"
	@echo "  make firebase-google-signin-android  SHA → Firebase (FLAVOR=staging|prod)"
	@echo "  make firebase-android-dist   APK staging → App Distribution (petsfollow-testers)"
	@echo "  make play-android-bundle     AAB prod → Google Play (API_BASE=https://… requis)"
	@echo "  make gcp-setup-stripe        secrets Stripe GCP (placeholders + instructions)"
	@echo "  make gcp-delete-seed-scheduler  supprime le Scheduler seed hebdo (reset = admin Pro)"
	@echo "  make gcp-retention-scheduler  Scheduler quotidien purge RGPD (RETENTION_PURGE_SECRET=…)"
	@echo "  make gcp-saas-invoices-scheduler  Scheduler mensuel brouillons SaaS Flux A C1"
	@echo "  make gcp-sales-branches-scheduler  Scheduler 10h/18h auto-branches (SALES_BRANCHES_AUTO_SECRET=…)"
	@echo ""
	@echo "Dev local — 2 terminaux :"
	@echo "  T1: make up-infra && make migrate && make seed && make api-dev"
	@echo "  T2: make nuxtjs-dev  →  http://localhost:3002"
	@echo "  (+ Flutter) make flutter-dev"

env:
	@test -f $(ENV_FILE) || cp .env.example $(ENV_FILE)

brand-sync:
	@bash scripts/brand-sync.sh

usecases-sync:
	@node scripts/sync-usecases.mjs

# Échoue si useCase/*.md a divergé du catalogue Nuxt (CI).
usecases-check:
	@node scripts/sync-usecases.mjs
	@git diff --exit-code -- nuxtjs/data/usecases/catalog.json \
		|| (echo "catalog.json obsolète — lancez make usecases-sync et committez" >&2; exit 1)

up: env brand-sync
	$(COMPOSE) up --build -d

up-infra: env
	$(COMPOSE) up -d db redis mailhog

down:
	$(COMPOSE) down

migrate: env
	@set -a && source $(ENV_FILE) && set +a && cd go && GOTOOLCHAIN=local go run ./cmd/petsfollow-api migrate

seed: env
	@set -a && source $(ENV_FILE) && set +a && cd go && GOTOOLCHAIN=local APP_ENV=$${APP_ENV:-local} go run ./cmd/petsfollow-api seed

# Import dictionnaire CNK. Ex.: make import-cnk FILE=go/testdata/cnk_sample.csv
# Optionnel: ARGS='--dry-run' ou ARGS='--deactivate-missing'
# FILE peut être relatif au repo ou absolu.
import-cnk: env
	@test -n "$(FILE)" || (echo "Usage: make import-cnk FILE=path.csv [ARGS='--dry-run']"; exit 1)
	@CNK_FILE="$(FILE)"; \
	  case "$$CNK_FILE" in /*) ;; *) CNK_FILE="$(CURDIR)/$$CNK_FILE" ;; esac; \
	  set -a && source $(ENV_FILE) && set +a && cd go && GOTOOLCHAIN=local go run ./cmd/petsfollow-api import-cnk --file=$$CNK_FILE $(ARGS)

seed-mass: env
	@set -a && source $(ENV_FILE) && set +a && cd go && GOTOOLCHAIN=local APP_ENV=$${APP_ENV:-local} go run ./cmd/petsfollow-api seed-mass

api-dev: env
	@set -a && source $(ENV_FILE) && set +a && cd go && GOTOOLCHAIN=local APP_ENV=$${APP_ENV:-local} MIGRATE_ON_BOOT=true DEV_SEED_ENABLED=true BILLING_MOCK_ENABLED=$${BILLING_MOCK_ENABLED:-true} BILLIT_ENABLED=$${BILLIT_ENABLED:-true} BILLIT_MOCK_ENABLED=$${BILLIT_MOCK_ENABLED:-true} AUTH_RATE_LIMIT_PER_MIN=$${AUTH_RATE_LIMIT_PER_MIN:-1000} PHARMACY_ENABLED=$${PHARMACY_ENABLED:-true} PRESCRIPTIONS_ENABLED=$${PRESCRIPTIONS_ENABLED:-true} go run ./cmd/petsfollow-api

# API Billit live (sandbox). Exige BILLIT_WEBHOOK_SECRET + BILLIT_SECRETS_BACKEND=local_enc + BILLIT_SECRETS_KEY.
# Ne pas confondre avec api-dev (mock on par défaut). Smoke : make billit-sandbox-smoke
api-billit-live: env
	@set -a && source $(ENV_FILE) && set +a; \
	  test -n "$${BILLIT_WEBHOOK_SECRET:-}" || (echo "BILLIT_WEBHOOK_SECRET required" >&2; exit 1); \
	  test "$${BILLIT_SECRETS_BACKEND:-}" = "local_enc" || (echo "BILLIT_SECRETS_BACKEND=local_enc required (not plain_dev)" >&2; exit 1); \
	  test -n "$${BILLIT_SECRETS_KEY:-}" || (echo "BILLIT_SECRETS_KEY required" >&2; exit 1); \
	  cd go && GOTOOLCHAIN=local APP_ENV=$${APP_ENV:-local} MIGRATE_ON_BOOT=true DEV_SEED_ENABLED=true \
	  BILLING_MOCK_ENABLED=$${BILLING_MOCK_ENABLED:-true} \
	  BILLIT_ENABLED=true BILLIT_MOCK_ENABLED=false \
	  BILLIT_WEBHOOK_SECRET=$${BILLIT_WEBHOOK_SECRET} \
	  BILLIT_SECRETS_BACKEND=local_enc BILLIT_SECRETS_KEY=$${BILLIT_SECRETS_KEY} \
	  BILLIT_BASE_URL=$${BILLIT_BASE_URL:-https://api.billit.be} \
	  BILLIT_RESELLER_REGISTER_URL=$${BILLIT_RESELLER_REGISTER_URL} \
	  AUTH_RATE_LIMIT_PER_MIN=$${AUTH_RATE_LIMIT_PER_MIN:-1000} \
	  go run ./cmd/petsfollow-api

nuxtjs-dev: env
	@set -a && source $(ENV_FILE) && set +a && cd nuxtjs && npm install && NUXT_PUBLIC_BILLIT_ENABLED=$${NUXT_PUBLIC_BILLIT_ENABLED:-true} NUXT_PUBLIC_PHARMACY_ENABLED=$${NUXT_PUBLIC_PHARMACY_ENABLED:-true} NUXT_PUBLIC_PRESCRIPTIONS_ENABLED=$${NUXT_PUBLIC_PRESCRIPTIONS_ENABLED:-true} npx nuxt dev --port $${PETSFOLLOW_NUXTJS_PORT:-3002} --host 0.0.0.0

# Flavor staging (package …mobile.staging). Device physique : API_BASE=http://<LAN>:8291 make flutter-dev
GOOGLE_SERVER_CLIENT_ID ?= 237481297060-90gihf09ec8pv2cc3jhnnodjo00vejde.apps.googleusercontent.com
API_BASE ?= http://10.0.2.2:8291
# SHA → Firebase : make firebase-google-signin-android SHA1=… [FLAVOR=staging|prod]
FLAVOR ?= staging
flutter-dev: env
	cd flutter && flutter pub get && flutter run --flavor staging \
		--dart-define=FLAVOR=staging \
		--dart-define=APP_ENV=staging \
		--dart-define=API_BASE=$(API_BASE) \
		--dart-define=GOOGLE_SERVER_CLIENT_ID=$(GOOGLE_SERVER_CLIENT_ID)

test-go:
	# -p 1 : les suites d'intégration partagent une seule base (seed tronque pendant que handlers lit).
	cd go && GOTOOLCHAIN=local go test -p 1 ./...

test-flutter:
	cd flutter && flutter pub get && flutter test

# Smoke API locale seedée (api-dev :8291). Skip propre si API down.
test-flutter-smoke:
	cd flutter && flutter pub get && flutter test test/smoke/ --dart-define=RUN_FLUTTER_SMOKE=true

test-nuxt:
	$(MAKE) usecases-check
	cd nuxtjs && npm install && npm test

# Garde-fou connexion / reset — à lancer avant deploy (DB seedée pour le volet Go).
test-auth:
	cd go && GOTOOLCHAIN=local go test ./internal/handlers/ -run 'TestAuth' -count=1
	cd nuxtjs && npm test -- tests/unit/useAuth.spec.ts

# Playwright P0 — nécessite API :8291 + Nuxt :3002 + seed.
test-e2e-p0:
	cd nuxtjs && npx playwright test --grep @p0

test-e2e:
	cd nuxtjs && npm run test:e2e

test: test-go test-nuxt test-flutter

smoke:
	@bash scripts/smoke-test.sh

# Gates Billit live (doc 34). Refuse si MOCK=true. Optionnel: BILLIT_SMOKE_PARTY_ID + BILLIT_SMOKE_API_KEY.
billit-sandbox-smoke:
	@bash scripts/smoke-billit-sandbox.sh

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

gcp-retention-scheduler:
	bash infra/gcp/setup-retention-scheduler.sh

gcp-saas-invoices-scheduler:
	bash infra/gcp/setup-saas-invoices-scheduler.sh

gcp-sales-branches-scheduler:
	bash infra/gcp/setup-sales-branches-scheduler.sh

# Alias obsolète → suppression du job Scheduler (plus de reset auto).
gcp-seed-scheduler: gcp-delete-seed-scheduler

gcp-delete-seed-scheduler:
	bash infra/gcp/delete-seed-scheduler.sh

gcp-deploy:
	@echo "→ Préférer push branche staging (CI + smoke + Playwright). Deploy nu = hors filet."
	gcloud builds submit --config=infra/gcp/cloudbuild.yaml .

gcp-domain:
	bash infra/gcp/setup-custom-domain.sh

firebase-flutter-setup:
	bash infra/firebase/setup-flutter-firebase.sh

# Usage : make firebase-google-signin-android SHA1=ED:B0:... [SHA256=...] [FLAVOR=staging|prod]
firebase-google-signin-android:
	@test -n "$(SHA1)" || (echo "Usage: make firebase-google-signin-android SHA1=.. [SHA256=..] [FLAVOR=staging|prod]" >&2; exit 1)
	FLAVOR=$(FLAVOR) bash infra/firebase/setup-google-signin-android.sh "$(SHA1)" $(if $(SHA256),"$(SHA256)",)

firebase-android-dist:
	bash infra/firebase/distribute-android.sh

play-android-bundle:
	bash infra/play/build-play-bundle.sh

gcp-smoke: smoke-staging
