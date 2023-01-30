package headscale

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

func (h *Headscale) doLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	provider := r.Form["provider"][0]
	nextURL := r.Form["next_url"][0]
	switch provider {
	case "Ali":
		h.doAliLogin(w, r, nextURL)
	}
}
func (h *Headscale) doAliLogin(w http.ResponseWriter, r *http.Request, nextURL string) {
	randomBlob := make([]byte, randomByteSize)
	if _, err := rand.Read(randomBlob); err != nil {
		log.Error().
			Caller().
			Msg("could not read 16 bytes from rand")
		http.Error(w, "生成注册随机数错误", http.StatusInternalServerError)
		return
	}
	stateStr := hex.EncodeToString(randomBlob)[:32]
	h.loginCache.Set(stateStr, time.Now(), registerCacheExpiration)

	extras := make([]oauth2.AuthCodeOption, 0, len(h.cfg.OIDC.ExtraParams))
	for k, v := range h.cfg.OIDC.ExtraParams {
		extras = append(extras, oauth2.SetAuthURLParam(k, v))
	}
	nextURL, err := url.PathUnescape(nextURL)
	toDistPath := "/"
	if strings.Contains(nextURL, "#") {
		toDistPath = strings.Split(nextURL, "#")[1]
	}
	nextURL = strings.Split(nextURL, "#")[0] + "?next_url=" + toDistPath

	if err != nil {
		log.Error().Msg("Next URL处理失败：" + err.Error())
	}
	log.Error().Msg("之后会跳转到：" + fmt.Sprintf(
		"%s/%s",
		strings.TrimSuffix(h.cfg.ServerURL, "/"),
		strings.TrimPrefix(nextURL, "/"),
	))

	adminOauth2Config := &oauth2.Config{
		ClientID:     h.cfg.OIDC.ClientID,
		ClientSecret: h.cfg.OIDC.ClientSecret,
		Endpoint:     h.oidcProvider.Endpoint(),
		RedirectURL: fmt.Sprintf(
			"%s/%s",
			strings.TrimSuffix(h.cfg.ServerURL, "/"),
			strings.TrimPrefix(nextURL, "/"),
		),
		Scopes: h.cfg.OIDC.Scope,
	}

	authURL := adminOauth2Config.AuthCodeURL(stateStr, extras...)
	log.Debug().Msgf("Redirecting to %s for authentication", authURL)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// 处理可能来自OIDC的callback
func (h *Headscale) getIDTokenFromOIDCCallback(
	w http.ResponseWriter,
	r *http.Request,
) (string, *oidc.IDToken) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		return "", nil
	}
	oauth2Token, err := h.oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		return "", nil
	}
	log.Trace().
		Caller().
		Str("code", code).
		Str("state", state).
		Msg("Got oidc callback")
	rawIDToken, rawIDTokenOK := oauth2Token.Extra("id_token").(string)
	if !rawIDTokenOK {
		return "", nil
	}
	verifier := h.oidcProvider.Verifier(&oidc.Config{ClientID: h.cfg.OIDC.ClientID})
	idToken, err := verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		return "", nil
	}
	return rawIDToken, idToken
}

// WebUI控制台鉴权中间件
func (h *Headscale) ConsoleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//检查是否OIDC callback
		if rawIDToken, idToken := h.getIDTokenFromOIDCCallback(w, r); idToken != nil {
			newQuery := r.URL.Query()
			newQuery.Del("code")
			newQuery.Del("state")
			nextURL := newQuery.Get("next_url")
			newQuery.Del("next_url")
			r.URL.RawQuery = newQuery.Encode()
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
			http.Redirect(w, r, "/admin#"+nextURL, http.StatusFound)
		} else {
			oidc_token, _ := r.Cookie("OIDC_Token")
			if oidc_token != nil {
				verifier := h.oidcProvider.Verifier(&oidc.Config{ClientID: h.cfg.OIDC.ClientID})
				idToken, err := verifier.Verify(r.Context(), oidc_token.Value)
				if err != nil {
					log.Error().
						Caller().
						Msg("could not verifyIDTokenForOIDCCallback")
					nextURL := r.URL.Path
					newQuery := r.URL.Query()
					newQuery.Add("next_url", nextURL)
					r.URL.RawQuery = newQuery.Encode()
					http.Redirect(w, r, "/login?"+r.URL.RawQuery, http.StatusFound)
					return
				}
				var claims IDTokenClaims
				err = idToken.Claims(&claims)
				if err != nil {
					log.Error().
						Caller().
						Msg("could not Extra Claims")
					nextURL := r.URL.Path
					newQuery := r.URL.Query()
					newQuery.Add("next_url", nextURL)
					r.URL.RawQuery = newQuery.Encode()
					http.Redirect(w, r, "/login?"+r.URL.RawQuery, http.StatusFound)
					return
				}
				next.ServeHTTP(w, r)
			} else {
				log.Error().Msg("未能从Cookie读取到OIDC Token！")
				nextURL := r.URL.Path
				newQuery := r.URL.Query()
				newQuery.Add("next_url", nextURL)
				r.URL.RawQuery = newQuery.Encode()
				http.Redirect(w, r, "/login?"+r.URL.RawQuery, http.StatusFound)
			}
		}
	})
}

// API鉴权结果响应
type APICheckRes struct {
	NeedReauth bool   `json:"needreauth"`
	Reason     string `json:"needreauthreason"`
}

// API鉴权中间件
func (h *Headscale) APIAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oidc_token, _ := r.Cookie("OIDC_Token")
		if oidc_token != nil {
			idToken, err := h.verifyIDTokenForOIDCCallback(r.Context(), w, oidc_token.Value)
			if err != nil {
				log.Error().
					Caller().
					Msg("could not verifyIDTokenForOIDCCallback")
				renderData := APICheckRes{
					NeedReauth: true,
					Reason:     "IDaaS无法校验",
				}
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&renderData)
				return
			}
			var claims IDTokenClaims
			err = idToken.Claims(&claims)
			if err != nil {
				log.Error().
					Caller().
					Msg("could not Extra Claims")
				http.Error(w, "OIDC Token解析Claim错误！", http.StatusInternalServerError)
				renderData := APICheckRes{
					NeedReauth: true,
					Reason:     "Token解析Claim错误",
				}
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(&renderData)
				return
			}
			next.ServeHTTP(w, r)
		} else {
			log.Error().Msg("未能从Cookie读取到OIDC Token！")
			renderData := APICheckRes{
				NeedReauth: true,
				Reason:     "未读取到Token",
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&renderData)
			return
		}
	})
}
