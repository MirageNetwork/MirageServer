package controller

import (
	_ "embed"
	"encoding/json"
	"net/http"
)

// 网络设置响应Data体
type NetSettingResData struct {
	FileSharing        bool   `json:"fileSharing"`
	ServicesCollection bool   `json:"servicesCollection"`
	HttpsEnabled       bool   `json:"httpsEnabled"`
	Provider           string `json:"provider"`
	MachineAuthNeeded  bool   `json:"machineAuthNeeded"`
	MaxKeyDurationDays int    `json:"maxKeyDurationDays"`
	NetworkLockEnabled bool   `json:"networkLockEnabled"`
}

// 查询网络设置API
func (h *Mirage) getNetSettingAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	user := h.verifyTokenIDandGetUser(writer, req)
	if user.CheckEmpty() {
		h.doAPIResponse(writer, "用户信息核对失败", nil)
		return
	}
	netsettingData := NetSettingResData{
		FileSharing:        false,         //未实现
		ServicesCollection: false,         //未实现
		HttpsEnabled:       false,         //未实现
		Provider:           "Mirage SaaS", //在个人版尚未开启更多验证方式时暂时统一设置
		MachineAuthNeeded:  false,         //未实现
		MaxKeyDurationDays: 180,
		NetworkLockEnabled: false, //未实现
	}
	netsettingData.MaxKeyDurationDays = int(user.Organization.ExpiryDuration)
	h.doAPIResponse(writer, "", netsettingData)
}

// 更新用户网络密钥过期时长
func (h *Mirage) ConsoleUpdateKeyExpiryAPI(
	writer http.ResponseWriter,
	req *http.Request,
) {
	user := h.verifyTokenIDandGetUser(writer, req)
	if user.CheckEmpty() {
		h.doAPIResponse(writer, "用户信息核对失败", nil)
		return
	}
	err := req.ParseForm()
	if err != nil {
		h.doAPIResponse(writer, "用户请求解析失败:"+err.Error(), nil)
		return
	}
	reqData := make(map[string]int)
	json.NewDecoder(req.Body).Decode(&reqData)
	newExpiryDuration := reqData["maxKeyDurationDays"]
	//	newExpiryDuration, err := strconv.Atoi(newExpiryDurationStr)
	if err != nil {
		h.doAPIResponse(writer, "从请求获取新值失败:"+err.Error(), nil)
		return
	}
	err = h.UpdateUserKeyExpiry(user, uint(newExpiryDuration))
	if err != nil {
		h.doAPIResponse(writer, "更新密钥过期时长失败:"+err.Error(), nil)
		return
	}
	h.doAPIResponse(writer, "", uint(newExpiryDuration))
}
