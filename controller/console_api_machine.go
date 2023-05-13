package controller

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

type machineData struct {
	Address                []string `json:"addresses"`
	AllowedIPs             []string `json:"allowedIPs"`
	ExtraIPs               []string `json:"extraIPs"`
	AdvertisedIPs          []string `json:"advertisedIPs"`
	HasSubnets             bool     `json:"hasSubnets"`
	AdvertisedExitNode     bool     `json:"advertisedExitNode"`
	AllowedExitNode        bool     `json:"allowedExitNode"`
	HasExitNode            bool     `json:"hasExitNode"` //未实现
	AllowedTags            []string `json:"allowedTags"` //未实现
	InvalidTags            []string `json:"invalidTags"` //未实现
	HasTags                bool     `json:"hasTags"`     //？？？移除？
	Endpoints              []string `json:"endpoints"`
	IpnVersion             string   `json:"ipnVersion"` //未实现
	Os                     string   `json:"os"`         //未实现
	Name                   string   `json:"name"`       //未实现
	Fqdn                   string   `json:"fqdn"`       //未实现
	Domain                 string   `json:"domain"`     //未实现
	Created                string   `json:"created"`    //未实现
	Hostname               string   `json:"hostname"`   //未实现
	MachineKey             string   `json:"machineKey"` //未实现
	NodeKey                string   `json:"nodeKey"`    //未实现
	Id                     string   `json:"id"`         //未实现
	StableId               string   `json:"stableId"`   //未实现
	User                   string   `json:"user"`       //未实现
	Creator                string   `json:"creator"`    //未实现
	Expires                string   `json:"expires"`
	NeverExpires           bool     `json:"neverExpires"`
	Authorized             bool     `json:"authorized"`             //未实现
	IsExternal             bool     `json:"isExternal"`             // ？？？        //未实现
	BrokenIPForwarding     bool     `json:"brokenIPForwarding"`     //未实现
	IsEphemeral            bool     `json:"isEphemeral"`            //未实现
	AvailableUpdateVersion string   `json:"availableUpdateVersion"` //未实现
	LastSeen               string   `json:"lastSeen"`               //未实现
	ConnectedToControl     bool     `json:"connectedToControl"`     //未实现
	AutomaticNameMode      bool     `json:"automaticNameMode"`
	TailnetLockKey         string   `json:"tailnetLockKey"`     //未实现
	ShareID                string   `json:"shareID"`            //未实现
	AcceptedShareCount     int      `json:"acceptedShareCount"` //未实现
	ParsedLinuxVersion     string   `json:"parsedLinuxVersion"` //未实现
}

type machineItem struct {
	Id                     string   `json:"id"`                     //done
	StableId               string   `json:"stableId"`               //未实现
	Name                   string   `json:"name"`                   //done
	Fqdn                   string   `json:"fqdn"`                   //未实现
	User                   string   `json:"user"`                   //done
	UserNameHead           string   `json:"usernamehead"`           // TODO
	Addresses              []string `json:"addresses"`              //done
	Os                     string   `json:"os"`                     //done
	Hostname               string   `json:"hostname"`               //done
	IpnVersion             string   `json:"ipnVersion"`             //done
	ConnectedToControl     bool     `json:"connectedToControl"`     //done
	AvailableUpdateVersion string   `json:"availableUpdateVersion"` //未实现
	LastSeen               string   `json:"lastSeen"`               //done
	Created                string   `json:"created"`                //done

	IsExternal   bool `json:"isExternal"`
	IsEphemeral  bool `json:"isEphemeral"`
	IsSharedOut  bool `json:"issharedout"`
	NeverExpires bool `json:"neverExpires"` //done

	AllowedIPs         []string `json:"allowedIPs"`
	ExtraIPs           []string `json:"extraIPs"`
	AdvertisedIPs      []string `json:"advertisedIPs"`
	HasSubnets         bool     `json:"hasSubnets"`
	AdvertisedExitNode bool     `json:"advertisedExitNode"`
	AllowedExitNode    bool     `json:"allowedExitNode"`
	AllowedTags        []string `json:"allowedTags"`
	InvalidTags        []string `json:"invalidTags"`
	HasTags            bool     `json:"hasTags"`

	Expires time.Time `json:"expires"`

	ExpiryDesc string `json:"expirydesc"`

	Endpoints         []string `json:"endpoints"`
	AutomaticNameMode bool     `json:"automaticNameMode"`
}

