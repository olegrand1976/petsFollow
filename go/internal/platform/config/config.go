package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	RedisAddr      string
	RedisKeyPrefix string
	JWTSigningKey  string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration
	LogLevel       string
	MigrateOnBoot  bool
	DevSeedEnabled bool
	SMTPHost       string
	SMTPPort       int
	SMTPFrom       string
	// SMTPUser / SMTPPass — auth PLAIN (OVH :587). Vides = MailHog / relay ouvert.
	SMTPUser                       string
	SMTPPass                       string
	HeartRateSeconds               int
	StripeSecretKey                string
	StripeWebhookSecret            string
	StripePriceAnnualOnetime       string
	StripePriceTriennialOnetime    string
	StripePriceQuinquennialOnetime string // legacy
	StripePriceMonthlySub          string
	StripePriceAnnualSub           string
	StripePriceTriennialSub        string
	StripePriceQuinquennialSub     string // legacy unused
	StripePriceAddonFamily         string // legacy
	StripePriceAddonKennel         string // legacy
	StripePriceAddonCarePlus       string // legacy
	StripePriceAddonHorse          string // legacy
	StripeSuccessURL               string
	StripeCancelURL                string
	APIPublicURL                   string
	ProPublicSiteURL               string
	BillingMockEnabled             bool
	GoogleOAuthClientID            string
	GCSMediaBucket                 string
	MediaLocalDir                  string
	LLITWebsiteURL                 string
	PetsAppDownloadURL             string
	FCMEnabled                     bool
	JourneyEmailEnabled            bool
	JourneyEmailInterval           time.Duration
	GeminiAPIKey                   string
	GeminiModel                    string
	GeminiLiteModel                string
	GeminiLiveModel                string
	PitchAnalyzerSecret            string
	ProductDigestSecret            string
	VertexProject                  string
	VertexLocation                 string
	// CareProPublicRegister enables POST /auth/register-care-pro (default off — admin creates care_pro).
	CareProPublicRegister bool
	// AuthRateLimitPerMin limite les endpoints auth publics par IP/minute (0 = désactivé).
	AuthRateLimitPerMin int
	// TrustedProxyHops — nombre de reverse proxies devant l'API (1 sur Cloud Run, 0 sans proxy).
	// Détermine quelle entrée de X-Forwarded-For sert de clé de rate limit (clients directs).
	TrustedProxyHops int
	// BFFProxySecret — secret partagé avec la BFF Nuxt (X-PF-Proxy-Secret). Vide = ignore X-PF-Client-IP.
	BFFProxySecret string
	// CORSAllowedOrigins — origines autorisées (séparées par des virgules). Défaut : site Pro public.
	CORSAllowedOrigins string
	// RetentionPurgeSecret protège POST /internal/retention/run (purge 3 ans d'inactivité RGPD).
	RetentionPurgeSecret string
	// PprofSecret protège /internal/debug/pprof/* (profils tas/goroutines/CPU).
	// Vide = routes non montées.
	PprofSecret string
	// SalesBranchesAutoSecret protège POST /internal/sales-branches-auto/run.
	SalesBranchesAutoSecret string
	// AiModuleFrictionSecret protège POST /internal/ai-module-friction/run.
	AiModuleFrictionSecret string
	// OpsNotifyEmail reçoit les leads « nouveau véto » + alertes auth ALERT/URGENT.
	OpsNotifyEmail string
	// CommercialContactPhone — fallback téléphone commercial (mail/PDF dossier) si profil vide.
	CommercialContactPhone string
	// SupportInboxEmail reçoit les nouveaux tickets bug-report (défaut barbara@petsfollow.app).
	SupportInboxEmail string
	// AuthHealthSecret protège POST /internal/auth-health/run.
	AuthHealthSecret string
	// MLMOrgEnabled exposes multi-depth downline UI; commissions remain flat until MLM billing ships.
	MLMOrgEnabled bool
	// PharmacyEnabled enables vet pharmacy module (CNK dictionary, stock, DAF) — default off.
	PharmacyEnabled bool
	// PharmacyExpirySecret protège POST /internal/pharmacy/expiry-run.
	PharmacyExpirySecret string
	// PharmacyWorkersEnabled starts Asynq server + enqueue for VAMReg (and future invoices.connect).
	PharmacyWorkersEnabled bool
	// VamregDryRun skips real declaration HTTP (default true until write credentials).
	VamregDryRun  bool
	VamregBaseURL string
	VamregAPIKey  string
	// VamregAfmps* — FAMHP readonly reference lists (ICD software-house, FAMHP-SEC-KEY).
	// Separate from VamregBaseURL/APIKey (declaration gateway) to avoid auth/path collisions.
	VamregAfmpsBaseURL string
	VamregAfmpsAPIKey  string
	// PrescriptionsEnabled enables veterinary prescription drafts + PDF preview — default off.
	PrescriptionsEnabled bool
	// PacsEnabled enables Orthanc PACS orchestration (status/wake/viewer proxy) — default off.
	PacsEnabled bool
	// PacsOrthancURL is the Orthanc Cloud Run / local base URL (no trailing slash).
	PacsOrthancURL      string
	PacsOrthancUser     string
	PacsOrthancPassword string
	// PacsOrthancUseIDToken sends a Google identity token (Cloud Run IAM invoker).
	PacsOrthancUseIDToken bool
	// ResearchEnabled enables petsFollow Research observatory (anonymized epi aggregates) — default off.
	ResearchEnabled bool
	// ResearchEtlSecret protects POST /internal/research-etl/run.
	ResearchEtlSecret string
	// ResearchAnonSalt salts practice_id_hash in research.anon_events (never exposed via API).
	ResearchAnonSalt string
	// VetNewsEnabled enables multi-source veterinary news ingest + dashboard widget — default off, tag dev.
	VetNewsEnabled bool
	// VetNewsSecret protects POST /internal/vet-news/run.
	VetNewsSecret string
	// ClientAIEnabled enables Flutter client AI (CR explain + triage 24/7) — default off, tag dev.
	ClientAIEnabled bool
	// AiCrAdvancedEnabled enables multi-agent CR improve + RAG knowledge base — default off, tag dev.
	AiCrAdvancedEnabled bool
	// GeminiEmbeddingModel is used for RAG chunk embeddings (768-dim).
	GeminiEmbeddingModel string
	// RagReindexSecret protects POST /internal/rag/reindex.
	RagReindexSecret string
	// SMSEnabled enables transactional client SMS via Telnyx (visit confirm/reminder/reschedule) — default off, tag dev.
	SMSEnabled bool
	// SMSDryRun logs SMS instead of calling Telnyx (default true until live credentials).
	SMSDryRun bool
	// Telnyx* — Messages API credentials; From is an optional long code / alphanumeric sender.
	TelnyxAPIKey             string
	TelnyxMessagingProfileID string
	TelnyxFrom               string
	// TelnyxPublicKey (base64, portail Telnyx » Auth) valide la signature Ed25519
	// des webhooks DLR / SMS entrants. Sans elle, aucun webhook n'est accepté.
	TelnyxPublicKey string
	// SMSDefaultRegion maps national numbers (0…) to a country prefix for E.164 (BE|FR|NL|LU).
	SMSDefaultRegion string
	// VisitRemindersSecret protège POST /internal/visit-reminders/run (rappel J-1).
	VisitRemindersSecret string
	// VisitReminderLookaheadHours — fin de fenêtre du rappel (défaut 30h ; début fixe now+1h).
	VisitReminderLookaheadHours int

	// BillitEnabled exposes invoicing routes (Billit reseller / Peppol).
	BillitEnabled bool
	// BillitMockEnabled uses the mock gateway (local/CI) — never call Billit live.
	BillitMockEnabled         bool
	BillitBaseURL             string
	BillitMasterPartyID       string
	BillitMasterAPIKey        string
	BillitResellerRegisterURL string
	BillitWebhookSecret       string
	BillitDefaultDocsIncluded int
	BillitSecretsBackend      string // local_enc | plain_dev
	BillitSecretsKey          string
	// InvoicingSaasEnabled réveille le Flux A (facturation SaaS LL-IT-SC → cabinet
	// via le compte Billit master). Off par défaut : l'abonnement Pro est facturé
	// hors application (commercial / compta). Billit ne sert que cabinet → clients.
	InvoicingSaasEnabled bool
	// InvoicingSaasPriceEURCents = SaaS Pro monthly HT in cents (Flux A draft). Default 8800 = 88 €.
	InvoicingSaasPriceEURCents int
	// InvoicingSaasAllowlist = UUIDs opted-in for Flux A without DB flag (comma-separated).
	InvoicingSaasAllowlist []string
	// SaasInvoicesSecret protège POST /internal/saas-invoices/run (cron brouillons Flux A).
	SaasInvoicesSecret string
	// InvoicingReconcileSecret protège POST /internal/invoicing-reconcile/run
	// (relecture des documents restés en `sending` après un webhook manqué).
	InvoicingReconcileSecret string
	// SeedNotifyStaff emails admin/commercial/commercial_manager after seed (staging reset).
	SeedNotifyStaff bool
	// AdminStagingSeedEnabled enables POST /admin/staging/seed (staging only — never prod).
	AdminStagingSeedEnabled bool
}

