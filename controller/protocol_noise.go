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
	NodeInfo  NaviNode
	Timestamp *time.Time
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
		now := time.Now().Round(time.Second)
		resp := NaviRegisterResponse{
			NodeInfo:  *node,
			Timestamp: &now,
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
	TrustNodesList map[string]string `json:"TrustNodesList"`
	Timestamp      *time.Time        `json:"Timestamp"`
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
		var machines []Machine
		var err error
		if node.NaviRegion.OrgID == 0 {
			machines, err = t.mirage.ListMachines()
		} else {
			machines, err = t.mirage.ListMachinesByOrgID(node.NaviRegion.OrgID)
		}
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Cannot list machines")
			http.Error(writer, "Internal error", http.StatusInternalServerError)
			return
		}
		log.Trace().Caller().Msgf("Navi node list for  %s prepared", node.ID)

		nodeList := make(map[string]string)
		for _, machine := range machines {
			nodeList[NodePublicKeyEnsurePrefix(machine.NodeKey)] = ""
		}
		now := time.Now().Round(time.Second)
		resp := PullNodesListResponse{
			TrustNodesList: nodeList,
			Timestamp:      &now,
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
