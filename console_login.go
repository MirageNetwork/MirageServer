package headscale

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
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

	rawIDToken, err := h.getIDTokenForOIDCCallback(r.Context(), w, code, state)
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

	tokencookie := &http.Cookie{
		Name:     "OIDC_Token",
		Value:    rawIDToken,
		Secure:   true,
		HttpOnly: true,
		Domain:   strings.Split(h.cfg.ServerURL, "://")[1],
		Expires:  idToken.Expiry,
		Path:     "/",
	}
	http.SetCookie(w, tokencookie)
	log.Info().Msg("将设置Cookie成功！")

	http.Redirect(w, r, "/admin", http.StatusFound)

}

func (h *Headscale) ConsoleLogout(
	w http.ResponseWriter,
	r *http.Request,
) {
	rToken, _ := r.Cookie("OIDC_Token")
	idtoken := rToken.Value
	delCookie := &http.Cookie{
		Name:     "OIDC_Token",
		Domain:   strings.Split(h.cfg.ServerURL, "://")[1],
		Expires:  time.Now().Add(time.Minute * 5),
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		Path:     "/",
	}

	http.SetCookie(w, delCookie)
	http.Redirect(w, r, h.cfg.OIDC.LogoutURL+"?id_token_hint="+idtoken+"&post_logout_redirect_uri="+h.cfg.ServerURL+"/admin/login", http.StatusFound)
}

func (h *Headscale) ConsoleLogoutCallback(
	w http.ResponseWriter,
	r *http.Request,
) {
	/*
		delCookie := &http.Cookie{
			Name:     "OIDC_Token",
			Domain:   strings.Split(h.cfg.ServerURL, "://")[1],
			Expires:  time.Now().Add(time.Minute * 5),
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			Path:     "/",
		}

		http.SetCookie(w, delCookie)
	*/
	renderResult(w, false, "您已经成功登出！", "/", "返回首页")
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
