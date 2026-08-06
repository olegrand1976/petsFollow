package seed

import (
	"context"

	"github.com/olegrand1976/petsFollow/go/internal/store"
)

func seedCommercialEmailTemplates(ctx context.Context, st *store.Store) error {
	// Full seed truncates sales.email_templates first — Upsert refreshes catalog copy.
	// Custom edits in non-truncate environments are overwritten by slug upsert.
	for _, t := range commercialEmailTemplateCatalog() {
		if _, err := st.UpsertEmailTemplateBySlug(ctx, t); err != nil {
			return err
		}
	}
	return nil
}

func commercialEmailTemplateCatalog() []store.EmailTemplateInput {
	active := true
	sig := `<p style="margin:28px 0 0;font-size:15px;line-height:1.55;color:#1B3A4B;">
  Cordialement,<br>
  <strong>{{commercial_name}}</strong><br>
  {{commercial_phone}} · {{commercial_email}}
</p>`
	cta := func(label, href string) string {
		return `<table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin:24px 0;">
  <tr><td style="background-color:#2A9D8F;border-radius:8px;">
    <a href="` + href + `" style="display:inline-block;padding:14px 28px;font-size:15px;font-weight:600;color:#FFFFFF;text-decoration:none;">` + label + `</a>
  </td></tr>
</table>`
	}

	return []store.EmailTemplateInput{
		{
			Slug: "intro_after_request", Name: "Suite à votre demande d’e-mail", Category: "intro", Locale: "fr", IsActive: &active,
			Subject: "{{practice_name}} — continuité de soins en 20 minutes",
			BodyHTML: `<p style="margin:0 0 16px;">Bonjour {{contact_name}},</p>
<p style="margin:0 0 16px;">Comme convenu, voici un bref aperçu de <strong>petsFollow</strong> : la continuité de soins prescrite, matérialisée en passeport digital de l’animal — Web pour le cabinet, mobile terrain et app propriétaire.</p>
<p style="margin:0 0 16px;">Le propriétaire paie le suivi (≤ 3,50 € / mois ; recommandé 95 € / 3 ans). L’idée n’est pas de vous vendre par e-mail : c’est de voir ensemble, en <strong>20 minutes</strong>, si le parcours colle à votre cabinet.</p>
` + cta("Choisir un créneau de 20 min", "{{cta_demo_url}}") + `
<p style="margin:0 0 8px;font-size:14px;color:#6B7280;">Deux options fréquentes : mardi midi ou jeudi en fin de journée — répondez simplement à cet e-mail avec votre préférence.</p>
` + sig,
		},
		{
			Slug: "rdv_confirm", Name: "Confirmation de rendez-vous", Category: "rdv", Locale: "fr", IsActive: &active,
			Subject: "Confirmé — démo petsFollow le {{appointment_at}}",
			BodyHTML: `<p style="margin:0 0 16px;">Bonjour {{contact_name}},</p>
<p style="margin:0 0 16px;">Rendez-vous noté pour <strong>{{practice_name}}</strong> :</p>
<p style="margin:0 0 16px;padding:14px 16px;background:#F7F9FB;border:1px solid #E2E6ED;border-radius:8px;"><strong>{{appointment_at}}</strong></p>
<p style="margin:0 0 16px;">En 20 minutes, on verra le passeport digital, la messagerie et comment vos clients s’activent — sans pitch commercial interminable.</p>
<p style="margin:0 0 16px;">Si le créneau ne convient plus, répondez à cet e-mail : on décale en 10 secondes.</p>
` + sig,
		},
		{
			Slug: "rdv_reminder_j1", Name: "Rappel J−1", Category: "rdv", Locale: "fr", IsActive: &active,
			Subject: "Demain — démo petsFollow {{appointment_at}}",
			BodyHTML: `<p style="margin:0 0 16px;">Bonjour {{contact_name}},</p>
<p style="margin:0 0 16px;">Petit rappel : notre échange de 20 minutes pour <strong>{{practice_name}}</strong> a lieu <strong>{{appointment_at}}</strong>.</p>
<p style="margin:0 0 16px;">Préparez éventuellement une question sur votre suivi actuel entre deux consultations — cela accélère la démo.</p>
` + cta("Ouvrir le lien / brochure", "{{cta_demo_url}}") + sig,
		},
		{
			Slug: "nurture_j1", Name: "Relance douce J+1", Category: "nurture", Locale: "fr", IsActive: &active,
			Subject: "{{contact_name}} — avez-vous pu parcourir le message ?",
			BodyHTML: `<p style="margin:0 0 16px;">Bonjour {{contact_name}},</p>
<p style="margin:0 0 16px;">Je vous avais envoyé un court message sur petsFollow pour {{practice_name}}. Pas d’urgence : je voulais simplement vérifier que vous l’aviez bien reçu.</p>
<p style="margin:0 0 16px;">Si 20 minutes cette semaine ou la suivante vous conviennent, indiquez-moi un créneau (ou répondez « plus tard » — je noterai).</p>
` + cta("Proposer 20 minutes", "{{cta_demo_url}}") + sig,
		},
		{
			Slug: "nurture_j3", Name: "Relance valeur J+3", Category: "nurture", Locale: "fr", IsActive: &active,
			Subject: "Passeport digital animal — Web + mobile pour {{practice_name}}",
			BodyHTML: `<p style="margin:0 0 16px;">Bonjour {{contact_name}},</p>
<p style="margin:0 0 16px;">Beaucoup de cabinets {{city}} suivent encore les patients au cas par cas (carnet, téléphone). petsFollow ajoute un <strong>passeport digital partagé</strong> entre le cabinet, les pros terrain et le propriétaire — sans remplacer votre PMS.</p>
<ul style="margin:0 0 16px;padding-left:20px;line-height:1.6;">
  <li>Web Pro pour le cabinet (CR IA inclus)</li>
  <li>Pro Light gratuit sur le terrain</li>
  <li>App client : le propriétaire paie le suivi (≤ 3,50 € / mois)</li>
</ul>
<p style="margin:0 0 16px;">20 minutes suffisent pour juger si ça vaut le coup pour vous.</p>
` + cta("Réserver 20 minutes", "{{cta_demo_url}}") + sig,
		},
		{
			Slug: "nurture_j7", Name: "Soft close J+7", Category: "nurture", Locale: "fr", IsActive: &active,
			Subject: "Dernier message — RDV ou on s’arrête là ?",
			BodyHTML: `<p style="margin:0 0 16px;">Bonjour {{contact_name}},</p>
<p style="margin:0 0 16px;">Je préfère un non clair qu’un peut-être. Pour {{practice_name}}, deux options :</p>
<ol style="margin:0 0 16px;padding-left:20px;line-height:1.6;">
  <li>On fixe 20 minutes cette semaine ou la suivante</li>
  <li>On s’arrête là — aucun souci, je ne relancerai plus</li>
</ol>
<p style="margin:0 0 16px;">Répondez simplement « RDV » ou « stop ».</p>
` + cta("Fixer un créneau", "{{cta_demo_url}}") + sig,
		},
		{
			Slug: "post_demo_next_steps", Name: "Après la démo — prochaines étapes", Category: "post", Locale: "fr", IsActive: &active,
			Subject: "Merci pour la démo — suite pour {{practice_name}}",
			BodyHTML: `<p style="margin:0 0 16px;">Bonjour {{contact_name}},</p>
<p style="margin:0 0 16px;">Merci pour votre temps. Pour activer petsFollow côté cabinet :</p>
<ol style="margin:0 0 16px;padding-left:20px;line-height:1.6;">
  <li>Compléter le profil Pro (cabinet / site)</li>
  <li>Inviter l’équipe (assistants, secrétaires)</li>
  <li>Activer les premiers patients — le client paie le suivi animal</li>
</ol>
<p style="margin:0 0 16px;">SaaS Pro : 834,71 € HTVA / an (ou 2 253,72 € / 3 ans, −10 %) + setup 320 € — facturation hors ligne. Je reste disponible pour l’onboarding.</p>
` + cta("Revoir la brochure", "{{cta_demo_url}}") + sig,
		},
		{
			Slug: "no_show_reschedule", Name: "No-show — replanifier", Category: "rdv", Locale: "fr", IsActive: &active,
			Subject: "On a manqué notre créneau — {{practice_name}}",
			BodyHTML: `<p style="margin:0 0 16px;">Bonjour {{contact_name}},</p>
<p style="margin:0 0 16px;">Nous n’avons pas pu nous connecter pour le créneau prévu. Aucun souci — la clinique passe souvent avant.</p>
<p style="margin:0 0 16px;">Je vous propose de rebloquer 20 minutes : répondez avec un jour qui vous arrange, ou choisissez via le lien.</p>
` + cta("Replanifier 20 minutes", "{{cta_demo_url}}") + sig,
		},
		{
			Slug: "reactivation_later", Name: "Pas le moment — rappel différé", Category: "reactivation", Locale: "fr", IsActive: &active,
			Subject: "Comme convenu — rappel petsFollow pour {{practice_name}}",
			BodyHTML: `<p style="margin:0 0 16px;">Bonjour {{contact_name}},</p>
<p style="margin:0 0 16px;">Vous m’aviez indiqué que ce n’était pas le bon moment. Je reviens comme promis, sans pression.</p>
<p style="margin:0 0 16px;">Si la continuité de soins / le passeport digital animal est à l’agenda cette année, 20 minutes clarifient vite si petsFollow est pertinent pour {{practice_name}}.</p>
` + cta("Reprendre contact", "{{cta_demo_url}}") + `
<p style="margin:16px 0 0;font-size:14px;color:#6B7280;">Sinon, ignorez ce message — ou utilisez le lien de désinscription en bas.</p>
` + sig,
		},
		{
			Slug: "onepager_leavebehind", Name: "Leave-behind 1 page", Category: "intro", Locale: "fr", IsActive: &active,
			Subject: "petsFollow en une page — {{practice_name}}",
			BodyHTML: `<p style="margin:0 0 16px;">Bonjour {{contact_name}},</p>
<p style="margin:0 0 16px;">Voici le leave-behind promis : une page pour garder le fil — continuité de soins prescrite, Web Pro + apps mobiles, client payeur ≤ 3,50 € / mois.</p>
` + cta("Ouvrir la page / brochure", "{{cta_demo_url}}") + `
<p style="margin:0 0 16px;">Pas de suite automatique de mon côté tant que vous ne le souhaitez pas.</p>
` + sig,
		},
	}
}
