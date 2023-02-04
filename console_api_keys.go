package headscale

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type KeysData struct {
	AuthKeys            []Key                `json:"authKeys"`
	InvalidAuthKeys     []InvalidKey         `json:"invalidAuthKeys"`     // 未实现
	ApiKeys             []Key                `json:"apiKeys"`             //未实现
	InvalidApiKeys      []InvalidKey         `json:"invalidApiKeys"`      //未实现
	OauthClients        []OauthClient        `json:"oauthClients"`        //未实现
	InvalidOauthClients []InvalidOauthClient `json:"invalidOauthClients"` //未实现
}

type InvalidOauthClient struct{} //未实现
type OauthClient struct{}        //未实现

type InvalidKey struct {
	KeyData Key    `json:"keyData"`
	Revoked string `json:"revoked"`
}

type Key struct {
	Id      string       `json:"id"`
	Created string       `json:"created"`
	Creator string       `json:"creator"`
	Expiry  string       `json:"expiry"`
	Type    string       `json:"type"` //"authkey","apikey"
	Authkey AuthKeyTypes `json:"authkey"`
	Apikey  ApiKeyTypes  `json:"apikey"`
}
type AuthKeyTypes struct {
	Reusable      bool `json:"reusable"`
	Ephemeral     bool `json:"ephemeral"`
	Preauthorized bool `json:"preauthorized"` //未实现，建议true
	ForAdminPanel bool `json:"forAdminPanel"` //未实现，未知含义，建议false
}
type ApiKeyTypes struct {
	Api string `json:"api"` //"control"
}

type GenKeyData struct {
	Id      string `json:"id"`
	FullKey string `json:"fullKey"`
	Created string `json:"created"`
	Expiry  string `json:"expiry"`
}

// 接受/admin/api/keys的Get请求，用于查询AuthKey
func (h *Headscale) CAPIGetKeys(
	w http.ResponseWriter,
	r *http.Request,
) {
	userName := h.verifyTokenIDandGetUser(w, r)
	if userName == "" {
		h.doAPIResponse(w, "用户信息核对失败", nil)
		return
	}

	authKeys, err := h.ListPreAuthKeys(userName)
	if err != nil {
		h.doAPIResponse(w, "授权密钥查询失败", nil)
		return
	}
	resData := KeysData{}
	resData.AuthKeys = make([]Key, 0)
	for _, key := range authKeys {
		tmpAuthKey := Key{
			Id:      key.Key[:12], //key.ID,
			Created: Time2SHString(*key.CreatedAt),
			Creator: key.User.Name,
			Expiry:  Time2SHString(*key.Expiration),
			Type:    "authkey",
			Authkey: AuthKeyTypes{
				Reusable:      key.Reusable,
				Ephemeral:     key.Ephemeral,
				Preauthorized: true,  //TODO
				ForAdminPanel: false, //TODO
			},
		}
		resData.AuthKeys = append(resData.AuthKeys, tmpAuthKey)
	}
	h.doAPIResponse(w, "", resData)
}

// 请求报文：{"keyData":{"type":"authkey","expirySeconds":7776000,"authkey":{"ephemeral":false,"reusable":false,"preauthorized":false}}}
type GenKeyREQ struct {
	KeyData REQKeyData `json:"keyData"`
}
type REQKeyData struct {
	Type          string       `json:"type"` //"authkey"
	ExpirySeconds uint64       `json:"expirySeconds"`
	Authkey       AuthKeyTypes `json:"authkey"`
}

// 接受/admin/api/keys的Post请求，用于创建AuthKey
func (h *Headscale) CAPIPostKeys(
	w http.ResponseWriter,
	r *http.Request,
) {
	userName := h.verifyTokenIDandGetUser(w, r)
	if userName == "" {
		h.doAPIResponse(w, "用户信息核对失败", nil)
		return
	}
	err := r.ParseForm()
	if err != nil {
		h.doAPIResponse(w, "用户请求解析失败:"+err.Error(), nil)
		return
	}
	reqData := GenKeyREQ{}
	json.NewDecoder(r.Body).Decode(&reqData)
	switch reqData.KeyData.Type {
	case "authkey":
		keyCfg := reqData.KeyData.Authkey
		keyExpiration := time.Now().Add(time.Duration(reqData.KeyData.ExpirySeconds) * time.Second)
		genedAuthKey, err := h.CreatePreAuthKey(userName, keyCfg.Reusable, keyCfg.Ephemeral, &keyExpiration, nil)
		if err != nil {
			h.doAPIResponse(w, "授权密钥创建失败", nil)
			return
		}
		resData := GenKeyData{
			Id:      genedAuthKey.Key[:12], //genedAuthKey.ID,
			FullKey: genedAuthKey.Key,
			Created: Time2SHString(*genedAuthKey.CreatedAt),
			Expiry:  Time2SHString(*genedAuthKey.Expiration),
		}
		h.doAPIResponse(w, "", resData)
	}
}

// 注销Key执行DELETE方法api/keys/:Id
func (h *Headscale) CAPIDelKeys(
	w http.ResponseWriter,
	r *http.Request,
) {
	userName := h.verifyTokenIDandGetUser(w, r)
	targetKeyID := strings.TrimPrefix(r.URL.Path, "/admin/api/keys/")
	allKeys, err := h.ListPreAuthKeys(userName)
	if err != nil {
		h.doAPIResponse(w, "查询用户密钥信息失败", nil)
		return
	}
	toDelKeys := make([]PreAuthKey, 0)
	for _, key := range allKeys {
		if key.Key[:12] == targetKeyID {
			toDelKeys = append(toDelKeys, key)
		}
	}
	if len(toDelKeys) == 0 {
		h.doAPIResponse(w, "该密钥不存在", nil)
		return
	} else if len(toDelKeys) > 1 {
		h.doAPIResponse(w, "存在多个密钥具备相同短形式（ID），请联系工作人员", nil)
		return
	}
	err = h.DestroyPreAuthKey(toDelKeys[0])
	if err != nil {
		h.doAPIResponse(w, "执行密钥删除失败", nil)
		return
	}
	h.doAPIResponse(w, "", targetKeyID)
}
