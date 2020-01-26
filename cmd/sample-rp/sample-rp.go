package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nerocrux/sample-rp/api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

var (
	version string
)

var (
	port            = flag.Uint("auth-port", 9001, "auth api port")
	verbose         = flag.Bool("verbose", false, "enable verbose log")
	debug           = flag.Bool("debug", false, "enable debug log")
	templateDir     = flag.String("template-dir", "/etc/rp/static/template", "set template directory")
	op              = flag.String("op", "", "set op")
	issuer          = flag.String("issuer", "", "set issuer")
	clientID        = flag.String("client-id", "", "set client ID")
	clientSecret    = flag.String("client-secret", "", "set client secret")
	redirectURI     = flag.String("redirect-uri", "", "set redirect uri")
	fixedSigningKey = flag.String("fixed-signing-key", "", "key for signing ID Token, will be deprecated")
)

var zapLogger *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	zapLogger, err = cfg.Build()
	if err != nil {
		log.Fatalf("failed to create zap logger: %v", err)
	}
}

func main() {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Parse()

	if *debug {
		zapLogger.Core().Enabled(zapcore.DebugLevel)
	} else if *verbose {
		zapLogger.Core().Enabled(zapcore.InfoLevel)
	}

	// root context notifies server shutdown by SIGINT or SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	sighupFn := func() {}
	sigintFn := func() {
		zapLogger.Info("Start shutdown gracefully ...")
		cancel()
	}

	go signalHandler(sigc, sighupFn, sigintFn)

	startServer(ctx)
}

func signalHandler(sigc <-chan os.Signal, sighupFn, sigintFn func()) {
	for sig := range sigc {
		switch sig {
		case syscall.SIGHUP:
			sighupFn()
		case syscall.SIGINT, syscall.SIGTERM:
			sigintFn()
		}
	}
}

func startServer(pctx context.Context) {
	eg, ctx := errgroup.WithContext(pctx)

	eg.Go(func() error {
		conf := api.ServerConfig{
			Port:            int(*port),
			ZapLogger:       zapLogger,
			TemplateDir:     *templateDir,
			OP:              *op,
			Issuer:          *issuer,
			ClientID:        *clientID,
			ClientSecret:    *clientSecret,
			RedirectURI:     *redirectURI,
			FixedSigningKey: *fixedSigningKey,
		}

		if err := api.RunServer(ctx, conf); err != nil {
			log.Fatalf("RunServer failed: %v", err)
		}
		return nil
	})

	zapLogger.Info("servers started")

	if err := eg.Wait(); err != nil {
		log.Fatalf("eg.Wait failed: %v", err)
	}
}
