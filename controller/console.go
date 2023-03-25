package controller

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type adminTemplateConfig struct {
	ErrorMsg     string                 `json:"errormsg"`
	Basedomain   string                 `json:"basedomain"`
	UserName     string                 `json:"username"`
	UserNameHead string                 `json:"usernamehead"`
	UserAccount  string                 `json:"useraccount"`
	OrgName      string                 `json:"orgname"`
	MList        map[string]machineItem `json:"mlist"`
}

// 提供获取用户信息的API
func (h *Mirage) ConsoleSelfAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	controlCodeCookie, _ := req.Cookie("miragecontrol")
	controlCode := controlCodeCookie.Value
	controlCodeC, controlCodeExpiration, ok := h.controlCodeCache.GetWithExpiration(controlCode)
	if !ok || controlCodeExpiration.Compare(time.Now()) != 1 {
		errRes := adminTemplateConfig{ErrorMsg: "验证Token失败"}
		err := json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	controlCodeItem := controlCodeC.(ControlCacheItem)
	user, err := h.GetUserByID(controlCodeItem.uid)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "查询用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}
	userName := user.Name
	userDisName := user.Display_Name
	userNameHead := string([]rune(userDisName)[0])

	userOrgName := user.Organization.Name
	// 用户所属组织将存于数据库，合并实现前，不设置配置文件中该项，均按个人用户处理

	renderData := adminTemplateConfig{
		Basedomain:   user.Organization.MagicDnsDomain,
		UserNameHead: userNameHead,
		UserName:     userDisName,
		UserAccount:  userName,
		OrgName:      userOrgName,
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(&renderData)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

// 验证Token并获取用户信息
func (h *Mirage) verifyTokenIDandGetUser(
	writer http.ResponseWriter,
	req *http.Request,
) *User {
	controlCodeCookie, err := req.Cookie("miragecontrol")
	if err == http.ErrNoCookie {
		errRes := adminTemplateConfig{ErrorMsg: "Token不存在"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return nil
	}
	controlCode := controlCodeCookie.Value
	controlCodeC, concontrolCodeExpiration, ok := h.controlCodeCache.GetWithExpiration(controlCode)
	if !ok || concontrolCodeExpiration.Compare(time.Now()) != 1 {
		errRes := adminTemplateConfig{ErrorMsg: "验证Token失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return nil
	}
	controlCodeItem := controlCodeC.(ControlCacheItem)
	user, err := h.GetUserByID(controlCodeItem.uid)
	if err != nil {
		errRes := adminTemplateConfig{ErrorMsg: "提取用户信息失败"}
		err = json.NewEncoder(writer).Encode(&errRes)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return nil
	}
	return user
}
