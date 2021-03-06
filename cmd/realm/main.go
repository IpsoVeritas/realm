//go:generate go generate ../../pkg/providers/bindata/files.go

package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	crypto "github.com/IpsoVeritas/crypto"
	httphandler "github.com/IpsoVeritas/httphandler"
	gormkeys "github.com/IpsoVeritas/keys/gorm"
	logger "github.com/IpsoVeritas/logger"
	realm "github.com/IpsoVeritas/realm"
	"github.com/IpsoVeritas/realm/pkg/api/rest"
	"github.com/IpsoVeritas/realm/pkg/providers/assets"
	"github.com/IpsoVeritas/realm/pkg/providers/bindata"
	"github.com/IpsoVeritas/realm/pkg/providers/dummy"
	filestore "github.com/IpsoVeritas/realm/pkg/providers/filestore"
	gormprvdr "github.com/IpsoVeritas/realm/pkg/providers/gorm"
	"github.com/IpsoVeritas/realm/pkg/providers/mailgun"
	"github.com/IpsoVeritas/realm/pkg/services"
	"github.com/IpsoVeritas/realm/pkg/version"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	colorable "github.com/mattn/go-colorable"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/tylerb/graceful"
	jose "gopkg.in/square/go-jose.v1"
)

func main() {
	loadEnv()
	viper.AutomaticEnv()
	viper.SetDefault("log_formatter", "text")
	viper.SetDefault("log_level", "debug")
	viper.SetDefault("prod", false)
	viper.SetDefault("addr", ":6593")
	viper.SetDefault("base", "http://localhost:6593")
	viper.SetDefault("cryptoprovider", "gorm")
	viper.SetDefault("gorm_dialect", "sqlite3")
	viper.SetDefault("gorm_options", "file:./realm.db?cache=shared")
	viper.SetDefault("cache", "file")
	viper.SetDefault("cache_dir", ".cache")
	viper.SetDefault("redis", "localhost:6379")
	viper.SetDefault("filestore_dir", ".files")
	viper.SetDefault("stats", "none")
	viper.SetDefault("adminui", "https://admin.integrity.app")
	viper.SetDefault("proxy_domain", "r.integrity.app")
	viper.SetDefault("proxy_endpoint", "https://proxy.svc.integrity.app")
	viper.SetDefault("email_provider", "dummy")
	viper.SetDefault("mailgun_config", "./mailgun.yml")
	viper.SetDefault("key", "./realm.pem")
	viper.SetDefault("allow_patching", false)

	if runtime.GOOS == "windows" && viper.GetString("log_formatter") == "text" {
		logger.SetOutput(colorable.NewColorableStdout())
		logger.SetLogrusFormatter(&logrus.TextFormatter{ForceColors: true})
	} else {
		logger.SetOutput(os.Stdout)
		logger.SetFormatter(viper.GetString("log_formatter"))
	}
	logger.SetLevel(viper.GetString("log_level"))
	logger.AddContext("service", "realm")
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

func loadEnv() {
	home, err := homedir.Dir()
	if err == nil {
		path := fmt.Sprintf("%s%s.config%sbrickchain/realm.env", home, string(os.PathSeparator), string(os.PathSeparator))
		if err = godotenv.Load(path); err == nil {
			logger.Infof("ENV variables loaded from  %s", path)
		}
	}
	current, err := os.Getwd()
	if err == nil {
		path := fmt.Sprintf("%s%s.env", current, string(os.PathSeparator))
		if err = godotenv.Load(path); err == nil {
			logger.Infof("ENV variables loaded from  %s", path)
		}
	}
}

func loadHandler() http.Handler {

	prod := viper.GetBool("prod")
	r := httphandler.NewRouter()

	wrapper := httphandler.NewWrapper(prod)

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

	email, err := loadEmail()
	if err != nil {
		logger.Fatal(err)
	}

	// TODO: build key set
	keyset := &jose.JsonWebKeySet{}

	bootRealmID, err := settings.Get("", "bootRealmID")
	if err != nil || bootRealmID == "" && viper.GetString("base") != "" {
		baseURL, err := url.Parse(viper.GetString("base"))
		if err != nil {
			logger.Fatal(errors.Wrap(err, "failed to parse base URL"))
		}

		bootRealmID = baseURL.Host
	}

	var key *jose.JsonWebKey
	if bootRealmID == "" || viper.GetString("base") == "" {
		_, err := os.Stat(viper.GetString("key"))
		if err != nil {
			key, err = crypto.NewKey()
			if err != nil {
				logger.Fatal(err)
			}

			kb, err := crypto.MarshalToPEM(key)
			if err != nil {
				logger.Fatal(err)
			}

			if err := ioutil.WriteFile(viper.GetString("key"), kb, 0600); err != nil {
				logger.Fatal(err)
			}
		} else {
			kb, err := ioutil.ReadFile(viper.GetString("key"))
			if err != nil {
				logger.Fatal(err)
			}

			key, err = crypto.UnmarshalPEM(kb)
			if err != nil {
				logger.Fatal(err)
			}
		}
	}

	base := viper.GetString("base")
	if base == "" {
		base = fmt.Sprintf("https://%s", bootRealmID)
		// contextProvider.SetBase(base)
	}

	contextProvider := services.NewRealmsServiceProvider(
		base,
		realms,
		actions,
		controllers,
		invites,
		mandates,
		mandateTickets,
		roles,
		settings,
		sks, kek[0:32],
		viper.GetString("realm_topic"),
		keyset,
		email,
		loadAssets(),
	)

	var bootContext *services.RealmService

	if err := contextProvider.LoadBootstrapRealm(bootRealmID); err != nil {
		logger.Infof("Bootstrap realm does not exist, setting up realm %s", bootRealmID)
		_, err = contextProvider.New(&realm.Realm{
			ID: bootRealmID,
		}, nil)
		if err != nil {
			logger.Fatal(err)
		}

		if err := contextProvider.LoadBootstrapRealm(bootRealmID); err != nil {
			logger.Fatal(err)
		}

		bootContext = contextProvider.Get(bootRealmID)

		pw, err := crypto.GenerateRandomString(16)
		if err != nil {
			logger.Fatal(err)
		}
		pw = strings.Replace(strings.Replace(pw, "-", "", -1), "_", "", -1)
		bootContext.Settings().Set("password", pw)
		bootContext.Settings().Set("bootstrapped", "false")

		logger.Infof("Bootstrap password: %s", pw)

	} else {
		bootContext = contextProvider.Get(bootRealmID)

		b, err := bootContext.Settings().Get("bootstrapped")
		if err != nil {
			if err := bootContext.Settings().Set("bootstrapped", "false"); err != nil {
				logger.Fatal(err)
			}
		}
		if b != "true" {
			pw, err := bootContext.Settings().Get("password")
			if err != nil {
				pw, err = crypto.GenerateRandomString(10)
				if err != nil {
					logger.Fatal(err)
				}
				bootContext.Settings().Set("password", pw)
			}

			logger.Infof("Bootstrap password: %s", pw)
		} else {
			// temporary code to refresh the descriptor after the name/id change
			bootRealm, err := bootContext.Realm()
			if err == nil {
				bootRealm.Descriptor.ID = bootRealmID
				bootContext.Set(bootRealm)
			}
		}
	}

	files, err := loadFilestore(base, wrapper, r)
	if err != nil {
		logger.Fatal(err)
	}

	contextProvider.SetFilestore(files)
	logger.Infof("Go to %s#/%s to manage your realm", viper.GetString("adminui"), bootRealmID)

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
	r.GET("/realm/v2/realms", wrapper.Wrap(realmsController.ListRealms))
	r.POST("/realm/v2/realms", wrapper.Wrap(realmsController.PostRealm))
	r.GET("/realm/v2/realms/:realmID", wrapper.Wrap(realmsController.GetRealm))
	r.PUT("/realm/v2/realms/:realmID", wrapper.Wrap(realmsController.UpdateRealm))
	r.DELETE("/realm/v2/realms/:realmID", wrapper.Wrap(realmsController.DeleteRealm))
	r.POST("/realm/v2/realms/:realmID/icon", wrapper.Wrap(realmsController.IconHandler))
	r.POST("/realm/v2/realms/:realmID/banner", wrapper.Wrap(realmsController.BannerHandler))

	// realm actions
	r.POST("/realm/v2/realms/:realmID/do/join", wrapper.Wrap(realmsController.JoinRealm))

	// bootstrap realm
	r.POST("/realm/v2/realms/:realmID/bootstrap", wrapper.Wrap(realmsController.Bootstrap))

	// mandate tickets
	mandateTicketController := rest.NewMandateTicketController(base, contextProvider)
	r.GET("/realm/v2/realms/:realmID/tickets/:ticketID/issue", wrapper.Wrap(mandateTicketController.IssueMandate))
	r.POST("/realm/v2/realms/:realmID/tickets/:ticketID/callback", wrapper.Wrap(mandateTicketController.IssueMandateCallback))

	// mandates
	mandatesController := rest.NewMandatesController(contextProvider)
	r.GET("/realm/v2/realms/:realmID/mandates/role/:roleName", wrapper.Wrap(mandatesController.List))
	r.GET("/realm/v2/realms/:realmID/mandates", wrapper.Wrap(mandatesController.List))
	r.GET("/realm/v2/realms/:realmID/mandate/:mandateID", wrapper.Wrap(mandatesController.Get))
	r.PUT("/realm/v2/realms/:realmID/mandates/:mandateID/revoke", wrapper.Wrap(mandatesController.Revoke))
	r.POST("/realm/v2/realms/:realmID/mandates/issue", wrapper.Wrap(mandatesController.Issue))

	// invites
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
	r.GET("/realm/v2/realms/:realmID/roles", wrapper.Wrap(rolesController.List))
	r.GET("/realm/v2/realms/:realmID/roles/:roleID", wrapper.Wrap(rolesController.Get))
	r.POST("/realm/v2/realms/:realmID/roles", wrapper.Wrap(rolesController.Set))
	r.POST("/realm/v2/realms/:realmID/roles/:roleID", wrapper.Wrap(rolesController.Set))
	r.PUT("/realm/v2/realms/:realmID/roles/:roleID", wrapper.Wrap(rolesController.Set))
	r.DELETE("/realm/v2/realms/:realmID/roles/:roleID", wrapper.Wrap(rolesController.Delete))

	// service listing
	servicesController := rest.NewServicesController(contextProvider)
	r.GET("/realm/v2/realms/:realmID/services", wrapper.Wrap(servicesController.ListServices))

	// cacheStore = loadCache()

	httphandler.SetExposedHeaders([]string{"Content-Language", "Content-Type", "X-Boot-Mode"})

	handler := httphandler.LoadMiddlewares(r, version.Version)

	return handler
}

func loadAssets() realm.AssetProvider {
	if viper.GetString("assets") != "" {
		return assets.NewAssetsProvider(viper.GetString("assets"))
	}

	return bindata.NewBindataProvider()
}

func loadFilestore(base string, wrapper *httphandler.Wrapper, r *httprouter.Router) (filestore.Filestore, error) {
	switch viper.GetString("filestore") {
	case "gcs":
		return filestore.NewGCS(viper.GetString("gcs_bucket"), viper.GetString("gcs_location"), viper.GetString("gcs_project"), viper.GetString("gcs_secret"))
	default:
		l := fmt.Sprintf("%s/realm/v2/files", base)
		f, err := filestore.NewFilesystem(l, viper.GetString("filestore_dir"))
		if err != nil {
			return nil, err
		}

		r.GET("/realm/v2/files/*filename", wrapper.Wrap(f.Handler))

		return f, nil
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