func IsUpdateAvailable(cur, latest string) bool {
	curV := strings.Split(strings.Split(cur, "-")[0], ".")
	latestV := strings.Split(strings.Split(latest, "-")[0], ".")
	for i := 0; i < len(latestV); i++ {
		curInt, _ := strconv.Atoi(curV[i])
		latestInt, _ := strconv.Atoi(latestV[i])
		if curInt < latestInt {
			return true
		}
	}
	return false
}

// 控制台获取设备信息列表的API
func (h *Mirage) ConsoleMachinesAPI(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, err := h.verifyTokenIDandGetUser(w, r)
	if err != nil || user.CheckEmpty() {
		h.doAPIResponse(w, "用户信息核对失败:"+err.Error(), nil)
		return
	}

	OrgMachines, err := h.ListMachinesByOrgID(user.OrganizationID)
	if err != nil {
		h.doAPIResponse(w, "查询用户节点列表失败", nil)
		return
	}

	mlist := make([]machineItem, 0)
	for _, machine := range OrgMachines {
		tz, _ := time.LoadLocation("Asia/Shanghai")
		tmpMachine := machineItem{
			Id:                 strconv.FormatInt(machine.ID, 10),
			Name:               machine.GivenName,
			User:               machine.User.Name,
			UserNameHead:       string([]rune(machine.User.Display_Name)[0]),
			Os:                 machine.HostInfo.OS,
			Hostname:           machine.HostInfo.Hostname,
			IpnVersion:         machine.HostInfo.IPNVersion,
			Created:            machine.CreatedAt.In(tz).Format("2006年01月02日 15:04:05"),
			LastSeen:           machine.LastSeen.In(tz).Format("2006年01月02日 15:04:05"),
			ConnectedToControl: machine.isOnline(),
			AllowedTags:        machine.ForcedTags,
			InvalidTags:        []string{},
			HasTags:            machine.ForcedTags != nil && len(machine.ForcedTags) > 0,

			IsEphemeral:  machine.isEphemeral(),
			NeverExpires: *machine.Expiry == time.Time{},
			Expires:      *machine.Expiry,

			Endpoints:         machine.Endpoints,
			AutomaticNameMode: machine.AutoGenName,
		}

		switch machine.HostInfo.OS {
		case "linux":
			if IsUpdateAvailable(machine.HostInfo.IPNVersion, h.cfg.ClientVersion.Linux.Version) {
				tmpMachine.AvailableUpdateVersion = strings.Split(h.cfg.ClientVersion.Linux.Version, "-")[0]
			}
		case "windows":
			if IsUpdateAvailable(machine.HostInfo.IPNVersion, h.cfg.ClientVersion.Win.Version) {
				tmpMachine.AvailableUpdateVersion = strings.Split(h.cfg.ClientVersion.Win.Version, "-")[0]
			}
		case "macOS":
			if h.cfg.ClientVersion.MacStore.Version != "" && IsUpdateAvailable(machine.HostInfo.IPNVersion, h.cfg.ClientVersion.MacStore.Version) {
				tmpMachine.AvailableUpdateVersion = strings.Split(h.cfg.ClientVersion.MacStore.Version, "-")[0]
			} else if IsUpdateAvailable(machine.HostInfo.IPNVersion, h.cfg.ClientVersion.MacTestFlight.Version) {
				tmpMachine.AvailableUpdateVersion = strings.Split(h.cfg.ClientVersion.MacTestFlight.Version, "-")[0]
			}
		case "iOS":
			if h.cfg.ClientVersion.IOSStore.Version != "" && IsUpdateAvailable(machine.HostInfo.IPNVersion, h.cfg.ClientVersion.IOSStore.Version) {
				tmpMachine.AvailableUpdateVersion = strings.Split(h.cfg.ClientVersion.IOSStore.Version, "-")[0]
			} else if IsUpdateAvailable(machine.HostInfo.IPNVersion, h.cfg.ClientVersion.IOSTestFlight.Version) {
				tmpMachine.AvailableUpdateVersion = strings.Split(h.cfg.ClientVersion.IOSTestFlight.Version, "-")[0]
			}
		case "android":
			if IsUpdateAvailable(machine.HostInfo.IPNVersion, h.cfg.ClientVersion.Android.Version) {
				tmpMachine.AvailableUpdateVersion = strings.Split(h.cfg.ClientVersion.Android.Version, "-")[0]
			}
		}

		if machine.User.Organization.EnableMagic {
			tmpMachine.Fqdn = machine.GivenName + "." + machine.User.Organization.MagicDnsDomain
		}
		// 处理路由部分
		machineRoutes, err := h.GetMachineRoutes(&machine)
		if err != nil {
			h.doAPIResponse(w, "查询设备路由失败", nil)
			return
		}
		for _, route := range machineRoutes {
			if route.isExitRoute() {
				if route.Advertised {
					tmpMachine.AdvertisedExitNode = true
					if route.Enabled {
						tmpMachine.AllowedExitNode = true
					}
				}
			} else {
				if route.Advertised {
					tmpMachine.HasSubnets = true
					routeV := netip.Prefix(route.Prefix).String()
					if err != nil {
						h.doAPIResponse(w, "子网路由地址转换失败", nil)
						return
					}
					tmpMachine.AdvertisedIPs = append(tmpMachine.AdvertisedIPs, routeV)
					if route.Enabled {
						tmpMachine.AllowedIPs = append(tmpMachine.AllowedIPs, routeV)
					} else {
						tmpMachine.ExtraIPs = append(tmpMachine.ExtraIPs, routeV)
					}
				}
			}
		}

		if !tmpMachine.NeverExpires {
			ExpiryDuration := time.Until(*machine.Expiry)
			tmpMachine.ExpiryDesc = convExpiryToStr(ExpiryDuration)
		}
		if machine.IPAddresses[0].Is4() {
			tmpMachine.Addresses = []string{
				machine.IPAddresses[0].String(),
				machine.IPAddresses[1].String()}
		} else if machine.IPAddresses[1].Is4() {
			tmpMachine.Addresses = []string{
				machine.IPAddresses[1].String(),
				machine.IPAddresses[0].String()}
		}
		mlist = append(mlist, tmpMachine)
	}

	h.doAPIResponse(w, "", struct {
		Machines []machineItem `json:"machines"`
	}{
		Machines: mlist,
	})
}

