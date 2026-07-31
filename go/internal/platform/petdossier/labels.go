package petdossier

import "strings"

type pdfLabels struct {
	Subtitle         string
	CTA              string
	Register         string
	Offer            string
	ContactTitle     string
	DefaultContact   string
	Phone            string
	Email            string
	Animal           string
	Name             string
	Species          string
	Breed            string
	Birth            string
	Weight           string
	Chip             string
	HealthBook       string
	Owner            string
	OwnerEmail       string
	RecentWeights    string
	HeartRate        string
	Alert            string
	CareReminders    string
	Visits           string
	ProConsulted     string
	Timeline         string
	Attachments      string
	HealthBookIncl   string
	NoAttachments    string
	FooterGenerated  string
	FooterCreateAcct string
	FooterCommercial string
}

func labelsFor(locale string) pdfLabels {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "en":
		return pdfLabels{
			Subtitle:         "Animal dossier — temporary share (24h)",
			CTA:              "Animal care professional: join petsFollow to follow your patients continuously.",
			Register:         "Sign up: ",
			Offer:            "Plans: ",
			ContactTitle:     "Your petsFollow contact",
			DefaultContact:   "petsFollow sales team",
			Phone:            "Phone: ",
			Email:            "Email: ",
			Animal:           "Animal",
			Name:             "Name",
			Species:          "Species",
			Breed:            "Breed",
			Birth:            "Birth date",
			Weight:           "Weight",
			Chip:             "Microchip",
			HealthBook:       "Health book",
			Owner:            "Owner",
			OwnerEmail:       "Owner email",
			RecentWeights:    "Recent weights",
			HeartRate:        "Heart rate",
			Alert:            " [alert]",
			CareReminders:    "Care reminders",
			Visits:           "Visits (practitioner)",
			ProConsulted:     "Practitioner: ",
			Timeline:         "Timeline",
			Attachments:      "Attachments",
			HealthBookIncl:   "• health-book.pdf (included in the pack)",
			NoAttachments:    "No attachments.",
			FooterGenerated:  "Generated on %s UTC — petsFollow — veterinary care continuity.",
			FooterCreateAcct: "Create a pro account: ",
			FooterCommercial: "Sales contact: ",
		}
	case "nl":
		return pdfLabels{
			Subtitle:         "Dossier dier — tijdelijke share (24u)",
			CTA:              "Dierenprofessional: sluit je aan bij petsFollow om je patiënten continu op te volgen.",
			Register:         "Registratie: ",
			Offer:            "Aanbod: ",
			ContactTitle:     "Uw petsFollow-contact",
			DefaultContact:   "petsFollow verkoopteam",
			Phone:            "Tel.: ",
			Email:            "E-mail: ",
			Animal:           "Dier",
			Name:             "Naam",
			Species:          "Soort",
			Breed:            "Ras",
			Birth:            "Geboorte",
			Weight:           "Gewicht",
			Chip:             "Chip",
			HealthBook:       "Gezondheidsboekje",
			Owner:            "Eigenaar",
			OwnerEmail:       "E-mail eigenaar",
			RecentWeights:    "Recente gewichten",
			HeartRate:        "Hartslag",
			Alert:            " [alert]",
			CareReminders:    "Zorgherinneringen",
			Visits:           "Bezoeken (raadpleger)",
			ProConsulted:     "Raadpleger: ",
			Timeline:         "Tijdlijn",
			Attachments:      "Bijlagen",
			HealthBookIncl:   "• gezondheidsboekje.pdf (inbegrepen in het pack)",
			NoAttachments:    "Geen bijlagen.",
			FooterGenerated:  "Gegenereerd op %s UTC — petsFollow — continuïteit van dierenverzorging.",
			FooterCreateAcct: "Pro-account aanmaken: ",
			FooterCommercial: "Commercieel contact: ",
		}
	case "es":
		return pdfLabels{
			Subtitle:         "Expediente animal — compartición temporal (24 h)",
			CTA:              "Profesional animal: únase a petsFollow para seguir a sus pacientes de forma continua.",
			Register:         "Registro: ",
			Offer:            "Oferta: ",
			ContactTitle:     "Su contacto petsFollow",
			DefaultContact:   "Equipo comercial petsFollow",
			Phone:            "Tel.: ",
			Email:            "Email: ",
			Animal:           "Animal",
			Name:             "Nombre",
			Species:          "Especie",
			Breed:            "Raza",
			Birth:            "Nacimiento",
			Weight:           "Peso",
			Chip:             "Chip",
			HealthBook:       "Cartilla",
			Owner:            "Propietario",
			OwnerEmail:       "Email del propietario",
			RecentWeights:    "Pesos recientes",
			HeartRate:        "Frecuencia cardíaca",
			Alert:            " [alerta]",
			CareReminders:    "Recordatorios de cuidados",
			Visits:           "Visitas (profesional consultado)",
			ProConsulted:     "Profesional: ",
			Timeline:         "Cronología",
			Attachments:      "Adjuntos",
			HealthBookIncl:   "• cartilla-salud.pdf (incluido en el pack)",
			NoAttachments:    "Sin adjuntos.",
			FooterGenerated:  "Generado el %s UTC — petsFollow — continuidad de cuidados veterinarios.",
			FooterCreateAcct: "Crear cuenta pro: ",
			FooterCommercial: "Contacto comercial: ",
		}
	case "et":
		return pdfLabels{
			Subtitle:         "Looma toimik — ajutine jagamine (24 h)",
			CTA:              "Loomaprofessionaal: liitu petsFollow’iga, et jälgida patsiente järjepidevalt.",
			Register:         "Registreerimine: ",
			Offer:            "Pakkumine: ",
			ContactTitle:     "Teie petsFollow kontakt",
			DefaultContact:   "petsFollow müügimeeskond",
			Phone:            "Tel: ",
			Email:            "E-post: ",
			Animal:           "Loom",
			Name:             "Nimi",
			Species:          "Liik",
			Breed:            "Tõug",
			Birth:            "Sünd",
			Weight:           "Kaal",
			Chip:             "Kiip",
			HealthBook:       "Tervisekaart",
			Owner:            "Omanik",
			OwnerEmail:       "Omaniku e-post",
			RecentWeights:    "Viimased kaalud",
			HeartRate:        "Südame löögisagedus",
			Alert:            " [häire]",
			CareReminders:    "Hoolduse meeldetuletused",
			Visits:           "Visiidid (konsulteeritud pro)",
			ProConsulted:     "Pro: ",
			Timeline:         "Ajajoon",
			Attachments:      "Manused",
			HealthBookIncl:   "• tervisekaart.pdf (paketis kaasas)",
			NoAttachments:    "Manuseid pole.",
			FooterGenerated:  "Loodud %s UTC — petsFollow — veterinaarse hoolduse järjepidevus.",
			FooterCreateAcct: "Loo pro konto: ",
			FooterCommercial: "Müügikontakt: ",
		}
	case "it":
		return pdfLabels{
			Subtitle:         "Cartella animale — condivisione temporanea (24h)",
			CTA:              "Professionista animale: unisciti a petsFollow per seguire i pazienti in continuo.",
			Register:         "Iscrizione: ",
			Offer:            "Offerta: ",
			ContactTitle:     "Il tuo contatto petsFollow",
			DefaultContact:   "Team commerciale petsFollow",
			Phone:            "Tel.: ",
			Email:            "Email: ",
			Animal:           "Animale",
			Name:             "Nome",
			Species:          "Specie",
			Breed:            "Razza",
			Birth:            "Nascita",
			Weight:           "Peso",
			Chip:             "Microchip",
			HealthBook:       "Libretto",
			Owner:            "Proprietario",
			OwnerEmail:       "Email proprietario",
			RecentWeights:    "Pesi recenti",
			HeartRate:        "Frequenza cardiaca",
			Alert:            " [allerta]",
			CareReminders:    "Promemoria cure",
			Visits:           "Visite (professionista consultato)",
			ProConsulted:     "Professionista: ",
			Timeline:         "Timeline",
			Attachments:      "Allegati",
			HealthBookIncl:   "• libretto-salute.pdf (incluso nel pack)",
			NoAttachments:    "Nessun allegato.",
			FooterGenerated:  "Generato il %s UTC — petsFollow — continuità delle cure veterinarie.",
			FooterCreateAcct: "Crea un account pro: ",
			FooterCommercial: "Contatto commerciale: ",
		}
	default: // fr
		return pdfLabels{
			Subtitle:         "Dossier animal — partage temporaire (24h)",
			CTA:              "Professionnel animalier : rejoignez petsFollow pour suivre vos patients en continu.",
			Register:         "Inscription : ",
			Offer:            "Offre : ",
			ContactTitle:     "Votre contact petsFollow",
			DefaultContact:   "Équipe commerciale petsFollow",
			Phone:            "Tél. : ",
			Email:            "Email : ",
			Animal:           "Animal",
			Name:             "Nom",
			Species:          "Espèce",
			Breed:            "Race",
			Birth:            "Naissance",
			Weight:           "Poids",
			Chip:             "Puce",
			HealthBook:       "Carnet",
			Owner:            "Propriétaire",
			OwnerEmail:       "Email propriétaire",
			RecentWeights:    "Poids récents",
			HeartRate:        "Fréquence cardiaque",
			Alert:            " [alerte]",
			CareReminders:    "Rappels de soins",
			Visits:           "Visites (pro consulté)",
			ProConsulted:     "Pro consulté : ",
			Timeline:         "Timeline",
			Attachments:      "Pièces jointes",
			HealthBookIncl:   "• carnet-sante.pdf (inclus dans le pack)",
			NoAttachments:    "Aucune pièce jointe.",
			FooterGenerated:  "Généré le %s UTC — petsFollow — continuité de soins vétérinaire.",
			FooterCreateAcct: "Créer un compte pro : ",
			FooterCommercial: "Contact commercial : ",
		}
	}
}
