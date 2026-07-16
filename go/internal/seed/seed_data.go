package seed

import (
	"time"

	"github.com/olegrand1976/petsFollow/go/internal/billing"
	"github.com/olegrand1976/petsFollow/go/pkg/kernel"
)

const (
	passwordVet    = "VetDemo123!"
	passwordClient = "ClientDemo123!"
	passwordAdmin  = "AdminDemo123!"

	// Token fixe pour /confirm-email?token=demo-confirm-email (compte vet.unverified@)
	demoEmailConfirmToken = "demo-confirm-email"
	// Token fixe pour /reset-password?token=demo-reset-password (compte vet.reset@)
	demoPasswordResetToken = "demo-reset-password"
)

type messageDef struct {
	senderRole string // "client" | "vet"
	body       string
	age        time.Duration // negative = in the past
	read       bool
}

type heartRateDef struct {
	status   kernel.SessionStatus
	tapCount int
	duration int
	bpm      int
	isAlert  bool
	age      time.Duration
}

type dossierEventDef struct {
	authorRole string // "vet" | "client"
	eventType  string
	content    string
	age        time.Duration
}

type careReminderDef struct {
	reminderType string
	title        string
	dueDays      int // days from now (negative = past)
	status       string
}

type visitDef struct {
	status      string
	notes       string
	source      string
	scheduledIn time.Duration // from now
}

type petDef struct {
	name          string
	species       string
	breed         string
	weightKg      float64
	paymentStatus string
	plan          billing.PlanCode
	billingMode   billing.BillingMode
	entitlement   billing.EntitlementStatus
	messages      []messageDef
	heartRates    []heartRateDef
	dossierEvents []dossierEventDef
	careReminders []careReminderDef
	visits        []visitDef
}

type clientDef struct {
	email              string
	fullName           string
	pets               []petDef
	seedDiscovery      bool
	extraPracticeVet   string // vet email for secondary practice link
}

type practiceDef struct {
	name              string
	vetEmail          string
	vetName           string
	phone             string
	address           string
	addressLine2      string
	city              string
	postalCode        string
	website           string
	availability      kernel.AvailabilityStatus
	autoReply         string
	clients           []clientDef
	notifyOnMessage   bool
	notifyOnHeartRate bool
	incompleteProfile  bool // profil cabinet non complété (onboarding)
	pendingEmailVerify bool // email non confirmé + token démo
	seedPasswordReset  bool // token démo reset password
}

// Comptes démo — mots de passe : VetDemo123! · ClientDemo123! · AdminDemo123!
//
// Vétos : vet.demo@ · vet.parc@ · vet.lyon@ · vet.onboarding@ · vet.unverified@ · vet.reset@ (reset MDP)
// Admin : admin.demo@
// Clients : client.demo@ · client.marie@ · client.paul@ · client.julie@ · client.thomas@ · client.vide@ (sans animal)

