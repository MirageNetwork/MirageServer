package controller

import (
	"encoding/binary"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/rs/zerolog/log"
	"tailscale.com/smallzstd"
	"tailscale.com/tailcfg"
	"tailscale.com/types/key"
)

func (h *Mirage) generateMapResponse(
	mapRequest tailcfg.MapRequest,
	machine *Machine,
) (*tailcfg.MapResponse, error) {
	log.Trace().
		Str("func", "generateMapResponse").
		Str("machine", mapRequest.Hostinfo.Hostname).
		Msg("Creating Map response")

	//cgao6: change to use User's DNSConfig
	node, err := h.toNode(*machine) //h.cfg.BaseDomain, h.cfg.DNSConfig)
	if err != nil {
		log.Error().
			Caller().
			Str("func", "generateMapResponse").
			Err(err).
			Msg("Cannot convert to node")

		return nil, err
	}

	peers, invalidNodeIDs, err := h.getValidPeers(machine)
	if invalidNodeIDs != nil {
		//log.Info().Msg("Should ignore invalidNodeIDs for current")
	}
	if err != nil {
		log.Error().
			Caller().
			Str("func", "generateMapResponse").
			Err(err).
			Msg("Cannot fetch peers")

		return nil, err
	}

	profiles := h.getMapResponseUserProfiles(*machine, peers)

	//cgao6: change to use User's DNSConfig
	nodePeers, err := h.toNodes(peers) //, h.cfg.BaseDomain, h.cfg.DNSConfig)
	if err != nil {
		log.Error().
			Caller().
			Str("func", "generateMapResponse").
			Err(err).
			Msg("Failed to convert peers to Tailscale nodes")

		return nil, err
	}

	//cgao6: use User's DNSconfig instead
	dnsConfig := getMapResponseDNSConfig(
		h.cfg.IPPrefixes, //
		//		h.cfg.DNSConfig,
		//		h.cfg.BaseDomain,
		*machine,
		peers,
	)

	now := time.Now()
	var org *Organization
	if machine.User.Organization.AclPolicy != nil {
		org = &machine.User.Organization
	} else {
		org, err = h.GetOrgnaizationByID(machine.User.OrganizationID)
		if err != nil {
			log.Error().
				Caller().
				Str("func", "generateMapResponse").
				Err(err).
				Msg("Failed to get organization of machine")

			return nil, err
		}
	}

	derpMap, err := h.LoadOrgDERPs(machine.User.OrganizationID)
	if err != nil {
		log.Error().
			Caller().
			Str("func", "generateMapResponse").
			Err(err).
			Msg("Failed to get DERP map of machine")
	}

	err = h.checkAndHandleAutogroupRules(machine, org)
	if err != nil {
		log.Error().
			Caller().
			Str("func", "generateMapResponse").
			Err(err).
			Msg("Failed to get machines of the user")
	}

	resp := tailcfg.MapResponse{
		KeepAlive: false,
		Node:      node,

		// TODO: Only send if updated
		DERPMap: derpMap, //cgao6: h.DERPMap,

		// TODO: Only send if updated
		Peers: nodePeers,

		// TODO(kradalby): Implement:
		// https://github.com/tailscale/tailscale/blob/main/tailcfg/tailcfg.go#L1351-L1374
		// PeersChanged
		// PeersRemoved
		// PeersChangedPatch
		// PeerSeenChange
		// OnlineChange

		// TODO: Only send if updated
		DNSConfig: dnsConfig,

		// TODO: Only send if updated
		Domain: org.Name,

		// Do not instruct clients to collect services, we do not
		// support or do anything with them
		CollectServices: "false",

		// TODO: Only send if updated
		PacketFilter: org.AclRules,

		UserProfiles: profiles,

		// TODO: Only send if updated
		SSHPolicy: org.SshPolicy,

		ControlTime: &now,

		Debug: &tailcfg.Debug{
			DisableLogTail:         true,
			SetRandomizeClientPort: "true",
		},
	}

	resp.ClientVersion = &tailcfg.ClientVersion{}

	switch true {
	case strings.Contains(mapRequest.Hostinfo.OS, "windows"):
		resp.ClientVersion.LatestVersion = h.cfg.ClientVersion.Win.Version
		resp.ClientVersion.NotifyURL = h.cfg.ClientVersion.Win.Url
	}

	log.Trace().
		Str("func", "generateMapResponse").
		Str("machine", mapRequest.Hostinfo.Hostname).
		// Interface("payload", resp).
		Msgf("Generated map response: %s", tailMapResponseToString(resp))

	return &resp, nil
}

func (h *Mirage) getMapResponseData(
	mapRequest tailcfg.MapRequest,
	machine *Machine,
) ([]byte, error) {
	mapResponse, err := h.generateMapResponse(mapRequest, machine)
	if err != nil {
		return nil, err
	}

	return h.marshalMapResponse(mapResponse, key.MachinePublic{}, mapRequest.Compress)

}

func (h *Mirage) getMapKeepAliveResponseData(
	mapRequest tailcfg.MapRequest,
	machine *Machine,
) ([]byte, error) {
	keepAliveResponse := tailcfg.MapResponse{
		KeepAlive: true,
	}

	return h.marshalMapResponse(keepAliveResponse, key.MachinePublic{}, mapRequest.Compress)

}

func (h *Mirage) marshalResponse(
	resp interface{},
	machineKey key.MachinePublic,
) ([]byte, error) {
	jsonBody, err := json.Marshal(resp)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Cannot marshal response")

		return nil, err
	}

	return jsonBody, nil

}

func (h *Mirage) marshalMapResponse(
	resp interface{},
	machineKey key.MachinePublic,
	compression string,
) ([]byte, error) {
	jsonBody, err := json.Marshal(resp)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Cannot marshal map response")
	}

	var respBody []byte
	if compression == ZstdCompression {
		respBody = zstdEncode(jsonBody)
	} else {
		respBody = jsonBody
	}

	data := make([]byte, reservedResponseHeaderSize)
	binary.LittleEndian.PutUint32(data, uint32(len(respBody)))
	data = append(data, respBody...)

	return data, nil
}

func zstdEncode(in []byte) []byte {
	encoder, ok := zstdEncoderPool.Get().(*zstd.Encoder)
	if !ok {
		panic("invalid type in sync pool")
	}
	out := encoder.EncodeAll(in, nil)
	_ = encoder.Close()
	zstdEncoderPool.Put(encoder)

	return out
}

var zstdEncoderPool = &sync.Pool{
	New: func() any {
		encoder, err := smallzstd.NewEncoder(
			nil,
			zstd.WithEncoderLevel(zstd.SpeedFastest))
		if err != nil {
			panic(err)
		}

		return encoder
	},
}
