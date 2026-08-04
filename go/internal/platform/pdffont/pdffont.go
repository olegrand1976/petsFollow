// Package pdffont centralise la police des PDF générés côté API (dossier
// animal, compte-rendu de consultation).
//
// Pourquoi une police embarquée plutôt que les polices « core » de gofpdf :
// Arial/Helvetica core sont encodées en cp1252 et passent par
// UnicodeTranslatorFromDescriptor(""), qui ne peut pas représenter le
// cyrillique. Sans police Unicode, un compte-rendu en ukrainien ou en russe —
// ou même un simple nom d'animal / de propriétaire en cyrillique dans un PDF
// français — ressort en caractères parasites.
//
// Liberation Sans est choisie parce qu'elle est *métriquement compatible avec
// Arial* : les largeurs d'avance des glyphes latins sont identiques, donc les
// gabarits existants (largeurs de cellules fixées en mm) ne se décalent pas.
// Elle couvre le latin étendu et le cyrillique complet, y compris les lettres
// propres à l'ukrainien (ґ, є, і, ї).
//
// Licence : SIL Open Font License 1.1 — voir fonts/LICENSE.txt.
package pdffont

import (
	"embed"

	"github.com/phpdave11/gofpdf"
)

//go:embed fonts/*.ttf
var fontFS embed.FS

// Family est le nom de famille à passer à pdf.SetFont.
const Family = "LiberationSans"

var styleFiles = map[string]string{
	"":  "fonts/LiberationSans-Regular.ttf",
	"B": "fonts/LiberationSans-Bold.ttf",
	"I": "fonts/LiberationSans-Italic.ttf",
}

// Register installe la famille Unicode sur pdf, pour les styles normal, gras et
// italique. À appeler juste après gofpdf.New(), avant tout SetFont.
//
// Les erreurs de police sont portées par l'objet gofpdf lui-même (pdf.Error()),
// vérifié à l'Output() par les appelants — inutile de renvoyer une erreur ici.
func Register(pdf *gofpdf.Fpdf) {
	for style, path := range styleFiles {
		data, err := fontFS.ReadFile(path)
		if err != nil {
			// embed garantit la présence des fichiers à la compilation ; un
			// échec ici ne peut venir que d'une corruption du binaire.
			pdf.SetError(err)
			return
		}
		pdf.AddUTF8FontFromBytes(Family, style, data)
	}
}

// Translate est le remplaçant de UnicodeTranslatorFromDescriptor("") : avec une
// police UTF-8, le texte est passé tel quel. Conservé sous forme de fonction
// pour que les générateurs gardent leur forme `tr(...)` autour des chaînes.
func Translate(s string) string { return s }
