package notify

// ✮ ⋆ ˚ email renderer — two profiles, one layout
// platform emails: radiooo logo (password reset, invites)
// station emails: station logo (show reminders, notifications)

import (
	"fmt"

	"github.com/matcornic/hermes/v2"
)

var platformHermes = hermes.Hermes{
	Product: hermes.Product{
		Name:      "radiooo",
		Link:      "https://radiooo.app",
		Logo:      "https://radiooo.app/logo.png",
		Copyright: "sent by radiooo",
	},
}

// RenderPlatformEmail renders with the radiooo branding ⊹ ࣪ ˖
func RenderPlatformEmail(email hermes.Email) (html string, plain string, err error) {
	html, err = platformHermes.GenerateHTML(email)
	if err != nil {
		return "", "", fmt.Errorf("render platform html: %w", err)
	}
	plain, err = platformHermes.GeneratePlainText(email)
	if err != nil {
		return "", "", fmt.Errorf("render platform text: %w", err)
	}
	return html, plain, nil
}

// RenderStationEmail renders with the station's own branding ⋆˙⟡
func RenderStationEmail(stationName, logoURL string, email hermes.Email) (html string, plain string, err error) {
	if logoURL == "" {
		logoURL = "https://radiooo.app/logo.png"
	}
	h := hermes.Hermes{
		Product: hermes.Product{
			Name:      stationName,
			Link:      "https://radiooo.app",
			Logo:      logoURL,
			Copyright: fmt.Sprintf("sent by %s via radiooo", stationName),
		},
	}
	html, err = h.GenerateHTML(email)
	if err != nil {
		return "", "", fmt.Errorf("render station html: %w", err)
	}
	plain, err = h.GeneratePlainText(email)
	if err != nil {
		return "", "", fmt.Errorf("render station text: %w", err)
	}
	return html, plain, nil
}
