package controller

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
	"tailscale.com/tailcfg"
)

func LoadDERPMapFromURL(addr string) (*tailcfg.DERPMap, error) {
	ctx, cancel := context.WithTimeout(context.Background(), HTTPReadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, addr, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout: HTTPReadTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var derpMap tailcfg.DERPMap
	err = json.Unmarshal(body, &derpMap)

	if len(derpMap.Regions) == 0 {
		log.Warn().
			Msg("DERP map is empty, not a single DERP map datasource was loaded correctly or contained a region")
	}

	return &derpMap, err
}
