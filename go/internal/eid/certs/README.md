# Certificats CA belges (Web eID)

Fichiers `.crt` (DER ou PEM) embarqués via `go:embed` pour valider les certificats
d’authentification des cartes eID belges.

Sources : copies alignées Ventura (`adant_iqc/config/web-eid/certs`) +
`https://certs.eid.belgium.be/` (belgiumrs*, citizen201304).

Mise à jour :

```bash
# depuis le dépôt Ventura ou le dépôt officiel
cp /path/to/citizen*.crt go/internal/eid/certs/
curl -fsSL -o go/internal/eid/certs/belgiumrs2.crt https://certs.eid.belgium.be/belgiumrs2.crt
```

En local, `WEB_EID_DISABLE_OCSP=true` (posé par `make api-dev`) désactive OCSP sans
désactiver la vérification de chaîne.
