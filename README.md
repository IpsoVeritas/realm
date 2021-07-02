# Getting started with realm-ng

## Compile realm-ng

Check out realm-ng from the git repository:

    go get github.com/Brickchain/realm

    cd $GOPATH/src/github.com/Brickchain/realm/cmd

Add a .env file with this content to that directory:

    EMAIL_PROVIDER=mailgun
    MAILGUN_CONFIG=dev.yml
    LOG_FORMATTER=dev
    GORM_DEBUG=false

You also need to add a dev.yml file, with a mailgun configuration. It looks like this:

    mailgun:
        mailgun_domain: "mg.brickchain.com"
        mailgun_api_key: <secret>
        mailgun_public_api_key: <secret>
        mailgun_from: "Brickchain <mailgun@mg.brickchain.com>"
        mailgun_testmode: "false"

To compile realm-ng:

    go build

## Run realm-ng

Run the realm by starting the executible from the build process:

    ./realm

By default, this starts the realm-ng with the proxy tunnel. See the output log for addresses. If you start realm-ng for the first time, the bootstrap password will also be logged. Use the admin interface URL to bootstrap the realm with the bootstrap password. Make sure to keep a copy of the admin mandate if you want to keep the realm for a while, during development.

Two files are created when starting the realm, a realm.pem file for the tunnel proxy. Keep this if you want to keep the same address for the realm during development. The other file is the realm.db file, that is the sqlite3 database for realm storage.

If you want to start the realm with using localhost addresses, set this environment variable before starting up the realm:

    export BASE=localhost:6593
