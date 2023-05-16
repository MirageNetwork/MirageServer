package controller

import (
	_ "embed"
	"net/http"
	"strconv"
	"time"
)

func (h *Mirage) ConsoleLogout(
	w http.ResponseWriter,
	r *http.Request,
) {
	controlCodeCookie, err := r.Cookie("miragecontrol")
	if err != http.ErrNoCookie {
		controlCode := controlCodeCookie.Value
		h.controlCodeCache.Delete(controlCode)
		delCookie := &http.Cookie{
			Name:     "miragecontrol",
			Domain:   h.cfg.ServerURL,
			Expires:  time.Now().Add(time.Minute * 5),
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(w, delCookie)
	}
	nextURL := r.URL.Query().Get("next_url")
	if nextURL == "" {
		nextURL = "/"
	}
	http.Redirect(w, r, nextURL, http.StatusFound)
}

func (h *Mirage) DexErrHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	dexErrCode, err := strconv.Atoi(r.URL.Query().Get("code"))
	if err != nil {
		dexErrCode = 500
	}
	dexErrDescription := r.URL.Query().Get("desc")
	h.ErrMessage(w, r, dexErrCode, dexErrDescription)
}
