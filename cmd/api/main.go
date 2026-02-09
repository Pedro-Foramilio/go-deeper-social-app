package main

import (
	"log"
	"time"

	"github.com/Pedro-Foramilio/social/internal/auth"
	"github.com/Pedro-Foramilio/social/internal/db"
	"github.com/Pedro-Foramilio/social/internal/env"
	"github.com/Pedro-Foramilio/social/internal/mailer"
	ratelimiter "github.com/Pedro-Foramilio/social/internal/rateLimiter"
	"github.com/Pedro-Foramilio/social/internal/store"
	"github.com/Pedro-Foramilio/social/internal/store/cache"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Printf("Error loading envs: %v\n", err)
	}

	dbConfig := dbConfig{
		addr:         env.GetString("DB_ADDR", "postgres://def:def@localhost/socialnetwork?sslmode=disable"),
		maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
		maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
		maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
	}

	redisCfg := redisConfig{
		addr:    env.GetString("REDIS_ADDR", "localhost:6379"),
		pw:      env.GetString("REDIS_PASSWORD", ""),
		db:      env.GetInt("REDIS_DB", 0),
		enabled: env.GetBool("REDIS_ENABLED", false),
	}

	cfg := config{
		addr: env.GetString("ADDR", ":8081"),
		db:   dbConfig,
		env:  env.GetString("ENV", "local"),
		mail: mailConfig{
			exp:       5 * time.Minute,
			apikey:    env.GetString("API_KEY", ""),
			fromEmail: env.GetString("FROM_EMAIL", "socialnetwork.com"),
			mailTrap: MailTrap{
				apikey: env.GetString("API_KEY_MAILTRAP", ""),
			},
		},
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
		auth: authConfig{
			basic: basicConfig{
				user: env.GetString("BASIC_AUTH_USER", ""),
				pass: env.GetString("BASIC_AUTH_PASS", ""),
			},
			token: tokenConfig{
				secret: env.GetString("JWT_SECRET", ""),
				exp:    time.Hour * 24 * 3,
				iss:    "socialnetwork",
			},
		},
		redisCfg: redisCfg,
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.GetInt("RATE_LIMITER_REQUESTS_COUNT", 20),
			TimeFrame:            time.Second * 5,
			Enabled:              env.GetBool("RATE_LIMITER_ENABLED", true),
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(
		dbConfig.addr,
		dbConfig.maxOpenConns,
		dbConfig.maxIdleConns,
		dbConfig.maxIdleTime,
	)

	var redis *redis.Client
	if redisCfg.enabled {
		logger.Info("redis enabled")
		redis = cache.NewRedisClient(
			redisCfg.addr,
			redisCfg.pw,
			redisCfg.db,
		)
	} else {
		logger.Info("Redis cache is disabled.")
	}

	if err != nil {
		logger.Panicf("Error connecting to the database: %v\n", err)
	}
	defer db.Close()
	logger.Info("Connected to the database successfully")

	store := store.NewStorage(db)

	// mailer := mailer.NewSendGrid(
	// 	cfg.mail.apikey,
	// 	cfg.mail.fromEmail,
	// )

	mailer, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.apikey, cfg.mail.fromEmail)

	jwtAuth := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)

	rateLimiter := ratelimiter.NewFixedWindowRateLimiter(cfg.rateLimiter.RequestsPerTimeFrame, cfg.rateLimiter.TimeFrame)

	app := &application{
		config:        cfg,
		store:         store,
		cacheStorage:  *cache.NewRedisStorage(redis),
		logger:        logger,
		mailer:        mailer,
		authenticator: jwtAuth,
		rateLimiter:   rateLimiter,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
