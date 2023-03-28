package controller

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"tailscale.com/tailcfg"
)

type UsersData struct {
	Users          []UserData `json:"users"`
	ExternalUsers  []UserData `json:"externalUsers"`
	CurrentUserID  int64      `json:"currentUserID"`
	OwnerID        int64      `json:"ownerID"`
	DomainHasOwner bool       `json:"domainHasOwner"`
}
type UserData struct {
	Id                 string    `json:"id"`
	StableId           string    `json:"stableId"`
	DisplayName        string    `json:"displayName"`
	LoginName          string    `json:"loginName"`
	DomainName         string    `json:"domainName"`   // company
	SharedDomain       bool      `json:"sharedDomain"` // ? false
	ProfilePicURL      string    `json:"profilePicURL"`
	Created            time.Time `json:"created"` //timestamp
	Role               string    `json:"role"`
	IsAdmin            bool      `json:"isAdmin"`
	IsOwner            bool      `json:"isOwner"`
	Status             string    `json:"status"`      // "active", 是否suspend
	DeviceCount        int       `json:"deviceCount"` // ? 0
	CanEditBilling     bool      `json:"canEditBilling"`
	NeedsOnboarding    bool      `json:"needsOnboarding"` // 待审核？
	LastSeen           time.Time `json:"lastSeen"`        // timestamp
	CurrentlyConnected bool      `json:"currentlyConnected"`
}

// 接受/admin/api/users的Get请求，用于查询用户
func (h *Mirage) CAPIGetUsers(
	w http.ResponseWriter,
	r *http.Request,
) {
	user := h.verifyTokenIDandGetUser(w, r)
	if user.CheckEmpty() {
		h.doAPIResponse(w, "用户信息核对失败", nil)
		return
	}

	resData := UsersData{
		ExternalUsers:  []UserData{},
		CurrentUserID:  user.ID,
		OwnerID:        user.ID, // TODO: 后续区分Admin和Owner后不再直接赋值当前用户ID而需要查询
		DomainHasOwner: true,    // TODO: 在不确定该项何时为false的情况前先默认为true
	}
	users, err := h.ListOrgUsers(user.OrganizationID)
	if err != nil {
		h.doAPIResponse(w, "用户列表获取失败:"+err.Error(), nil)
		return
	}
	for _, u := range users {
		devCount := 0
		lastSeen := u.CreatedAt
		currentlyConnected := false

		userMachines, err := h.ListMachinesByUser(u.ID)
		if err == nil {
			devCount = len(userMachines)
			for _, m := range userMachines {
				if m.LastSeen.After(lastSeen) {
					lastSeen = *m.LastSeen
				}
				if !currentlyConnected && m.isOnline() {
					currentlyConnected = true
				}
			}
		}
		resData.Users = append(resData.Users, UserData{
			Id:                 strconv.FormatInt(u.ID, 10),
			StableId:           u.StableID,
			DisplayName:        u.Display_Name,
			LoginName:          u.Name,
			DomainName:         u.Organization.Name,
			SharedDomain:       false, //???
			ProfilePicURL:      "",    // TODO: 我们暂时没存头像
			Created:            u.CreatedAt.UTC(),
			Role:               RoleStr[u.Role],
			IsAdmin:            u.Role == RoleOwner, // TODO: 后续区分Admin和Owner后需要变为不等比较
			IsOwner:            u.Role == RoleOwner,
			Status:             "active",            // TODO: 在添加suspend功能后需要变为判断
			DeviceCount:        devCount,            // TODO: 不知为啥官方目前是0，我们做下统计
			CanEditBilling:     u.Role == RoleOwner, // TODO: 后续有更多角色时需要变更
			NeedsOnboarding:    false,               // TODO: 后续增加新成员需approve后需要使用
			LastSeen:           lastSeen.UTC().Round(time.Second),
			CurrentlyConnected: currentlyConnected,
		})
	}
	h.doAPIResponse(w, "", resData)
}

// 请求报文：
type UserActionREQ struct {
	UserID string `json:"userID"`
	Action string `json:"action"` //"restore_user", "suspend_user", "delete_user", "set_owner","set_member"
}

// 接受/admin/api/users的Post请求，用于对用户操作
func (h *Mirage) CAPIPostUsers(
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
	reqData := UserActionREQ{}
	json.NewDecoder(r.Body).Decode(&reqData)

	switch reqData.Action {
	case "set_owner":
		if user.Role != RoleOwner {
			h.doAPIResponse(w, "权限不足", nil)
			return
		}
		targetUID, err := strconv.ParseInt(reqData.UserID, 10, 64)
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
		targetUID, err := strconv.ParseInt(reqData.UserID, 10, 64)
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
