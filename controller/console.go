package controller

import (
	_ "embed"
	"fmt"
	"net/http"
	"time"
)

type selfData struct {
	Basedomain   string `json:"basedomain"`
	UserName     string `json:"username"`
	UserNameHead string `json:"usernamehead"`
	UserAccount  string `json:"useraccount"`
	OrgName      string `json:"orgname"`
}

// 提供获取用户信息的API
func (h *Mirage) ConsoleSelfAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	user, err := h.verifyTokenIDandGetUser(writer, req)
	if err != nil || user.CheckEmpty() {
		h.doAPIResponse(writer, "用户信息核对失败:"+err.Error(), nil)
		return
	}
	userName := user.Name
	userDisName := user.Display_Name
	userNameHead := string([]rune(userDisName)[0])

	userOrgName := user.Organization.Name
	// 用户所属组织将存于数据库，合并实现前，不设置配置文件中该项，均按个人用户处理

	resData := selfData{
		Basedomain:   user.Organization.MagicDnsDomain,
		UserNameHead: userNameHead,
		UserName:     userDisName,
		UserAccount:  userName,
		OrgName:      userOrgName,
	}
	h.doAPIResponse(writer, "", resData)
}

// 验证Token并获取用户信息
func (h *Mirage) verifyTokenIDandGetUser(
	writer http.ResponseWriter,
	req *http.Request,
) (*User, error) {
	controlCodeCookie, err := req.Cookie("miragecontrol")
	if err == http.ErrNoCookie {
		return nil, fmt.Errorf("Token不存在")
	}
	controlCode := controlCodeCookie.Value
	controlCodeC, concontrolCodeExpiration, ok := h.controlCodeCache.GetWithExpiration(controlCode)
	if !ok || concontrolCodeExpiration.Compare(time.Now()) != 1 {
		return nil, fmt.Errorf("验证Token失败")
	}
	controlCodeItem := controlCodeC.(ControlCacheItem)
	user, err := h.GetUserByID(controlCodeItem.uid)
	if err != nil {
		return nil, fmt.Errorf("提取用户信息失败")
	}
	return user, nil
}
