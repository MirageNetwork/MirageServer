package controller

import (
	_ "embed"
	"net/http"
)

// 网络设置响应Data体
type SubscriptionResData struct {
	BillingUsage BillingUsageData `json:"billingUsage"`
}

type BillingUsageData struct {
	Usage                UsageList     `json:"usage"`
	Allowance            AllowanceList `json:"allowance"`
	ReferralDeviceCredit int           `json:"referralDeviceCredit"`
}

type UsageList struct {
	Users         int `json:"users"`
	AdminUsers    int `json:"admin_users"`
	AclNamedUsers int `json:"acl_named_users"`
	Devices       int `json:"devices"`
	Subnets       int `json:"subnets"`
}

type AllowanceList struct {
	Users         Allowance `json:"users"`
	AdminUsers    Allowance `json:"admin_users"`
	AclNamedUsers Allowance `json:"acl_named_users"`
	Devices       Allowance `json:"devices"`
	Subnets       Allowance `json:"subnets"`
}

type Allowance struct {
	Total struct {
		Amount    int  `json:"amount"`
		Unlimited bool `json:"unlimited"`
	} `json:"total"`
}

// 查询订阅用量API
func (h *Mirage) CAPIGetSubscription(
	w http.ResponseWriter,
	r *http.Request,
) {
	user := h.verifyTokenIDandGetUser(w, r)
	if user.CheckEmpty() {
		h.doAPIResponse(w, "用户信息核对失败", nil)
		return
	}

	unlimitAllowance := Allowance{
		Total: struct {
			Amount    int  `json:"amount"`
			Unlimited bool `json:"unlimited"`
		}{
			Amount:    65535,
			Unlimited: true,
		},
	}

	allowanceList := AllowanceList{
		Users:         unlimitAllowance,
		AdminUsers:    unlimitAllowance,
		AclNamedUsers: unlimitAllowance,
		Devices:       unlimitAllowance,
		Subnets:       unlimitAllowance,
	}

	orgUsers, err := h.ListOrgUsers(user.OrganizationID)
	if err != nil {
		h.doAPIResponse(w, "获取组织用户信息失败", nil)
		return
	}
	orgMachines, err := h.ListMachinesByOrgID(user.OrganizationID)
	if err != nil {
		h.doAPIResponse(w, "获取组织设备信息失败", nil)
		return
	}

	userNum := 0
	adminUserNum := 0
	devNum := 0
	subnetNum := 0
	if orgUsers != nil {
		userNum = len(orgUsers)
		for _, u := range orgUsers {
			if u.Role == RoleOwner { //TODO
				adminUserNum++
			}
		}
	}
	if orgMachines != nil {
		devNum = len(orgMachines)
		for _, m := range orgMachines {
			routes, err := h.GetEnabledRoutes(&m)
			if err != nil && routes != nil && len(routes) > 0 {
				subnetNum++
			}
		}
	}

	usageList := UsageList{
		Users:         userNum,
		AdminUsers:    adminUserNum,
		AclNamedUsers: 0, //TODO
		Devices:       devNum,
		Subnets:       subnetNum,
	}

	subscriptionData := SubscriptionResData{
		BillingUsage: BillingUsageData{
			Usage:                usageList,
			Allowance:            allowanceList,
			ReferralDeviceCredit: 0,
		},
	}
	h.doAPIResponse(w, "", subscriptionData)
}
