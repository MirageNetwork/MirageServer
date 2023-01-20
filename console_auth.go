package headscale

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

func (h *Headscale) GoOIDCLogin(w http.ResponseWriter, r *http.Request) {
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

	adminOauth2Config := &oauth2.Config{
		ClientID:     h.cfg.OIDC.ClientID,
		ClientSecret: h.cfg.OIDC.ClientSecret,
		Endpoint:     h.oidcProvider.Endpoint(),
		RedirectURL: fmt.Sprintf(
			"%s/login/callback",
			strings.TrimSuffix(h.cfg.ServerURL, "/"),
		),
		Scopes: h.cfg.OIDC.Scope,
	}

	authURL := adminOauth2Config.AuthCodeURL(stateStr, extras...)
	log.Debug().Msgf("Redirecting to %s for authentication", authURL)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// WebUI控制台鉴权中间件
func (h *Headscale) ConsoleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oidc_token, _ := r.Cookie("OIDC_Token")
		if oidc_token != nil {
			log.Info().Msg(oidc_token.Value)
			idToken, err := h.verifyIDTokenForOIDCCallback(r.Context(), w, oidc_token.Value)
			if err != nil {
				log.Error().
					Caller().
					Msg("could not verifyIDTokenForOIDCCallback")
				h.GoOIDCLogin(w, r)
				return
			}
			var claims IDTokenClaims
			err = idToken.Claims(&claims)
			if err != nil {
				log.Error().
					Caller().
					Msg("could not Extra Claims")
				http.Error(w, "OIDC Token解析Claim错误！", http.StatusInternalServerError)
				renderResult(w, true, "OIDC Token解析Claim错误！", "回到首页", "/")
			}
			log.Info().Msg(claims.Username + "Token认证成功！")
			next.ServeHTTP(w, r)
		} else {
			log.Error().Msg("未能从Cookie读取到OIDC Token！")
			h.GoOIDCLogin(w, r)
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
			log.Info().Msg(oidc_token.Value)
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
