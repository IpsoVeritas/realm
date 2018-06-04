//go:generate go generate ../../pkg/providers/bindata/files.go

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Brickchain/go-crypto.v2"
	"github.com/Brickchain/go-document.v2"
	httphandler "github.com/Brickchain/go-httphandler.v2"
	logger "github.com/Brickchain/go-logger.v1"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tylerb/graceful"
	realm "gitlab.brickchain.com/brickchain/realm-ng"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/api/rest"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/providers/assets"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/providers/bindata"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/providers/dummy"
	gormprvdr "gitlab.brickchain.com/brickchain/realm-ng/pkg/providers/gorm"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/providers/mailgun"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/proxy"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/services"
	"gitlab.brickchain.com/brickchain/realm-ng/pkg/version"
	"gitlab.brickchain.com/libs/go-cache.v1"
	filestore "gitlab.brickchain.com/libs/go-filestore.v1"
	gormkeys "gitlab.brickchain.com/libs/go-keys.v1/gorm"
	pubsub "gitlab.brickchain.com/libs/go-pubsub.v1"
	jose "gopkg.in/square/go-jose.v1"
)

func main() {
	_ = godotenv.Load(".env")
	viper.AutomaticEnv()
	viper.SetDefault("log_formatter", "text")
	viper.SetDefault("log_level", "debug")
	viper.SetDefault("prod", false)
	viper.SetDefault("addr", ":6593")
	viper.SetDefault("base", "")
	viper.SetDefault("config", "./dev.yml")
	viper.SetDefault("cryptoprovider", "gorm")
	viper.SetDefault("gorm_dialect", "sqlite3")
	viper.SetDefault("gorm_options", "file:./realm.db?cache=shared")
	viper.SetDefault("cache", "file")
	viper.SetDefault("cache_dir", ".cache")
	viper.SetDefault("redis", "localhost:6379")
	viper.SetDefault("filestore_dir", ".files")
	viper.SetDefault("stats", "none")
	viper.SetDefault("adminui", "https://actions.brickchain.com/realm-admin-ng/master/")
	viper.SetDefault("proxy_domain", "r.integrity.app")
	viper.SetDefault("proxy_endpoint", "https://larsdev.plusintegrity.com/push/register")
	viper.SetDefault("email_provider", "dummy")

	logger.SetOutput(os.Stdout)
	logger.SetFormatter(viper.GetString("log_formatter"))
	logger.SetLevel(viper.GetString("log_level"))
	logger.AddContext("service", "realm-ng")
	logger.AddContext("version", version.Version)

	w = logger.GetLogger().Logger.Writer()
	defer w.Close()

	addr := viper.GetString("addr")
	server := &graceful.Server{
		Timeout: time.Duration(15) * time.Second,
		Server: &http.Server{
			Addr:        addr,
			Handler:     loadHandler(),
			ReadTimeout: time.Duration(10) * time.Second,
			ErrorLog:    log.New(w, "server", 0),
		},
	}

	logger.Debugf("server starting at %s", addr)
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}
}

var w *io.PipeWriter
var db *gorm.DB
var cacheStore cache.Cache

