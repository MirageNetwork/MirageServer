package controller

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/rs/zerolog/log"
)

func (h *Mirage) ListIdps(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.doAPIResponse(w, "", h.cfg.IdpList)
}

// API调用的统一响应发报
// @msg 响应状态：成功时data不为nil则忽略，自动设置为success，否则拼接error-{msg}
// @data 响应数据：key值为data的json对象
func (h *Mirage) doAPIResponse(writer http.ResponseWriter, msg string, data interface{}) {
	res := APIResponse{}
	if msg == "" {
		res.Status = "success"
		res.Data = data
	} else {
		res.Status = "error-" + msg
	}
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	err := json.NewEncoder(writer).Encode(&res)
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

//go:embed templates/BadCode.html
var errTemplate string

// cgao6: 用这个向前端返回错误页面
func (h *Mirage) ErrMessage(
	w http.ResponseWriter,
	r *http.Request,
	code int,
	msg string,
) {
	errT := template.Must(template.New("err").Parse(errTemplate))

	config := map[string]interface{}{
		"ErrCode": code,
		"ErrMsg":  msg,
	}

	var payload bytes.Buffer
	if err := errT.Execute(&payload, config); err != nil {
		log.Error().
			Str("handler", "ErrMessage").
			Err(err).
			Msg("Could not render Error template")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Could not render Error template"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	_, err := w.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

//go:embed templates/memberNoConsole.html
var noConsoleTemplate string

// cgao6: 用这个向前端返回普通用户无权登录控制台
func (h *Mirage) renderNoConsole(
	w http.ResponseWriter,
	r *http.Request,
	userName string,
	orgName string,
) {
	noConsoleT := template.Must(template.New("noConsole").Parse(noConsoleTemplate))

	config := map[string]interface{}{
		"UserName": userName,
		"OrgName":  orgName,
	}

	var payload bytes.Buffer
	if err := noConsoleT.Execute(&payload, config); err != nil {
		log.Error().
			Str("handler", "ErrMessage").
			Err(err).
			Msg("Could not render noConsole template")

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Could not render noConsole template"))
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
