SHELL := /bin/bash
COMPOSE_FILE := deploy/docker-compose.yml
ENV_FILE     := .env
COMPOSE      := docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE)

.DEFAULT_GOAL := help

.PHONY: help env brand-sync usecases-sync usecases-check up up-infra up-pacs down migrate seed seed-mass api-dev api-billit-live pacs-demo-seed nuxtjs-dev flutter-dev test test-go test-flutter test-flutter-smoke test-nuxt test-auth test-e2e test-e2e-p0 smoke smoke-prod billit-sandbox-smoke billit-saas-master-smoke invoicing-staging-cleanup staging-quality-cleanup gcp-staging-quality-cleanup gcp-setup gcp-setup-prod gcp-github gcp-setup-media gcp-setup-stripe gcp-bff-proxy-secret gcp-retention-scheduler gcp-visit-reminders-scheduler gcp-saas-invoices-scheduler gcp-invoicing-reconcile-scheduler gcp-sales-branches-scheduler gcp-pharmacy-expiry-scheduler gcp-research-etl-scheduler gcp-vet-news-scheduler gcp-product-digest-scheduler gcp-product-digest-weekly-scheduler gcp-all-schedulers gcp-seed-scheduler gcp-delete-seed-scheduler gcp-deploy gcp-deploy-prod gcp-domain gcp-domain-prod gcp-smoke firebase-flutter-setup firebase-google-signin-android firebase-android-dist play-android-bundle play-android-bundle-internal play-android-bundle-prod import-cnk rotate-pet-documents

help:
	@echo "petsFollow — commandes"
	@echo "  make env            .env.example → .env"
	@echo "  make brand-sync     tokens CSS + Dart"
	@echo "  make up             stack Docker complète"
	@echo "  make up-infra       db + redis + mailhog"
	@echo "  make up-pacs        Orthanc local (:8042) — démo Imagerie"
	@echo "  make migrate        migrations API"
	@echo "  make seed           seed demo"
	@echo "  make seed-mass      densifie la DB (après seed) — volume démo prod-like"
	@echo "  make api-dev        API Go (bloque le terminal, port 8291)"
	@echo "  make pacs-demo-seed upload RX démo sur pet client.demo"
	@echo "  make api-billit-live  API Go Billit live (MOCK=false — secrets requis)"
	@echo "  make billit-sandbox-smoke  gates webhook+routes live (doc 34)"
	@echo "  make billit-saas-master-smoke  Flux A draft (+send) via BILLIT_MASTER_*"
	@echo "  make nuxtjs-dev     Web Pro Nuxt (autre terminal, port 3002)"
	@echo "  make flutter-dev    Flutter pets staging (émulateur) + Google Sign-In"
	@echo "  make test-go        tests Go"
	@echo "  make test-e2e-p0    Playwright @p0 (API+Nuxt locaux)"
	@echo "  make test-e2e       Playwright suite complète"
	@echo "  make test-auth      garde-fou login/forgot (Go + Vitest)"
	@echo "  make smoke          smoke API MVP (SMOKE_PROFILE=full)"
	@echo "  make smoke-prod     smoke API prod non-mutatif (health/ready/plans)"
	@echo "  make staging-quality-cleanup  purge smoke/e2e (PF_CLEANUP_TARGET=staging DATABASE_URL=…)"
	@echo "  make gcp-staging-quality-cleanup  idem via Cloud SQL staging (gcloud + proxy)"
	@echo "  make gcp-setup      Bootstrap staging (SQL/secrets/bucket)"
	@echo "  make gcp-setup-prod Bootstrap prod (SQL petsfollow-db-prod + secrets *-prod)"
	@echo "  make gcp-deploy     Cloud Build staging"
	@echo "  make gcp-deploy-prod  Cloud Build production (cloudbuild-prod.yaml)"
	@echo "  make gcp-domain     Domaines staging (LB + cert)"
	@echo "  make gcp-domain-prod Domaines petsfollow.app + api.petsfollow.app"
	@echo "  make firebase-flutter-setup  apps Firebase Android/iOS (staging + prod)"
	@echo "  make firebase-google-signin-android  SHA → Firebase (FLAVOR=staging|prod)"
	@echo "  make firebase-android-dist   APK staging → App Distribution (petsfollow-testers)"
	@echo "  make play-android-bundle-internal  AAB Play + API staging (Internal testing)"
	@echo "  make play-android-bundle-prod      AAB Play + API prod (api.petsfollow.app)"
	@echo "  make play-android-bundle     AAB Play (API_BASE=https://… requis)"
	@echo "  make gcp-setup-stripe        secrets Stripe GCP (placeholders + instructions)"
	@echo "  make gcp-bff-proxy-secret    secret BFF→API IP client (ATTACH=1 → API+Nuxt)"
	@echo "  make gcp-delete-seed-scheduler  supprime le Scheduler seed hebdo (reset = admin Pro)"
	@echo "  make gcp-retention-scheduler  Scheduler quotidien purge RGPD (RETENTION_PURGE_SECRET=…)"
	@echo "  make gcp-visit-reminders-scheduler  Scheduler quotidien rappel J-1 RDV (VISIT_REMINDERS_SECRET=…)"
	@echo "  make gcp-saas-invoices-scheduler  Scheduler mensuel brouillons SaaS Flux A C1"
	@echo "  make gcp-invoicing-reconcile-scheduler  Scheduler horaire relecture factures en vol (INVOICING_RECONCILE_SECRET=…)"
	@echo "  make gcp-sales-branches-scheduler  Scheduler 10h/18h auto-branches (SALES_BRANCHES_AUTO_SECRET=…)"
	@echo "  make gcp-pharmacy-expiry-scheduler Scheduler quotidien auto-quarantaine lots (PHARMACY_EXPIRY_SECRET=…)"
	@echo "  make gcp-research-etl-scheduler Scheduler 6h ETL Research (RESEARCH_ETL_SECRET=… RESEARCH_ANON_SALT=…)"
	@echo "  make gcp-vet-news-scheduler Scheduler veille news 06:00 (VET_NEWS_SECRET=…)"
	@echo "  make gcp-product-digest-scheduler Scheduler digest produit 18:00 (PRODUCT_DIGEST_SECRET=…)"
	@echo "  make gcp-product-digest-weekly-scheduler Scheduler digest hebdo samedi 08:00"
	@echo "  make gcp-all-schedulers  Tous les schedulers (PETSFOLLOW_GCP_ENV=prod pour prod)"
	@echo ""
	@echo "Dev local — 2 terminaux :"
	@echo "  T1: make up-infra && make up-pacs && make migrate && make seed && make api-dev"
	@echo "  T2: make nuxtjs-dev  →  http://localhost:3002"
	@echo "  Démo PACS: make pacs-demo-seed → fiche animal ?tab=imaging"
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

