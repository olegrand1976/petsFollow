package email

import (
	"fmt"
	"html"
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

type confirmRegistrationContent struct {
	Lang        string
	Tagline     string
	Greeting    string
	Intro       string
	CTALabel    string
	Expiry      string
	Disclaimer  string
	Preheader   string
	ConfirmURL  string
}

func renderConfirmRegistration(c confirmRegistrationContent) string {
	lang := html.EscapeString(c.Lang)
	if lang == "" {
		lang = "fr"
	}
	tagline := html.EscapeString(c.Tagline)
	greeting := html.EscapeString(c.Greeting)
	intro := html.EscapeString(c.Intro)
	ctaLabel := html.EscapeString(c.CTALabel)
	expiry := html.EscapeString(c.Expiry)
	disclaimer := html.EscapeString(c.Disclaimer)
	preheader := html.EscapeString(c.Preheader)
	confirmURL := html.EscapeString(c.ConfirmURL)

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
                petsFollow <span style="color:%s;font-weight:600;">Pro</span>
              </div>
              <div style="margin-top:8px;font-size:14px;color:rgba(255,255,255,0.88);letter-spacing:0.01em;">%s</div>
            </td>
          </tr>
          <tr>
            <td style="background-color:%s;padding:40px;border-left:1px solid %s;border-right:1px solid %s;">
              <p style="margin:0 0 12px;font-size:20px;font-weight:600;color:%s;line-height:1.3;">%s</p>
              <p style="margin:0 0 28px;font-size:16px;line-height:1.65;color:%s;">%s</p>
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" align="center" style="margin:0 auto 28px;">
                <tr>
                  <td align="center" style="background-color:%s;border-radius:8px;">
                    <a href="%s" target="_blank" style="display:inline-block;padding:15px 36px;font-size:16px;font-weight:600;color:#FFFFFF;text-decoration:none;border-radius:8px;mso-padding-alt:0;">%s</a>
                  </td>
                </tr>
              </table>
              <p style="margin:0;font-size:14px;line-height:1.55;color:%s;text-align:center;">%s</p>
            </td>
          </tr>
          <tr>
            <td style="background-color:%s;border:1px solid %s;border-top:none;border-radius:0 0 12px 12px;padding:28px 40px;text-align:center;">
              <p style="margin:0;font-size:13px;line-height:1.55;color:%s;">%s</p>
              <p style="margin:16px 0 0;font-size:12px;color:%s;">petsFollow &mdash; LL-IT-SC</p>
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
		colorGold, tagline,
		colorSurface, colorBorder, colorBorder,
		colorText, greeting,
		colorText, intro,
		colorAccent, confirmURL, ctaLabel,
		colorMuted, expiry,
		colorBg, colorBorder,
		colorMuted, disclaimer,
		colorMuted,
	)
}
