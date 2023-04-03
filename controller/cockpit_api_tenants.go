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

func (c *Cockpit) GetTenantByID(id int64) (*Organization, error) {
	org := &Organization{}
	err := c.db.Where(&Organization{ID: id}).Take(org).Error
	if err != nil {
		return nil, err
	}
	return org, err
}

func (c *Cockpit) HardDeleteMachine(machine *Machine) error {
	if err := c.db.Unscoped().Delete(&machine).Error; err != nil {
		return err
	}

	return nil
}

func (c *Cockpit) ListPreAuthKeys(userID int64) ([]PreAuthKey, error) {
	keys := []PreAuthKey{}
	if err := c.db.Preload("User").Preload("ACLTags").Where(&PreAuthKey{UserID: userID}).Find(&keys).Error; err != nil {
		return nil, err
	}

	return keys, nil
}
func (c *Cockpit) DestroyPreAuthKey(pak PreAuthKey) error {
	return c.db.Transaction(func(db *gorm.DB) error {

		if result := db.Unscoped().Delete(pak); result.Error != nil {
			return result.Error
		}

		return nil
	})
}

func (c *Cockpit) DestroyUser(user *User) error {
	keys, err := c.ListPreAuthKeys(user.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	for _, key := range keys {
		err = c.DestroyPreAuthKey(key)
		if err != nil {
			return err
		}
	}

	return c.db.Unscoped().Delete(&user).Error
}

func (c *Cockpit) DestroyTenant(tenant *Organization) error {
	machines, err := c.ListMachinesByTenantID(tenant.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	for _, machine := range machines {
		err = c.HardDeleteMachine(&machine)
		if err != nil {
			return err
		}
	}

	users, err := c.ListTenantUsers(tenant.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	for _, user := range users {
		err = c.DestroyUser(&user)
		if err != nil {
			return err
		}
	}

	return c.db.Unscoped().Delete(&tenant).Error
}

func (c *Cockpit) GetUser(name string, orgID int64) (*User, error) {
	user := User{}
	err := c.db.Where(&User{
		Name:           name,
		OrganizationID: orgID,
	}).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

func (c *Cockpit) TransferOwner(srcId, destId int64) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Select("Role").Updates(&User{ID: srcId, Role: RoleMember})
		if result.Error != nil || result.RowsAffected == 0 {
			return ErrChangeUserRole
		}
		result = tx.Select("Role").Updates(&User{ID: destId, Role: RoleOwner})
		if result.Error != nil || result.RowsAffected == 0 {
			return ErrChangeUserRole
		}
		return nil
	})
}

func (c *Cockpit) UpdateTenant(org *Organization) error {
	err := c.db.Save(org).Error
	if err != nil {
		return err
	}
	return nil
}

// 请求报文：
type TenantActionREQ struct {
	TenantID string           `json:"tenantID"`
	Action   string           `json:"action"`
	NewValue TenantUpdateData `json:"newValue"`
}

type TenantUpdateData struct {
	MagicDomain string `json:"magicDomain"`
	Owner       string `json:"owner"`
	Provider    string `json:"provider"`
	Name        string `json:"name"`
}

// 接受/admin/api/users的Post请求，用于对用户操作
func (c *Cockpit) CAPIPostTenants(
	w http.ResponseWriter,
	r *http.Request,
) {
	err := r.ParseForm()
	if err != nil {
		c.doAPIResponse(w, "用户请求解析失败:"+err.Error(), nil)
		return
	}
	reqData := TenantActionREQ{}
	json.NewDecoder(r.Body).Decode(&reqData)

	targetTenantID, err := strconv.ParseInt(reqData.TenantID, 10, 64)
	if err != nil {
		c.doAPIResponse(w, "目标租户ID解析失败:"+err.Error(), nil)
		return
	}
	targetTenant, err := c.GetTenantByID(targetTenantID)
	if err != nil {
		c.doAPIResponse(w, "目标租户获取失败:"+err.Error(), nil)
		return
	}

	switch reqData.Action {
	case "delete_tenant":
		// 删除租户
		if err = c.DestroyTenant(targetTenant); err != nil {
			c.doAPIResponse(w, "目标租户删除失败:"+err.Error(), nil)
			return
		}
		c.doAPIResponse(w, "", nil)
		return
	case "update_tenant":
		// 更新租户配置
		if reqData.NewValue.MagicDomain != "" {
			targetTenant.MagicDnsDomain = reqData.NewValue.MagicDomain
		}

		if reqData.NewValue.Name != "" {
			targetTenant.Name = reqData.NewValue.Name
		}
		switch reqData.NewValue.Provider {
		case "Microsoft", "Github", "Google", "Apple", "WXScan":
			targetTenant.Provider = reqData.NewValue.Provider
		default:
			c.doAPIResponse(w, "目标租户更新失败:不支持的Provider", nil)
		}
		c.UpdateTenant(targetTenant)

		// 更新租户Owner
		newOwner, err := c.GetUser(reqData.NewValue.Owner, targetTenant.ID)
		if err != nil {
			c.doAPIResponse(w, "目标租户更新失败:新Owner不存在", nil)
			return
		}
		users, err := c.ListTenantUsers(targetTenant.ID)
		if err != nil {
			c.doAPIResponse(w, "目标租户更新失败:获取租户用户列表失败", nil)
			return
		}
		for _, user := range users {
			if user.Role == RoleOwner {
				if user.ID != newOwner.ID {
					err = c.TransferOwner(user.ID, newOwner.ID)
					if err != nil {
						c.doAPIResponse(w, "目标租户更新失败:更改Owner失败", nil)
						return
					}
				}
				break
			}
		}
		c.doAPIResponse(w, "", nil)
		return
	}
}
