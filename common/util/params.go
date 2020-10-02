package util

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

func (s *Params) GetCategories() []string {
	return s.categories
}

func (s *Params) GetHttpPort() int {
	return s.httpPort
}

func (s *Params) GetSecret() string {
	return s.secret
}

func (s *Params) GetAlwaysStartHttpServer() bool {
	return s.alwaysStartHttpServer
}

func (s *Params) GetLogLevel() string {
	return s.logLevel
}

func (s *Params) GetRootPath() string {
	return s.rootPath
}
