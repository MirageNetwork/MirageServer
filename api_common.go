package headscale

import (
	"github.com/rs/zerolog/log"
	"tailscale.com/tailcfg"
)

func (h *Headscale) generateMapResponse(
	mapRequest tailcfg.MapRequest,
	machine *Machine,
) (*tailcfg.MapResponse, error) {
	log.Trace().
		Str("func", "generateMapResponse").
		Str("machine", mapRequest.Hostinfo.Hostname).
		Msg("Creating Map response")
	node, err := h.toNode(*machine, h.cfg.BaseDomain, h.cfg.DNSConfig)
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

	nodePeers, err := h.toNodes(peers, h.cfg.BaseDomain, h.cfg.DNSConfig)
	if err != nil {
		log.Error().
			Caller().
			Str("func", "generateMapResponse").
			Err(err).
			Msg("Failed to convert peers to Tailscale nodes")

		return nil, err
	}

	dnsConfig := getMapResponseDNSConfig(
		h.cfg.DNSConfig,
		h.cfg.BaseDomain,
		*machine,
		peers,
	)

	resp := tailcfg.MapResponse{
		KeepAlive: false,
		Node:      node,
		Peers:     nodePeers,
		//		PeersChanged: nodePeers,
		//		PeersRemoved: invalidNodeIDs,
		DNSConfig:    dnsConfig,
		Domain:       h.cfg.BaseDomain,
		PacketFilter: h.aclRules,
		SSHPolicy:    h.sshPolicy,
		DERPMap:      h.DERPMap,
		UserProfiles: profiles,
		Debug: &tailcfg.Debug{
			DisableLogTail:      !h.cfg.LogTail.Enabled,
			RandomizeClientPort: h.cfg.RandomizeClientPort,
		},
	}

	log.Trace().
		Str("func", "generateMapResponse").
		Str("machine", mapRequest.Hostinfo.Hostname).
		// Interface("payload", resp).
		Msgf("Generated map response: %s", tailMapResponseToString(resp))

	return &resp, nil
}