# Orthanc PACS local (port 8042) — démo Imagerie.
up-pacs: env
	$(COMPOSE) up -d orthanc
	@echo "→ Orthanc http://localhost:$${PETSFOLLOW_ORTHANC_PORT:-8042} (petsfollow/petsfollow)"

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

# One-shot : re-clé les documents animaux dont l'URL de stockage a fuité.
# Optionnel: ARGS='--dry-run'
rotate-pet-documents: env
	@set -a && source $(ENV_FILE) && set +a && cd go && GOTOOLCHAIN=local go run ./cmd/petsfollow-api rotate-pet-documents $(ARGS)

api-dev: env
	@set -a && source $(ENV_FILE) && set +a && cd go && GOTOOLCHAIN=local APP_ENV=$${APP_ENV:-local} MIGRATE_ON_BOOT=true DEV_SEED_ENABLED=true BILLING_MOCK_ENABLED=$${BILLING_MOCK_ENABLED:-true} BILLIT_ENABLED=$${BILLIT_ENABLED:-true} BILLIT_MOCK_ENABLED=$${BILLIT_MOCK_ENABLED:-true} AUTH_RATE_LIMIT_PER_MIN=$${AUTH_RATE_LIMIT_PER_MIN:-1000} TRUSTED_PROXY_HOPS=$${TRUSTED_PROXY_HOPS:-0} BFF_PROXY_SECRET=$${BFF_PROXY_SECRET:-dev-bff-proxy-secret} PHARMACY_ENABLED=$${PHARMACY_ENABLED:-true} PRESCRIPTIONS_ENABLED=$${PRESCRIPTIONS_ENABLED:-true} PACS_ENABLED=$${PACS_ENABLED:-true} PACS_ORTHANC_URL=$${PACS_ORTHANC_URL:-http://localhost:8042} PACS_ORTHANC_USER=$${PACS_ORTHANC_USER:-petsfollow} PACS_ORTHANC_PASSWORD=$${PACS_ORTHANC_PASSWORD:-petsfollow} PACS_ORTHANC_USE_ID_TOKEN=$${PACS_ORTHANC_USE_ID_TOKEN:-false} RESEARCH_ENABLED=$${RESEARCH_ENABLED:-true} RESEARCH_ETL_SECRET=$${RESEARCH_ETL_SECRET:-dev-research-etl} RESEARCH_ANON_SALT=$${RESEARCH_ANON_SALT:-petsfollow-research-dev-salt} VET_NEWS_ENABLED=$${VET_NEWS_ENABLED:-true} VET_NEWS_SECRET=$${VET_NEWS_SECRET:-dev-vet-news} CLIENT_AI_ENABLED=$${CLIENT_AI_ENABLED:-true} VAMREG_DRY_RUN=$${VAMREG_DRY_RUN:-true} SMS_ENABLED=$${SMS_ENABLED:-true} SMS_DRY_RUN=$${SMS_DRY_RUN:-true} VISIT_REMINDERS_SECRET=$${VISIT_REMINDERS_SECRET:-dev-visit-reminders} go run ./cmd/petsfollow-api

# Upload fixture DICOM sur pet client.demo (API + Orthanc requis).
pacs-demo-seed: env
	@bash scripts/pacs-demo-seed.sh

# API Billit live (sandbox Access Point). Exige BILLIT_WEBHOOK_SECRET + BILLIT_SECRETS_BACKEND=local_enc + BILLIT_SECRETS_KEY.
# Ne pas confondre avec api-dev (mock on par défaut). Smoke : make billit-sandbox-smoke
# Prod live (api.billit.be) : BILLIT_BASE_URL=https://api.billit.be make api-billit-live — whitelist Billit requise.
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
	  BILLIT_BASE_URL=$${BILLIT_BASE_URL:-https://api.sandbox.billit.be} \
	  BILLIT_RESELLER_REGISTER_URL=$${BILLIT_RESELLER_REGISTER_URL:-https://my.sandbox.billit.be/Account/Register} \
	  AUTH_RATE_LIMIT_PER_MIN=$${AUTH_RATE_LIMIT_PER_MIN:-1000} \
	  go run ./cmd/petsfollow-api

nuxtjs-dev: env
	@set -a && source $(ENV_FILE) && set +a && cd nuxtjs && npm install && BFF_PROXY_SECRET=$${BFF_PROXY_SECRET:-dev-bff-proxy-secret} NUXT_PUBLIC_BILLIT_ENABLED=$${NUXT_PUBLIC_BILLIT_ENABLED:-true} NUXT_PUBLIC_PHARMACY_ENABLED=$${NUXT_PUBLIC_PHARMACY_ENABLED:-true} NUXT_PUBLIC_PRESCRIPTIONS_ENABLED=$${NUXT_PUBLIC_PRESCRIPTIONS_ENABLED:-true} NUXT_PUBLIC_PACS_ENABLED=$${NUXT_PUBLIC_PACS_ENABLED:-true} NUXT_PUBLIC_RESEARCH_ENABLED=$${NUXT_PUBLIC_RESEARCH_ENABLED:-true} NUXT_PUBLIC_VET_NEWS_ENABLED=$${NUXT_PUBLIC_VET_NEWS_ENABLED:-true} NUXT_PUBLIC_SITES_UI_ENABLED=$${NUXT_PUBLIC_SITES_UI_ENABLED:-true} npx nuxt dev --port $${PETSFOLLOW_NUXTJS_PORT:-3002} --host 0.0.0.0

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
		--dart-define=GOOGLE_SERVER_CLIENT_ID=$(GOOGLE_SERVER_CLIENT_ID) \
		--dart-define=CLIENT_AI_ENABLED=$${CLIENT_AI_ENABLED:-true}

test-go:
	# -p 1 : les suites d'intégration partagent une seule base (seed tronque pendant que handlers lit).
	cd go && GOTOOLCHAIN=local go test -p 1 ./...

# Benchmarks des chemins chauds (réponses JSON, JWT, i18n). Comparer deux runs :
# make bench-go > /tmp/before.txt … make bench-go > /tmp/after.txt puis benchstat.
bench-go:
	cd go && GOTOOLCHAIN=local go test ./... -run '^$$' -bench . -benchmem

go-vuln:
	cd go && GOTOOLCHAIN=local go run golang.org/x/vuln/cmd/govulncheck@latest ./...

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

smoke-prod:
	SMOKE_PROFILE=prod PETSFOLLOW_API_URL=$${PETSFOLLOW_API_URL:-https://api.petsfollow.app} bash scripts/smoke-test.sh

# Gates Billit live (doc 34). Refuse si MOCK=true. Optionnel: BILLIT_SMOKE_PARTY_ID + BILLIT_SMOKE_API_KEY.
billit-sandbox-smoke:
	@bash scripts/smoke-billit-sandbox.sh

# Purge docs practice draft/rejected/sending des cabinets *@petsfollow.test.
# PF_CLEANUP_TARGET=staging DATABASE_URL=… make invoicing-staging-cleanup
# Options : CLEANUP_ARGS='--connections' ou CLEANUP_ARGS='--dry-run'
invoicing-staging-cleanup:
	@test -n "$${DATABASE_URL:-}" || (echo "DATABASE_URL required" >&2; exit 1)
	@bash scripts/cleanup-staging-invoicing.sh $${CLEANUP_ARGS:-}

# Purge artefacts smoke/e2e staging (messages, BP/labs, prospects, users smoke*, factu draft).
# PF_CLEANUP_TARGET=staging DATABASE_URL=… make staging-quality-cleanup
# Options : CLEANUP_ARGS='--dry-run' ou CLEANUP_ARGS='--skip-invoicing'
# Via Cloud SQL staging (WIF/gcloud) : make gcp-staging-quality-cleanup
staging-quality-cleanup:
	@test -n "$${DATABASE_URL:-}" || (echo "DATABASE_URL required" >&2; exit 1)
	@bash scripts/cleanup-staging-quality.sh $${CLEANUP_ARGS:-}

gcp-staging-quality-cleanup:
	@bash infra/gcp/cleanup-staging-quality.sh $${CLEANUP_ARGS:-}

# Flux A live : 1 cabinet draft master (+ send si BILLIT_SMOKE_SAAS_SEND=1).
billit-saas-master-smoke:
	@bash scripts/smoke-billit-saas-master.sh

# Smoke staging mutatif → purge quality systématique derrière (même en échec).
smoke-staging:
	@set +e; PETSFOLLOW_API_URL=https://api.petsfollow.ll-it-sc.be bash scripts/smoke-test.sh; rc=$$?; \
	echo "→ Purge artefacts smoke/e2e (staging DB)"; \
	bash infra/gcp/cleanup-staging-quality.sh || echo "WARN: purge quality échouée — relancer make gcp-staging-quality-cleanup" >&2; \
	exit $$rc

# Pharmacie S6 : receipt → DAF → VAMReg dry-run → movements (local ou PETSFOLLOW_API_URL=staging).
smoke-pharmacy-s6:
	@bash scripts/smoke-pharmacy-s6.sh

# Mutatif sur staging → purge quality systématique derrière (même en échec).
smoke-pharmacy-s6-staging:
	@set +e; PETSFOLLOW_API_URL=https://api.petsfollow.ll-it-sc.be bash scripts/smoke-pharmacy-s6.sh; rc=$$?; \
	echo "→ Purge artefacts smoke/e2e (staging DB)"; \
	bash infra/gcp/cleanup-staging-quality.sh || echo "WARN: purge quality échouée — relancer make gcp-staging-quality-cleanup" >&2; \
	exit $$rc

gcp-github:
	bash infra/gcp/setup-github-deploy.sh

gcp-setup:
	bash infra/gcp/setup-gcp.sh

gcp-setup-prod:
	bash infra/gcp/setup-gcp-prod.sh

gcp-setup-media:
	bash infra/gcp/setup-gcs-media.sh

gcp-setup-stripe:
	bash infra/gcp/setup-stripe-secrets.sh

# Secret BFF→API (X-PF-Client-IP). ATTACH=1 monte sur petsfollow-api + petsfollow-nuxtjs.
# Prod : PETSFOLLOW_GCP_ENV=prod make gcp-bff-proxy-secret ATTACH=1
gcp-bff-proxy-secret:
	@if [ "$${ATTACH:-}" = "1" ]; then \
	  bash infra/gcp/setup-bff-proxy-secret.sh --attach-run; \
	else \
	  bash infra/gcp/setup-bff-proxy-secret.sh; \
	fi

gcp-retention-scheduler:
	bash infra/gcp/setup-retention-scheduler.sh

gcp-visit-reminders-scheduler:
	bash infra/gcp/setup-visit-reminders-scheduler.sh

gcp-saas-invoices-scheduler:
	bash infra/gcp/setup-saas-invoices-scheduler.sh

gcp-invoicing-reconcile-scheduler:
	bash infra/gcp/setup-invoicing-reconcile-scheduler.sh

gcp-sales-branches-scheduler:
	bash infra/gcp/setup-sales-branches-scheduler.sh

gcp-pharmacy-expiry-scheduler:
	bash infra/gcp/setup-pharmacy-expiry-scheduler.sh

gcp-research-etl-scheduler:
	bash infra/gcp/setup-research-etl-scheduler.sh

gcp-vet-news-scheduler:
	bash infra/gcp/setup-vet-news-scheduler.sh

gcp-product-digest-scheduler:
	bash infra/gcp/setup-product-digest-scheduler.sh

gcp-product-digest-weekly-scheduler:
	bash infra/gcp/setup-product-digest-weekly-scheduler.sh

gcp-all-schedulers:
	bash infra/gcp/setup-all-schedulers.sh

# Alias obsolète → suppression du job Scheduler (plus de reset auto).
gcp-seed-scheduler: gcp-delete-seed-scheduler

gcp-delete-seed-scheduler:
	bash infra/gcp/delete-seed-scheduler.sh

gcp-deploy:
	@echo "→ Préférer push branche staging (CI + smoke + Playwright). Deploy nu = hors filet."
	gcloud builds submit --config=infra/gcp/cloudbuild.yaml .

# Production : préférer workflow_dispatch deploy-gcp-prod.yml. Garde-fou SQL FIXME dans le YAML.
gcp-deploy-prod:
	@echo "→ Deploy prod (services *-prod). Prérequis : SQL/DNS/secrets — doc 10-GCP § Production."
	gcloud builds submit --config=infra/gcp/cloudbuild-prod.yaml .

gcp-domain:
	bash infra/gcp/setup-custom-domain.sh

gcp-domain-prod:
	PETSFOLLOW_GCP_ENV=prod bash infra/gcp/setup-custom-domain.sh

firebase-flutter-setup:
	bash infra/firebase/setup-flutter-firebase.sh

# Usage : make firebase-google-signin-android SHA1=ED:B0:... [SHA256=...] [FLAVOR=staging|prod]
firebase-google-signin-android:
	@test -n "$(SHA1)" || (echo "Usage: make firebase-google-signin-android SHA1=.. [SHA256=..] [FLAVOR=staging|prod]" >&2; exit 1)
	FLAVOR=$(FLAVOR) bash infra/firebase/setup-google-signin-android.sh "$(SHA1)" $(if $(SHA256),"$(SHA256)",)

firebase-android-dist:
	bash infra/firebase/distribute-android.sh

# Generic : API_BASE=https://… make play-android-bundle
play-android-bundle:
	bash infra/play/build-play-bundle.sh

# Play Internal testing — package prod, API staging (seedable).
play-android-bundle-internal:
	@set -a; source infra/play/api-bases.sh; set +a; \
	ALLOW_STAGING_API=1 API_BASE="$${PETSFOLLOW_API_STAGING}" bash infra/play/build-play-bundle.sh

# Play Production track — API prod (après deploy main / GCP prod).
play-android-bundle-prod:
	@set -a; source infra/play/api-bases.sh; set +a; \
	API_BASE="$${PETSFOLLOW_API_PROD}" bash infra/play/build-play-bundle.sh

gcp-smoke: smoke-staging
