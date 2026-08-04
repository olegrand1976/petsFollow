// Package sms envoie les SMS transactionnels client via Telnyx.
// Même philosophie que fcm : constructeur à dégradation gracieuse (l'API
// démarre toujours), no-op quand le module est désactivé.
package sms

import (
	"context"
	"log"

	"github.com/olegrand1976/petsFollow/go/internal/platform/config"
)

// Result décrit l'issue d'un envoi ; DryRun distingue le log local d'un envoi réel.
type Result struct {
	ProviderMessageID string
	DryRun            bool
}

type Sender interface {
	// Send envoie un SMS à un numéro E.164. Le texte est déjà localisé.
	Send(ctx context.Context, to, text string) (Result, error)
}

// NopSender absorbe les envois quand le module SMS est désactivé ou mal configuré.
// DryRun=true : aucun message n'a quitté le système, le journal doit le refléter
// (jamais un statut « sent » pour un envoi qui n'a pas eu lieu).
type NopSender struct{}

func (NopSender) Send(context.Context, string, string) (Result, error) {
	return Result{DryRun: true}, nil
}

// NewFromConfig retourne le sender adapté à la config. Live sans clé API →
// no-op loggé plutôt qu'un boot en échec (ValidateSMS bloque déjà ce cas
// hors environnements dev).
func NewFromConfig(cfg config.Config) Sender {
	if !cfg.SMSEnabled {
		log.Printf("sms: disabled (SMS_ENABLED=false)")
		return NopSender{}
	}
	if !cfg.SMSDryRun && cfg.TelnyxAPIKey == "" {
		log.Printf("sms: live mode without TELNYX_API_KEY — falling back to no-op")
		return NopSender{}
	}
	if cfg.SMSDryRun {
		log.Printf("sms: dry-run mode (SMS_DRY_RUN=true) — messages logged, not sent")
	}
	return &TelnyxClient{
		APIKey:             cfg.TelnyxAPIKey,
		MessagingProfileID: cfg.TelnyxMessagingProfileID,
		From:               cfg.TelnyxFrom,
		DryRun:             cfg.SMSDryRun,
	}
}