func Load() Config {
	return Config{
		HTTPAddr:                       envOr("HTTP_ADDR", ":8080"),
		DatabaseURL:                    envOr("DATABASE_URL", "postgres://petsfollow:petsfollow@localhost:5437/petsfollow?sslmode=disable"),
		RedisAddr:                      envOr("REDIS_ADDR", "localhost:6382"),
		RedisKeyPrefix:                 envOr("REDIS_KEY_PREFIX", "petsfollow:"),
		JWTSigningKey:                  envOr("JWT_SIGNING_KEY", "dev-change-me"),
		JWTAccessTTL:                   envDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:                  envDuration("JWT_REFRESH_TTL", 30*24*time.Hour),
		LogLevel:                       envOr("LOG_LEVEL", "info"),
		MigrateOnBoot:                  envBool("MIGRATE_ON_BOOT"),
		DevSeedEnabled:                 envBool("DEV_SEED_ENABLED"),
		SMTPHost:                       envOr("SMTP_HOST", "localhost"),
		SMTPPort:                       envInt("SMTP_PORT", 1027),
		SMTPFrom:                       envOr("SMTP_FROM", "petsFollow <noreply@petsfollow.test>"),
		SMTPUser:                       envOr("SMTP_USER", ""),
		SMTPPass:                       envOr("SMTP_PASS", ""),
		HeartRateSeconds:               envInt("HEARTRATE_DURATION_SEC", 60),
		StripeSecretKey:                envOr("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:            envOr("STRIPE_WEBHOOK_SECRET", "whsec_test"),
		StripePriceAnnualOnetime:       envOr("STRIPE_PRICE_ANNUAL_ONETIME", ""),
		StripePriceTriennialOnetime:    envOr("STRIPE_PRICE_TRIENNIAL_ONETIME", ""),
		StripePriceQuinquennialOnetime: envOr("STRIPE_PRICE_QUINQUENNIAL_ONETIME", ""),
		StripePriceMonthlySub:          envOr("STRIPE_PRICE_MONTHLY_SUB", ""),
		StripePriceAnnualSub:           envOr("STRIPE_PRICE_ANNUAL_SUB", ""),
		StripePriceTriennialSub:        envOr("STRIPE_PRICE_TRIENNIAL_SUB", ""),
		StripePriceQuinquennialSub:     envOr("STRIPE_PRICE_QUINQUENNIAL_SUB", ""),
		StripePriceAddonFamily:         envOr("STRIPE_PRICE_ADDON_FAMILY", ""),
		StripePriceAddonKennel:         envOr("STRIPE_PRICE_ADDON_KENNEL", ""),
		StripePriceAddonCarePlus:       envOr("STRIPE_PRICE_ADDON_CARE_PLUS", ""),
		StripePriceAddonHorse:          envOr("STRIPE_PRICE_ADDON_HORSE", ""),
		StripeSuccessURL:               envOr("STRIPE_SUCCESS_URL", "petsfollow://payment/success"),
		StripeCancelURL:                envOr("STRIPE_CANCEL_URL", "petsfollow://payment/cancel"),
		APIPublicURL:                   envOr("PETSFOLLOW_API_PUBLIC_URL", "http://localhost:8291"),
		ProPublicSiteURL:               envOr("PETSFOLLOW_PUBLIC_SITE_URL", "http://localhost:3002"),
		// Mock billing uniquement sur opt-in explicite — jamais inféré de l'absence de clé Stripe.
		BillingMockEnabled:  envBool("BILLING_MOCK_ENABLED"),
		GoogleOAuthClientID: envOr("GOOGLE_OAUTH_CLIENT_ID", ""),
		GCSMediaBucket:      envOr("GCS_MEDIA_BUCKET", ""),
		MediaLocalDir:       envOr("MEDIA_LOCAL_DIR", "./data/uploads"),
		LLITWebsiteURL:      envOr("LLIT_WEBSITE_URL", "https://ll-it-sc.be"),
		PetsAppDownloadURL:  envOr("PETS_APP_DOWNLOAD_URL", "https://appdistribution.firebase.google.com/testerapps/1:237481297060:android:cfda5c59a08bfd6dc9d231"),
		// FCM enabled by default; ADC (GOOGLE_APPLICATION_CREDENTIALS / Cloud Run SA) required to actually send.
		FCMEnabled:              envBoolDefault("FCM_ENABLED", true),
		JourneyEmailEnabled:     envBoolDefault("JOURNEY_EMAIL_ENABLED", true),
		JourneyEmailInterval:    envDuration("JOURNEY_EMAIL_INTERVAL", time.Hour),
		GeminiAPIKey:            envOr("GEMINI_API_KEY", ""),
		GeminiModel:             envOr("GEMINI_MODEL", "gemini-3.6-flash"),
		GeminiLiteModel:         envOr("GEMINI_LITE_MODEL", "gemini-3.5-flash-lite"),
		GeminiLiveModel:         envOr("GEMINI_LIVE_MODEL", "gemini-2.5-flash-native-audio-preview-09-2025"),
		GeminiEmbeddingModel:    envOr("GEMINI_EMBEDDING_MODEL", "text-embedding-004"),
		PitchAnalyzerSecret:     envOr("PITCH_ANALYZER_SECRET", ""),
		ProductDigestSecret:     envOr("PRODUCT_DIGEST_SECRET", ""),
		VertexProject:           envOr("VERTEX_PROJECT", ""),
		VertexLocation:          envOr("VERTEX_LOCATION", "europe-west9"),
		CareProPublicRegister:   envBool("CARE_PRO_PUBLIC_REGISTER"),
		AuthRateLimitPerMin:     envInt("AUTH_RATE_LIMIT_PER_MIN", 60),
		TrustedProxyHops:        envInt("TRUSTED_PROXY_HOPS", 1),
		BFFProxySecret:          envOr("BFF_PROXY_SECRET", ""),
		CORSAllowedOrigins:      envOr("CORS_ALLOWED_ORIGINS", ""),
		RetentionPurgeSecret:    envOr("RETENTION_PURGE_SECRET", ""),
		PprofSecret:             envOr("PPROF_SECRET", ""),
		SalesBranchesAutoSecret: envOr("SALES_BRANCHES_AUTO_SECRET", ""),
		AiModuleFrictionSecret:  envOr("AI_MODULE_FRICTION_SECRET", ""),
		OpsNotifyEmail:          envOr("OPS_NOTIFY_EMAIL", ""),
		CommercialContactPhone:  envOr("COMMERCIAL_CONTACT_PHONE", ""),
		SupportInboxEmail:       envOr("SUPPORT_INBOX_EMAIL", "barbara@petsfollow.app"),
		AuthHealthSecret:        envOr("AUTH_HEALTH_SECRET", ""),
		MLMOrgEnabled:           envBool("MLM_ORG_ENABLED"),
		PharmacyEnabled:         envBool("PHARMACY_ENABLED"),
		PharmacyExpirySecret:    envOr("PHARMACY_EXPIRY_SECRET", ""),
		PharmacyWorkersEnabled:  envBool("PHARMACY_WORKERS_ENABLED"),
		VamregDryRun:            envBoolDefault("VAMREG_DRY_RUN", true),
		VamregBaseURL:           envOr("VAMREG_BASE_URL", ""),
		VamregAPIKey:            envOr("VAMREG_API_KEY", ""),
		VamregAfmpsBaseURL:      envOr("VAMREG_AFMPS_BASE_URL", ""),
		VamregAfmpsAPIKey:       envOr("VAMREG_AFMPS_API_KEY", ""),
		PrescriptionsEnabled:    envBool("PRESCRIPTIONS_ENABLED"),
		PacsEnabled:             envBool("PACS_ENABLED"),
		PacsOrthancURL:          envOr("PACS_ORTHANC_URL", ""),
		PacsOrthancUser:         envOr("PACS_ORTHANC_USER", "petsfollow"),
		PacsOrthancPassword:     envOr("PACS_ORTHANC_PASSWORD", ""),
		PacsOrthancUseIDToken:   envBool("PACS_ORTHANC_USE_ID_TOKEN"),
		ResearchEnabled:         envBool("RESEARCH_ENABLED"),
		ResearchEtlSecret:       envOr("RESEARCH_ETL_SECRET", ""),
		ResearchAnonSalt:        envOr("RESEARCH_ANON_SALT", ""),
		VetNewsEnabled:          envBool("VET_NEWS_ENABLED"),
		VetNewsSecret:           envOr("VET_NEWS_SECRET", ""),
		ClientAIEnabled:         envBool("CLIENT_AI_ENABLED"),
		AiCrAdvancedEnabled:     envBool("AI_CR_ADVANCED_ENABLED"),
		RagReindexSecret:        envOr("RAG_REINDEX_SECRET", ""),

		// SMS transactionnel (Telnyx) : off par défaut ; dry-run tant que les creds live ne sont pas montés.
		SMSEnabled:                  envBool("SMS_ENABLED"),
		SMSDryRun:                   envBoolDefault("SMS_DRY_RUN", true),
		TelnyxAPIKey:                envOr("TELNYX_API_KEY", ""),
		TelnyxMessagingProfileID:    envOr("TELNYX_MESSAGING_PROFILE_ID", ""),
		TelnyxFrom:                  envOr("TELNYX_FROM", ""),
		TelnyxPublicKey:             envOr("TELNYX_PUBLIC_KEY", ""),
		SMSDefaultRegion:            envOr("SMS_DEFAULT_REGION", "BE"),
		VisitRemindersSecret:        envOr("VISIT_REMINDERS_SECRET", ""),
		VisitReminderLookaheadHours: envInt("VISIT_REMINDER_LOOKAHEAD_HOURS", 30),

		// Billit : off par défaut ; mock uniquement opt-in (comme BILLING_MOCK_ENABLED).
		// Environnements isolés (doc Access Point) :
		//   sandbox  → https://api.sandbox.billit.be  (+ my.sandbox.billit.be)
		//   production → https://api.billit.be         (+ my.billit.be)
		// Staging GCP pose la sandbox via deploy-run-args ; prod / défaut code = API live.
		// Clés sandbox ≠ prod — ne jamais réutiliser une ApiKey entre les deux.
		BillitEnabled:              envBool("BILLIT_ENABLED"),
		BillitMockEnabled:          envBool("BILLIT_MOCK_ENABLED"),
		BillitBaseURL:              envOr("BILLIT_BASE_URL", "https://api.billit.be"),
		BillitMasterPartyID:        envOr("BILLIT_MASTER_PARTY_ID", ""),
		BillitMasterAPIKey:         envOr("BILLIT_MASTER_API_KEY", ""),
		BillitResellerRegisterURL:  envOr("BILLIT_RESELLER_REGISTER_URL", "https://my.billit.be/account/PetsFollow/Register"),
		BillitWebhookSecret:        envOr("BILLIT_WEBHOOK_SECRET", ""),
		BillitDefaultDocsIncluded:  envInt("BILLIT_DEFAULT_DOCS_INCLUDED", 50),
		BillitSecretsBackend:       envOr("BILLIT_SECRETS_BACKEND", "plain_dev"),
		BillitSecretsKey:           envOr("BILLIT_SECRETS_KEY", ""),
		InvoicingSaasEnabled:       envBool("INVOICING_SAAS_ENABLED"),
		InvoicingSaasPriceEURCents: envInt("INVOICING_SAAS_PRICE_EUR_CENTS", 8800),
		InvoicingSaasAllowlist:     envCSV("INVOICING_SAAS_ALLOWLIST"),
		SaasInvoicesSecret:         envOr("SAAS_INVOICES_SECRET", ""),
		InvoicingReconcileSecret:   envOr("INVOICING_RECONCILE_SECRET", ""),
		SeedNotifyStaff:            envBool("SEED_NOTIFY_STAFF"),
		AdminStagingSeedEnabled:    envBool("ADMIN_STAGING_SEED_ENABLED"),
	}
}

// ValidateResearch refuses RESEARCH_ENABLED without salt + ETL secret outside seedable envs.
func (c Config) ValidateResearch() error {
	if !c.ResearchEnabled {
		return nil
	}
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	switch env {
	case "local", "development", "dev", "test":
		return nil
	}
	if c.DevSeedEnabled {
		return nil
	}
	if strings.TrimSpace(c.ResearchAnonSalt) == "" {
		return errors.New("RESEARCH_ANON_SALT required when RESEARCH_ENABLED outside local/dev/test")
	}
	if strings.TrimSpace(c.ResearchEtlSecret) == "" {
		return errors.New("RESEARCH_ETL_SECRET required when RESEARCH_ENABLED outside local/dev/test")
	}
	return nil
}

// ValidateBillit refuses unsafe Billit configs outside DEV_SEED (prod/staging).
func (c Config) ValidateBillit() error {
	if !c.BillitEnabled {
		return nil
	}
	if !c.BillitMockEnabled {
		if strings.TrimSpace(c.BillitBaseURL) == "" {
			return errors.New("BILLIT_BASE_URL required when BILLIT_MOCK_ENABLED=false")
		}
		if strings.TrimSpace(c.BillitWebhookSecret) == "" {
			return errors.New("BILLIT_WEBHOOK_SECRET required when BILLIT_MOCK_ENABLED=false")
		}
	}
	if c.DevSeedEnabled {
		return nil
	}
	if c.BillitSecretsBackend == "" || c.BillitSecretsBackend == "plain_dev" {
		return errors.New("BILLIT_SECRETS_BACKEND=plain_dev refused outside DEV_SEED_ENABLED (use local_enc + BILLIT_SECRETS_KEY)")
	}
	if c.BillitSecretsBackend == "local_enc" && strings.TrimSpace(c.BillitSecretsKey) == "" {
		return errors.New("BILLIT_SECRETS_KEY required when BILLIT_SECRETS_BACKEND=local_enc")
	}
	return nil
}

// ValidateVamreg refuses live VAMReg without a gateway URL and API key (never silent dry-run).
// Also rejects using the AFMPS readonly lists base URL as a declaration gateway.
func (c Config) ValidateVamreg() error {
	if c.VamregDryRun {
		return nil
	}
	if strings.TrimSpace(c.VamregBaseURL) == "" {
		return errors.New("VAMREG_BASE_URL required when VAMREG_DRY_RUN=false")
	}
	if strings.TrimSpace(c.VamregAPIKey) == "" {
		return errors.New("VAMREG_API_KEY required when VAMREG_DRY_RUN=false")
	}
	if vamregBaseURLLooksLikeAFMPSReadonlyLists(c.VamregBaseURL) {
		return errors.New("VAMREG_BASE_URL must not be the AFMPS readonly /vamreg/api host when VAMREG_DRY_RUN=false (use VAMREG_AFMPS_* for lists)")
	}
	return nil
}

// ValidateSMS refuses live SMS without Telnyx credentials outside seedable envs
// (never silent dry-run) and requires the reminder-job secret in live mode.
func (c Config) ValidateSMS() error {
	if !c.SMSEnabled || c.SMSDryRun {
		return nil
	}
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	switch env {
	case "local", "development", "dev", "test":
		return nil
	}
	if c.DevSeedEnabled {
		return nil
	}
	if strings.TrimSpace(c.TelnyxAPIKey) == "" {
		return errors.New("TELNYX_API_KEY required when SMS_DRY_RUN=false")
	}
	if strings.TrimSpace(c.TelnyxMessagingProfileID) == "" && strings.TrimSpace(c.TelnyxFrom) == "" {
		return errors.New("TELNYX_MESSAGING_PROFILE_ID or TELNYX_FROM required when SMS_DRY_RUN=false")
	}
	if strings.TrimSpace(c.VisitRemindersSecret) == "" {
		return errors.New("VISIT_REMINDERS_SECRET required when SMS_DRY_RUN=false")
	}
	// Sans clé publique, les webhooks sont tous rejetés : un STOP entrant serait
	// perdu, donc l'opt-out légal ne serait pas honoré. Exigée en envoi réel.
	if strings.TrimSpace(c.TelnyxPublicKey) == "" {
		return errors.New("TELNYX_PUBLIC_KEY required when SMS_DRY_RUN=false (inbound STOP must be honoured)")
	}
	return nil
}

// vamregBaseURLLooksLikeAFMPSReadonlyLists detects the FAMHP software-house lists host
// (ICD readonly) so it cannot be reused as the provisional declaration POST base.
func vamregBaseURLLooksLikeAFMPSReadonlyLists(base string) bool {
	u := strings.ToLower(strings.TrimSpace(base))
	u = strings.TrimRight(u, "/")
	if strings.Contains(u, "fagg-afmps.be") && strings.Contains(u, "/vamreg/api") {
		return true
	}
	// Bare path coincidence with default lists base.
	if strings.HasSuffix(u, "/vamreg/api") {
		return true
	}
	return false
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envBool(k string) bool {
	return os.Getenv(k) == "true" || os.Getenv(k) == "1"
}

func envBoolDefault(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v == "true" || v == "1"
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func envCSV(k string) []string {
	raw := strings.TrimSpace(os.Getenv(k))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