type MachineDebugInfo struct {
	MappingVariesByDestIP bool                `json:"mappingVariesByDestIP"`
	HairPinning           bool                `json:"hairPinning"`
	IPv6                  bool                `json:"ipv6"`
	UDP                   bool                `json:"udp"`
	UPnP                  bool                `json:"upnp"`
	PMP                   bool                `json:"pmp"`
	PCP                   bool                `json:"pcp"`
	Latency               map[string]*Latency `json:"latency"` // "derp区域编号"-结构体
}
type Latency struct {
	RegionName string  `json:"regionName"`
	Preferred  bool    `json:"preferred"`
	LatencyMs  float64 `json:"latencyMs"`
}

func (m *Mirage) ConsoleMachineDebugAPI(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, err := m.verifyTokenIDandGetUser(w, r)
	if err != nil || user.CheckEmpty() {
		m.doAPIResponse(w, "用户信息核对失败:"+err.Error(), nil)
		return
	}

	targetMIPStr := r.URL.Query().Get("ip")
	if targetMIPStr == "" {
		m.doAPIResponse(w, "用户查询IP为空", nil)
		return
	}
	targetMIP, err := netip.ParseAddr(targetMIPStr)
	if err != nil {
		m.doAPIResponse(w, "用户请求IP解析失败", nil)
		return
	}
	targetMachine := m.GetMachineByIP(targetMIP)
	if targetMachine == nil || targetMachine.User.OrganizationID != user.OrganizationID {
		m.doAPIResponse(w, "组织内无此设备", nil)
		return
	}
	resData := &MachineDebugInfo{
		MappingVariesByDestIP: targetMachine.HostInfo.NetInfo.MappingVariesByDestIP.EqualBool(true),
		HairPinning:           targetMachine.HostInfo.NetInfo.HairPinning.EqualBool(true),
		IPv6:                  targetMachine.HostInfo.NetInfo.WorkingIPv6.EqualBool(true),
		UDP:                   targetMachine.HostInfo.NetInfo.WorkingUDP.EqualBool(true),
		UPnP:                  targetMachine.HostInfo.NetInfo.UPnP.EqualBool(true),
		PMP:                   targetMachine.HostInfo.NetInfo.PMP.EqualBool(true),
		PCP:                   targetMachine.HostInfo.NetInfo.PCP.EqualBool(true),
		Latency:               make(map[string]*Latency),
	}
	derpMap, err := m.LoadOrgDERPs(targetMachine.User.OrganizationID)
	if err != nil {
		m.doAPIResponse(w, "获取组织中继信息失败", nil)
		return
	}

	if targetMachine.HostInfo.NetInfo != nil && targetMachine.HostInfo.NetInfo.PreferredDERP != 0 {
		for derpname, latency := range targetMachine.HostInfo.NetInfo.DERPLatency {
			ipver := strings.Split(derpname, "-")[1]
			derpRegionIDStr := strings.Split(derpname, "-")[0]
			derpRegionID, _ := strconv.Atoi(derpRegionIDStr)

			if resData.Latency[derpRegionIDStr] == nil {
				resData.Latency[derpRegionIDStr] = &Latency{
					RegionName: derpMap.Regions[derpRegionID].RegionName,
					Preferred:  targetMachine.HostInfo.NetInfo.PreferredDERP == derpRegionID,
				}
			}

			if ipver == "v4" {
				if peerlatency, ok := targetMachine.HostInfo.NetInfo.DERPLatency[derpRegionIDStr+"-v6"]; ok {
					if latency < peerlatency {
						resData.Latency[derpRegionIDStr].LatencyMs = latency * 1000
					}
				} else {
					resData.Latency[derpRegionIDStr].LatencyMs = latency * 1000
				}
			} else {
				if peerlatency, ok := targetMachine.HostInfo.NetInfo.DERPLatency[derpRegionIDStr+"-v4"]; ok {
					if latency < peerlatency {
						resData.Latency[derpRegionIDStr].LatencyMs = latency * 1000
					}
				} else {
					resData.Latency[derpRegionIDStr].LatencyMs = latency * 1000
				}
			}
		}
	}
	m.doAPIResponse(w, "", resData)
}

