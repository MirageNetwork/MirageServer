package controller

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var esClient *elasticsearch.Client

func InitESLogger(cfg *Config) error {
	// 初始化 Elasticsearch 客户端
	escfg := elasticsearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Addresses: []string{cfg.ESURL},
		APIKey:    cfg.ESKey,
	}
	var err error
	esClient, err = elasticsearch.NewClient(escfg)
	if err != nil {
		return err
	}

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	logger := zerolog.New(zerolog.MultiLevelWriter(os.Stdout, esClientWriter{})).With().Timestamp().Logger()

	// 将自定义的 io.Writer 注册到 Logger 中
	logger = logger.Level(zerolog.DebugLevel)
	log.Logger = logger
	return nil
}

type esClientWriter struct{}

func (w esClientWriter) Write(p []byte) (n int, err error) {
	req := esapi.IndexRequest{
		Index: "mirage-server",
		Body:  strings.NewReader(string(p)),
	}
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	return len(p), nil
}

func (w esClientWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	switch l {
	case zerolog.InfoLevel:
		req := esapi.IndexRequest{
			Index: "mirage-server",
			Body:  strings.NewReader(string(p)),
		}
		res, err := req.Do(context.Background(), esClient)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()
	}
	return len(p), nil
}
