package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type indexScreenParameters struct {
}

func (s *Server) IndexEndpoint(rw http.ResponseWriter, req *http.Request) error {
	var err error

	if req.Method == http.MethodPost {
		if err = req.ParseForm(); err != nil {
			return errors.Wrap(err, "cannot parse post form")
		}
		state, err := generateRandomString(10)
		if err != nil {
			return errors.Wrap(err, "failed generate state")
		}
		nonce, err := generateRandomString(10)
		if err != nil {
			return errors.Wrap(err, "failed generate nonce")
		}
		url := s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oidc.Nonce(nonce))

		wh := rw.Header()
		wh.Set("Location", url)
		rw.WriteHeader(http.StatusFound)
	}

	if req.Method == http.MethodGet {
		if err := s.templates.index.Execute(rw, indexScreenParameters{}); err != nil {
			return errors.Wrap(err, "failed render template")
		}
	}
	return nil
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, 128)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	var b64 = base64.URLEncoding.WithPadding(base64.NoPadding)
	raw := b64.EncodeToString(b)
	for _, s := range []string{"-", "_"} {
		raw = strings.Replace(raw, s, "", -1)
	}
	if len(raw) < length {
		return "", errors.New("randomly generated string is too short")
	}
	return raw[0:length], nil
}