var demoPractices = []practiceDef{
	{
		name:              "Cabinet VetPlus Demo",
		vetEmail:          "vet.demo@petsfollow.test",
		vetName:           "Dr Martin Demo",
		phone:             "01 23 45 67 89",
		address:           "12 avenue des Vétérinaires",
		addressLine2:      "Bâtiment B",
		city:              "Paris",
		postalCode:        "75015",
		website:           "https://vetplus.demo.petsfollow.test",
		availability:      kernel.AvailabilityAvailable,
		notifyOnMessage:   true,
		notifyOnHeartRate: true,
		clients: []clientDef{
			{
				email:            "client.demo@petsfollow.test",
				fullName:         "Sophie Demo",
				seedDiscovery:    true,
				extraPracticeVet: "vet.parc@petsfollow.test",
				pets: []petDef{
					{
						name:          "Rex",
						species:       "dog",
						breed:         "Labrador",
						weightKg:      32.5,
						paymentStatus: "active",
						plan:          billing.PlanTriennial,
						billingMode:   billing.ModeSubscription,
						entitlement:   billing.StatusActive,
						messages: []messageDef{
							{senderRole: "client", body: "Bonjour docteur, Rex tousse un peu depuis hier soir.", age: -72 * time.Hour, read: true},
							{senderRole: "vet", body: "Bonjour Sophie. Pas de fièvre ni de fatigue ? Je peux vous recevoir demain matin.", age: -70 * time.Hour, read: true},
							{senderRole: "client", body: "Non, il mange normalement. Demain 10h convient.", age: -68 * time.Hour, read: true},
							{senderRole: "vet", body: "Parfait, rendez-vous confirmé. Continuez le suivi cardiaque en attendant.", age: -67 * time.Hour, read: false},
						},
						heartRates: []heartRateDef{
							{status: kernel.SessionValidated, tapCount: 72, duration: 60, bpm: 72, age: -7 * 24 * time.Hour},
							{status: kernel.SessionValidated, tapCount: 68, duration: 60, bpm: 68, age: -3 * 24 * time.Hour},
							{status: kernel.SessionPendingValidation, tapCount: 74, duration: 60, bpm: 74, age: -2 * time.Hour},
						},
						dossierEvents: []dossierEventDef{
							{authorRole: "vet", eventType: "note", content: "Suivi cardiaque post-op. Fréquence stable.", age: -14 * 24 * time.Hour},
						},
						careReminders: []careReminderDef{
							{reminderType: "vaccination", title: "Vaccination", dueDays: 120, status: "pending"},
							{reminderType: "deworming", title: "Vermifuge", dueDays: 14, status: "pending"},
							{reminderType: "vet_check", title: "Contrôle vétérinaire", dueDays: -30, status: "done"},
							{reminderType: "dental", title: "Soins dentaires", dueDays: 200, status: "pending"},
						},
						visits: []visitDef{
							{status: "confirmed", notes: "Contrôle post-op cardiaque", source: "vet", scheduledIn: 7 * 24 * time.Hour},
						},
					},
					{
						name:          "Bella",
						species:       "cat",
						breed:         "Européen",
						weightKg:      4.2,
						paymentStatus: "active",
						plan:          billing.PlanAnnual,
						billingMode:   billing.ModeOneTime,
						entitlement:   billing.StatusActive,
						messages: []messageDef{
							{senderRole: "client", body: "Bella a fait son relevé ce matin, tout semble normal.", age: -5 * time.Hour, read: false},
						},
						heartRates: []heartRateDef{
							{status: kernel.SessionValidated, tapCount: 120, duration: 60, bpm: 120, age: -5 * time.Hour},
						},
					},
					{
						name:          "Spirit",
						species:       "horse",
						breed:         "Pur-sang",
						weightKg:      480,
						paymentStatus: "active",
						plan:          billing.PlanAnnual,
						billingMode:   billing.ModeOneTime,
						entitlement:   billing.StatusActive,
						careReminders: []careReminderDef{
							{reminderType: "farrier", title: "Maréchal-ferrant", dueDays: 21, status: "pending"},
							{reminderType: "fecal_egg", title: "Coproscopie", dueDays: 60, status: "pending"},
						},
					},
				},
			},
			{
				email:    "client.vide@petsfollow.test",
				fullName: "Luc Moreau",
				pets:     nil,
			},
		},
	},
	{
		name:              "Clinique du Parc",
		vetEmail:          "vet.parc@petsfollow.test",
		vetName:           "Dr Claire Parc",
		phone:             "01 45 67 89 01",
		address:           "8 rue du Parc",
		addressLine2:      "1er étage",
		city:              "Boulogne-Billancourt",
		postalCode:        "92100",
		website:           "https://clinique-parc.demo.petsfollow.test",
		availability:      kernel.AvailabilityAvailable,
		notifyOnMessage:   true,
		notifyOnHeartRate: true,
		clients: []clientDef{
			{
				email:    "client.marie@petsfollow.test",
				fullName: "Marie Leclerc",
				pets: []petDef{
					{
						name:          "Mimi",
						species:       "cat",
						breed:         "Européen",
						weightKg:      3.8,
						paymentStatus: "active",
						plan:          billing.PlanTriennial,
						billingMode:   billing.ModeSubscription,
						entitlement:   billing.StatusActive,
						messages: []messageDef{
							{senderRole: "client", body: "Mimi a moins d'appétit depuis 2 jours.", age: -48 * time.Hour, read: true},
							{senderRole: "vet", body: "Merci pour l'info. Le dernier relevé cardiaque est dans la norme. Surveillez l'hydratation.", age: -46 * time.Hour, read: true},
						},
						heartRates: []heartRateDef{
							{status: kernel.SessionValidated, tapCount: 110, duration: 60, bpm: 110, age: -10 * 24 * time.Hour},
							{status: kernel.SessionValidated, tapCount: 115, duration: 60, bpm: 115, age: -2 * 24 * time.Hour},
						},
					},
					{
						name:          "Chouchou",
						species:       "cat",
						breed:         "Persan",
						weightKg:      5.1,
						paymentStatus: "active",
						plan:          billing.PlanQuinquennial,
						billingMode:   billing.ModeOneTime,
						entitlement:   billing.StatusActive,
						messages: []messageDef{
							{senderRole: "client", body: "Alerte sur le dernier relevé de Chouchou, pouvez-vous regarder ?", age: -3 * time.Hour, read: false},
						},
						heartRates: []heartRateDef{
							{status: kernel.SessionValidated, tapCount: 180, duration: 60, bpm: 180, isAlert: true, age: -3 * time.Hour},
							{status: kernel.SessionValidated, tapCount: 130, duration: 60, bpm: 130, age: -5 * 24 * time.Hour},
						},
						dossierEvents: []dossierEventDef{
							{authorRole: "vet", eventType: "alert", content: "Tachycardie détectée — contrôle clinique recommandé sous 48h.", age: -3 * time.Hour},
						},
					},
				},
			},
			{
				email:    "client.paul@petsfollow.test",
				fullName: "Paul Bernard",
				pets: []petDef{
					{
						name:          "Max",
						species:       "dog",
						breed:         "Golden Retriever",
						weightKg:      28.0,
						paymentStatus: "active",
						plan:          billing.PlanAnnual,
						billingMode:   billing.ModeSubscription,
						entitlement:   billing.StatusActive,
						messages: []messageDef{
							{senderRole: "client", body: "Max est très actif après la promenade, c'est normal pour le BPM ?", age: -24 * time.Hour, read: true},
							{senderRole: "vet", body: "Oui, attendez 30 min au repos avant un relevé pour une mesure fiable.", age: -23 * time.Hour, read: true},
						},
						heartRates: []heartRateDef{
							{status: kernel.SessionValidated, tapCount: 95, duration: 60, bpm: 95, age: -4 * 24 * time.Hour},
							{status: kernel.SessionValidated, tapCount: 78, duration: 60, bpm: 78, age: -1 * 24 * time.Hour},
						},
					},
				},
			},
		},
	},
	{
		name:              "Centre Cardio Animaux Lyon",
		vetEmail:          "vet.lyon@petsfollow.test",
		vetName:           "Dr Antoine Lyon",
		phone:             "04 78 00 12 34",
		address:           "25 cours Vitton",
		addressLine2:      "",
		city:              "Lyon",
		postalCode:        "69006",
		website:           "https://cardio-lyon.demo.petsfollow.test",
		availability:      kernel.AvailabilityUnavailable,
		autoReply:         "Je suis en consultation. Pour les urgences cardiaques, contactez le service d'astreinte au 04 00 00 00 00.",
		notifyOnMessage:   true,
		notifyOnHeartRate: true,
		clients: []clientDef{
			{
				email:    "client.julie@petsfollow.test",
				fullName: "Julie Martin",
				pets: []petDef{
					{
						name:          "Oscar",
						species:       "dog",
						breed:         "Cavalier King Charles",
						weightKg:      8.5,
						paymentStatus: "active",
						plan:          billing.PlanTriennial,
						billingMode:   billing.ModeSubscription,
						entitlement:   billing.StatusActive,
						messages: []messageDef{
							{senderRole: "client", body: "Oscar est sous traitement cardiaque, je fais les relevés comme convenu.", age: -120 * time.Hour, read: true},
							{senderRole: "vet", body: "Excellent. Je vois une légère hausse sur le relevé de mardi, on en reparle au prochain contrôle.", age: -118 * time.Hour, read: true},
							{senderRole: "client", body: "D'accord, je programme le prochain relevé demain matin.", age: -12 * time.Hour, read: false},
						},
						heartRates: []heartRateDef{
							{status: kernel.SessionValidated, tapCount: 88, duration: 60, bpm: 88, age: -30 * 24 * time.Hour},
							{status: kernel.SessionValidated, tapCount: 102, duration: 60, bpm: 102, isAlert: true, age: -7 * 24 * time.Hour},
							{status: kernel.SessionValidated, tapCount: 92, duration: 60, bpm: 92, age: -2 * 24 * time.Hour},
							{status: kernel.SessionPendingValidation, tapCount: 98, duration: 60, bpm: 98, age: -1 * time.Hour},
						},
						dossierEvents: []dossierEventDef{
							{authorRole: "vet", eventType: "diagnosis", content: "Insuffisance mitrale stade B2 — suivi mensuel.", age: -60 * 24 * time.Hour},
							{authorRole: "vet", eventType: "note", content: "Augmentation modérée du BPM au repos — ajuster si persistance.", age: -7 * 24 * time.Hour},
						},
					},
				},
			},
			{
				email:    "client.thomas@petsfollow.test",
				fullName: "Thomas Durand",
				pets: []petDef{
					{
						name:          "Luna",
						species:       "cat",
						breed:         "Siamois",
						weightKg:      3.5,
						paymentStatus: "active",
						plan:          billing.PlanAnnual,
						billingMode:   billing.ModeOneTime,
						entitlement:   billing.StatusActive,
						messages: []messageDef{
							{senderRole: "client", body: "Premier relevé de Luna effectué, merci pour l'accompagnement.", age: -6 * time.Hour, read: true},
						},
						heartRates: []heartRateDef{
							{status: kernel.SessionValidated, tapCount: 125, duration: 60, bpm: 125, age: -6 * time.Hour},
						},
					},
					{
						name:          "Nico",
						species:       "dog",
						breed:         "Beagle",
						weightKg:      12.0,
						paymentStatus: "pending_payment",
						plan:          billing.PlanTriennial,
						billingMode:   billing.ModeSubscription,
						entitlement:   billing.StatusPending,
						messages: []messageDef{
							{senderRole: "client", body: "Je finalise le paiement pour Nico cette semaine.", age: -2 * 24 * time.Hour, read: true},
							{senderRole: "vet", body: "Pas de souci, le dossier sera activé dès confirmation Stripe.", age: -47 * time.Hour, read: false},
						},
					},
				},
			},
		},
	},
	{
		name:               "Cabinet Onboarding Demo",
		vetEmail:           "vet.onboarding@petsfollow.test",
		vetName:            "Dr Emma Nouveau",
		phone:              "",
		address:            "",
		city:               "",
		postalCode:         "",
		website:            "",
		availability:       kernel.AvailabilityAvailable,
		notifyOnMessage:    true,
		notifyOnHeartRate:  true,
		incompleteProfile:  true,
	},
	{
		name:               "Cabinet Email Pending",
		vetEmail:           "vet.unverified@petsfollow.test",
		vetName:            "Dr Pending Verify",
		phone:              "",
		address:            "",
		city:               "",
		postalCode:         "",
		website:            "",
		availability:       kernel.AvailabilityAvailable,
		notifyOnMessage:    true,
		notifyOnHeartRate:  true,
		incompleteProfile:  true,
		pendingEmailVerify: true,
	},
	{
		name:              "Cabinet Reset Demo",
		vetEmail:          "vet.reset@petsfollow.test",
		vetName:           "Dr Reset Demo",
		phone:             "01 00 00 00 00",
		address:           "1 rue Reset",
		city:              "Lille",
		postalCode:        "59000",
		website:           "",
		availability:      kernel.AvailabilityAvailable,
		notifyOnMessage:   true,
		notifyOnHeartRate: true,
		seedPasswordReset: true,
	},
}
