package controller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"tailscale.com/tailcfg"
)

type NaviRegion struct {
	ID         int    `gorm:"primary_key;unique;not null" json:"RegionID"`
	OrgID      int64  `gorm:";not null" json:"OrgID"` // 0代表全局向导
	RegionCode string `gorm:"not null" json:"RegionCode"`
	RegionName string `gorm:"not null" json:"RegionName"`
	//这个不知道有何用 Avoid bool `json:",omitempty"`
}
type NaviNode struct {
	ID      string `gorm:"primary_key;unique;not null" json:"Name"` //映射到DERPNode的Name
	NaviKey string `json:"NaviKey"`                                 //记录DERPNode的MachineKey公钥

	NaviRegionID int         `gorm:"not null" json:"RegionID"`                       //映射到DERPNode的RegionID
	NaviRegion   *NaviRegion `gorm:"foreignKey:NaviRegionID;references:ID" json:"-"` //映射到DERPNode的RegionID

	HostName string `json:"HostName"` //这个不需要独有，但是否必须域名呢？
	//这个不用？ CertName string `json:",omitempty"`

	IPv4 string `json:"IPv4"` // 不是ipv4地址则失效，为none则禁用ipv4
	IPv6 string `json:"IPv6"` // 不是ipv6地址则失效，为none则禁用ipv6

	NoSTUN   bool `json:"NoSTUN"`   //禁用STUN
	STUNPort int  `json:"STUNPort"` //0代表3478，-1代表禁用

	NoDERP   bool `json:"NoDERP"`   //禁用DERP
	DERPPort int  `json:"DERPPort"` //0代表443

	SSHAddr     string `json:"SSHAddr"`     //SSH地址
	SSHPwd      string `json:"SSHPwd"`      //SSH口令
	DNSProvider string `json:"DNSProvider"` //DNS服务商
	DNSID       string `json:"DNSID"`       //DNS服务商的ID
	DNSKey      string `json:"DNSKey"`      //DNS服务商的Key

	Arch    string     `json:"Arch"` //所在环境架构，x86_64或aarch64
	Statics NaviStatus `json:"Statics"`
}

func (c *Cockpit) toDERPRegion(nr NaviRegion) (tailcfg.DERPRegion, error) {
	nodes := c.ListNaviNodes(nr.ID)
	derpNodes, err := c.toDERPNodes(nodes)
	if err != nil {
		return tailcfg.DERPRegion{}, err
	}
	return tailcfg.DERPRegion{
		RegionID:   nr.ID,
		RegionCode: nr.RegionCode,
		RegionName: nr.RegionName,
		Nodes:      derpNodes,
	}, nil
}

func (m *Cockpit) toDERPNodes(nodes []NaviNode) ([]*tailcfg.DERPNode, error) {
	derpNodes := make([]*tailcfg.DERPNode, len(nodes))
	for index, node := range nodes {
		derpNode, err := m.toDERPNode(node)
		if err != nil {
			return nil, err
		}
		derpNodes[index] = derpNode
	}
	return derpNodes, nil

}

func (c *Cockpit) toDERPNode(node NaviNode) (*tailcfg.DERPNode, error) {
	derp := &tailcfg.DERPNode{
		Name:     node.ID,
		RegionID: node.NaviRegionID,
		HostName: node.HostName,
		IPv4:     node.IPv4,
		IPv6:     node.IPv6,
		STUNPort: node.STUNPort,
		STUNOnly: node.NoDERP,
		DERPPort: node.DERPPort,
	}
	if node.NoSTUN {
		derp.STUNPort = -1
	}
	return derp, nil
}

func (c *Cockpit) ListNaviRegions() []NaviRegion {
	naviRegions := []NaviRegion{}
	if err := c.db.Find(&naviRegions).Error; err != nil {
		return nil
	}
	return naviRegions
}

func (c *Cockpit) GetNaviRegion(id int) *NaviRegion {
	naviRegion := NaviRegion{}
	if err := c.db.First(&naviRegion, id).Error; err != nil {
		return nil
	}
	return &naviRegion
}

func (c *Cockpit) CreateNaviRegion(naviRegion *NaviRegion) *NaviRegion {
	if err := c.db.Create(naviRegion).Error; err != nil {
		return nil
	}
	return naviRegion
}

func (c *Cockpit) UpdateNaviRegion(naviRegion *NaviRegion) *NaviRegion {
	if err := c.db.Save(naviRegion).Error; err != nil {
		return nil
	}
	return naviRegion
}

func (c *Cockpit) ListNaviNodes(regionID int) []NaviNode {
	naviNodes := []NaviNode{}
	if err := c.db.Preload("NaviRegion").Where("navi_region_id = ?", regionID).Find(&naviNodes).Error; err != nil {
		return nil
	}
	return naviNodes
}

func (c *Cockpit) GetNaviNode(id string) *NaviNode {
	naviNode := NaviNode{}
	if err := c.db.Preload("NaviRegion").First(&naviNode, "id = ?", id).Error; err != nil {
		return nil
	}
	return &naviNode
}

func (c *Cockpit) CreateNaviNode(naviNode *NaviNode) *NaviNode {
	if err := c.db.Create(naviNode).Error; err != nil {
		return nil
	}
	return naviNode
}

