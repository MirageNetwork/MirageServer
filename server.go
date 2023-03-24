package main

import (
	"os"
	"time"

	"MirageNetwork/MirageServer/controller"

	"github.com/efekarakus/termcolor"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	MirageDateTimeFormat   = "2006-01-02 15:04:05"
	SocketWritePermissions = 0o666
)

func main() {
	sysAddr, set := os.LookupEnv("MIRAGE_SYS_ADDR")
	if !set {
		sysAddr = ":8081"
	}

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

	datapool := controller.DataPool{}
	err := datapool.OpenDB()
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("Error opening database")
	}

	ctrlChn := make(chan controller.CtrlMsg)
	msgChn := make(chan controller.CtrlMsg)
	err = datapool.InitCockpitDB()
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("Error initializing cockpit database")
	}
	cockpit, err := controller.NewCockpit(sysAddr, ctrlChn, msgChn, datapool.DB())
	if err != nil {
		log.Fatal().Caller().Err(err).Msg("Error initializing cockpit")
	}

	go cockpit.Run()

	log.Info().Caller().Msg("Cockpit is ready on " + sysAddr + "")

	/*
		err = controller.LoadConfig()
		if err != nil {
			log.Fatal().Caller().Err(err).Msgf("Error loading config")
		}

		cfg, err := controller.GetMirageConfig(srvAddr, serverURL)
		if err != nil {
			log.Fatal().Caller().Err(err).Msgf("Error loading config")
		}
	*/
	for {
		select {
		case cockpitMsg := <-ctrlChn:
			log.Trace().Caller().Msgf("Cockpit message received: %s", cockpitMsg.Msg)
			switch cockpitMsg.Msg {
			case "start":
				err = datapool.InitMirageDB()
				if err != nil {
					log.Error().Caller().Err(err).Msgf("Error initializing Mirage DB")
					msgChn <- controller.CtrlMsg{
						Msg: "error",
						Err: err,
					}
					break
				}

				app, err := controller.NewMirage(cockpitMsg.SysCfg, datapool.DB())
				if err != nil {
					log.Error().Caller().Err(err).Msg("Error initializing Mirage")
					msgChn <- controller.CtrlMsg{
						Msg: "error",
						Err: err,
					}
					break
				}

				err = app.Serve(ctrlChn)
				if err != nil {
					log.Error().Caller().Err(err).Msg("Error starting server")
					msgChn <- controller.CtrlMsg{
						Msg: "error",
						Err: err,
					}
					break
				}
			}
		}
	}
}
