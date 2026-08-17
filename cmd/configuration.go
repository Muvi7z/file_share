package cmd

import (
	"fmt"
	"net"
	"time"
)

const (
	envPostgresDB                 = "POSTGRES_DB"
	envPostgresHost               = "POSTGRES_HOST"
	envPostgresPort               = "POSTGRES_PORT"
	envPostgresUser               = "POSTGRES_USER"
	envPostgresPassword           = "POSTGRES_PASSWORD"
	envPostgresSslMode            = "POSTGRES_SSL_MODE"
	envPostgresMaxIdleConnections = "POSTGRES_MAX_IDLE_CONNECTIONS"
	envPostgresMaxOpenConnections = "POSTGRES_MAX_OPEN_CONNECTIONS"

	envExternalRedisPort      = "EXTERNAL_REDIS_PORT"
	envExternalRedisHost      = "EXTERNAL_REDIS_HOST"
	envRedisConnectionTimeout = "REDIS_CONNECTION_TIMEOUT"

	envServerHost = "SERVER_HOST"
	envServerPort = "SERVER_PORT"
	envPosterDir  = "POSTER_DIR"

	envJWTSecret = "JWT_SECRET"
)

func newFromEnv() (*configuration, error) {
	c := &configuration{}

	pc := &postgresConfiguration{}

	rc := &redisConfiguration{}

	var err error
	pc.user, err = getStringFromEnv(envPostgresUser)
	if err != nil {
		return nil, fmt.Errorf("error getting postgres user: %s", err)
	}

	pc.host, err = getStringFromEnv(envPostgresHost)
	if err != nil {
		return nil, fmt.Errorf("error getting postgres host: %s", err)
	}

	pc.port = getStringFromEnvOrDefault(envPostgresPort, "5432")

	rc.port = getStringFromEnvOrDefault(envExternalRedisPort, "6379")
	rc.host = getStringFromEnvOrDefault(envExternalRedisHost, "localhost")
	rc.connectionTimeout, err = getDurationFromEnvOrDefault(envRedisConnectionTimeout, time.Second*10)
	if err != nil {
		return nil, fmt.Errorf("error getting redis connection timeout: %s", err)
	}

	pc.password, err = getStringFromEnv(envPostgresPassword)
	if err != nil {
		return nil, fmt.Errorf("error getting postgres password: %s", err)
	}
	c.posterDir, err = getStringFromEnv(envPosterDir)
	if err != nil {
		return nil, fmt.Errorf("error getting poster dir: %s", err)
	}

	pc.sslmode = getStringFromEnvOrDefault(envPostgresSslMode, "disable")

	pc.db, err = getStringFromEnv(envPostgresDB)
	if err != nil {
		return nil, fmt.Errorf("error getting postgres db: %s", err)
	}

	pc.maxIdleConnections, err = getIntValueFromEnv(envPostgresMaxIdleConnections, 10)
	if err != nil {
		return nil, fmt.Errorf("error getting postgres max idle connections: %s", err)
	}

	pc.maxOpenConnections, err = getIntValueFromEnv(envPostgresMaxOpenConnections, 10)
	if err != nil {
		return nil, fmt.Errorf("error getting postgres max open connections: %s", err)
	}

	c.postgresConfiguration = pc

	sc := &serverConfiguration{}

	sc.host = getStringFromEnvOrDefault(envServerHost, "localhost")

	sc.port, err = getIntValueFromEnv(envServerPort, 5432)
	if err != nil {
		return nil, fmt.Errorf("error getting postgres port: %s", err)
	}

	c.serverConfiguration = sc

	c.redisConfiguration = rc

	c.jwtSecret = getStringFromEnvOrDefault(envJWTSecret, "default-secret-key-change-me")

	return c, nil
}

type configuration struct {
	postgresConfiguration *postgresConfiguration
	serverConfiguration   *serverConfiguration
	redisConfiguration    *redisConfiguration
	posterDir             string
	jwtSecret             string
}

type redisConfiguration struct {
	host              string
	port              string
	connectionTimeout time.Duration
}

type postgresConfiguration struct {
	db                 string
	host               string
	port               string
	user               string
	password           string
	sslmode            string
	maxOpenConnections int64
	maxIdleConnections int64
}

type serverConfiguration struct {
	host string
	port int64
}

func (c *configuration) GetPostgresConfiguration() *postgresConfiguration {
	return c.postgresConfiguration
}
func (c *configuration) GetPosterDir() string {
	return c.posterDir
}

func (pc *postgresConfiguration) GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", pc.user, pc.password, pc.host, pc.port, pc.db, pc.sslmode)
}

func (rc *redisConfiguration) GetAddress() string {
	return net.JoinHostPort(rc.host, rc.port)
}

func (rc *redisConfiguration) GetConnectionTimeout() time.Duration {
	return rc.connectionTimeout
}

func (pc *postgresConfiguration) GetMaxIdleConns() int {
	return int(pc.maxIdleConnections)
}

func (pc *postgresConfiguration) GetMaxOpenConns() int {
	return int(pc.maxOpenConnections)
}

func (c *configuration) GetServerConfiguration() *serverConfiguration {
	return c.serverConfiguration
}

func (sc *serverConfiguration) GetAddress() string {
	return fmt.Sprintf("%s:%d", sc.host, sc.port)
}

func (c *configuration) GetJWTSecret() string {
	return c.jwtSecret
}
