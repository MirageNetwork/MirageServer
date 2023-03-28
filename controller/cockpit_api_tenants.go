package controller

import (
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"tailscale.com/tailcfg"
)

type TenantsData struct {
	Tenants []TenantData `json:"tenants"`
	// ExternalUsers  []UserData `json:"externalUsers"`
	// CurrentUserID  int64      `json:"currentUserID"`
	// OwnerID        int64      `json:"ownerID"`
	// DomainHasOwner bool       `json:"domainHasOwner"`
}
type TenantData struct {
	Id          string    `json:"id"`
	StableId    string    `json:"stableId"`
	Name        string    `json:"name"`
	MagicDomain string    `json:"magicDomain"`
	Provider    string    `json:"provider"`
	Owner       string    `json:"owner"`
	Created     time.Time `json:"created"`
	Status      string    `json:"status"` // "active", 是否suspend
	UserCount   int       `json:"userCount"`
	AdminCount  int       `json:"adminCount"`
	DeviceCount int       `json:"deviceCount"`
	SubnetCount int       `json:"subnetCount"`

	NeedsOnboarding    bool      `json:"needsOnboarding"` // 待审核？
	LastSeen           time.Time `json:"lastSeen"`        // timestamp
	CurrentlyConnected bool      `json:"currentlyConnected"`
}

func (c *Cockpit) ListTenants() ([]Organization, error) {
	tenants := []Organization{}
	if err := c.db.Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}
func (c *Cockpit) ListTenantUsers(orgID int64) ([]User, error) {
	var users []User
	err := c.db.Preload("Organization").Where(&User{
		OrganizationID: orgID,
	}).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (c *Cockpit) ListMachinesByTenantID(orgId int64) ([]Machine, error) {
	machines := []Machine{}
	userIds := []int64{}
	c.db.Model(&User{}).Where(&User{
		OrganizationID: orgId,
	}).Select("id").Find(&userIds)
	if len(userIds) == 0 {
		return machines, nil
	}
	scopeFunc := func(tx *gorm.DB) *gorm.DB {
		if len(userIds) == 1 {
			return tx.Where("user_id = ?", userIds[0])
		} else {
			return tx.Where("user_id in ?", userIds)
		}
	}
	if err := c.db.Preload("AuthKey").Preload("AuthKey.User").Preload("User").Preload("User.Organization").Scopes(scopeFunc).Find(&machines).Error; err != nil {
		return nil, err
	}

	return machines, nil
}

func (c *Cockpit) GetEnabledRoutes(machine *Machine) ([]netip.Prefix, error) {
	routes := []Route{}

	err := c.db.
		Preload("Machine").
		Where("machine_id = ? AND advertised = ? AND enabled = ?", machine.ID, true, true).
		Find(&routes).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().
			Caller().
			Err(err).
			Str("machine", machine.Hostname).
			Msg("Could not get enabled routes for machine")

		return nil, err
	}

	prefixes := []netip.Prefix{}
	for _, route := range routes {
		prefixes = append(prefixes, netip.Prefix(route.Prefix))
	}

	return prefixes, nil
}

