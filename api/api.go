package api

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/nerocrux/sample-rp/api/middleware"
	"github.com/nerocrux/sample-rp/auth"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type ServerConfig struct {
	Port         int
	TemplateDir  string
	ZapLogger    *zap.Logger
	OP           string
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	CertsURL     string
}

func RunServer(ctx context.Context, conf ServerConfig) error {
	server := auth.NewServer(
		conf.ZapLogger,
		conf.TemplateDir,
		conf.OP,
		conf.Issuer,
		conf.ClientID,
		conf.ClientSecret,
		conf.RedirectURI,
		conf.CertsURL,
	)

	mux := http.NewServeMux()
	mux.Handle("/callback", middleware.HTTPLogger(server.CallbackEndpoint, conf.ZapLogger))
	mux.Handle("/", middleware.HTTPLogger(server.IndexEndpoint, conf.ZapLogger))
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	srv := http.Server{
		Handler: mux,
	}
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := srv.Serve(lis); err != nil {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}
		if err := lis.Close(); err != nil {
			return err
		}
		return nil
	})

	return eg.Wait()
}
