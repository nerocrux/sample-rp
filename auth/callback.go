package auth

import (
	"context"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    string `json:"expires_in"`
}

type claims struct {
	Subject         string   `json:"sub"`
	Audience        []string `json:"aud"`
	ExpiredAt       int      `json:"exp"`
	IssuedAt        int      `json:"iat"`
	Issuer          string   `json:"iss"`
	Nonce           string   `json:"nonce"`
	AccessTokenHash string   `json:"at_hash"`
}

type callbackScreenParameters struct {
	Subject         string
	Audience        []string
	KeyID           string
	Issuer          string
	Nonce           string
	AccessTokenHash string
	Err             string
}

func (s *Server) CallbackEndpoint(rw http.ResponseWriter, req *http.Request) error {
	// 1. Obtain AuthCode
	code := getCodeFromCallback(req)

	// TODO Validate state

	// 2. Exchange token with AuthCode
	ctx := context.Background()
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		err := errors.Wrap(err, "exchange error")
		s.renderErrorScreen(rw, err)
		return err
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		err := errors.Wrap(err, "cannot extract ID Token from response")
		s.renderErrorScreen(rw, err)
		return err
	}

	// 3. Verify & parse ID Token
	keySet := oidc.NewRemoteKeySet(context.TODO(), s.certsURL)
	verifier := oidc.NewVerifier(s.issuer, keySet, &oidc.Config{
		ClientID:             s.oauthConfig.ClientID,
		SupportedSigningAlgs: []string{"ES256"},
	})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		err := errors.Wrap(err, "failed verify ID Token")
		s.renderErrorScreen(rw, err)
		return err
	}
	claims := claims{}
	if err := idToken.Claims(&claims); err != nil {
		err := errors.Wrap(err, "failed parse ID Token")
		s.renderErrorScreen(rw, err)
		return err
	}

	/*
		// 4. (TODO) Get Userinfo from OP /userinfo endpoint
		client := s.oauthConfig.Client(oauth2.NoContext, token)
		resp, err := client.Get(s.userinfoEndpoint)
		if err != nil {
			return errors.Wrap(err, "client get error")
		}

		defer resp.Body.Close()
		byteArray, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(byteArray))
	*/

	// 5. Render template to show token information
	s.renderSuccessScreen(rw, claims)
	return nil
}

func (s *Server) renderErrorScreen(rw http.ResponseWriter, err error) {
	if err := s.templates.callback.Execute(rw, callbackScreenParameters{
		Subject:  "",
		Audience: nil,
		Err:      err.Error(),
	}); err != nil {
		log.Fatalf("failed to render error screen")
	}
}

func (s *Server) renderSuccessScreen(rw http.ResponseWriter, claims claims) {
	if err := s.templates.callback.Execute(rw, callbackScreenParameters{
		Nonce:           claims.Nonce,
		AccessTokenHash: claims.AccessTokenHash,
		Subject:         claims.Subject,
		Issuer:          claims.Issuer,
		Audience:        claims.Audience,
		Err:             "",
	}); err != nil {
		log.Fatalf("failed to render callback screen")
	}
}

func getCodeFromCallback(req *http.Request) string {
	return req.URL.Query().Get("code")
}
