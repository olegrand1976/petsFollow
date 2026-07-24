package email

import (
	"fmt"
	"html"
	"strings"
)

const (
	colorPrimary  = "#1B3A4B"
	colorAccent   = "#2A9D8F"
	colorGold     = "#E9C46A"
	colorBg       = "#F7F9FB"
	colorSurface  = "#FFFFFF"
	colorText     = "#1B3A4B"
	colorMuted    = "#6B7280"
	colorBorder   = "#E2E6ED"
	fontStack     = "'DM Sans', Arial, Helvetica, sans-serif"
	emailMaxWidth = 600
)

type brandAssets struct {
	LLITLogoURL    string
	LLITWebsiteURL string
	SiteURL        string
}

type brandedEmailContent struct {
	Lang             string
	ProductLabel     string // default "petsFollow Pro"; use "petsFollow" for client emails
	Tagline          string
	Greeting         string
	Intro            string
	Detail           string // optional plain-text block (e.g. message preview)
	CTALabel         string
	CTAURL           string
	Expiry           string
	Disclaimer       string
	Preheader        string
	FooterPoweredBy  string
	FooterVisit      string
	UnsubscribeLabel string
	UnsubscribeURL   string
	Brand            brandAssets
}

func renderBrandedEmail(c brandedEmailContent) string {
	lang := html.EscapeString(c.Lang)
	if lang == "" {
		lang = "fr"
	}
	productLabel := strings.TrimSpace(c.ProductLabel)
	if productLabel == "" {
		productLabel = "petsFollow Pro"
	}
	// Split "petsFollow Pro" for gold accent on the second word when present.
	productHTML := html.EscapeString(productLabel)
	if productLabel == "petsFollow Pro" {
		productHTML = `petsFollow <span style="color:` + colorGold + `;font-weight:600;">Pro</span>`
	} else if productLabel == "petsFollow" {
		productHTML = `petsFollow`
	} else {
		productHTML = html.EscapeString(productLabel)
	}
	tagline := html.EscapeString(c.Tagline)
	greeting := html.EscapeString(c.Greeting)
	intro := html.EscapeString(c.Intro)
	detail := html.EscapeString(c.Detail)
	ctaLabel := html.EscapeString(c.CTALabel)
	ctaURL := html.EscapeString(c.CTAURL)
	expiry := html.EscapeString(c.Expiry)
	disclaimer := html.EscapeString(c.Disclaimer)
	preheader := html.EscapeString(c.Preheader)
	footerPowered := html.EscapeString(c.FooterPoweredBy)
	footerVisit := html.EscapeString(c.FooterVisit)
	unsubLabel := html.EscapeString(c.UnsubscribeLabel)
	unsubURL := html.EscapeString(c.UnsubscribeURL)
	llitLogo := html.EscapeString(c.Brand.LLITLogoURL)
	llitURL := html.EscapeString(c.Brand.LLITWebsiteURL)
	siteURL := html.EscapeString(c.Brand.SiteURL)

	unsubBlock := ""
	if c.UnsubscribeURL != "" && c.UnsubscribeLabel != "" {
		unsubBlock = fmt.Sprintf(`
              <p style="margin:12px 0 0;font-size:12px;line-height:1.5;color:%s;">
                <a href="%s" target="_blank" style="color:%s;text-decoration:underline;">%s</a>
              </p>`, colorMuted, unsubURL, colorMuted, unsubLabel)
	}

	ctaBlock := ""
	if c.CTAURL != "" && c.CTALabel != "" {
		ctaBlock = fmt.Sprintf(`
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin:0 auto 28px;">
                <tr>
                  <td align="center" style="background-color:%s;border-radius:8px;">
                    <a href="%s" target="_blank" style="display:inline-block;padding:15px 36px;font-size:16px;font-weight:600;color:#FFFFFF;text-decoration:none;border-radius:8px;mso-padding-alt:0;">%s</a>
                  </td>
                </tr>
              </table>`, colorAccent, ctaURL, ctaLabel)
	}

	detailBlock := ""
	if strings.TrimSpace(c.Detail) != "" {
		detailBlock = fmt.Sprintf(`
              <div style="margin:0 0 28px;padding:16px 18px;background-color:%s;border:1px solid %s;border-radius:8px;font-size:15px;line-height:1.6;color:%s;text-align:left;white-space:pre-wrap;">%s</div>`,
			colorBg, colorBorder, colorText, detail)
	}

	expiryBlock := ""
	if strings.TrimSpace(c.Expiry) != "" {
		expiryBlock = fmt.Sprintf(`<p style="margin:0;font-size:14px;line-height:1.55;color:%s;text-align:center;">%s</p>`, colorMuted, expiry)
	}

	llitLogoBlock := ""
	if c.Brand.LLITLogoURL != "" {
		llitLogoBlock = fmt.Sprintf(`
              <a href="%s" target="_blank" style="display:inline-block;margin:18px 0 8px;text-decoration:none;">
                <img src="%s" alt="LL-IT Software &amp; Computer" width="160" style="display:block;width:160px;max-width:70%%;height:auto;border:0;margin:0 auto;" />
              </a>`, llitURL, llitLogo)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="color-scheme" content="light">
  <meta name="supported-color-schemes" content="light">
  <title>petsFollow Pro</title>
  <!--[if mso]>
  <style type="text/css">
    body, table, td { font-family: Arial, Helvetica, sans-serif !important; }
  </style>
  <![endif]-->
</head>
<body style="margin:0;padding:0;background-color:%s;font-family:%s;-webkit-font-smoothing:antialiased;">
  <div style="display:none;max-height:0;overflow:hidden;mso-hide:all;">%s</div>
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="background-color:%s;">
    <tr>
      <td align="center" style="padding:40px 16px;">
        <table role="presentation" width="%d" cellspacing="0" cellpadding="0" border="0" style="max-width:%dpx;width:100%%;">
          <tr>
            <td style="background-color:%s;background:linear-gradient(135deg, %s 0%%, %s 100%%);border-radius:12px 12px 0 0;padding:36px 40px;text-align:center;">
              <div style="font-size:30px;font-weight:700;color:#FFFFFF;letter-spacing:-0.03em;line-height:1.2;">
                %s
              </div>
              <div style="margin-top:8px;font-size:14px;color:rgba(255,255,255,0.88);letter-spacing:0.01em;">%s</div>
            </td>
          </tr>
          <tr>
            <td style="background-color:%s;padding:40px;border-left:1px solid %s;border-right:1px solid %s;">
              <p style="margin:0 0 12px;font-size:20px;font-weight:600;color:%s;line-height:1.3;">%s</p>
              <p style="margin:0 0 28px;font-size:16px;line-height:1.65;color:%s;">%s</p>
              %s
              %s
              %s
            </td>
          </tr>
          <tr>
            <td style="background-color:%s;border:1px solid %s;border-top:none;border-radius:0 0 12px 12px;padding:28px 40px;text-align:center;">
              <p style="margin:0;font-size:13px;line-height:1.55;color:%s;">%s</p>
              %s
              <p style="margin:8px 0 0;font-size:12px;color:%s;">%s</p>
              <p style="margin:6px 0 0;font-size:12px;color:%s;">
                <a href="%s" target="_blank" style="color:%s;text-decoration:none;">%s</a>
              </p>
              <p style="margin:10px 0 0;font-size:12px;color:%s;">
                <a href="%s" target="_blank" style="color:%s;text-decoration:underline;">petsFollow</a>
                &nbsp;&middot;&nbsp;
                <a href="%s" target="_blank" style="color:%s;text-decoration:underline;">LL-IT Software &amp; Computer</a>
              </p>
              %s
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,
		lang,
		colorBg, fontStack, preheader,
		colorBg,
		emailMaxWidth, emailMaxWidth,
		colorPrimary, colorPrimary, colorAccent,
		productHTML, tagline,
		colorSurface, colorBorder, colorBorder,
		colorText, greeting,
		colorText, intro,
		detailBlock,
		ctaBlock,
		expiryBlock,
		colorBg, colorBorder,
		colorMuted, disclaimer,
		llitLogoBlock,
		colorMuted, footerPowered,
		colorMuted, llitURL, colorAccent, footerVisit,
		colorMuted, siteURL, colorMuted, llitURL, colorMuted,
		unsubBlock,
	)
}

// Deprecated alias kept for older tests — prefer renderBrandedEmail.
func renderConfirmRegistration(c confirmRegistrationContent) string {
	return renderBrandedEmail(brandedEmailContent{
		Lang:       c.Lang,
		Tagline:    c.Tagline,
		Greeting:   c.Greeting,
		Intro:      c.Intro,
		CTALabel:   c.CTALabel,
		CTAURL:     c.ConfirmURL,
		Expiry:     c.Expiry,
		Disclaimer: c.Disclaimer,
		Preheader:  c.Preheader,
		FooterPoweredBy: "petsFollow — LL-IT Software & Computer",
		FooterVisit:     "Visit LL-IT Software & Computer",
		Brand: brandAssets{
			LLITWebsiteURL: "https://ll-it-sc.be",
			SiteURL:        "https://petsfollow.ll-it-sc.be",
		},
	})
}

type confirmRegistrationContent struct {
	Lang       string
	Tagline    string
	Greeting   string
	Intro      string
	CTALabel   string
	Expiry     string
	Disclaimer string
	Preheader  string
	ConfirmURL string
}
