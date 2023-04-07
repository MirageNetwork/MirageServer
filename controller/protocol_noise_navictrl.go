package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"tailscale.com/control/controlclient"
	"tailscale.com/net/tsdial"
	"tailscale.com/types/key"
	"tailscale.com/util/singleflight"
)

// 生成连接司南的Noise客户端
func (m *Mirage) GetNaviNoiseClient(naviPub key.MachinePublic, naviHostname string, naviDERPPort int) (*controlclient.NoiseClient, error) {
	dialer := &tsdial.Dialer{Logf: log.Logger.Printf}
	var sfGroup singleflight.Group[struct{}, *controlclient.NoiseClient]
	nc, err, _ := sfGroup.Do(struct{}{}, func() (*controlclient.NoiseClient, error) {
		log.Trace().Caller().Msg("creating new noise client")
		var nc *controlclient.NoiseClient
		var err error
		if naviDERPPort == 0 {
			nc, err = controlclient.NewNoiseClient(*m.noisePrivateKey, naviPub, "https://"+naviHostname, dialer, nil)
		} else {
			nc, err = controlclient.NewNoiseClient(*m.noisePrivateKey, naviPub, "https://"+naviHostname+":"+strconv.Itoa(naviDERPPort), dialer, nil)
		}
		if err != nil {
			return nil, err
		}
		return nc, nil
	})
	if err != nil {
		return nil, err
	}
	return nc, nil
}

func decodeNoiseResponse(res *http.Response, v any) error {
	defer res.Body.Close()
	msg, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("%d: %v", res.StatusCode, string(msg))
	}
	return json.Unmarshal(msg, v)
}

type NodesChange struct {
	SeqNum     int
	AddNode    string
	RemoveNode string
}

func (m *Mirage) sendNodesChange(navi *NaviNode, addNode, removeNode string) error {
	m.DERPseqnum[navi.ID]++
	request := NodesChange{
		SeqNum:     m.DERPseqnum[navi.ID],
		AddNode:    addNode,
		RemoveNode: removeNode,
	}
	url := fmt.Sprintf("https://%s/ctrl/nodes", navi.HostName)
	bodyData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("node change request: %w", err)
	}
	req, err := http.NewRequestWithContext(m.ctx, "POST", url, bytes.NewReader(bodyData))
	if err != nil {
		return fmt.Errorf("node change request: %w", err)
	}
	res, err := m.DERPNCs[navi.ID].Do(req)
	if err != nil {
		return fmt.Errorf("node change request: %w", err)
	}
	if res.StatusCode != 200 {
		msg, _ := io.ReadAll(res.Body)
		res.Body.Close()
		return fmt.Errorf("node change request: http %d: %.200s",
			res.StatusCode, strings.TrimSpace(string(msg)))
	}
	return nil
}

func (m *Mirage) NotifyNaviOrgNodesChange(orgID int64, addNode, removeNode string) {
	//TODO
	nrs := m.ListNaviRegions()
	for _, nr := range nrs {
		if nr.OrgID == orgID || nr.OrgID == 0 {
			nns := m.ListNaviNodes(nr.ID)
			for _, nn := range nns {
				err := m.sendNodesChange(&nn, addNode, removeNode)
				if err != nil {
					log.Error().
						Caller().
						Err(err).
						Msg("Cannot send nodes change")
				}
			}
		}
	}
}

func (m *Mirage) getOrgNodesKey(orgID int64) ([]string, error) {
	var machines []Machine
	var err error
	if orgID == 0 {
		machines, err = m.ListMachines()
	} else {
		machines, err = m.ListMachinesByOrgID(orgID)
	}
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Cannot list machines")
		return nil, err
	}

	nodeList := make([]string, 0)
	for _, machine := range machines {
		nodeList = append(nodeList, machine.NodeKey)
	}
	return nodeList, nil
}
