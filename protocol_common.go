package Mirage

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"tailscale.com/tailcfg"
	"tailscale.com/types/key"
)

const (
	// The CapabilityVersion is used by Tailscale clients to indicate
	// their codebase version. Tailscale clients can communicate over TS2021
	// from CapabilityVersion 28, but we only have good support for it
	// since https://github.com/tailscale/tailscale/pull/4323 (Noise in any HTTPS port).
	//
	// Related to this change, there is https://github.com/tailscale/tailscale/pull/5379,
	// where CapabilityVersion 39 is introduced to indicate #4323 was merged.
	//
	// See also https://github.com/tailscale/tailscale/blob/main/tailcfg/tailcfg.go
	NoiseCapabilityVersion = 39
)

// KeyHandler provides the Mirage pub key
// Listens in /key.
func (h *Mirage) KeyHandler(
	writer http.ResponseWriter,
	req *http.Request,
) {
	// New Tailscale clients send a 'v' parameter to indicate the CurrentCapabilityVersion
	clientCapabilityStr := req.URL.Query().Get("v")

	log.Debug().
		Str("handler", "/key").
		Str("v", clientCapabilityStr).
		Msg("New noise client")
	_, err := strconv.Atoi(clientCapabilityStr) // cgao6: 版本号暂时不判断不使用
	if err != nil {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusBadRequest)
		_, err := writer.Write([]byte("Wrong params"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		return
	}

	// TS2021 (Tailscale v2 protocol) requires to have a different key
	// cgao6: 我们只支持不低于39版本的客户端
	//if clientCapabilityVersion >= NoiseCapabilityVersion {
	resp := tailcfg.OverTLSPublicKeyResponse{
		LegacyPublicKey: key.MachinePublic{},
		PublicKey:       h.noisePrivateKey.Public(),
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(resp)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}

	return
}

// cgao6: HS原本版本中存在诸多不合理的处理，这里我们需要根据自己的理解使用自己的版本
func (h *Mirage) handleRegisterCommon(
	writer http.ResponseWriter,
	req *http.Request,
	registerRequest tailcfg.RegisterRequest,
	machineKey key.MachinePublic,
) {
	now := time.Now().UTC()
	// 这一步目前考虑不使用MachineKey
	machine, _ := h.GetMachineByNodeKey(registerRequest.NodeKey)

	// 机器已存在，意味着：
	// - 正常使用(NodeKey一致、未过期、未设置要求过期)
	// - NodeKey一致但设置过期
	// - NodeKey一致但已过期（此时应该当做不一致处理，因为旧的过期的本该删除）
	// - 无一致NodeKey
	// 后两种当新的处理，后续认证过后我们会再将原先的同用户node替掉或者创建新的
	if machine != nil {
		if !registerRequest.Expiry.IsZero() &&
			registerRequest.Expiry.UTC().Before(now) {
			h.handleMachineLogOutCommon(writer, *machine, machineKey)
			return
		}
		if !machine.isExpired() {
			h.handleMachineValidRegistrationCommon(writer, *machine, machineKey)
			return
		}
	}
	//cgao6: 因为除去NodeKey一致（正常）和NodeKey一致（请求过期）两种外我们预计同样处理，故后续不用再判断

	//cgao6: 授权密钥注册模式 //TODO: 后续需要对授权密钥注册进行检查
	if registerRequest.Auth.AuthKey != "" {
		h.handleAuthKeyCommon(writer, registerRequest, machineKey)
		return
	}

	// TODO: cgao6: 我们需要对Followup的使用要进一步思索
	// 这里是非常有价值修复的问题所在
	if registerRequest.Followup != "" {
		aCode := registerRequest.Followup[len(registerRequest.Followup)-12:]
		if _, ok := h.aCodeCache.Get(aCode); ok {
			log.Debug().
				Caller().
				Str("machine", registerRequest.Hostinfo.Hostname).
				Str("machine_key", machineKey.ShortString()).
				Str("node_key", registerRequest.NodeKey.ShortString()).
				Str("node_key_old", registerRequest.OldNodeKey.ShortString()).
				Str("follow_up", registerRequest.Followup).
				Msg("Machine is waiting for interactive login")

			longPollChan := make(chan string)
			h.longPollChanPool[aCode] = longPollChan
			select {
			case <-req.Context().Done():
				fmt.Println("DEBUG: 客户端断开long poll")
				return
			case loginNoticeMsg := <-longPollChan:
				delete(h.longPollChanPool, aCode)
				if loginNoticeMsg == "ok" {
					h.sendLoginSuccess(writer, machineKey)
				}
				return
			}
		}
	}

	log.Info().
		Caller().
		Str("machine", registerRequest.Hostinfo.Hostname).
		Str("machine_key", machineKey.ShortString()).
		Str("node_key", registerRequest.NodeKey.ShortString()).
		Str("node_key_old", registerRequest.OldNodeKey.ShortString()).
		Str("follow_up", registerRequest.Followup).
		Msg("New machine not yet in the database")

	// TODO: 原本对机器的givenName的随机数模式并不优雅，要改成模仿TS的做法（即在后面加-<递增数字>的方法）
	// 而且要在实际入库时做

	// 创建aCode缓存用来后续注册使用
	// 因为过期时间取决于用户的过期设置，故此处不必记录！
	// TODO: 创建ACode时是否要记录MachineKey？？？
	log.Debug().Caller().Str("machine", registerRequest.Hostinfo.Hostname).Msg("The node seems to be new, sending auth url")
	aCode := h.GenACode()
	stateCode := h.GenStateCode()
	h.aCodeCache.Set(
		aCode,
		ACacheItem{
			stateCode: stateCode,
			uid:       -1,
			mKey:      machineKey,
			regReq:    registerRequest,
		},
		time.Now().AddDate(0, 1, 0).Sub(time.Now()),
	)
	h.stateCodeCache.Set(
		stateCode,
		StateCacheItem{
			nextURL:    "/a/" + aCode,
			uid:        -1,
			machineKey: machineKey,
		},
		time.Now().AddDate(0, 1, 0).Sub(time.Now()),
	)
	// 创建新acode时，将原先机器对应的controlCode全部清除
	if machineControlCodeC, ok := h.machineControlCodeCache.Get(machineKey.String()); ok {
		for _, controlCode := range machineControlCodeC.(MachineControlCodeCacheItem).controlCodes {
			h.controlCodeCache.Delete(controlCode)
		}
	}

	h.SendACode(writer, aCode, registerRequest, machineKey)

	return
}

type ACacheItem struct {
	stateCode string
	mKey      key.MachinePublic
	regReq    tailcfg.RegisterRequest
	uid       tailcfg.UserID
}

func (h *Mirage) GenACode() string {
	randomBlob := make([]byte, 6)
	if _, err := rand.Read(randomBlob); err != nil {
		log.Error().
			Caller().
			Msg("could not read 6 bytes from rand")
		return ""
	}
	stateStr := hex.EncodeToString(randomBlob)[:12]
	return stateStr
}

func (h *Mirage) GenStateCode() string {
	const letterBytes = "_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 25)
	b[0] = 'm'
	b[1] = 'n'
	b[2] = '-'

	index := make([]byte, 22)
	rand.Read(index)
	for i := 0; i < 22; i++ {
		b[i+3] = letterBytes[index[i]&63]
	}
	return string(b)
}

// cgao6: 用来测试longpoll解决方案，返回空authURL值
func (h *Mirage) sendLoginSuccess(
	writer http.ResponseWriter,
	machineKey key.MachinePublic,
) {
	resp := tailcfg.RegisterResponse{}

	resp.AuthURL = ""

	respBody, err := h.marshalResponse(resp, machineKey)
	if err != nil {
		log.Error().Caller().Err(err).Msg("Cannot encode message")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(respBody)
	if err != nil {
		log.Error().Caller().Err(err).Msg("Failed to write response")
	}

	log.Info().Caller().Msg("Successfully Empty authURL")
}

// cgao6: 替代handleNewMachineCommon，处理新设备注册，变更了返回值
func (h *Mirage) SendACode(
	writer http.ResponseWriter,
	aCode string,
	registerRequest tailcfg.RegisterRequest,
	machineKey key.MachinePublic,
) {
	resp := tailcfg.RegisterResponse{}

	resp.AuthURL = fmt.Sprintf(
		"%s/a/%s",
		strings.TrimSuffix(h.cfg.ServerURL, "/"),
		aCode,
	)

	respBody, err := h.marshalResponse(resp, machineKey)
	if err != nil {
		log.Error().Caller().Err(err).Msg("Cannot encode message")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(respBody)
	if err != nil {
		log.Error().Caller().Err(err).Msg("Failed to write response")
	}

	log.Info().Caller().Str("AuthURL", resp.AuthURL).Str("machine", registerRequest.Hostinfo.Hostname).Msg("Successfully sent auth url")
}

// handleAuthKeyCommon contains the logic to manage auth key client registration
// It is used both by the legacy and the new Noise protocol.
//
// TODO: check if any locks are needed around IP allocation.
func (h *Mirage) handleAuthKeyCommon(
	writer http.ResponseWriter,
	registerRequest tailcfg.RegisterRequest,
	machineKey key.MachinePublic,
) {
	log.Debug().
		Str("func", "handleAuthKeyCommon").
		Str("machine", registerRequest.Hostinfo.Hostname).
		Msgf("Processing auth key for %s", registerRequest.Hostinfo.Hostname)
	resp := tailcfg.RegisterResponse{}

	pak, err := h.checkKeyValidity(registerRequest.Auth.AuthKey)
	if err != nil {
		log.Error().
			Caller().
			Str("func", "handleAuthKeyCommon").
			Str("machine", registerRequest.Hostinfo.Hostname).
			Err(err).
			Msg("Failed authentication via AuthKey")
		resp.MachineAuthorized = false

		respBody, err := h.marshalResponse(resp, machineKey)
		if err != nil {
			log.Error().
				Caller().
				Str("func", "handleAuthKeyCommon").
				Str("machine", registerRequest.Hostinfo.Hostname).
				Err(err).
				Msg("Cannot encode message")
			http.Error(writer, "Internal server error", http.StatusInternalServerError)

			return
		}

		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(http.StatusUnauthorized)
		_, err = writer.Write(respBody)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		log.Error().
			Caller().
			Str("func", "handleAuthKeyCommon").
			Str("machine", registerRequest.Hostinfo.Hostname).
			Msg("Failed authentication via AuthKey")

		return
	}

	log.Debug().
		Str("func", "handleAuthKeyCommon").
		Str("machine", registerRequest.Hostinfo.Hostname).
		Msg("Authentication key was valid, proceeding to acquire IP addresses")

	nodeKey := NodePublicKeyStripPrefix(registerRequest.NodeKey)

	// retrieve machine information if it exist
	// The error is not important, because if it does not
	// exist, then this is a new machine and we will move
	// on to registration.
	//machine, _ := h.GetMachineByAnyKey(machineKey, registerRequest.NodeKey, registerRequest.OldNodeKey)
	machine, _ := h.GetMachineByAnyKey(key.MachinePublic{}, registerRequest.NodeKey, registerRequest.OldNodeKey)
	if machine != nil {
		log.Trace().
			Caller().
			Str("machine", machine.Hostname).
			Msg("machine was already registered before, refreshing with new auth key")

		machine.NodeKey = nodeKey
		machine.AuthKeyID = uint(pak.ID)
		err := h.RefreshMachine(machine, registerRequest.Expiry)
		if err != nil {
			log.Error().
				Caller().
				Str("machine", machine.Hostname).
				Err(err).
				Msg("Failed to refresh machine")

			return
		}

		aclTags := pak.GetAclTags()
		if len(aclTags) > 0 {
			// This conditional preserves the existing behaviour, although SaaS would reset the tags on auth-key login
			err = h.SetTags(machine, aclTags)

			if err != nil {
				log.Error().
					Caller().
					Str("machine", machine.Hostname).
					Strs("aclTags", aclTags).
					Err(err).
					Msg("Failed to set tags after refreshing machine")

				return
			}
		}
	} else {
		now := time.Now().UTC()

		givenName, err := h.GenerateGivenName(registerRequest.Hostinfo.BackendLogID, registerRequest.Hostinfo.Hostname)
		//		givenName, err := h.GenerateGivenName(MachinePublicKeyStripPrefix(machineKey), registerRequest.Hostinfo.Hostname)
		if err != nil {
			log.Error().
				Caller().
				Str("func", "RegistrationHandler").
				Str("hostinfo.name", registerRequest.Hostinfo.Hostname).
				Err(err)

			return
		}

		machineToRegister := Machine{
			Hostname:       registerRequest.Hostinfo.Hostname,
			GivenName:      givenName,
			UserID:         pak.User.ID,
			MachineKey:     MachinePublicKeyStripPrefix(machineKey),
			RegisterMethod: RegisterMethodAuthKey,
			Expiry:         &registerRequest.Expiry,
			NodeKey:        nodeKey,
			LastSeen:       &now,
			AuthKeyID:      uint(pak.ID),
			ForcedTags:     pak.GetAclTags(),
		}

		machine, err = h.RegisterMachine(
			machineToRegister,
		)
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("could not register machine")
			http.Error(writer, "Internal server error", http.StatusInternalServerError)

			return
		}
	}

	err = h.UsePreAuthKey(pak)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to use pre-auth key")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)

		return
	}

	resp.MachineAuthorized = true
	resp.User = *pak.User.toTailscaleUser()
	// Provide LoginName when registering with pre-auth key
	// Otherwise it will need to exec `tailscale up` twice to fetch the *LoginName*
	resp.Login = *pak.User.toTailscaleLogin()

	respBody, err := h.marshalResponse(resp, machineKey)
	if err != nil {
		log.Error().
			Caller().
			Str("func", "handleAuthKeyCommon").
			Str("machine", registerRequest.Hostinfo.Hostname).
			Err(err).
			Msg("Cannot encode message")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)

		return
	}
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(respBody)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}

	log.Info().
		Str("func", "handleAuthKeyCommon").
		Str("machine", registerRequest.Hostinfo.Hostname).
		Str("ips", strings.Join(machine.IPAddresses.ToStringSlice(), ", ")).
		Msg("Successfully authenticated via AuthKey")
}

