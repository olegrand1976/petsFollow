package config

import (
	"os"
	"strconv"
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
	HeartRateMinBPM                int
	HeartRateMaxBPM                int
	HeartRateSeconds               int
	StripeSecretKey                string
	StripeWebhookSecret            string
	StripePriceAnnualOnetime       string
	StripePriceTriennialOnetime    string
	StripePriceQuinquennialOnetime string
	StripePriceAnnualSub           string
	StripePriceTriennialSub        string
	StripePriceQuinquennialSub     string
	StripePriceAddonFamily         string
	StripePriceAddonKennel         string
	StripePriceAddonCarePlus       string
	StripePriceAddonHorse          string
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
}

func Load() Config {
	return Config{
		HTTPAddr:                       envOr("HTTP_ADDR", ":8080"),
		DatabaseURL:                    envOr("DATABASE_URL", "postgres://petsfollow:petsfollow@localhost:5437/petsfollow?sslmode=disable"),
		RedisAddr:                      envOr("REDIS_ADDR", "localhost:6382"),
		RedisKeyPrefix:                 envOr("REDIS_KEY_PREFIX", "petsfollow:"),
		JWTSigningKey:                  envOr("JWT_SIGNING_KEY", "dev-change-me"),
		JWTAccessTTL:                   envDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:                  envDuration("JWT_REFRESH_TTL", 7*24*time.Hour),
		LogLevel:                       envOr("LOG_LEVEL", "info"),
		MigrateOnBoot:                  envBool("MIGRATE_ON_BOOT"),
		DevSeedEnabled:                 envBool("DEV_SEED_ENABLED"),
		SMTPHost:                       envOr("SMTP_HOST", "localhost"),
		SMTPPort:                       envInt("SMTP_PORT", 1027),
		SMTPFrom:                       envOr("SMTP_FROM", "petsFollow <noreply@petsfollow.test>"),
		HeartRateMinBPM:                envInt("HEARTRATE_MIN_BPM", 60),
		HeartRateMaxBPM:                envInt("HEARTRATE_MAX_BPM", 140),
		HeartRateSeconds:               envInt("HEARTRATE_DURATION_SEC", 60),
		StripeSecretKey:                envOr("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret:            envOr("STRIPE_WEBHOOK_SECRET", "whsec_test"),
		StripePriceAnnualOnetime:       envOr("STRIPE_PRICE_ANNUAL_ONETIME", ""),
		StripePriceTriennialOnetime:    envOr("STRIPE_PRICE_TRIENNIAL_ONETIME", ""),
		StripePriceQuinquennialOnetime: envOr("STRIPE_PRICE_QUINQUENNIAL_ONETIME", ""),
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
		BillingMockEnabled:             envBool("BILLING_MOCK_ENABLED") || envOr("STRIPE_SECRET_KEY", "") == "",
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
	}
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
