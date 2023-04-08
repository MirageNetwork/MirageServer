package controller

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// 发送可信节点变更请求
func (m *Mirage) sendNodesChange(navi *NaviNode, addNode, removeNode string) error {
	m.DERPseqnum[navi.ID]++
	request := NodesChange{
		SeqNum:     m.DERPseqnum[navi.ID],
		AddNode:    addNode,
		RemoveNode: removeNode,
	}
	url := fmt.Sprintf("https://%s:%d/ctrl/nodes", navi.HostName, navi.DERPPort)
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

// 通知租户内（及全局）司南可信节点变更
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

// 获取租户内节点NodeKey列表
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

type NaviStatus struct {
	CounterUptimeSec uint64 `json:"counter_uptime_sec"`
	Derp             struct {
		Accepts                     uint64 `json:"accepts"`
		BytesReceived               uint64 `json:"bytes_received"`
		BytesSent                   uint64 `json:"bytes_sent"`
		CounterPacketsDroppedReason struct {
			GoneDisconnected uint64 `json:"gone_disconnected"`
			GoneNotHere      uint64 `json:"gone_not_here"`
			UnknownDest      uint64 `json:"unknown_dest"`
			UnknownDestOnFwd uint64 `json:"unknown_dest_on_fwd"`
			WriteError       uint64 `json:"write_error"`
		} `json:"counter_packets_dropped_reason"`
		CounterTotalDupClientConns uint64 `json:"counter_total_dup_client_conns"`
		GaugeClientsLocal          uint64 `json:"gauge_clients_local"`
		GaugeClientsTotal          uint64 `json:"gauge_clients_total"`
		GaugeCurrentConnections    uint64 `json:"gauge_current_connections"`
		GotPing                    uint64 `json:"got_ping"`
		SentPong                   uint64 `json:"sent_pong"`
		HomeMovesIn                uint64 `json:"home_moves_in"`
		HomeMovesOut               uint64 `json:"home_moves_out"`
		PacketsDropped             uint64 `json:"packets_dropped"`
		PacketsForwarded_in        uint64 `json:"packets_forwarded_in"`
		PacketsForwarded_out       uint64 `json:"packets_forwarded_out"`
		PacketsReceived            uint64 `json:"packets_received"`
		PacketsSent                uint64 `json:"packets_sent"`
		Version                    string `json:"version"`
	} `json:"derp"`
	Latency int64 `json:"latency"`
}

func (ns *NaviStatus) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, ns)
	case string:
		return json.Unmarshal([]byte(src), ns)
	default:
		return fmt.Errorf("cannot convert %T to NaviStatus", src)
	}
}

func (ns NaviStatus) Value() (driver.Value, error) {
	return json.Marshal(ns)
}

func (m *Mirage) updateNaviStatus(navi *NaviNode) error {
	req204, err := http.NewRequestWithContext(m.ctx, "GET", fmt.Sprintf("https://%s:%d/generate_204", navi.HostName, navi.DERPPort), nil)
	if err != nil {
		return fmt.Errorf("update navi status request: %w", err)
	}
	start := time.Now()
	res204, err := http.DefaultClient.Do(req204)
	latency := time.Since(start)
	if err != nil {
		navi.Statics = NaviStatus{
			Latency: -1,
		}
		m.db.Save(navi)
		return fmt.Errorf("update navi status request: %w", err)
	}

	if res204.StatusCode != http.StatusNoContent {
		msg, _ := io.ReadAll(res204.Body)
		res204.Body.Close()
		navi.Statics = NaviStatus{
			Latency: -1,
		}
		m.db.Save(navi)
		return fmt.Errorf("update navi status request: http %d: %.200s",
			res204.StatusCode, strings.TrimSpace(string(msg)))
	}

	if navi.NaviKey == "" {
		//TODO: 非受控节点只检查204状态
		navi.Statics = NaviStatus{
			Latency: latency.Milliseconds(),
		}
		err = m.db.Save(navi).Error
		return err
	}

	url := fmt.Sprintf("https://%s:%d/ctrl/vars", navi.HostName, navi.DERPPort)
	req, err := http.NewRequestWithContext(m.ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("update navi status request: %w", err)
	}
	res, err := m.DERPNCs[navi.ID].Do(req)
	if err != nil {
		return fmt.Errorf("update navi status request: %w", err)
	}
	if res.StatusCode != 200 {
		msg, _ := io.ReadAll(res.Body)
		res.Body.Close()
		return fmt.Errorf("update navi status request: http %d: %.200s",
			res.StatusCode, strings.TrimSpace(string(msg)))
	}

	var status NaviStatus
	err = decodeNoiseResponse(res, &status)
	if err != nil {
		return fmt.Errorf("update navi status request: %w", err)
	}

	log.Debug().Msg(fmt.Sprintf("Navi %s status: %v", navi.HostName, status))
	navi.Statics = status
	navi.Statics.Latency = latency.Milliseconds()
	m.db.Save(navi)

	return nil
}

func (m *Mirage) refreshAllNaviStatus() {
	nrs := m.ListNaviRegions()
	for _, nr := range nrs {
		nns := m.ListNaviNodes(nr.ID)
		for _, nn := range nns {
			err := m.updateNaviStatus(&nn)
			if err != nil {
				log.Error().
					Caller().
					Err(err).
					Msg("Cannot update navi status")
			}
		}
	}
}

func (m *Mirage) refreshNaviStatusPoller(ticker *time.Ticker) {
	for range ticker.C {
		m.refreshAllNaviStatus()
	}
}
