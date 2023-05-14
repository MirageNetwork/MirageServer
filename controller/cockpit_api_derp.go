package controller

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/sftp"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

type NaviDeployREQ struct {
	RegionCode string `json:"RegionCode"`
	RegionName string `json:"RegionName"`

	NaviNode NaviNode `json:"NaviNode"`
}

// 接受/cockpit/api/derp/query的Get请求，用于进行DERP查询
func (c *Cockpit) CAPIQueryDERP(
	w http.ResponseWriter,
	r *http.Request,
) {
	resData := []struct {
		Region NaviRegion `json:"Region"`
		Nodes  []NaviNode `json:"Nodes"`
	}{}
	naviRegions := c.ListNaviRegions()
	for _, naviRegion := range naviRegions {
		if naviRegion.OrgID == 0 {
			naviNodes := c.ListNaviNodes(naviRegion.ID)
			for index := range naviNodes { // 清除掉敏感信息
				if naviNodes[index].NaviKey != "" && naviNodes[index].Arch == "external" {
					naviNodes[index].Arch = "unknown"
				}
				naviNodes[index].NaviKey = ""
				naviNodes[index].SSHPwd = ""
				naviNodes[index].DNSKey = ""
			}
			resData = append(resData, struct {
				Region NaviRegion `json:"Region"`
				Nodes  []NaviNode `json:"Nodes"`
			}{
				Region: naviRegion,
				Nodes:  naviNodes,
			})
		}
	}
	c.doAPIResponse(w, "", resData)
}

