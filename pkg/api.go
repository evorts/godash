package pkg

import "github.com/alexedwards/scs/v2"

var (
	api *API
)

type (
	API struct {
		log     LoggingManager
		config  ConfigManager
		session *scs.SessionManager
		crypt   *Crypt
		hash   *Crypt
	}
	ErrorResponse struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Reasons map[string]string `json:"reasons"`
		Details []interface{}     `json:"details,omitempty"`
	}
	Response struct {
		RequestID string                 `json:"request_id"`
		Status    int                    `json:"status"`
		Content   map[string]interface{} `json:"content,omitempty"`
		Error     *ErrorResponse         `json:"error,omitempty"`
	}
)

func NewAPI(config ConfigManager, session *scs.SessionManager, logging LoggingManager, crypt *Crypt, hash *Crypt) *API {
	return &API{
		config:  config,
		session: session,
		log:     logging,
		crypt:   crypt,
		hash: hash,
	}
}

func (a *API) Logger() LoggingManager {
	return a.log
}

func (a *API) Config() ConfigManager {
	return a.config
}

func (a *API) Session() *scs.SessionManager {
	return a.session
}

func (a *API) Crypt() *Crypt {
	return a.crypt
}

func (a *API) Hash() *Crypt {
	return a.hash
}