func loadHandler() http.Handler {

	prod := viper.GetBool("prod")
	r := httphandler.NewRouter()

	var err error
	if db == nil {
		db, err = gorm.Open(viper.GetString("gorm_dialect"), viper.GetString("gorm_options"))
		if err != nil {
			logger.Fatal(err)
		}
		if viper.GetBool("gorm_debug") {
			db.LogMode(true)
		}
		db.SetLogger(log.New(w, "database", 0))
		db.DB().SetMaxIdleConns(1)
		db.DB().SetMaxOpenConns(5)
	}

	sks, err := gormkeys.NewGormStoredKeyService(db)
	if err != nil {
		logger.Fatal(err)
	}

	kek := sha256.Sum256([]byte(viper.GetString("kek")))
	if len(kek) != 32 {
		logger.Fatal(fmt.Errorf("SHA256 hash of KEK needs to be 32 byte long, current hash is %d bytes", len(kek)))
	}

	// initialize services
	realms, err := gormprvdr.NewGormRealmService(db)
	if err != nil {
		logger.Fatal(err)
	}

	actions, err := gormprvdr.NewGormActionService(db)
	if err != nil {
		logger.Fatal(err)
	}

	controllers, err := gormprvdr.NewGormControllerService(db)
	if err != nil {
		logger.Fatal(err)
	}

	invites, err := gormprvdr.NewGormInviteService(db)
	if err != nil {
		logger.Fatal(err)
	}

	mandates, err := gormprvdr.NewGormMandateService(db)
	if err != nil {
		logger.Fatal(err)
	}

	mandateTickets, err := gormprvdr.NewGormMandateTicketService(db)
	if err != nil {
		logger.Fatal(err)
	}

	roles, err := gormprvdr.NewGormRoleService(db)
	if err != nil {
		logger.Fatal(err)
	}

	settings, err := gormprvdr.NewGormSettingService(db)
	if err != nil {
		logger.Fatal(err)
	}

	files, err := loadFilestore(r)
	if err != nil {
		logger.Fatal(err)
	}

	pubsub, err := loadPubSub()
	if err != nil {
		logger.Fatal(err)
	}

	email, err := loadEmail()
	if err != nil {
		logger.Fatal(err)
	}

	// TODO: build key set
	keyset := &jose.JsonWebKeySet{}

	bootRealmName, err := settings.Get("", "bootRealmName")
	if err != nil || bootRealmName == "" && viper.GetString("base") != "" {
		baseURL, err := url.Parse(viper.GetString("base"))
		if err != nil {
			logger.Fatal(errors.Wrap(err, "failed to parse base URL"))
		}

		bootRealmName = baseURL.Host
	}

	contextProvider := services.NewRealmsServiceProvider(
		viper.GetString("base"),
		realms,
		actions,
		controllers,
		invites,
		mandates,
		mandateTickets,
		roles,
		settings,
		files,
		sks, kek[0:32],
		pubsub,
		viper.GetString("realmtopic"),
		keyset,
		email,
		loadAssets(),
	)

	var bootContext *services.RealmService

	// logger.Infof("Bootstrap realm name: %s", bootRealmName)
	if err := contextProvider.LoadBootstrapRealm(bootRealmName); err != nil {
		if bootRealmName != "" {
			logger.Infof("Bootstrap realm does not exist, setting up realm %s", bootRealmName)
			_, err = contextProvider.New(&realm.Realm{
				ID:         bootRealmName,
				Name:       bootRealmName,
				AdminRoles: []string{fmt.Sprintf("admin@%s", bootRealmName)},
			}, nil)
			if err != nil {
				logger.Fatal(err)
			}
		} else {

			key, err := crypto.NewKey()
			if err != nil {
				logger.Fatal(err)
			}

			name := fmt.Sprintf("%s.%s", crypto.Thumbprint(key), viper.GetString("proxy_domain"))
			contextProvider.SetBase(fmt.Sprintf("https://%s", name))

			rd, err := contextProvider.New(&realm.Realm{
				Name: name,
			}, key)
			if err != nil {
				logger.Fatal(err)
			}

			bootRealmName = rd.Name
			settings.Set("", "bootRealmName", bootRealmName)
			logger.Infof("Created bootstrap realm: %s", bootRealmName)
		}

		if err := contextProvider.LoadBootstrapRealm(bootRealmName); err != nil {
			logger.Fatal(err)
		}

		bootContext = contextProvider.Get(bootRealmName)

		pw, err := crypto.GenerateRandomString(10)
		if err != nil {
			logger.Fatal(err)
		}
		bootContext.Settings().Set("password", pw)
		bootContext.Settings().Set("bootstrapped", "false")

		logger.Infof("Bootstrap password: %s", pw)

		for _, roleName := range []string{"admin", "guest", "services"} {
			if err := bootContext.Roles().Set(document.NewRole(fmt.Sprintf("%s@%s", roleName, bootRealmName))); err != nil {
				logger.Fatal("Failed to create role %s:", roleName, err)
			}
		}
	} else {
		bootContext = contextProvider.Get(bootRealmName)

		b, _ := bootContext.Settings().Get("bootstrapped")
		if b != "true" {
			pw, err := bootContext.Settings().Get("password")
			if err == nil {
				logger.Infof("Bootstrap password: %s", pw)
			}
		}
	}

	base := viper.GetString("base")
	if base == "" {
		base = fmt.Sprintf("https://%s", bootRealmName)
		contextProvider.SetBase(base)
	}

	logger.Infof("Go to %s?realm=%s to manage your Realm", viper.GetString("adminui"), bootRealmName)

	wrapper := httphandler.NewWrapper(prod)

	// Add bootstrap check middleware
	bootstrapped := false
	wrapper.AddMiddleware(func(req httphandler.Request, res httphandler.Response) (httphandler.Response, error) {
		if !bootstrapped {
			b, err := bootContext.Settings().Get("bootstrapped")
			if err == nil && b == "true" {
				bootstrapped = true
			}
		}

		if !bootstrapped {
			res.Header().Set("X-Boot-Mode", "Yes")
		}

		return res, nil
	})

	// Version handler
	r.GET("/", wrapper.Wrap(rest.Version))

	configController := rest.NewConfigController(contextProvider)
	r.GET("/realm/v2/realms/:realmID/config", wrapper.Wrap(configController.Config))

	authController := rest.NewAuthController(contextProvider)
	r.GET("/realm/v2/realms/:realmID/auth", wrapper.Wrap(authController.Authenticated))

	// .well-known
	wellKnown := rest.NewWellKnownHandler(base, contextProvider)
	r.GET("/.well-known/realm/realm.json", wrapper.Wrap(wellKnown.WellKnown))
	r.GET("/realm/v2/realms/:realmID/realm.json", wrapper.Wrap(wellKnown.WellKnownForRealm))

	// realms
	realmsController := rest.NewRealmsController(base, contextProvider, keyset)
	r.POST("/realm/v2/realms/:realmID/bootstrap", wrapper.Wrap(realmsController.Bootstrap))

	// mandate tickets
	mandateTicketController := rest.NewMandateTicketController(base, contextProvider)
	r.GET("/realm/v2/realms/:realmID/tickets/:ticketID/issue", wrapper.Wrap(mandateTicketController.IssueMandate))
	r.POST("/realm/v2/realms/:realmID/tickets/:ticketID/callback", wrapper.Wrap(mandateTicketController.IssueMandateCallback))

	mandatesController := rest.NewMandatesController(contextProvider)
	r.GET("/realm/v2/realms/:realmID/mandates/role/:roleName", wrapper.Wrap(mandatesController.List))
	r.GET("/realm/v2/realms/:realmID/mandates", wrapper.Wrap(mandatesController.List))

	invitesController := rest.NewInvitesController(contextProvider)
	r.GET("/realm/v2/realms/:realmID/invites/role/:roleName", wrapper.Wrap(invitesController.List))
	r.GET("/realm/v2/realms/:realmID/invites", wrapper.Wrap(invitesController.List))
	r.GET("/realm/v2/realms/:realmID/invites/id/:inviteID", wrapper.Wrap(invitesController.Get))
	r.POST("/realm/v2/realms/:realmID/invites", wrapper.Wrap(invitesController.Set))
	r.POST("/realm/v2/realms/:realmID/invites/id/:inviteID", wrapper.Wrap(invitesController.Set))
	r.PUT("/realm/v2/realms/:realmID/invites/id/:inviteID", wrapper.Wrap(invitesController.Set))
	r.DELETE("/realm/v2/realms/:realmID/invites/id/:inviteID", wrapper.Wrap(invitesController.Delete))
	r.PUT("/realm/v2/realms/:realmID/invites/id/:inviteID/send", wrapper.Wrap(invitesController.Send))
	r.GET("/realm/v2/realms/:realmID/invites/id/:inviteID/fetch", wrapper.Wrap(invitesController.Fetch))
	r.POST("/realm/v2/realms/:realmID/invites/id/:inviteID/callback", wrapper.Wrap(invitesController.Callback))

	// controllers
	controllersController := rest.NewControllersController(contextProvider)
	r.GET("/realm/v2/realms/:realmID/controllers", wrapper.Wrap(controllersController.ListControllers))
	r.GET("/realm/v2/realms/:realmID/controllers/id/:controllerID", wrapper.Wrap(controllersController.GetController))
	r.POST("/realm/v2/realms/:realmID/controllers", wrapper.Wrap(controllersController.Set))
	r.POST("/realm/v2/realms/:realmID/controllers/id/:controllerID", wrapper.Wrap(controllersController.Set))
	r.PUT("/realm/v2/realms/:realmID/controllers/id/:controllerID", wrapper.Wrap(controllersController.Set))
	r.DELETE("/realm/v2/realms/:realmID/controllers/id/:controllerID", wrapper.Wrap(controllersController.Delete))
	r.POST("/realm/v2/realms/:realmID/controllers/bind", wrapper.Wrap(controllersController.Bind))
	r.POST("/realm/v2/realms/:realmID/controllers/id/:controllerID/actions", wrapper.Wrap(controllersController.UpdateActions))

	// roles
	rolesController := rest.NewRolesController(contextProvider)
	r.GET("/realm/v2/realms/:realmID/roles", wrapper.Wrap(rolesController.ListRoles))
	r.GET("/realm/v2/realms/:realmID/roles/:roleID", wrapper.Wrap(rolesController.GetRole))

	// service listing
	servicesController := rest.NewServicesController(contextProvider)
	r.GET("/realm/v2/realms/:realmID/services", wrapper.Wrap(servicesController.ListServices))

	// cacheStore = loadCache()

	httphandler.SetExposedHeaders([]string{"Content-Language", "Content-Type", "X-Boot-Mode"})

	handler := httphandler.LoadMiddlewares(r, version.Version)

	if strings.HasSuffix(bootRealmName, viper.GetString("proxy_domain")) {
		go func() {
			for {
				logger.Debug("Connecting to proxy..")
				key, err := bootContext.Key()
				if err != nil {
					logger.Fatal(err)
				}

				p, err := proxy.NewProxy(key, viper.GetString("proxy_endpoint"), viper.GetString("proxy_token"))
				if err != nil {
					logger.Error(err)
				} else {
					if err := p.Subscribe(handler); err != nil {
						logger.Error(err)
					}
				}

				time.Sleep(time.Second * 1)
			}
		}()
	}

	return handler
}

