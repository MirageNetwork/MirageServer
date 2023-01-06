package headscale

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

func (h *Headscale) ConsoleLogin(
	w http.ResponseWriter,
	r *http.Request,
) {
	code, state, err := validateOIDCCallbackParams(w, r)
	if err != nil {
		log.Error().
			Caller().
			Msg("could not validateOIDCCallbackParams")
		http.Error(w, "OIDC返回参数错误", http.StatusInternalServerError)

		return
	}

	rawIDToken, accessToken, err := h.getIDTokenForOIDCCallback(r.Context(), w, code, state)
	if err != nil {
		log.Error().
			Caller().
			Msg("could not getIDTokenForOIDCCallback")
		http.Error(w, "OIDC返回获取IDToken错误", http.StatusInternalServerError)

		return
	}

	idToken, err := h.verifyIDTokenForOIDCCallback(r.Context(), w, rawIDToken)
	if err != nil {
		log.Error().
			Caller().
			Msg("could not verifyIDTokenForOIDCCallback")
		http.Error(w, "OIDC校验IDToken错误", http.StatusInternalServerError)

		return
	}
	log.Info().Msg(idToken.Expiry.String())

	tokencookie := &http.Cookie{
		Name:     "OIDC_Token",
		Value:    rawIDToken,
		Secure:   true,
		HttpOnly: true,
		Domain:   "sdp.ipv4.uk",
		Expires:  idToken.Expiry,
		Path:     "/admin",
	}
	http.SetCookie(w, tokencookie)
	atcookie := &http.Cookie{
		Name:     "AccessToken",
		Value:    accessToken,
		Secure:   true,
		HttpOnly: true,
		Domain:   "sdp.ipv4.uk",
		Path:     "/admin",
	}
	http.SetCookie(w, atcookie)

	http.Redirect(w, r, "/admin", http.StatusFound)

}

func doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	client := http.DefaultClient
	if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		client = c
	}
	return client.Do(req)
}
func (h *Headscale) revokeToken(ctx context.Context, token string) error {
	values := url.Values{}
	values.Set("id_type_hint", "access_token")
	values.Set("token", token)
	req, err := http.NewRequest(http.MethodPost, h.cfg.OIDC.RevokeURL, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(h.cfg.OIDC.ClientID, h.cfg.OIDC.ClientSecret)

	resp, err := doRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "Error contacting revocation endpoint")
	}
	if code := resp.StatusCode; code != 200 {
		defer resp.Body.Close()
		_, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.New(fmt.Sprintf("ReturnCode: %v Failed: %v", code, err))
		}
		return errors.New(fmt.Sprintf("ReturnCode: %v Failed: %v", code, err))
	}
	return nil
}

func (h *Headscale) ConsoleLogout(
	w http.ResponseWriter,
	r *http.Request,
) {
	at, _ := r.Cookie("AccessToken")
	if at == nil {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		err := h.revokeToken(r.Context(), at.Value)
		if err != nil {
			log.Error().Msg(err.Error())
		}
		idt, _ := r.Cookie("OIDC_Token")
		if idt == nil {
			http.Redirect(w, r, "/", http.StatusFound)
		}
		http.Redirect(w, r, h.cfg.OIDC.LogoutURL+"?id_token_hint="+idt.Value+"&post_logout_redirect_uri="+h.cfg.ServerURL+"/logout/callback", http.StatusFound)
	}
}

func (h *Headscale) ConsoleLogoutCallback(
	w http.ResponseWriter,
	r *http.Request,
) {
	delTokenCookie := http.Cookie{Name: "OIDC_Token", Value: "", Path: "/admin", MaxAge: -1}
	http.SetCookie(w, &delTokenCookie)
	http.Redirect(w, r, "/admin", http.StatusFound)
}

//go:embed admin/welcome.html
var welcomeHTML string

func (h *Headscale) ConsoleWelcome(
	writer http.ResponseWriter,
	req *http.Request,
) {
	welcomeT := template.Must(template.New("welcome").Parse(welcomeHTML))
	var payload bytes.Buffer
	if err := welcomeT.Execute(&payload, nil); err != nil {
		log.Error().
			Str("handler", "welcomeHTML").
			Err(err).
			Msg("Could not render welcome HTML")

		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusInternalServerError)
		_, err := writer.Write([]byte("Could not render welcome index template"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		return
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}