func (h *Mirage) ConsoleMachinesUpdateAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	user, err := h.verifyTokenIDandGetUser(writer, req)
	if err != nil || user.CheckEmpty() {
		h.doAPIResponse(writer, "用户信息核对失败:"+err.Error(), nil)
		return
	}
	err = req.ParseForm()
	if err != nil {
		h.doAPIResponse(writer, "用户请求解析失败:"+err.Error(), nil)
		return
	}
	reqData := make(map[string]interface{})
	json.NewDecoder(req.Body).Decode(&reqData)
	reqMID, ok := reqData["mid"].(string)
	if !ok {
		h.doAPIResponse(writer, "用户请求mid解析失败", nil)
		return
	}
	MachineID, err := strconv.ParseInt(reqMID, 0, 64)
	if err != nil {
		h.doAPIResponse(writer, "用户请求mid处理失败", nil)
		return
	}
	toUpdateMachine, err := h.GetMachineByID(MachineID)
	if err != nil {
		h.doAPIResponse(writer, "查询用户设备失败", nil)
		return
	}
	/*
		if toUpdateMachine.User.ID != user.ID {
			h.doAPIResponse(writer, "用户没有该权限", nil)
			return
		}
	*/
	reqState, ok := reqData["state"].(string)
	if !ok {
		h.doAPIResponse(writer, "用户请求state解析失败", nil)
		return
	}

	switch reqState {
	case "set-expires": //切换密钥永不过期设置
		msg, err := h.setMachineExpiry(toUpdateMachine)
		if err != nil {
			h.doAPIResponse(writer, msg, nil)
		} else {
			resData := machineData{
				NeverExpires: *toUpdateMachine.Expiry == time.Time{},
				Expires:      msg,
			}
			h.doAPIResponse(writer, "", resData)
		}
	case "rename-node": //设置设备名称
		newName := reqData["nodeName"].(string)
		msg, _, err := h.setMachineName(toUpdateMachine, newName)
		if err != nil {
			h.doAPIResponse(writer, msg, nil)
		} else {
			resData := machineData{
				AutomaticNameMode: toUpdateMachine.AutoGenName,
				Name:              toUpdateMachine.GivenName,
				Hostname:          toUpdateMachine.Hostname,
				NeverExpires:      *toUpdateMachine.Expiry == time.Time{},
				Expires:           msg,
			}
			h.doAPIResponse(writer, "", resData)
		}
	case "set-route-settings": //设置子网转发及出口节点
		allowedIPsInterface := reqData["allowedIPs"].([]interface{})
		allowExitNode := reqData["allowedExitNode"].(bool)

		allowedIPs := new([]string)
		for _, ip := range allowedIPsInterface {
			*allowedIPs = append(*allowedIPs, ip.(string))
		}

		msg, err := h.setMachineSubnet(toUpdateMachine, allowExitNode, *allowedIPs)
		if err != nil {
			h.doAPIResponse(writer, msg, nil)
			return
		} else {
			resData := machineData{
				AutomaticNameMode: toUpdateMachine.AutoGenName,
				Name:              toUpdateMachine.GivenName,
				Hostname:          toUpdateMachine.Hostname,
				NeverExpires:      *toUpdateMachine.Expiry == time.Time{},
				Expires:           msg,
			}
			machineRoutes, err := h.GetMachineRoutes(toUpdateMachine)
			if err != nil {
				h.doAPIResponse(writer, "查询设备路由失败", nil)
				return
			}
			for _, route := range machineRoutes {
				if route.isExitRoute() {
					if route.Advertised {
						resData.AdvertisedExitNode = true
						if route.Enabled {
							resData.AllowedExitNode = true
						}
					}
				} else {
					if route.Advertised {
						resData.HasSubnets = true
						routeV := netip.Prefix(route.Prefix).String()
						if err != nil {
							h.doAPIResponse(writer, "子网路由地址转换失败", nil)
							return
						}
						resData.AdvertisedIPs = append(resData.AdvertisedIPs, routeV)
						if route.Enabled {
							resData.AllowedIPs = append(resData.AllowedIPs, routeV)
						} else {
							resData.ExtraIPs = append(resData.ExtraIPs, routeV)
						}
					}
				}
			}
			h.doAPIResponse(writer, "", resData)
		}
	case "set-tags": //设置设备标签
		reqTags := reqData["tags"].([]interface{})
		setTags := make([]string, len(reqTags))
		for i, tag := range reqTags {
			setTags[i] = tag.(string)
		}
		msg, err := h.setMachineTags(toUpdateMachine, setTags)
		if err != nil {
			h.doAPIResponse(writer, msg, nil)
		} else {
			invalidTags := []string{}
			allowedTags := []string{}
			org, err := h.GetOrgnaizationByID(user.OrganizationID)
			if err != nil {
				h.doAPIResponse(writer, msg, nil)
			}
			for _, tag := range setTags {
				if _, ok := org.AclPolicy.TagOwners[tag]; ok {
					allowedTags = append(allowedTags, tag)
				} else {
					invalidTags = append(invalidTags, tag)
				}
			}
			resData := machineData{
				AutomaticNameMode: toUpdateMachine.AutoGenName,
				Name:              toUpdateMachine.GivenName,
				Hostname:          toUpdateMachine.Hostname,
				NeverExpires:      *toUpdateMachine.Expiry == time.Time{},
				Expires:           msg,
				HasTags:           len(setTags) > 0,
				AllowedTags:       allowedTags,
				InvalidTags:       invalidTags,
			}
			h.doAPIResponse(writer, "", resData)
		}
	}
}

