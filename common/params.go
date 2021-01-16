package common

import (
	"flag"
	"strings"
)

type Params struct {
	categories            []string
	httpPort              int
	secret                string
	alwaysStartHttpServer bool
	logLevel              string
	rootPath              string
}

func NewEmptyParams() *Params {
	return &Params{
		categories:            []string{},
		httpPort:              0,
		secret:                "",
		alwaysStartHttpServer: false,
		logLevel:              "",
		rootPath:              "",
	}
}

func ParseParams() *Params {
	categories := flag.String("categories", "", "Comma separated categories. Each category in format <name>:<shortcut> e.g. Good:G")
	categoryArr := strings.Split(*categories, ",")
	httpPort := flag.Int("httpPort", 8080, "HTTP Server port for Chrome Cast")
	secret := flag.String("secret", "", "Override default random secret for casting")
	alwaysStartHttpServer := flag.Bool("alwaysStartHttpServer", false, "Always start HTTP server. Not only when casting.")
	logLevel := flag.String("logLevel", "INFO", "Log level: ERROR, WARN, INFO, DEBUG, Trace")

	flag.Parse()
	rootPath := flag.Arg(0)

	return &Params{
		categories:            categoryArr,
		httpPort:              *httpPort,
		secret:                *secret,
		alwaysStartHttpServer: *alwaysStartHttpServer,
		logLevel:              *logLevel,
		rootPath:              rootPath,
	}
}

func (s *Params) Categories() []string {
	return s.categories
}

func (s *Params) HttpPort() int {
	return s.httpPort
}

func (s *Params) Secret() string {
	return s.secret
}

func (s *Params) AlwaysStartHttpServer() bool {
	return s.alwaysStartHttpServer
}

func (s *Params) LogLevel() string {
	return s.logLevel
}

func (s *Params) RootPath() string {
	return s.rootPath
}
