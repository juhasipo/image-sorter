package util

import (
	"flag"
	"strings"
)

type Params struct {
	Categories            []string
	HttpPort              int
	Secret                string
	AlwaysStartHttpServer bool
	LogLevel              string
	RootPath              string
}

func GetAppParams() *Params {
	categories := flag.String("categories", "", "Comma separated categories. Each category in format <name>:<shortcut> e.g. Good:G")
	categoryArr := strings.Split(*categories, ",")
	httpPort := flag.Int("httpPort", 8080, "HTTP Server port for Chrome Cast")
	secret := flag.String("secret", "", "Override default random secret for casting")
	alwaysStartHttpServer := flag.Bool("alwaysStartHttpServer", false, "Always start HTTP server. Not only when casting.")
	logLevel := flag.String("logLevel", "INFO", "Log level: ERROR, WARN, INFO, DEBUG, Trace")

	flag.Parse()
	rootPath := flag.Arg(0)

	return &Params{
		Categories:            categoryArr,
		HttpPort:              *httpPort,
		Secret:                *secret,
		AlwaysStartHttpServer: *alwaysStartHttpServer,
		LogLevel:              *logLevel,
		RootPath:              rootPath,
	}
}