// 删除设备API
func (h *Mirage) ConsoleRemoveMachineAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	user, err := h.verifyTokenIDandGetUser(writer, req)
	if err != nil || user.CheckEmpty() {
		h.doAPIResponse(writer, "用户信息核对失败:"+err.Error(), nil)
		return
	}
	UserMachines, err := h.ListMachinesByUser(user.ID)
	if err != nil {
		h.doAPIResponse(writer, "用户设备检索失败:"+err.Error(), nil)
		return
	}
	err = req.ParseForm()
	if err != nil {
		h.doAPIResponse(writer, "用户请求解析失败:"+err.Error(), nil)
		return
	}
	reqData := make(map[string]string)
	json.NewDecoder(req.Body).Decode(&reqData)
	wantRemoveID := reqData["mid"]
	for _, machine := range UserMachines {
		if strconv.FormatInt(machine.ID, 10) == wantRemoveID {
			err = h.HardDeleteMachine(&machine)
			if err != nil {
				h.doAPIResponse(writer, "用户设备删除失败:"+err.Error(), nil)
				return
			}
			h.NotifyNaviOrgNodesChange(user.OrganizationID, "", machine.NodeKey)

			h.doAPIResponse(writer, "", nil)
			return
		}
	}
	h.doAPIResponse(writer, "未找到目标设备", nil)
}

