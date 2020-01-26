package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    string `json:"expires_in"`
}

type claims struct {
	Subject   string   `json:"sub"`
	Audience  []string `json:"aud"`
	AuthTime  int      `json:"auth_time"`
	ExpiredAt int      `json:"exp"`
	IssuedAt  int      `json:"iat"`
	Issuer    string   `json:"iss"`
	JTI       string   `json:"jti"`
	Nonce     string   `json:"nonce"`
}

type callbackScreenParameters struct {
	Subject  string
	Audience []string
	Err      string
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
	keySet, err := getStaticKeySet(s.fixedSigningKey)
	if err != nil {
		err := errors.Wrap(err, "cannot fetch public key from file")
		s.renderErrorScreen(rw, err)
		return err
	}
	verifier := oidc.NewVerifier(s.issuer, keySet, &oidc.Config{ClientID: s.oauthConfig.ClientID})
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
		Subject:  claims.Subject,
		Audience: claims.Audience,
		Err:      "",
	}); err != nil {
		log.Fatalf("failed to render callback screen")
	}
}

func getCodeFromCallback(req *http.Request) string {
	return req.URL.Query().Get("code")
}

type fixedKeySet struct {
	key *rsa.PublicKey
}

func getStaticKeySet(fixedSigningKey string) (oidc.KeySet, error) {
	content, err := ioutil.ReadFile(fixedSigningKey)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(content)
	if block == nil {
		return nil, errors.New("invalid public key data")
	}
	if block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid public key type : %s", block.Type)
	}
	keyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key, ok := keyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}
	return &fixedKeySet{
		key: key,
	}, nil
}

func (ks *fixedKeySet) VerifySignature(ctx context.Context, jwt string) (p []byte, err error) {
	jws, err := jose.ParseSigned(jwt)
	if err != nil {
		return nil, fmt.Errorf("oidc: malformed jwt: %v", err)
	}
	payload, err := jws.Verify(ks.key)
	if err == nil {
		return payload, nil
	}
	return nil, errors.Wrap(err, "cannot decode signature")
}
