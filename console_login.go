package headscale

import (
	_ "embed"
	"net/http"
	"strings"
	"time"
)

/*
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
*/
func (h *Headscale) ConsoleLogout(
	w http.ResponseWriter,
	r *http.Request,
) {
	controlCodeCookie, err := r.Cookie("miragcontrol")
	if err != http.ErrNoCookie {
		controlCode := controlCodeCookie.Value
		h.controlCodeCache.Delete(controlCode)
		delCookie := &http.Cookie{
			Name:     "miragecontrol",
			Domain:   strings.Split(h.cfg.ServerURL, "://")[1],
			Expires:  time.Now().Add(time.Minute * 5),
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(w, delCookie)
	}
	http.Redirect(w, r, "/", http.StatusFound)
	//http.Redirect(w, r, h.cfg.OIDC.LogoutURL+"?id_token_hint="+idtoken+"&post_logout_redirect_uri="+h.cfg.ServerURL+"/login", http.StatusFound)
}