// 接受/admin/api/users的Get请求，用于查询用户
func (c *Cockpit) CAPIGetTenant(
	w http.ResponseWriter,
	r *http.Request,
) {
	resData := TenantsData{
		Tenants: []TenantData{},
	}

	tenants, err := c.ListTenants()
	if err != nil {
		c.doAPIResponse(w, "租户列表获取失败:"+err.Error(), nil)
		return
	}

	for _, tenant := range tenants {
		tenantUsers, err := c.ListTenantUsers(tenant.ID)
		if err != nil {
			c.doAPIResponse(w, "租户用户列表获取失败:"+err.Error(), nil)
			return
		}
		tenantOwnerAccount := ""
		tenantUserCount := len(tenantUsers)
		tenantAdminCount := 0
		tenantDeviceCount := 0
		tenantSubnetCount := 0
		lastSeen := time.Time{}
		currentlyConnected := false
		for _, user := range tenantUsers {
			if user.Role == RoleOwner {
				tenantOwnerAccount = user.Name
			}
			if user.Role == RoleOwner {
				tenantAdminCount++
			}
		}
		tenantMachines, err := c.ListMachinesByTenantID(tenant.ID)
		if err != nil {
			c.doAPIResponse(w, "租户设备列表获取失败:"+err.Error(), nil)
			return
		}
		if tenantMachines != nil {
			tenantDeviceCount = len(tenantMachines)
		}
		for _, machine := range tenantMachines {
			if r, err := c.GetEnabledRoutes(&machine); err == nil && r != nil && len(r) > 0 {
				tenantSubnetCount++
			}
			if machine.LastSeen.After(lastSeen) {
				lastSeen = *machine.LastSeen
			}
			if !currentlyConnected && machine.isOnline() {
				currentlyConnected = true
			}
		}

		tmpTenant := TenantData{
			Id:                 strconv.FormatInt(tenant.ID, 10),
			StableId:           tenant.StableID,
			Name:               tenant.Name,
			MagicDomain:        tenant.MagicDnsDomain,
			Provider:           tenant.Provider,
			Owner:              tenantOwnerAccount,
			Created:            tenant.CreatedAt,
			Status:             "active",
			UserCount:          tenantUserCount,
			AdminCount:         tenantAdminCount,
			DeviceCount:        tenantDeviceCount,
			SubnetCount:        tenantSubnetCount,
			NeedsOnboarding:    false,              //TODO
			LastSeen:           lastSeen,           // timestamp
			CurrentlyConnected: currentlyConnected, //TODO
		}
		resData.Tenants = append(resData.Tenants, tmpTenant)
	}

	c.doAPIResponse(w, "", resData)
}

// 请求报文：
type TenantActionREQ struct {
	TenantID string `json:"tenantID"`
	Action   string `json:"action"`
}

// 接受/admin/api/users的Post请求，用于对用户操作
func (h *Mirage) CAPIPostTenants(
	w http.ResponseWriter,
	r *http.Request,
) {
	user := h.verifyTokenIDandGetUser(w, r)
	if user.CheckEmpty() {
		h.doAPIResponse(w, "用户信息核对失败", nil)
		return
	}
	err := r.ParseForm()
	if err != nil {
		h.doAPIResponse(w, "用户请求解析失败:"+err.Error(), nil)
		return
	}
	reqData := TenantActionREQ{}
	json.NewDecoder(r.Body).Decode(&reqData)

	switch reqData.Action {
	case "set_owner":
		if user.Role != RoleOwner {
			h.doAPIResponse(w, "权限不足", nil)
			return
		}
		targetUID, err := strconv.ParseInt(reqData.TenantID, 10, 64)
		if err != nil {
			h.doAPIResponse(w, "目标用户ID解析失败:"+err.Error(), nil)
			return
		}
		err = h.TransferOwner(tailcfg.UserID(user.ID), tailcfg.UserID(targetUID))
		if err != nil {
			h.doAPIResponse(w, "修改用户角色失败:"+err.Error(), nil)
			return
		}
		h.doAPIResponse(w, "", nil)
	case "delete_user":
		targetUID, err := strconv.ParseInt(reqData.TenantID, 10, 64)
		if err != nil {
			h.doAPIResponse(w, "目标用户ID解析失败:"+err.Error(), nil)
			return
		}
		targetUser, err := h.GetUserByID(tailcfg.UserID(targetUID))
		if err != nil {
			h.doAPIResponse(w, "目标用户信息获取失败:"+err.Error(), nil)
			return
		}
		if targetUser.Role == RoleOwner {
			h.doAPIResponse(w, "无法删除Owner，请联系我们", nil)
			return
		}
		mlist, err := h.ListMachinesByUser(targetUID)
		if err != nil {
			h.doAPIResponse(w, "目标用户设备列表获取失败:"+err.Error(), nil)
			return
		}
		for _, m := range mlist {
			if m.ForcedTags != nil && len([]string(m.ForcedTags)) > 0 {
				continue
			}
			err = h.HardDeleteMachine(&m)
			if err != nil {
				h.doAPIResponse(w, "目标用户设备删除失败:"+err.Error(), nil)
				return
			}
		}
		err = h.DestroyUser(targetUser.Name, targetUser.Organization.Name, targetUser.Organization.Provider)
		if err != nil {
			h.doAPIResponse(w, "目标用户删除失败:"+err.Error(), nil)
			return
		}
		h.doAPIResponse(w, "", nil)
	}
}