// 切换设备密钥是否禁用过期
func (h *Mirage) setMachineExpiry(machine *Machine) (string, error) {
	if (*machine.Expiry != time.Time{}) {
		err := h.RefreshMachine(machine, time.Time{})
		if err != nil {
			return "设备密钥过期禁用失败", err
		} else {
			return "", err
		}
	} else {
		expiryDuration := time.Hour * 24 * time.Duration(machine.User.Organization.ExpiryDuration)
		newExpiry := time.Now().Add(expiryDuration)
		err := h.RefreshMachine(machine, newExpiry)
		if err != nil {
			return "设备密钥过期启用失败", err
		} else {
			return convExpiryToStr(expiryDuration), nil
		}
	}
}

// 三个返回值：msg、nowName、err
func (h *Mirage) setMachineName(machine *Machine, newName string) (string, string, error) {
	newGiveName, err := h.setAutoGenName(machine, newName)
	if err != nil {
		return "设置主机名失败", "", err
	}
	return "", newGiveName, nil
}

func (h *Mirage) setMachineTags(machine *Machine, tags []string) (string, error) {
	err := h.SetTags(machine, tags)
	if err != nil {
		return "设置设备标签失败", err
	}
	return "", nil
}

func (h *Mirage) setMachineSubnet(machine *Machine, ExitNodeEnable bool, allowedIPs []string) (string, error) {
	machineRoutes, err := h.GetMachineRoutes(machine)
	if err != nil {
		return "获取设备路由设置失败", err
	}
	for _, r := range machineRoutes {
		if r.isExitRoute() {
			if ExitNodeEnable {
				err = h.EnableRoute(uint64(r.ID))
			} else {
				err = h.DisableRoute(uint64(r.ID))
			}
			if err != nil {
				return "设置设备出口节点状态失败", err
			}
		} else {
			err = h.DisableRoute(uint64(r.ID))
			if err != nil {
				return "设置设备出口节点状态失败", err
			}
		}
	}
	err = h.enableRoutes(machine, allowedIPs...)
	if err != nil {
		return "设置设备子网路由状态失败", err
	}
	return "", nil
}
