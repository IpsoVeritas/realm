package rest

import (
	httphandler "github.com/Brickchain/go-httphandler.v2"
	"github.com/pkg/errors"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/services"

	"net/http"
)

type ConfigController struct {
	contextProvider *services.RealmsServiceProvider
}

func NewConfigController(contextProvider *services.RealmsServiceProvider) *ConfigController {
	return &ConfigController{
		contextProvider: contextProvider,
	}
}

// Config returns the realm config
func (c *ConfigController) Config(req httphandler.Request) httphandler.Response {

	realmID := req.Params().ByName("realmID")
	if realmID == "" {
		return httphandler.NewErrorResponse(http.StatusBadRequest, errors.New("No realm specified"))
	}

	context := c.contextProvider.Get(realmID)

	realmData, err := context.Realm()
	if err != nil {
		return httphandler.NewErrorResponse(http.StatusInternalServerError, errors.Wrap(err, "failed to get realm"))
	}

	cfg := realm.RealmConfig{
		ServicesFeed: "https://firebasestorage.googleapis.com/v0/b/integrity-autobinder-staging.appspot.com/o/services.json?alt=media",
		AdminRoles:   realmData.AdminRoles,
	}

	return httphandler.NewJsonResponse(http.StatusOK, cfg)
}