// 接受/cockpit/api/derp/add的Post请求，用于进行DERP登记以及部署
func (c *Cockpit) CAPIAddDERP(
	w http.ResponseWriter,
	r *http.Request,
) {
	reqData := NaviDeployREQ{}
	json.NewDecoder(r.Body).Decode(&reqData)

	if reqData.NaviNode.SSHAddr == "" {
		// 非受管模式
		// 开始处理服务端数据信息
		if reqData.NaviNode.NaviRegionID == -1 {
			naviRegion := &NaviRegion{
				OrgID:      0,
				RegionCode: reqData.RegionCode,
				RegionName: reqData.RegionName,
			}
			naviRegion = c.CreateNaviRegion(naviRegion)
			if naviRegion == nil {
				c.doAPIResponse(w, "创建区域失败", nil)
				return
			}
			reqData.NaviNode.NaviRegionID = naviRegion.ID
		} else {
			naviRegion := c.GetNaviRegion(reqData.NaviNode.NaviRegionID)
			if naviRegion == nil {
				c.doAPIResponse(w, "区域不存在", nil)
				return
			}
		}

		derpid := uuid.New().String()
		reqData.NaviNode.ID = derpid
		reqData.NaviNode.Arch = "external"
		naviNode := c.CreateNaviNode(&reqData.NaviNode)
		if naviNode == nil {
			c.doAPIResponse(w, "新建司南档案失败", nil)
			return
		}
		c.CAPIQueryDERP(w, r)
		return
	}

	remoteAuth := []ssh.AuthMethod{}
	if reqData.NaviNode.SSHPwd != "" {
		remoteAuth = append(remoteAuth, ssh.Password(reqData.NaviNode.SSHPwd))
	} else {
		var keyData []byte

		sysCfg := c.GetSysCfg()
		if sysCfg != nil && sysCfg.NaviDeployKey != "" {
			keyData = []byte(sysCfg.NaviDeployKey)
		} else {
			c.doAPIResponse(w, "不存在远程主机认证信息", nil)
			return
		}

		pemBlock, _ := pem.Decode(keyData)
		if pemBlock == nil {
			c.doAPIResponse(w, "解析私钥失败", nil)
			return
		}

		ecdsaKey, err := x509.ParseECPrivateKey(pemBlock.Bytes)
		if err != nil {
			c.doAPIResponse(w, "解析私钥失败:"+err.Error(), nil)
			return
		}
		pk, err := ssh.NewSignerFromKey(ecdsaKey)
		if err != nil {
			c.doAPIResponse(w, "解析私钥失败:"+err.Error(), nil)
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
		c.doAPIResponse(w, "连接远程主机失败:"+err.Error(), nil)
		return
	}
	archCheckSession, err := client.NewSession()
	if err != nil {
		c.doAPIResponse(w, "创建会话失败:"+err.Error(), nil)
		return
	}
	defer archCheckSession.Close()

	// 检查目标机处理器架构以便传送对应版本
	arch, err := archCheckSession.Output("arch")
	if err != nil {
		c.doAPIResponse(w, "执行命令失败:"+err.Error(), nil)
		return
	}
	archStr := strings.TrimSuffix(string(arch), "\n")
	if archStr != "x86_64" && archStr != "aarch64" {
		c.doAPIResponse(w, "不支持的处理器架构:"+archStr, nil)
		return
	}

	// 开始处理服务端数据信息
	if reqData.NaviNode.NaviRegionID == -1 {
		naviRegion := &NaviRegion{
			OrgID:      0,
			RegionCode: reqData.RegionCode,
			RegionName: reqData.RegionName,
		}
		naviRegion = c.CreateNaviRegion(naviRegion)
		if naviRegion == nil {
			c.doAPIResponse(w, "创建区域失败", nil)
			return
		}
		reqData.NaviNode.NaviRegionID = naviRegion.ID
	} else {
		naviRegion := c.GetNaviRegion(reqData.NaviNode.NaviRegionID)
		if naviRegion == nil {
			c.doAPIResponse(w, "区域不存在", nil)
			return
		}
	}

	// TODO: 是否需要检查目标机曾部署过司南
	derpid := uuid.New().String()
	reqData.NaviNode.ID = derpid
	reqData.NaviNode.Arch = archStr
	naviNode := c.CreateNaviNode(&reqData.NaviNode)
	if naviNode == nil {
		c.doAPIResponse(w, "新建司南档案失败", nil)
		return
	}

	//TODO: 司南建档成功后在目标机执行部署启动
	// 停止服务
	systemdStopSession, err := client.NewSession()
	if err != nil {
		c.doAPIResponse(w, "创建会话失败:"+err.Error(), nil)
		return
	}
	defer systemdStopSession.Close()

	_, err = systemdStopSession.Output("systemctl stop MirageNavi")
	if err != nil {
		c.doAPIResponse(w, "执行命令失败:"+err.Error(), nil)
		return
	}

	err = sshSendFile(client, "download/"+archStr+"/MirageNavi", "/usr/local/bin/MirageNavi")
	if err != nil {
		c.doAPIResponse(w, "传送司南客户端到目标服务器失败:"+err.Error(), nil)
		return
	}
	// 进行赋权
	chmodSession, err := client.NewSession()
	if err != nil {
		c.doAPIResponse(w, "创建会话失败:"+err.Error(), nil)
		return
	}
	defer chmodSession.Close()

	_, err = chmodSession.Output("chmod +x /usr/local/bin/MirageNavi")
	if err != nil {
		c.doAPIResponse(w, "执行命令失败:"+err.Error(), nil)
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
Environment=MIRAGE_CTRL_URL=https://` + c.GetSysCfg().ServerURL + `
Environment=MIRAGE_NAVI_ID=` + derpid + `
	
[Install]
WantedBy=multi-user.target`

	// 将文本写入文件
	err = os.WriteFile("download/"+derpid+".service.tmp", []byte(serviceScript), 0644)
	if err != nil {
		c.doAPIResponse(w, "创建临时服务文件失败:"+err.Error(), nil)
		return
	}
	err = sshSendFile(client, "download/"+derpid+".service.tmp", "/etc/systemd/system/MirageNavi.service")
	if err != nil {
		c.doAPIResponse(w, "传送服务文件到目标服务器失败:"+err.Error(), nil)
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
		c.doAPIResponse(w, "创建会话失败:"+err.Error(), nil)
		return
	}
	defer systemdReloadSession.Close()

	_, err = systemdReloadSession.Output("systemctl daemon-reload")
	if err != nil {
		c.doAPIResponse(w, "执行命令失败:"+err.Error(), nil)
		return
	}

	// 启动服务
	systemdEnableSession, err := client.NewSession()
	if err != nil {
		c.doAPIResponse(w, "创建会话失败:"+err.Error(), nil)
		return
	}
	defer systemdEnableSession.Close()

	_, err = systemdEnableSession.Output("systemctl enable --now MirageNavi")
	if err != nil {
		c.doAPIResponse(w, "执行命令失败:"+err.Error(), nil)
		return
	}

	c.CAPIQueryDERP(w, r)
}

func genSSHKeypair() (priKey, pubKey string, err error) {
	var privateKey *ecdsa.PrivateKey
	var publicKey ssh.PublicKey

	if privateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader); err != nil {
		return
	}
	if publicKey, err = ssh.NewPublicKey(privateKey.Public()); err != nil {
		return
	}

	pubKey = string(ssh.MarshalAuthorizedKey(publicKey))
	var priKeyData []byte
	if priKeyData, err = x509.MarshalECPrivateKey(privateKey); err != nil {
		return
	}
	priKey = string(pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: priKeyData,
	}))

	return
}

func sshSendFile(sshClient *ssh.Client, localFilePath string, remoteFilePath string) error {
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	remoteFile, err := sftpClient.Create(remoteFilePath)
	if err != nil {
		return err
	}
	defer remoteFile.Close()

	localFile, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer localFile.Close()

	n, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return err
	}

	localFileInfo, err := os.Stat(localFilePath)
	if err != nil {
		return err
	}
	if n != localFileInfo.Size() {
		return errors.New("文件大小不一致，传输失败")
	}
	return nil
}

func (c *Cockpit) CAPIDelNaviNode(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	naviID, ok := vars["id"]
	if !ok {
		c.doAPIResponse(w, "未指定司南ID", nil)
		return
	}

	naviNode := c.GetNaviNode(naviID)
	if naviNode == nil {
		c.doAPIResponse(w, "未找到该司南档案", nil)
		return
	}
	// TODO: 是否有必要做远程关闭？

	if err := c.db.Delete(&naviNode).Error; err != nil {
		c.doAPIResponse(w, "数据库删除司南节点失败:"+err.Error(), nil)
		return
	}
	delete(c.App.DERPNCs, naviID)
	delete(c.App.DERPseqnum, naviID)
	//	c.App.LoadDERPMapFromURL(c.App.cfg.DERPURL)

	c.CtrlChn <- CtrlMsg{
		Msg: "set-last-update",
	}
	c.CAPIQueryDERP(w, r)
}
