package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"tailscale.com/tailcfg"
)

// // NoiseRegistrationHandler handles the actual registration process of a machine.
func (t *ts2021App) NoiseRegistrationHandler(
	writer http.ResponseWriter,
	req *http.Request,
) {
	log.Trace().Caller().Msgf("Noise registration handler for client %s", req.RemoteAddr)
	if req.Method != http.MethodPost {
		http.Error(writer, "Wrong method", http.StatusMethodNotAllowed)

		return
	}
	body, _ := io.ReadAll(req.Body)
	registerRequest := tailcfg.RegisterRequest{}
	if err := json.Unmarshal(body, &registerRequest); err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Cannot parse RegisterRequest")
		http.Error(writer, "Internal error", http.StatusInternalServerError)

		return
	}

	t.mirage.handleRegisterCommon(writer, req, registerRequest, t.conn.Peer())
}

type NaviRegisterRequest struct {
	ID        string
	Timestamp *time.Time
}

type NaviRegisterResponse struct {
	NodeInfo  NaviNode
	Timestamp *time.Time
}

// 司南注册noise协议接口
func (t *ts2021App) NoiseNaviRegisterHandler(
	writer http.ResponseWriter,
	req *http.Request,
) {
	log.Trace().Caller().Msgf("Noise registration handler for Navi %s", req.RemoteAddr)
	if req.Method != http.MethodPost {
		http.Error(writer, "Wrong method", http.StatusMethodNotAllowed)

		return
	}
	body, _ := io.ReadAll(req.Body)
	registerRequest := NaviRegisterRequest{}
	if err := json.Unmarshal(body, &registerRequest); err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Cannot parse RegisterRequest")
		http.Error(writer, "Internal error", http.StatusInternalServerError)

		return
	}

	node := t.mirage.GetNaviNode(registerRequest.ID)
	if node == nil {
		log.Warn().Caller().Msgf("Navi node %s not found", registerRequest.ID)
		http.Error(writer, "Navi node not found", http.StatusNotFound)
		return
	}
	if node.NaviKey == "" || node.NaviKey == MachinePublicKeyStripPrefix(t.conn.Peer()) {
		node.NaviKey = MachinePublicKeyStripPrefix(t.conn.Peer())
		node := t.mirage.UpdateNaviNode(node)
		if node == nil {
			log.Warn().Caller().Msgf("Navi node %s update failed", registerRequest.ID)
			http.Error(writer, "Internal error", http.StatusInternalServerError)
			return
		}
		log.Trace().Caller().Msgf("Navi node %s registered", node.ID)
		now := time.Now().Round(time.Second)
		resp := NaviRegisterResponse{
			NodeInfo:  *node,
			Timestamp: &now,
		}
		respBody, err := t.mirage.marshalResponse(resp, t.conn.Peer())
		if err != nil {
			log.Error().
				Caller().
				Str("func", "handleNaviRegister").
				Err(err).
				Msg("Cannot encode message")
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)
		_, err = writer.Write(respBody)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		log.Info().
			Str("func", "handleNaviRegister").
			Str("derpID", registerRequest.ID).
			Msg("Successfully register Navi node")

		return
	}

	log.Error().
		Caller().
		Msg("Navi node not created yet or key mismatch")
	http.Error(writer, "Internal error", http.StatusInternalServerError)
	return
}
