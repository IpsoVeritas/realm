module github.com/IpsoVeritas/realm

go 1.16

replace github.com/IpsoVeritas/document => ../document

replace github.com/IpsoVeritas/httphandler => ../httphandler

replace github.com/IpsoVeritas/crypto => ../crypto

replace github.com/IpsoVeritas/logger => ../logger

replace github.com/IpsoVeritas/keys => ../keys

require (
	cloud.google.com/go/storage v1.10.0
	github.com/DataDog/datadog-go v0.0.0-20170727083428-a420eee23bee // indirect
	github.com/IpsoVeritas/crypto v0.0.0-20181010203950-c229a2b23e68
	github.com/IpsoVeritas/document v0.0.0-20180814075806-099bc71d4b53
	github.com/IpsoVeritas/httphandler v0.0.0-20180917092253-de7d59aef300
	github.com/IpsoVeritas/keys v0.0.0-20180614130935-e07793b924eb
	github.com/IpsoVeritas/logger v0.0.0-20180912100710-b76d97958f28
	github.com/armon/consul-api v0.0.0-20180202201655-eb2c6b5be1b6 // indirect
	github.com/coreos/etcd v3.3.10+incompatible // indirect
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/denisenkom/go-mssqldb v0.10.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/facebookgo/ensure v0.0.0-20200202191622-63f1cf65ac4c // indirect
	github.com/facebookgo/stack v0.0.0-20160209184415-751773369052 // indirect
	github.com/facebookgo/subset v0.0.0-20200203212716-c811ad88dec4 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/gorilla/handlers v1.5.1 // indirect
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/jinzhu/gorm v0.0.0-20170504140837-9acaa33324bb
	github.com/jinzhu/inflection v0.0.0-20170102125226-1c35d901db3d // indirect
	github.com/jinzhu/now v1.1.2 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/jonboulle/clockwork v0.1.0 // indirect
	github.com/julienschmidt/httprouter v1.3.0
	github.com/lib/pq v0.0.0-20170603225454-8837942c3e09 // indirect
	github.com/mailgun/mailgun-go v2.0.0+incompatible
	github.com/mattn/go-colorable v0.1.0
	github.com/mattn/go-sqlite3 v1.14.7 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.13.0 // indirect
	github.com/pascaldekloe/goe v0.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v0.9.3 // indirect
	github.com/prometheus/procfs v0.0.0-20190522114515-bc1a522cf7b1 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/spf13/viper v1.8.1
	github.com/subosito/twilio v0.0.2-0.20160901001414-ef2f13504366
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/tylerb/graceful v1.2.16-0.20170221171003-d72b0151351a
	github.com/ugorji/go v1.1.4 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	github.com/xordataexchange/crypt v0.0.3-0.20170626215501-b2862e3d0a77 // indirect
	go.etcd.io/bbolt v1.3.2 // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	google.golang.org/api v0.50.1-0.20210702115825-985b53fdf9cd
	gopkg.in/resty.v1 v1.12.0
	gopkg.in/square/go-jose.v1 v1.1.2
)
