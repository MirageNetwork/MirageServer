package controller

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/rs/zerolog/log"
)

//go:embed console_html/downloads/index.html
var downloadsTemplate string

func (m *Mirage) getClientsInfo() *ClientVersionInfo {
	cfg := []SysConfig{}
	err := m.db.Find(&cfg).Error
	if err != nil || cfg == nil || len(cfg) == 0 {
		return nil
	}
	return &cfg[0].ClientVersion
}

type DownloadLinks struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

type DownloadInfo struct {
	MacOS   DownloadLinks `json:"macos"`
	IOS     DownloadLinks `json:"ios"`
	Windows DownloadLinks `json:"windows"`
	Android DownloadLinks `json:"android"`
	Linux   DownloadLinks `json:"linux"`
}

func (m *Mirage) sendDownloadsPage(
	w http.ResponseWriter,
	r *http.Request,
) {
	downloadsPageT := template.Must(template.New("connectDevice").Parse(downloadsTemplate))

	clientsInfo := m.getClientsInfo()

	config := map[string]interface{}{
		"DownloadDetails": DownloadInfo{
			MacOS: DownloadLinks{
				Primary:   clientsInfo.MacStore.Url,
				Secondary: clientsInfo.MacTestFlight.Url,
			},
			IOS: DownloadLinks{
				Primary:   clientsInfo.IOSStore.Url,
				Secondary: clientsInfo.IOSTestFlight.Url,
			},
			Windows: DownloadLinks{
				Primary: clientsInfo.Win.Url,
			},
			Linux: DownloadLinks{
				Primary:   clientsInfo.Linux.Version,
				Secondary: m.cfg.ServerURL,
			},
			Android: DownloadLinks{
				Primary: clientsInfo.Android.Url,
			},
		}}

	var payload bytes.Buffer
	if err := downloadsPageT.Execute(&payload, config); err != nil {
		log.Error().
			Str("handler", "downloadPageRender").
			Err(err).
			Msg("Could not render download page template")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Could not render download page template"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}
