package controller

import (
	_ "embed"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

// 接受/cockpit/api/publish的Post请求，用于进行客户端发布
// 根据请求类型不同，可能是json报文发送来的版本号和URL，也可能是form表单发送来的版本号和文件流
func (c *Cockpit) CAPIPublishClient(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	osType, ok := vars["os"]
	if !ok {
		c.doAPIResponse(w, "未指定客户端类型", nil)
		return
	}

	sysCfg := c.GetSysCfg()
	if sysCfg == nil {
		c.doAPIResponse(w, "获取系统配置失败", nil)
		return
	}

	reqData := ClientVer{}

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		json.NewDecoder(r.Body).Decode(&reqData)
	} else {
		r.ParseMultipartForm(64 << 20)
		mForm := r.MultipartForm

		fileName := mForm.File["file"][0].Filename
		if strings.HasPrefix(osType, "navi") {
			fileName = "MirageNavi"
		}

		reqData.Version = mForm.Value["version"][0]
		file, _, err := r.FormFile("file")
		if err != nil {
			c.doAPIResponse(w, "文件解析失败:"+err.Error(), nil)
			return
		}
		defer file.Close()
		fileData, err := io.ReadAll(file)
		if err != nil {
			c.doAPIResponse(w, "文件读取失败:"+err.Error(), nil)
			return
		}

		_, err = os.Stat("download")
		if err != nil {
			err = os.Mkdir("download", os.ModePerm)
			if err != nil {
				c.doAPIResponse(w, "下载文件夹创建失败:"+err.Error(), nil)
				return
			}
		}

		switch osType {
		case "navi_x86_64":
			_, err = os.Stat("download/x86_64")
			if err != nil {
				err = os.Mkdir("download/x86_64", os.ModePerm)
				if err != nil {
					c.doAPIResponse(w, "下载文件夹创建失败:"+err.Error(), nil)
					return
				}
			}
			fileName = "x86_64/" + fileName
		case "navi_aarch64":
			_, err = os.Stat("download/aarch64")
			if err != nil {
				err = os.Mkdir("download/aarch64", os.ModePerm)
				if err != nil {
					c.doAPIResponse(w, "下载文件夹创建失败:"+err.Error(), nil)
					return
				}
			}
			fileName = "aarch64/" + fileName
		}

		newFile, err := os.Create("download/" + fileName)
		if err != nil {
			c.doAPIResponse(w, "文件创建失败:"+err.Error(), nil)
			return
		}
		defer newFile.Close()
		_, err = newFile.Write(fileData)
		if err != nil {
			c.doAPIResponse(w, "文件写入失败:"+err.Error(), nil)
			return
		}
		reqData.Url = "https://" + sysCfg.ServerURL + "/download/" + fileName
	}

	if reqData.Url == "" || reqData.Version == "" && osType != "linux" {
		c.doAPIResponse(w, "客户端发布请求处理失败", nil)
		return
	}

	switch osType {
	case "win":
		sysCfg.ClientVersion.Win = reqData
	case "ios_store":
		sysCfg.ClientVersion.IOSStore = reqData
	case "ios_test":
		sysCfg.ClientVersion.IOSTestFlight = reqData
	case "navi_x86_64":
		sysCfg.ClientVersion.NaviAMD64 = reqData.Version
	case "navi_aarch64":
		sysCfg.ClientVersion.NaviAARCH64 = reqData.Version
	case "linux":
		sysCfg.ClientVersion.Linux.Url = reqData.Url
		if reqData.Version != "" {
			sysCfg.ClientVersion.Linux.RepoCred = reqData.Version
			if reqData.Version == "clear" {
				sysCfg.ClientVersion.Linux.RepoCred = ""
			}
		}
	default:
		c.doAPIResponse(w, "未支持的客户端类型", nil)
		return
	}
	if err := c.db.Save(sysCfg).Error; err != nil {
		c.doAPIResponse(w, "更新客户端信息失败", nil)
		return
	}

	if osType == "linux" {
		go c.BuildLinuxClient()
	}

	if c.serviceState {
		newCfg, err := c.GetSysCfg().toSrvConfig()
		if err != nil {
			c.doAPIResponse(w, "更新系统配置失败", nil)
			return
		}
		c.CtrlChn <- CtrlMsg{
			Msg:    "update-config",
			SysCfg: newCfg,
		}
	}

	c.GetSettingGeneral(w, r)
}

type PublishInfoData struct {
	UploadURL     string            `json:"upload_url"`
	ClientVersion ClientVersionInfo `json:"client_version"`
}

func (c *Cockpit) GetPublishInfo(
	w http.ResponseWriter,
	r *http.Request,
) {
	sysCfg := c.GetSysCfg()
	if sysCfg == nil {
		c.doAPIResponse(w, "获取系统配置失败", nil)
		return
	}
	c.doAPIResponse(w, "", PublishInfoData{
		UploadURL:     "https://" + sysCfg.ServerURL + "/cockpit/api/publish",
		ClientVersion: sysCfg.ClientVersion,
	})
}
