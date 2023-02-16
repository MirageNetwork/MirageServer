package Mirage

import (
	_ "embed"
	"net/http"
	"strings"
	"time"
)

func (h *Mirage) ConsoleLogout(
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
}