func (c *Cockpit) UpdateNaviNode(naviNode *NaviNode) *NaviNode {
	if err := c.db.Save(naviNode).Error; err != nil {
		return nil
	}
	return naviNode
}

// cgao6: 以下为Mirage的实现
func (m *Mirage) LoadDERPMapFromURL(addr string) (*tailcfg.DERPMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), HTTPReadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout: HTTPReadTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var derpMap tailcfg.DERPMap
	err = json.Unmarshal(body, &derpMap)

	if len(derpMap.Regions) == 0 {
		log.Warn().
			Msg("DERP map is empty, not a single DERP map datasource was loaded correctly or contained a region")
	}

	//cgao6: TEMP
	derpMap = tailcfg.DERPMap{
		Regions: make(map[int]*tailcfg.DERPRegion),
	} //TODO: 临时

	// 从数据库读取DERP信息
	naviRegions := m.ListNaviRegions()
	if len(naviRegions) != 0 {
		for _, nr := range naviRegions {
			derpRegion, err := m.toDERPRegion(nr)
			if err != nil {
				log.Error().Err(err).Msg("Cannot convert NaviRegion to DERPRegion")
				return nil, err
			}
			derpMap.Regions[derpRegion.RegionID] = &derpRegion
		}
	}

	return &derpMap, err
}

func (m *Mirage) LoadOrgDERPs(orgID int64) (*tailcfg.DERPMap, error) {
	derpMap := &tailcfg.DERPMap{
		Regions: make(map[int]*tailcfg.DERPRegion),
	}

	// 从数据库读取DERP信息
	naviRegions := m.ListNaviRegions()
	if len(naviRegions) != 0 {
		for _, nr := range naviRegions {
			if nr.OrgID == 0 || nr.OrgID == orgID {
				derpRegion, err := m.toDERPRegion(nr)
				if err != nil {
					log.Error().Err(err).Msg("Cannot convert NaviRegion to DERPRegion")
					return nil, err
				}
				derpMap.Regions[derpRegion.RegionID] = &derpRegion
			}
		}
	}

	return derpMap, nil
}

func (m *Mirage) toDERPRegion(nr NaviRegion) (tailcfg.DERPRegion, error) {
	nodes := m.ListNaviNodes(nr.ID)
	derpNodes, err := m.toDERPNodes(nodes)
	if err != nil {
		return tailcfg.DERPRegion{}, err
	}
	return tailcfg.DERPRegion{
		RegionID:   nr.ID,
		RegionCode: nr.RegionCode,
		RegionName: nr.RegionName,
		Nodes:      derpNodes,
	}, nil
}
func (m *Mirage) toDERPNodes(nodes []NaviNode) ([]*tailcfg.DERPNode, error) {
	derpNodes := make([]*tailcfg.DERPNode, len(nodes))
	for index, node := range nodes {
		derpNode, err := m.toDERPNode(node)
		if err != nil {
			return nil, err
		}
		derpNodes[index] = derpNode
	}
	return derpNodes, nil

}

func (m *Mirage) toDERPNode(node NaviNode) (*tailcfg.DERPNode, error) {
	derp := &tailcfg.DERPNode{
		Name:     node.ID,
		RegionID: node.NaviRegionID,
		HostName: node.HostName,
		IPv4:     node.IPv4,
		IPv6:     node.IPv6,
		STUNPort: node.STUNPort,
		STUNOnly: node.NoDERP,
		DERPPort: node.DERPPort,
	}
	if node.NoSTUN {
		derp.STUNPort = -1
	}
	return derp, nil
}

func (m *Mirage) ListNaviRegions() []NaviRegion {
	naviRegions := []NaviRegion{}
	if err := m.db.Find(&naviRegions).Error; err != nil {
		return nil
	}
	return naviRegions
}

func (m *Mirage) GetNaviRegion(id int) *NaviRegion {
	naviRegion := NaviRegion{}
	if err := m.db.First(&naviRegion, id).Error; err != nil {
		return nil
	}
	return &naviRegion
}

func (m *Mirage) CreateNaviRegion(naviRegion *NaviRegion) *NaviRegion {
	if err := m.db.Create(naviRegion).Error; err != nil {
		return nil
	}
	return naviRegion
}
func (m *Mirage) UpdateNaviRegion(naviRegion *NaviRegion) *NaviRegion {
	if err := m.db.Save(naviRegion).Error; err != nil {
		return nil
	}
	return naviRegion
}

func (m *Mirage) ListNaviNodes(regionID int) []NaviNode {
	naviNodes := []NaviNode{}
	if err := m.db.Preload("NaviRegion").Where("navi_region_id = ?", regionID).Find(&naviNodes).Error; err != nil {
		return nil
	}
	return naviNodes
}

func (m *Mirage) GetNaviNode(id string) *NaviNode {
	naviNode := NaviNode{}
	if err := m.db.Preload("NaviRegion").First(&naviNode, "id = ?", id).Error; err != nil {
		return nil
	}
	return &naviNode
}
func (m *Mirage) CreateNaviNode(naviNode *NaviNode) *NaviNode {
	if err := m.db.Create(naviNode).Error; err != nil {
		return nil
	}
	return naviNode
}

func (m *Mirage) UpdateNaviNode(naviNode *NaviNode) *NaviNode {
	if err := m.db.Save(naviNode).Error; err != nil {
		return nil
	}
	return naviNode
}
