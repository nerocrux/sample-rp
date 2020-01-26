package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"text/template"

	"github.com/coreos/go-oidc"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	STATUS_OK     = "ok"
	STATUS_FAILED = "failed"
)

type Templates struct {
	index    *template.Template
	callback *template.Template
}

type Server struct {
	zapLogger        *zap.Logger
	templates        Templates
	oauthConfig      oauth2.Config
	issuer           string
	userinfoEndpoint string
	fixedSigningKey  string
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewServer(l *zap.Logger, templateDir, op, issuer, clientID, clientSecret, redirectURI, fixedSigningKey string) *Server {
	server := &Server{}
	server.zapLogger = l
	server.templates = Templates{
		index:    template.Must(template.New("layout.html.tpl").ParseFiles(path.Join(templateDir, "layout.html.tpl"), path.Join(templateDir, "index.html.tpl"))),
		callback: template.Must(template.New("layout.html.tpl").ParseFiles(path.Join(templateDir, "layout.html.tpl"), path.Join(templateDir, "callback.html.tpl"))),
	}

	server.fixedSigningKey = fixedSigningKey
	server.issuer = issuer
	server.oauthConfig = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  op + "/authorize",
			TokenURL: op + "/token",
		},
		RedirectURL: redirectURI,
		Scopes: []string{
			//"user:profile:read",
			//"user:email:read",
			oidc.ScopeOpenID,
			oidc.ScopeOfflineAccess,
		},
	}
	server.userinfoEndpoint = op + "userinfo"

	return server
}

func errorHandler(rw http.ResponseWriter, status int, message string) {
	rw.WriteHeader(status)
	switch status {
	case http.StatusNotFound:
		fmt.Fprint(rw, "404 Not Found.")
	case http.StatusBadRequest:
		fmt.Fprint(rw, "Bad Request: "+message)
	default:
		fmt.Fprint(rw, "Internal Server Error")
	}
}

func setCookie(rw http.ResponseWriter, name string, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		HttpOnly: true,
	}
	http.SetCookie(rw, cookie)
}

func getCookie(req *http.Request, key string) (string, error) {
	cookie, err := req.Cookie(key)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func writeErrorResponse(rw http.ResponseWriter, status int, message string) {
	res, _ := json.Marshal(&Response{
		Status:  STATUS_FAILED,
		Message: message,
	})
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	rw.Write(res)
}

func writeSuccessResponse(rw http.ResponseWriter) {
	res, _ := json.Marshal(&Response{
		Status: STATUS_OK,
	})
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(res)
}
