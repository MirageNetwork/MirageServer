package controller

import (
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

type NaviQueryRes struct {
	DeployPub     string            `json:"DeployPub"`
	BannedRegions []int             `json:"BannedRegions"`
	Regions       []NaviQueryRegion `json:"Regions"`
}
type NaviQueryRegion struct {
	Region NaviRegion `json:"Region"`
	Nodes  []NaviNode `json:"Nodes"`
}

// 接受/admin/api/derp/query的Get请求，用于进行DERP查询
func (m *Mirage) CAPIQueryDERP(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, err := m.verifyTokenIDandGetUser(w, r)
	if err != nil || user.CheckEmpty() {
		m.doAPIResponse(w, "用户信息核对失败:"+err.Error(), nil)
		return
	}

	resData := NaviQueryRes{}
	naviRegions := m.ListNaviRegions()
	for _, naviRegion := range naviRegions {
		if naviRegion.OrgID == 0 || naviRegion.OrgID == user.OrganizationID {
			naviNodes := m.ListNaviNodes(naviRegion.ID)
			for index := range naviNodes { // 清除掉敏感信息
				if naviNodes[index].NaviRegion.OrgID == 0 {
					latency := naviNodes[index].Statics.Latency
					naviNodes[index].Statics = NaviStatus{
						Latency: latency,
					}
					naviNodes[index].Arch = "common"
				} else if naviNodes[index].NaviKey != "" && naviNodes[index].Arch == "external" {
					naviNodes[index].Arch = "unknown"
				}
				naviNodes[index].NaviKey = ""
				naviNodes[index].SSHPwd = ""
				naviNodes[index].DNSKey = ""
			}
			resData.Regions = append(resData.Regions, NaviQueryRegion{
				Region: naviRegion,
				Nodes:  naviNodes,
			})
		}
	}

	org, err := m.GetOrgnaizationByID(user.OrganizationID)
	if err != nil {
		m.doAPIResponse(w, "用户组织查询失败:"+err.Error(), nil)
		return
	}

	if org.NaviDeployKey == "" {
		pri, pub, err := genSSHKeypair()
		if err != nil {
			log.Error().Msg(err.Error())
			m.doAPIResponse(w, "用户组织部署密钥生成失败:"+err.Error(), nil)
			return
		}
		org.NaviDeployPub = pub
		org.NaviDeployKey = pri
		err = m.db.Save(&org).Error
		if err != nil {
			log.Error().Msg(err.Error())
			m.doAPIResponse(w, "用户组织部署密钥生成失败:"+err.Error(), nil)
			return
		}
	}

	resData.BannedRegions = []int{}
	for id := range org.NaviBanList {
		resData.BannedRegions = append(resData.BannedRegions, id)
	}

	resData.DeployPub = org.NaviDeployPub

	m.doAPIResponse(w, "", resData)
}

// 接受/admin/api/derp/add的Post请求，用于进行DERP登记以及部署
func (m *Mirage) CAPIAddDERP(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, err := m.verifyTokenIDandGetUser(w, r)
	if err != nil || user.CheckEmpty() {
		m.doAPIResponse(w, "用户信息核对失败:"+err.Error(), nil)
		return
	}

	reqData := NaviDeployREQ{}
	json.NewDecoder(r.Body).Decode(&reqData)

	if reqData.NaviNode.SSHAddr == "" {
		// 非受管模式
		// 开始处理服务端数据信息
		if reqData.NaviNode.NaviRegionID == -1 {
			naviRegion := &NaviRegion{
				OrgID:      user.OrganizationID,
				RegionCode: reqData.RegionCode,
				RegionName: reqData.RegionName,
			}
			naviRegion = m.CreateNaviRegion(naviRegion)
			if naviRegion == nil {
				m.doAPIResponse(w, "创建区域失败", nil)
				return
			}
			reqData.NaviNode.NaviRegionID = naviRegion.ID
		} else {
			naviRegion := m.GetNaviRegion(reqData.NaviNode.NaviRegionID)
			if naviRegion == nil {
				m.doAPIResponse(w, "区域不存在", nil)
				return
			}
			if naviRegion.OrgID == 0 || naviRegion.OrgID != user.OrganizationID {
				m.doAPIResponse(w, "该区域无权限", nil)
				return
			}
		}

		derpid := uuid.New().String()
		reqData.NaviNode.ID = derpid
		reqData.NaviNode.Arch = "external"
		naviNode := m.CreateNaviNode(&reqData.NaviNode)
		if naviNode == nil {
			m.doAPIResponse(w, "新建司南档案失败", nil)
			return
		}

		m.setOrgLastStateChangeToNow(user.OrganizationID)

		m.CAPIQueryDERP(w, r)
		return
	}

	remoteAuth := []ssh.AuthMethod{}
	if reqData.NaviNode.SSHPwd != "" {
		remoteAuth = append(remoteAuth, ssh.Password(reqData.NaviNode.SSHPwd))
	} else {
		var keyData []byte
		if user.Organization.NaviDeployKey != "" {
			keyData = []byte(user.Organization.NaviDeployKey)
		} else {
			m.doAPIResponse(w, "不存在远程主机认证信息", nil)
			return
		}

		pemBlock, _ := pem.Decode(keyData)
		if pemBlock == nil {
			m.doAPIResponse(w, "解析私钥失败", nil)
			return
		}

		ecdsaKey, err := x509.ParseECPrivateKey(pemBlock.Bytes)
		if err != nil {
			m.doAPIResponse(w, "解析私钥失败:"+err.Error(), nil)
			return
		}
		pk, err := ssh.NewSignerFromKey(ecdsaKey)
		if err != nil {
			m.doAPIResponse(w, "解析私钥失败:"+err.Error(), nil)
			return
		}
		remoteAuth = append(remoteAuth, ssh.PublicKeys(pk))
	}

	client, err := ssh.Dial("tcp", reqData.NaviNode.SSHAddr, &ssh.ClientConfig{
		User:            "root",
		Auth:            remoteAuth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		m.doAPIResponse(w, "连接远程主机失败:"+err.Error(), nil)
		return
	}
	archCheckSession, err := client.NewSession()
	if err != nil {
		m.doAPIResponse(w, "创建会话失败:"+err.Error(), nil)
		return
	}
	defer archCheckSession.Close()

	// 检查目标机处理器架构以便传送对应版本
	arch, err := archCheckSession.Output("arch")
	if err != nil {
		m.doAPIResponse(w, "执行命令失败:"+err.Error(), nil)
		return
	}
	archStr := strings.TrimSuffix(string(arch), "\n")
	if archStr != "x86_64" && archStr != "aarch64" {
		m.doAPIResponse(w, "不支持的处理器架构:"+archStr, nil)
		return
	}

	// 开始处理服务端数据信息
	if reqData.NaviNode.NaviRegionID == -1 {
		naviRegion := &NaviRegion{
			OrgID:      user.OrganizationID,
			RegionCode: reqData.RegionCode,
			RegionName: reqData.RegionName,
		}
		naviRegion = m.CreateNaviRegion(naviRegion)
		if naviRegion == nil {
			m.doAPIResponse(w, "创建区域失败", nil)
			return
		}
		reqData.NaviNode.NaviRegionID = naviRegion.ID
	} else {
		naviRegion := m.GetNaviRegion(reqData.NaviNode.NaviRegionID)
		if naviRegion == nil {
			m.doAPIResponse(w, "区域不存在", nil)
			return
		}
		if naviRegion.OrgID == 0 || naviRegion.OrgID != user.OrganizationID {
			m.doAPIResponse(w, "该区域无权限", nil)
			return
		}
	}

	// TODO: 是否需要检查目标机曾部署过司南
	derpid := uuid.New().String()
	reqData.NaviNode.ID = derpid
	reqData.NaviNode.Arch = archStr
	naviNode := m.CreateNaviNode(&reqData.NaviNode)
	if naviNode == nil {
		m.doAPIResponse(w, "新建司南档案失败", nil)
		return
	}

	//TODO: 司南建档成功后在目标机执行部署启动
	// 停止服务
	systemdStopSession, err := client.NewSession()
	if err != nil {
		m.doAPIResponse(w, "创建会话失败:"+err.Error(), nil)
		return
	}
	defer systemdStopSession.Close()

	_, err = systemdStopSession.Output("systemctl stop MirageNavi")
	if err != nil {
		m.doAPIResponse(w, "执行命令失败:"+err.Error(), nil)
		return
	}

	err = sshSendFile(client, "download/"+archStr+"/MirageNavi", "/usr/local/bin/MirageNavi")
	if err != nil {
		m.doAPIResponse(w, "传送司南客户端到目标服务器失败:"+err.Error(), nil)
		return
	}
	// 进行赋权
	chmodSession, err := client.NewSession()
	if err != nil {
		m.doAPIResponse(w, "创建会话失败:"+err.Error(), nil)
		return
	}
	defer chmodSession.Close()

	_, err = chmodSession.Output("chmod +x /usr/local/bin/MirageNavi")
	if err != nil {
		m.doAPIResponse(w, "执行命令失败:"+err.Error(), nil)
		return
	}

	serviceScript :=
		`[Unit]
Description=Mirage Navigation Node Service
	
[Service]
ExecStart=/usr/local/bin/MirageNavi -ctrl-url ${MIRAGE_CTRL_URL} -id ${MIRAGE_NAVI_ID} >> ${LOG_DIR}/MirageNavi.log 2>&1
Restart=always
User=root
Group=root
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=LOG_DIR=/var/log
Environment=MIRAGE_CTRL_URL=https://` + m.cfg.ServerURL + `
Environment=MIRAGE_NAVI_ID=` + derpid + `
	
[Install]
WantedBy=multi-user.target`

	// 将文本写入文件
	err = os.WriteFile("download/"+derpid+".service.tmp", []byte(serviceScript), 0644)
	if err != nil {
		m.doAPIResponse(w, "创建临时服务文件失败:"+err.Error(), nil)
		return
	}
	err = sshSendFile(client, "download/"+derpid+".service.tmp", "/etc/systemd/system/MirageNavi.service")
	if err != nil {
		m.doAPIResponse(w, "传送服务文件到目标服务器失败:"+err.Error(), nil)
		return
	}
	err = os.Remove("download/" + derpid + ".service.tmp")
	if err != nil {
		log.Warn().Caller().Err(err).Msg("删除服务临时文件失败")
		return
	}

	// 重置服务配置
	systemdReloadSession, err := client.NewSession()
	if err != nil {
		m.doAPIResponse(w, "创建会话失败:"+err.Error(), nil)
		return
	}
	defer systemdReloadSession.Close()

	_, err = systemdReloadSession.Output("systemctl daemon-reload")
	if err != nil {
		m.doAPIResponse(w, "执行命令失败:"+err.Error(), nil)
		return
	}

	// 启动服务
	systemdEnableSession, err := client.NewSession()
	if err != nil {
		m.doAPIResponse(w, "创建会话失败:"+err.Error(), nil)
		return
	}
	defer systemdEnableSession.Close()

	_, err = systemdEnableSession.Output("systemctl enable --now MirageNavi")
	if err != nil {
		m.doAPIResponse(w, "执行命令失败:"+err.Error(), nil)
		return
	}

	m.setOrgLastStateChangeToNow(user.OrganizationID)

	m.CAPIQueryDERP(w, r)
}

func (m *Mirage) CAPIDelNaviNode(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, err := m.verifyTokenIDandGetUser(w, r)
	if err != nil || user.CheckEmpty() {
		m.doAPIResponse(w, "用户信息核对失败:"+err.Error(), nil)
		return
	}

	vars := mux.Vars(r)
	naviID, ok := vars["id"]
	if !ok {
		m.doAPIResponse(w, "未指定司南ID", nil)
		return
	}

	naviNode := m.GetNaviNode(naviID)
	if naviNode == nil {
		m.doAPIResponse(w, "未找到该司南档案", nil)
		return
	}
	if naviNode.NaviRegion.OrgID == 0 || naviNode.NaviRegion.OrgID != user.OrganizationID {
		m.doAPIResponse(w, "没有权限", nil)
		return
	}
	// TODO: 是否有必要做远程关闭？

	if err := m.db.Delete(&naviNode).Error; err != nil {
		m.doAPIResponse(w, "数据库删除司南节点失败:"+err.Error(), nil)
		return
	}
	delete(m.DERPNCs, naviID)
	delete(m.DERPseqnum, naviID)
	//	c.App.LoadDERPMapFromURL(c.App.cfg.DERPURL)

	m.setOrgLastStateChangeToNow(user.OrganizationID)

	m.CAPIQueryDERP(w, r)
}

func (m *Mirage) CAPISwitchRegionBan(
	w http.ResponseWriter,
	r *http.Request,
) {
	user, err := m.verifyTokenIDandGetUser(w, r)
	if err != nil || user.CheckEmpty() {
		m.doAPIResponse(w, "用户信息核对失败:"+err.Error(), nil)
		return
	}

	vars := mux.Vars(r)
	regionIDStr, ok := vars["id"]
	if !ok {
		m.doAPIResponse(w, "未指定区域ID", nil)
		return
	}
	regionID, err := strconv.Atoi(regionIDStr)
	if err != nil {
		m.doAPIResponse(w, "区域ID格式错误", nil)
		return
	}

	region := m.GetNaviRegion(regionID)
	if region == nil || region.OrgID != 0 {
		m.doAPIResponse(w, "未找到该区域档案", nil)
		return
	}

	org, err := m.GetOrgnaizationByID(user.OrganizationID)
	if err != nil {
		m.doAPIResponse(w, "查询用户所属组织失败", nil)
		return
	}
	if _, ok := org.NaviBanList[regionID]; ok {
		delete(org.NaviBanList, regionID)
	} else {
		if org.NaviBanList == nil {
			org.NaviBanList = make(map[int]struct{})
		}
		org.NaviBanList[regionID] = struct{}{}
	}

	if err := m.db.Save(org).Error; err != nil {
		m.doAPIResponse(w, "数据库更新组织禁用区域信息失败:"+err.Error(), nil)
		return
	}

	m.setOrgLastStateChangeToNow(org.ID)

	m.CAPIQueryDERP(w, r)
}
