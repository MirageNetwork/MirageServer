package main

import (
	"fmt"
	"os"
	"time"

	"github.com/MirageNetwork/Mirage"
	"github.com/efekarakus/termcolor"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Version = "0.1.0"

const (
	MirageDateTimeFormat   = "2006-01-02 15:04:05"
	SocketWritePermissions = 0o666
)

func getMirageApp() (*Mirage.Mirage, error) {
	cfg, err := Mirage.GetMirageConfig()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to load configuration while creating mirage instance: %w",
			err,
		)
	}

	app, err := Mirage.NewMirage(cfg)
	if err != nil {
		return nil, err
	}

	// We are doing this here, as in the future could be cool to have it also hot-reload

	if cfg.ACL.PolicyPath != "" {
		aclPath := Mirage.AbsolutePathFromConfigPath(cfg.ACL.PolicyPath)
		err = app.LoadACLPolicy(aclPath)
		if err != nil {
			log.Fatal().
				Str("path", aclPath).
				Err(err).
				Msg("Could not load the ACL policy")
		}
	}

	return app, nil
}

func main() {
	var colors bool
	switch l := termcolor.SupportLevel(os.Stderr); l {
	case termcolor.Level16M:
		colors = true
	case termcolor.Level256:
		colors = true
	case termcolor.LevelBasic:
		colors = true
	case termcolor.LevelNone:
		colors = false
	default:
		// no color, return text as is.
		colors = false
	}

	// Adhere to no-color.org manifesto of allowing users to
	// turn off color in cli/services
	if _, noColorIsSet := os.LookupEnv("NO_COLOR"); noColorIsSet {
		colors = false
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    !colors,
	})

	err := Mirage.LoadConfig("", false)
	if err != nil {
		log.Fatal().Caller().Err(err).Msgf("Error loading config")
	}

	app, err := getMirageApp()
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("Error initializing")
	}

	err = app.Serve()
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("Error starting server")
	}
}