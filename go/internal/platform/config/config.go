package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr                       string
	DatabaseURL                    string
	RedisAddr                      string
	RedisKeyPrefix                 string
	JWTSigningKey                  string
	JWTAccessTTL                   time.Duration
	JWTRefreshTTL                  time.Duration
	LogLevel                       string
	MigrateOnBoot                  bool
	DevSeedEnabled                 bool
	SMTPHost                       string
	SMTPPort                       int
	SMTPFrom                       string
	// SMTPUser / SMTPPass — auth PLAIN (OVH :587). Vides = MailHog / relay ouvert.
	SMTPUser string
	SMTPPass string
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
	// CORSAllowedOrigins — origines autorisées (séparées par des virgules). Défaut : site Pro public.
	CORSAllowedOrigins string
	// RetentionPurgeSecret protège POST /internal/retention/run (purge 3 ans d'inactivité RGPD).
	RetentionPurgeSecret string
	// SalesBranchesAutoSecret protège POST /internal/sales-branches-auto/run.
	SalesBranchesAutoSecret string
	// AiModuleFrictionSecret protège POST /internal/ai-module-friction/run.
	AiModuleFrictionSecret string
	// OpsNotifyEmail reçoit les leads « nouveau véto » + alertes auth ALERT/URGENT.
	OpsNotifyEmail string
	// CommercialContactPhone — fallback téléphone commercial (mail/PDF dossier) si profil vide.
	CommercialContactPhone string
	// SupportInboxEmail reçoit les nouveaux tickets bug-report (défaut support@petsfollow.app).
	SupportInboxEmail string
	// AuthHealthSecret protège POST /internal/auth-health/run.
	AuthHealthSecret string
	// MLMOrgEnabled exposes multi-depth downline UI; commissions remain flat until MLM billing ships.
	MLMOrgEnabled bool
	// PharmacyEnabled enables vet pharmacy module (CNK dictionary, stock, DAF) — default off.
	PharmacyEnabled bool
	// PharmacyExpirySecret protège POST /internal/pharmacy/expiry-run.
	PharmacyExpirySecret string
	// PrescriptionsEnabled enables veterinary prescription drafts + PDF preview — default off.
	PrescriptionsEnabled bool

	// BillitEnabled exposes invoicing routes (Billit reseller / Peppol).
	BillitEnabled bool
	// BillitMockEnabled uses the mock gateway (local/CI) — never call Billit live.
	BillitMockEnabled bool
	BillitBaseURL             string
	BillitMasterPartyID       string
	BillitMasterAPIKey        string
	BillitResellerRegisterURL string
	BillitWebhookSecret       string
	BillitDefaultDocsIncluded int
	BillitSecretsBackend      string // local_enc | plain_dev
	BillitSecretsKey          string
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
		BillingMockEnabled:             envBool("BILLING_MOCK_ENABLED"),
		GoogleOAuthClientID:            envOr("GOOGLE_OAUTH_CLIENT_ID", ""),
		GCSMediaBucket:                 envOr("GCS_MEDIA_BUCKET", ""),
		MediaLocalDir:                  envOr("MEDIA_LOCAL_DIR", "./data/uploads"),
		LLITWebsiteURL:                 envOr("LLIT_WEBSITE_URL", "https://ll-it-sc.be"),
		PetsAppDownloadURL:             envOr("PETS_APP_DOWNLOAD_URL", "https://appdistribution.firebase.google.com/testerapps/1:237481297060:android:cfda5c59a08bfd6dc9d231"),
		// FCM enabled by default; ADC (GOOGLE_APPLICATION_CREDENTIALS / Cloud Run SA) required to actually send.
		FCMEnabled:           envBoolDefault("FCM_ENABLED", true),
		JourneyEmailEnabled:  envBoolDefault("JOURNEY_EMAIL_ENABLED", true),
		JourneyEmailInterval: envDuration("JOURNEY_EMAIL_INTERVAL", time.Hour),
		GeminiAPIKey:         envOr("GEMINI_API_KEY", ""),
		GeminiModel:          envOr("GEMINI_MODEL", "gemini-3.6-flash"),
		GeminiLiteModel:      envOr("GEMINI_LITE_MODEL", "gemini-3.5-flash-lite"),
		GeminiLiveModel:      envOr("GEMINI_LIVE_MODEL", "gemini-2.5-flash-native-audio-preview-09-2025"),
		PitchAnalyzerSecret:  envOr("PITCH_ANALYZER_SECRET", ""),
		ProductDigestSecret:  envOr("PRODUCT_DIGEST_SECRET", ""),
		VertexProject:        envOr("VERTEX_PROJECT", ""),
		VertexLocation:       envOr("VERTEX_LOCATION", "europe-west9"),
		CareProPublicRegister: envBool("CARE_PRO_PUBLIC_REGISTER"),
		AuthRateLimitPerMin:   envInt("AUTH_RATE_LIMIT_PER_MIN", 60),
		CORSAllowedOrigins:    envOr("CORS_ALLOWED_ORIGINS", ""),
		RetentionPurgeSecret:      envOr("RETENTION_PURGE_SECRET", ""),
		SalesBranchesAutoSecret:   envOr("SALES_BRANCHES_AUTO_SECRET", ""),
		AiModuleFrictionSecret:    envOr("AI_MODULE_FRICTION_SECRET", ""),
		OpsNotifyEmail:         envOr("OPS_NOTIFY_EMAIL", ""),
		CommercialContactPhone: envOr("COMMERCIAL_CONTACT_PHONE", ""),
		SupportInboxEmail:      envOr("SUPPORT_INBOX_EMAIL", "support@petsfollow.app"),
		AuthHealthSecret:       envOr("AUTH_HEALTH_SECRET", ""),
		MLMOrgEnabled:           envBool("MLM_ORG_ENABLED"),
		PharmacyEnabled:         envBool("PHARMACY_ENABLED"),
		PharmacyExpirySecret:    envOr("PHARMACY_EXPIRY_SECRET", ""),
		PrescriptionsEnabled:    envBool("PRESCRIPTIONS_ENABLED"),

		// Billit : off par défaut ; mock uniquement opt-in (comme BILLING_MOCK_ENABLED).
		BillitEnabled:           envBool("BILLIT_ENABLED"),
		BillitMockEnabled:       envBool("BILLIT_MOCK_ENABLED"),
		BillitBaseURL:           envOr("BILLIT_BASE_URL", "https://api.billit.be"),
		BillitMasterPartyID:     envOr("BILLIT_MASTER_PARTY_ID", ""),
		BillitMasterAPIKey:      envOr("BILLIT_MASTER_API_KEY", ""),
		BillitResellerRegisterURL: envOr("BILLIT_RESELLER_REGISTER_URL", "https://my.billit.be/account/PetsFollow/Register"),
		BillitWebhookSecret:       envOr("BILLIT_WEBHOOK_SECRET", ""),
		BillitDefaultDocsIncluded: envInt("BILLIT_DEFAULT_DOCS_INCLUDED", 50),
		BillitSecretsBackend:      envOr("BILLIT_SECRETS_BACKEND", "plain_dev"),
		BillitSecretsKey:          envOr("BILLIT_SECRETS_KEY", ""),
		SeedNotifyStaff:         envBool("SEED_NOTIFY_STAFF"),
		AdminStagingSeedEnabled: envBool("ADMIN_STAGING_SEED_ENABLED"),
	}
}

// ValidateBillit refuses unsafe Billit configs outside DEV_SEED (prod/staging).
func (c Config) ValidateBillit() error {
	if !c.BillitEnabled {
		return nil
	}
	if !c.BillitMockEnabled {
		return errors.New("BILLIT_ENABLED without BILLIT_MOCK_ENABLED requires a live Billit client (not implemented yet — keep BILLIT_MOCK_ENABLED=true or disable BILLIT_ENABLED)")
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
