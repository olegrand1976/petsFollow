# Certificats CA belges (Web eID)

Fichiers `.crt` (**PEM** après sync) embarqués via `go:embed` pour valider les
certificats d’authentification des cartes eID belges (Citizen **et** Foreigner).
Le loader Go accepte encore le DER (fichiers AIA historiques non re-sync).

## Dernière sync

**2026-08-14** — sync complète (index miroir + sondage AIA).

| Famille | Approx. | Fichiers |
|--------|---------|----------|
| Root CA | **5** | `belgiumrca.crt`, `belgiumrca2`–`4`, `belgiumrca6` (pas de Root CA5 publié) |
| Citizen CA | **~332** | `citizen*`, `citizenYYYYMM`, `eidcYYYYMM` (AIA) |
| Foreigner CA | **~93** | `foreigner*` (miroir) + `eidfYYYYMM` (AIA) |
| **Total unique** | **~430** | empreinte SHA-256 DER |

Les cartes « Foreigner » (titres de séjour / non-citoyens) ne chaînent pas sur Citizen CA :
sans `eidf*` / `foreigner*` + `belgiumrca6`, Web eID renvoie `CERTIFICATE_NOT_TRUSTED`
→ `eid_cert_untrusted`.

`belgiumrs*.crt` du miroir sont des doublons bit-à-bit de `belgiumrca*` → retirés au dédup.

**Lots expirés** : une part importante des CA historiques a un `notAfter` passé. On les
**conserve volontairement** (cartes long-vives / chaînes anciennes). Ne pas purger
par date — Go n’utilise un intermédiaire périmé que s’il n’existe pas de chemin valide.

## Sources

1. `http://certs.eid.belgium.be/` — listing HTML (racines, lots `citizen*`, `foreigner*`)
2. `http://crt.eidpki.belgium.be/eid/` — AIA **sans** index : `brca6.crt` → `belgiumrca6.crt`,
   lots `eidfYYYYMM.crt` (Foreigner), `eidcYYYYMM.crt` (Citizen)

Doc publique : https://eid.belgium.be/ — Viewer / Web eID (écosystème BOSA).

> **HTTP clair** : les miroirs BOSA sont en `http://`. Lancer `sync-eid-certs` depuis un
> réseau de confiance (pas un hotspot public) — risque MITM sur le contenu téléchargé.

Patterns AIA sans hit (2026-08-14) : `eidaYYYYMM.crt`, `foreignerYYYYMM.crt` sur l’hôte AIA
(les `foreigner*` historiques restent sur le miroir). `belgiumrca5` / `brca5` → 404.

## Re-sync

```bash
# Depuis la racine du monorepo
bash scripts/sync-eid-certs.sh
# ou: make sync-eid-certs

# Aperçu sans écrire
bash scripts/sync-eid-certs.sh --dry-run

# Vérifier le trust store embarqué
cd go && go test ./internal/eid/ -count=1
```

Le script : parse l’index miroir → télécharge racines/citizen/foreigner → sonde l’AIA
(`brca*`, `eidf*`/`eidc*` mois 01–12, 2018–2026) → vérifie avec `openssl x509` →
stocke / réécrit en **PEM** → déduplique par empreinte SHA-256 DER.

En local, `WEB_EID_DISABLE_OCSP=true` (posé par `make api-dev`) désactive OCSP sans
désactiver la vérification de chaîne.

Ne pas committer de certificats feuille personnels (cartes individuelles) — uniquement des CA.
