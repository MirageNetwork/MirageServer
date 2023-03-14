package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

const (
	randomByteSize = 16

	errEmptyOIDCCallbackParams = Error("empty OIDC callback params")
	errNoOIDCIDToken           = Error("could not extract ID Token for OIDC callback")
	errOIDCAllowedDomains      = Error("authenticated principal does not match any allowed domain")
	errOIDCAllowedGroups       = Error("authenticated principal is not in any allowed group")
	errOIDCAllowedUsers        = Error("authenticated principal does not match any allowed user")
	errOIDCInvalidMachineState = Error(
		"requested machine state key expired before authorisation completed",
	)
	errOIDCNodeKeyMissing = Error("could not get node key from cache")
)

type IDTokenClaims struct {
	Name     string   `json:"name,omitempty"`
	Groups   []string `json:"groups,omitempty"`
	Email    string   `json:"email,omitempty"`
	Phone    string   `json:"phone_number"`
	Username string   `json:"preferred_username,omitempty"`
}

func (h *Mirage) initOIDC() error {
	var err error
	// grab oidc config if it hasn't been already
	if h.oauth2Config == nil {
		h.oidcProvider, err = oidc.NewProvider(context.Background(), h.cfg.OIDC.Issuer)

		if err != nil {
			log.Error().
				Err(err).
				Caller().
				Msgf("Could not retrieve OIDC Config: %s", err.Error())

			return err
		}

		h.oauth2Config = &oauth2.Config{
			ClientID:     h.cfg.OIDC.ClientID,
			ClientSecret: h.cfg.OIDC.ClientSecret,
			Endpoint:     h.oidcProvider.Endpoint(),
			RedirectURL: fmt.Sprintf(
				"https://%s/a/oauth_response",
				h.cfg.ServerURL,
			),
			Scopes: h.cfg.OIDC.Scope,
		}
	}

	return nil
}

type oidcCallbackTemplateConfig struct {
	User string
	Verb string
}

func validateOIDCCallbackParams(
	writer http.ResponseWriter,
	req *http.Request,
) (string, string, error) {
	code := req.URL.Query().Get("code")
	state := req.URL.Query().Get("state")

	if code == "" || state == "" {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusBadRequest)
		_, err := writer.Write([]byte("Wrong params"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		return "", "", errEmptyOIDCCallbackParams
	}

	return code, state, nil
}

func (h *Mirage) getIDTokenForOIDCCallback(
	ctx context.Context,
	writer http.ResponseWriter,
	code, state string,
) (string, error) {
	oauth2Token, err := h.oauth2Config.Exchange(ctx, code)
	if err != nil {
		log.Error().
			Err(err).
			Caller().
			Msg("Could not exchange code for token")
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusBadRequest)
		_, werr := writer.Write([]byte("Could not exchange code for token"))
		if werr != nil {
			log.Error().
				Caller().
				Err(werr).
				Msg("Failed to write response")
		}

		return "", err
	}

	log.Trace().
		Caller().
		Str("code", code).
		Str("state", state).
		Msg("Got oidc callback")

	rawIDToken, rawIDTokenOK := oauth2Token.Extra("id_token").(string)
	if !rawIDTokenOK {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusBadRequest)
		_, err := writer.Write([]byte("Could not extract ID Token"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		return "", errNoOIDCIDToken
	}

	return rawIDToken, nil
}

func (h *Mirage) verifyIDTokenForOIDCCallback(
	ctx context.Context,
	writer http.ResponseWriter,
	rawIDToken string,
) (*oidc.IDToken, error) {
	verifier := h.oidcProvider.Verifier(&oidc.Config{ClientID: h.cfg.OIDC.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		log.Error().
			Err(err).
			Caller().
			Msg("failed to verify id token")
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusBadRequest)
		_, werr := writer.Write([]byte("Failed to verify id token"))
		if werr != nil {
			log.Error().
				Caller().
				Err(werr).
				Msg("Failed to write response")
		}

		return nil, err
	}

	return idToken, nil
}

func extractIDTokenClaims(
	writer http.ResponseWriter,
	idToken *oidc.IDToken,
) (*IDTokenClaims, error) {
	var claims IDTokenClaims
	if err := idToken.Claims(&claims); err != nil {
		log.Error().
			Err(err).
			Caller().
			Msg("Failed to decode id token claims")
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusBadRequest)
		_, werr := writer.Write([]byte("Failed to decode id token claims"))
		if werr != nil {
			log.Error().
				Caller().
				Err(werr).
				Msg("Failed to write response")
		}

		return nil, err
	}

	return &claims, nil
}