func (h *Mirage) handleMachineLogOutCommon(
	writer http.ResponseWriter,
	machine Machine,
	machineKey key.MachinePublic,
) {
	resp := tailcfg.RegisterResponse{}

	log.Info().
		Str("machine", machine.Hostname).
		Msg("Client requested logout")

	err := h.ExpireMachine(&machine)
	if err != nil {
		log.Error().
			Caller().
			Str("func", "handleMachineLogOutCommon").
			Err(err).
			Msg("Failed to expire machine")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)

		return
	}

	resp.AuthURL = ""
	resp.MachineAuthorized = false
	resp.NodeKeyExpired = true
	resp.User = *machine.User.toTailscaleUser()
	respBody, err := h.marshalResponse(resp, machineKey)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Cannot encode message")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(respBody)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")

		return
	}

	if machine.isEphemeral() {
		err = h.HardDeleteMachine(&machine)
		if err != nil {
			log.Error().
				Err(err).
				Str("machine", machine.Hostname).
				Msg("Cannot delete ephemeral machine from the database")
		}

		return
	}

	log.Info().
		Caller().
		Str("machine", machine.Hostname).
		Msg("Successfully logged out")
}

func (h *Mirage) handleMachineValidRegistrationCommon(
	writer http.ResponseWriter,
	machine Machine,
	machineKey key.MachinePublic,
) {
	resp := tailcfg.RegisterResponse{}

	// The machine registration is valid, respond with redirect to /map
	log.Debug().
		Caller().
		Str("machine", machine.Hostname).
		Msg("Client is registered and we have the current NodeKey. All clear to /map")

	resp.AuthURL = ""
	resp.MachineAuthorized = true
	resp.User = *machine.User.toTailscaleUser()
	resp.Login = *machine.User.toTailscaleLogin()

	respBody, err := h.marshalResponse(resp, machineKey)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Cannot encode message")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(respBody)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}

	log.Info().
		Caller().
		Str("machine", machine.Hostname).
		Msg("Machine successfully authorized")
}

