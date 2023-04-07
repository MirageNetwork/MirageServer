package controller

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"tailscale.com/tailcfg"
	"tailscale.com/types/key"
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

	if registerRequest.Auth.Provider == "Mirage" {
		t.mirage.handleRegisterNavi(writer, req, registerRequest, t.conn.Peer())
		return
	}

	t.mirage.handleRegisterCommon(writer, req, registerRequest, t.conn.Peer())
}

type NaviRegisterResponse struct {
	NaviInfo   NaviNode
	TrustNodes []string
	Timestamp  *time.Time
}

// 司南注册noise协议接口
func (m *Mirage) handleRegisterNavi(
	writer http.ResponseWriter,
	req *http.Request,
	registerRequest tailcfg.RegisterRequest,
	naviKey key.MachinePublic,
) {
	log.Trace().Caller().Msgf("Noise registration handler for Navi %s", req.RemoteAddr)

	node := m.GetNaviNode(registerRequest.Auth.LoginName)
	if node == nil {
		log.Warn().Caller().Msgf("Navi node %s not found", registerRequest.Auth.LoginName)
		http.Error(writer, "Navi node not found", http.StatusNotFound)
		return
	}
	if node.NaviKey == "" || node.NaviKey == MachinePublicKeyStripPrefix(naviKey) {
		node.NaviKey = MachinePublicKeyStripPrefix(naviKey)
		node := m.UpdateNaviNode(node)
		if node == nil {
			log.Warn().Caller().Msgf("Navi node %s update failed", registerRequest.Auth.LoginName)
			http.Error(writer, "Internal error", http.StatusInternalServerError)
			return
		}
		log.Trace().Caller().Msgf("Navi node %s registered", node.ID)

		trustNodesKeys, err := m.getOrgNodesKey(node.NaviRegion.OrgID)
		if err != nil {
			log.Error().Caller().Err(err).Msg("Failed to get trust nodes key")
			http.Error(writer, "Internal error", http.StatusInternalServerError)
			return
		}
		now := time.Now().Round(time.Second)
		resp := NaviRegisterResponse{
			NaviInfo:   *node,
			TrustNodes: trustNodesKeys,
			Timestamp:  &now,
		}
		respBody, err := m.marshalResponse(resp, naviKey)
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

		nc, err := m.GetNaviNoiseClient(naviKey, node.HostName, node.DERPPort)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to get Navi Noise client")
			return
		}
		m.DERPNCs[node.ID] = nc
		m.DERPseqnum[node.ID] = 0 //初始化Noise client及序列号

		_, err = writer.Write(respBody)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		log.Info().
			Str("func", "handleNaviRegister").
			Str("derpID", registerRequest.Auth.LoginName).
			Msg("Successfully register Navi node")

		return
	}

	log.Error().
		Caller().
		Msg("Navi node not created yet or key mismatch")
	http.Error(writer, "Internal error", http.StatusInternalServerError)
	return
}

type PullNodesListResponse struct {
	TrustNodes []string   `json:"TrustNodes"`
	Timestamp  *time.Time `json:"Timestamp"`
}

func (t *ts2021App) NoiseNaviPollNodesListHandler(
	writer http.ResponseWriter,
	req *http.Request,
) {
	log.Trace().Caller().Msgf("Noise NodesListPoll handler for Navi %s", req.RemoteAddr)
	if req.Method != http.MethodPost {
		http.Error(writer, "Wrong method", http.StatusMethodNotAllowed)

		return
	}
	body, _ := io.ReadAll(req.Body)
	pollReq := tailcfg.MapRequest{}
	if err := json.Unmarshal(body, &pollReq); err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Cannot parse PollNodesListRequest")
		http.Error(writer, "Internal error", http.StatusInternalServerError)

		return
	}

	node := t.mirage.GetNaviNode(pollReq.Hostinfo.BackendLogID)
	if node == nil {
		log.Warn().Caller().Msgf("Navi node %s not found", pollReq.Hostinfo.BackendLogID)
		http.Error(writer, "Navi node not found", http.StatusNotFound)
		return
	}
	if node.NaviKey == MachinePublicKeyStripPrefix(t.conn.Peer()) {
		trustNodesKeys, err := t.mirage.getOrgNodesKey(node.NaviRegion.OrgID)
		if err != nil {
			log.Error().Caller().Err(err).Msg("Failed to get trust nodes key")
			http.Error(writer, "Internal error", http.StatusInternalServerError)
			return
		}
		now := time.Now().Round(time.Second)
		resp := PullNodesListResponse{
			TrustNodes: trustNodesKeys,
			Timestamp:  &now,
		}

		respBody, err := t.mirage.marshalResponse(resp, t.conn.Peer())
		if err != nil {
			log.Error().
				Caller().
				Str("func", "NoiseNaviPollNodesListHandler").
				Err(err).
				Msg("Cannot encode message")
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusOK)

		t.mirage.DERPseqnum[node.ID] = 0 //序列号置零

		_, err = writer.Write(respBody)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		log.Info().
			Str("func", "NoiseNaviPollNodesListHandler").
			Str("derpID", pollReq.Hostinfo.BackendLogID).
			Msg("Successfully return Navi trust nodes list")
		return
	}

	log.Error().
		Caller().
		Msg("Navi node not created yet or key mismatch")
	http.Error(writer, "Internal error", http.StatusInternalServerError)
	return
}