// validateOIDCAllowedDomains checks that if AllowedDomains is provided,
// that the authenticated principal ends with @<alloweddomain>.
func validateOIDCAllowedDomains(
	writer http.ResponseWriter,
	allowedDomains []string,
	claims *IDTokenClaims,
) error {
	if len(allowedDomains) > 0 {
		if at := strings.LastIndex(claims.Email, "@"); at < 0 ||
			!IsStringInSlice(allowedDomains, claims.Email[at+1:]) {
			log.Error().Msg("authenticated principal does not match any allowed domain")
			writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
			writer.WriteHeader(http.StatusBadRequest)
			_, err := writer.Write([]byte("unauthorized principal (domain mismatch)"))
			if err != nil {
				log.Error().
					Caller().
					Err(err).
					Msg("Failed to write response")
			}

			return errOIDCAllowedDomains
		}
	}

	return nil
}

// validateOIDCAllowedGroups checks if AllowedGroups is provided,
// and that the user has one group in the list.
// claims.Groups can be populated by adding a client scope named
// 'groups' that contains group membership.
func validateOIDCAllowedGroups(
	writer http.ResponseWriter,
	allowedGroups []string,
	claims *IDTokenClaims,
) error {
	if len(allowedGroups) > 0 {
		for _, group := range allowedGroups {
			if IsStringInSlice(claims.Groups, group) {
				return nil
			}
		}

		log.Error().Msg("authenticated principal not in any allowed groups")
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusBadRequest)
		_, err := writer.Write([]byte("unauthorized principal (allowed groups)"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		return errOIDCAllowedGroups
	}

	return nil
}

// validateOIDCAllowedUsers checks that if AllowedUsers is provided,
// that the authenticated principal is part of that list.
func validateOIDCAllowedUsers(
	writer http.ResponseWriter,
	allowedUsers []string,
	claims *IDTokenClaims,
) error {
	if len(allowedUsers) > 0 &&
		!IsStringInSlice(allowedUsers, claims.Email) {
		log.Error().Msg("authenticated principal does not match any allowed user")
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusBadRequest)
		_, err := writer.Write([]byte("unauthorized principal (user mismatch)"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		return errOIDCAllowedUsers
	}

	return nil
}

func getUserName(
	writer http.ResponseWriter,
	claims *IDTokenClaims,
	stripEmaildomain bool,
) (string, string, error) {
	/* cgao6 change to use phone

	userName, err := NormalizeToFQDNRules(
		claims.Email,
		stripEmaildomain,
	)
	if err != nil {
		log.Error().Err(err).Caller().Msgf("couldn't normalize email")
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusInternalServerError)
		_, werr := writer.Write([]byte("couldn't normalize email"))
		if werr != nil {
			log.Error().
				Caller().
				Err(werr).
				Msg("Failed to write response")
		}

		return "", err
	}
	*/
	userName := strings.ReplaceAll(strings.TrimPrefix(claims.Phone, "+86"), " ", "")
	userDisName := claims.Name
	return userName, userDisName, nil
}

func (h *Mirage) findOrCreateNewUserForOIDCCallback(
	userName string,
	userDisName string,
	orgId int64,
) (*User, error) {
	user, err := h.GetUser(userName)
	if errors.Is(err, ErrUserNotFound) {
		user, err = h.CreateUser(userName, userDisName, orgId)
		if err != nil {
			log.Error().
				Err(err).
				Caller().
				Msgf("could not create new user '%s'", userName)
			return nil, err
		}
	} else if err != nil {
		log.Error().
			Caller().
			Err(err).
			Str("user", userName).
			Msg("could not find or create user")
		return nil, err
	}
	return user, nil
}