func (h *Mirage) handleMachineRefreshKeyCommon(
	writer http.ResponseWriter,
	registerRequest tailcfg.RegisterRequest,
	machine Machine,
	machineKey key.MachinePublic,
) {
	resp := tailcfg.RegisterResponse{}

	log.Info().
		Caller().
		Str("machine", machine.Hostname).
		Msg("We have the OldNodeKey in the database. This is a key refresh")
	machine.NodeKey = NodePublicKeyStripPrefix(registerRequest.NodeKey)

	if err := h.db.Save(&machine).Error; err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to update machine key in the database")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)

		return
	}

	resp.AuthURL = ""
	resp.User = *machine.User.toTailscaleUser()
	respBody, err := h.marshalResponse(resp, machineKey)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Cannot encode message")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(respBody)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}

	log.Info().
		Caller().
		Str("node_key", registerRequest.NodeKey.ShortString()).
		Str("old_node_key", registerRequest.OldNodeKey.ShortString()).
		Str("machine", machine.Hostname).
		Msg("Node key successfully refreshed")
}

func (h *Mirage) handleMachineExpiredOrLoggedOutCommon(
	writer http.ResponseWriter,
	registerRequest tailcfg.RegisterRequest,
	machine Machine,
	machineKey key.MachinePublic,
) {
	resp := tailcfg.RegisterResponse{}

	if registerRequest.Auth.AuthKey != "" {
		h.handleAuthKeyCommon(writer, registerRequest, machineKey)

		return
	}

	// The client has registered before, but has expired or logged out
	log.Trace().
		Caller().
		Str("machine", machine.Hostname).
		Str("machine_key", machineKey.ShortString()).
		Str("node_key", registerRequest.NodeKey.ShortString()).
		Str("node_key_old", registerRequest.OldNodeKey.ShortString()).
		Msg("Machine registration has expired or logged out. Sending a auth url to register")

	if h.oauth2Config != nil {
		resp.AuthURL = fmt.Sprintf("%s/oidc/register/%s",
			strings.TrimSuffix(h.cfg.ServerURL, "/"),
			registerRequest.NodeKey)
	} else {
		resp.AuthURL = fmt.Sprintf("%s/register/%s",
			strings.TrimSuffix(h.cfg.ServerURL, "/"),
			registerRequest.NodeKey)
	}

	respBody, err := h.marshalResponse(resp, machineKey)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Cannot encode message")
		http.Error(writer, "Internal server error", http.StatusInternalServerError)

		return
	}

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(respBody)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}

	log.Trace().
		Caller().
		Str("machine_key", machineKey.ShortString()).
		Str("node_key", registerRequest.NodeKey.ShortString()).
		Str("node_key_old", registerRequest.OldNodeKey.ShortString()).
		Str("machine", machine.Hostname).
		Msg("Machine logged out. Sent AuthURL for reauthentication")
}