func loadAssets() realm.AssetProvider {
	if viper.GetString("assets") != "" {
		return assets.NewAssetsProvider(viper.GetString("assets"))
	}

	return bindata.NewBindataProvider()
}

func loadFilestore(r *httprouter.Router) (filestore.Filestore, error) {
	switch viper.GetString("filestore") {
	case "gcs":
		return filestore.NewGCS(viper.GetString("gcs_bucket"), viper.GetString("gcs_location"), viper.GetString("gcs_project"), viper.GetString("gcs_secret"))
	default:
		l := fmt.Sprintf("%s/realm/v2/files", viper.GetString("base"))
		f, err := filestore.NewFilesystem(l, viper.GetString("filestore_dir"))
		if err != nil {
			return nil, err
		}

		//r.GET("/realm/v2/files/*filename", httphandler.Wrap(f.Handler))

		return f, nil
	}
}

func loadPubSub() (pubsub.PubSubInterface, error) {
	switch viper.GetString("pubsub") {
	case "google":
		return pubsub.NewGCloudPubSub(viper.GetString("pubsub_project_id"), viper.GetString("pubsub_credentials"))
	case "redis":
		return pubsub.NewRedisPubSub(viper.GetString("redis"))
	default:
		return pubsub.NewInmemPubSub(), nil
	}
}

func loadEmail() (realm.EmailProvider, error) {
	switch viper.GetString("email_provider") {
	case "mailgun":
		return mailgun.NewMailgunProvider(viper.GetString("mailgun_config"))
	default:
		return dummy.NewDummyEmailProvider()
	}
}
