package headscale

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"

	"github.com/rs/zerolog/log"
)

//go:embed admin/console.html
var adminHTML string

func (h *Headscale) ConsolePanel(
	writer http.ResponseWriter,
	req *http.Request,
) {
	adminT := template.Must(template.New("admin").Parse(adminHTML))
	var payload bytes.Buffer
	if err := adminT.Execute(&payload, nil); err != nil {
		log.Error().
			Str("handler", "adminHTML").
			Err(err).
			Msg("Could not render admin HTML")

		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusInternalServerError)
		_, err := writer.Write([]byte("Could not render admin index template"))
		if err != nil {
			log.Error().
				Caller().
				Err(err).
				Msg("Failed to write response")
		}

		return
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
}

//go:embed admin/result.html
var resultHTML string
var resultTemplate = template.Must(template.New("result").Parse(resultHTML))

type resultTemplateConfig struct {
	ERROR   bool
	Msg     string
	Next    string
	NextMsg string
}

// 用来为控制台网页生成结果页
func renderResult(
	writer http.ResponseWriter,
	isErr bool, Msg string, Next string, NextMsg string,
) error {
	var payload bytes.Buffer
	if err := resultTemplate.Execute(&payload, resultTemplateConfig{
		ERROR:   isErr,
		Msg:     Msg,
		Next:    Next,
		NextMsg: NextMsg,
	}); err != nil {
		log.Error().
			Str("func", "renderResult").
			Str("type", "Result").
			Err(err).
			Msg("Could not generate result page")
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		writer.WriteHeader(http.StatusInternalServerError)
		_, werr := writer.Write([]byte("服务器把结果页搞丢了~_~"))
		if werr != nil {
			log.Error().
				Caller().
				Err(werr).
				Msg("Failed to write response")
		}
		return err
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write(payload.Bytes())
	if err != nil {
		log.Error().
			Caller().
			Err(err).
			Msg("Failed to write response")
	}
	return nil
}
