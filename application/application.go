package application

import (
	amqp "github.com/kaellybot/kaelly-amqp"
	"github.com/kaellybot/kaelly-notifier/models/constants"
	"github.com/kaellybot/kaelly-notifier/repositories/almanaxes"
	"github.com/kaellybot/kaelly-notifier/repositories/feeds"
	"github.com/kaellybot/kaelly-notifier/repositories/twitch"
	"github.com/kaellybot/kaelly-notifier/repositories/youtube"
	"github.com/kaellybot/kaelly-notifier/services/discord"
	"github.com/kaellybot/kaelly-notifier/services/notifiers"
	"github.com/kaellybot/kaelly-notifier/utils/databases"
	"github.com/kaellybot/kaelly-notifier/utils/insights"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func New() (*Impl, error) {
	// misc
	broker := amqp.New(constants.RabbitMQClientID, viper.GetString(constants.RabbitMQAddress),
		amqp.WithBindings(notifiers.GetBinding()))
	db := databases.New()
	probes := insights.NewProbes(broker.IsConnected, db.IsConnected)
	prom := insights.NewPrometheusMetrics()

	// Repositories
	almanaxRepo := almanaxes.New(db)
	feedRepo := feeds.New(db)
	twitchRepo := twitch.New(db)
	youtubeRepo := youtube.New(db)

	// services
	discordService, errDisc := discord.New(viper.GetString(constants.DiscordToken))
	if errDisc != nil {
		log.Fatal().Err(errDisc).Msgf("Discord connection failed, shutting down.")
	}

	notifierService := notifiers.New(broker, discordService, almanaxRepo,
		feedRepo, twitchRepo, youtubeRepo)

	return &Impl{
		broker:          broker,
		db:              db,
		probes:          probes,
		prom:            prom,
		notifierService: notifierService,
	}, nil
}

func (app *Impl) Run() error {
	app.probes.ListenAndServe()
	app.prom.ListenAndServe()

	if err := app.db.Run(); err != nil {
		return err
	}

	if err := app.broker.Run(); err != nil {
		return err
	}

	app.notifierService.Consume()
	return nil
}

func (app *Impl) Shutdown() {
	app.broker.Shutdown()
	app.db.Shutdown()
	app.prom.Shutdown()
	app.probes.Shutdown()
	log.Info().Msgf("Application is no longer running")
}
