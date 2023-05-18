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

// mapResponseStreamState tracks state associated with a stream of MapResponse messages,
// which may optionally send only deltas from the previous message.
type mapResponseStreamState struct {
	// peersByID is the peers sent in the last stream message,
	// for comparison in generating deltas in the new message.
	peersByID map[int64]Machine
}

func (h *Mirage) generateMapResponse(
	mapRequest tailcfg.MapRequest,
	machine *Machine,
	streamState *mapResponseStreamState,
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

	// in this function, we get the peers and the organization(with aclRules updated), organization was set to field :machine.User.Organization
	peers, invalidNodeIDs, err := h.getValidPeers(machine)
	if invalidNodeIDs != nil {
		log.Trace().Msg("Should ignore invalidNodeIDs for current")
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

	//cgao6: use User's DNSconfig instead
	dnsConfig := getMapResponseDNSConfig(
		h.cfg.IPPrefixes, //
		//		h.cfg.DNSConfig,
		//		h.cfg.BaseDomain,
		*machine,
		peers,
	)

	now := time.Now()
	org := &machine.User.Organization

	derpMap, err := h.LoadOrgDERPs(machine.User.OrganizationID)
	if err != nil {
		log.Error().
			Caller().
			Str("func", "generateMapResponse").
			Err(err).
			Msg("Failed to get DERP map of machine")
	}

	resp := tailcfg.MapResponse{
		KeepAlive: false,
		Node:      node,

		// TODO: Only send if updated
		DERPMap: derpMap, //cgao6: h.DERPMap,

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

	toNodes := func(machines Machines) ([]*tailcfg.Node, error) {
		return h.toNodes(machines) //, h.cfg.BaseDomain, h.cfg.DNSConfig)
	}
	resp, err = applyMapResponseDelta(resp, streamState, peers, toNodes)
	if err != nil {
		log.Error().
			Caller().
			Str("func", "generateMapResponse").
			Err(err).
			Msg("Cannot apply map response deltas")

		return nil, err
	}
	resp.ClientVersion = &tailcfg.ClientVersion{}

	if mapRequest.Hostinfo.OS == "windows" {
		if IsUpdateAvailable(mapRequest.Hostinfo.IPNVersion, h.cfg.ClientVersion.Win.Version) {
			resp.ClientVersion.RunningLatest = false
			resp.ClientVersion.LatestVersion = strings.Split(h.cfg.ClientVersion.Win.Version, "-")[0]
			resp.ClientVersion.NotifyURL = h.cfg.ClientVersion.Win.Url
		} else {
			resp.ClientVersion.RunningLatest = true
		}
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
	streamState *mapResponseStreamState,
) ([]byte, error) {
	mapResponse, err := h.generateMapResponse(mapRequest, machine, streamState)
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

// applyMapResponseDelta returns a modified MapResponse
// with fields modified which make use of delta (send on changes).
//
// mapResponse the current mapResponse with delta fields not set.
// streamState optional previous state of mapResponse sent in this stream. Set to nil for a "full update" (no deltas).
// currentPeers list of peers currently available for the node that this mapResponse is for.
// toNodes a function to convert the Headscale Machines structure to Tailscale Nodes structure.
func applyMapResponseDelta(
	mapResponse tailcfg.MapResponse,
	streamState *mapResponseStreamState,
	currentPeers Machines,
	toNodes func(Machines) ([]*tailcfg.Node, error)) (tailcfg.MapResponse, error) {

	// Peer delta
	currentPeersByID := machinesByID(currentPeers)

	if streamState.peersByID == nil {
		// 1st map, send full nodes
		nodePeers, err := toNodes(currentPeers)
		if err != nil {
			return tailcfg.MapResponse{}, err
		}
		mapResponse.Peers = nodePeers
	} else {
		// Update PeersChanged with any peers which were removed or changed
		var peersChanged []Machine
		for id, peer := range currentPeersByID {
			previousPeer, hadPrevious := streamState.peersByID[id]
			if !hadPrevious || previousPeer.LastSuccessfulUpdate.Before(*peer.LastSuccessfulUpdate) {
				peersChanged = append(peersChanged, peer)
			}
		}
		nodesChanged, err := toNodes(peersChanged)
		if err != nil {
			return tailcfg.MapResponse{}, err
		}
		mapResponse.PeersChanged = nodesChanged

		// Update PeersRemoved with any peers which are no longer present
		for id := range streamState.peersByID {
			if _, has := currentPeersByID[id]; !has {
				mapResponse.PeersRemoved = append(mapResponse.PeersRemoved, tailcfg.NodeID(id))
			}
		}
	}

	// Update streamState for use in the next message
	streamState.peersByID = currentPeersByID

	// TODO(kallen): Also Implement the following deltas for even smaller
	// message sizes:
	//
	// PeersChangedPatch
	// PeerSeenChange
	// OnlineChange

	return mapResponse, nil
}
