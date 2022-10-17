# PhotoGallery

Simple MVP of web-application written in Go, without MVC go web frameworks.
However, it works pretty well and implements all the usual stuff you're used to seeing in a web applications.

The frontend is made on views rendered by Go Templates with no JS at all. Yes, it works fast, but look... meeeeeh backendish :D

So the PhotoGallery model can handle multiple users and provide those users with the ability to create multiple galleries and edit them. Users can upload images, delete images in their galleries, and delete an entire gallery at once.

The galleries themselves are public, so you can share your photos with your friends! Awesome!

I think this app is pretty solid in terms of security: at least we have protection against SQL infections provided to us by the default html/template package, user passwords are encrypted with salt and pepper, and we also have CSRF protection in middleware by validating the csrf-token in every request to the server.

# Install

A simple `go build` that should automatically install all dependencies from go.sum.

The PhotoGallery model works with PostgreSQL using the Gorm module, and any version of Postgress above 9 should work fine.

To be able to run PhotoGallery, you must have access to a PostgreSQL instance with the database, port number, and credentials specified in the configuration file.


# Getting started
The development environment can be started with the default configuration provided as an example `.config` json file at the root of the repository.

If the configuration file is missing, the development environment can still be started with the built-in default configuration, the value of which is equivalent to this example:

```
{
    "port": 3000,
    "env": "dev",
    "pepper": "secret-random-string-dev",
    "hamc_key": "secret-hmac-key-dev",
    "database": {
        "host": "localhost",
		"port": 5432,
		"user": "admin",
		"password": "qwerty",
		"name": "photogallery_dev"
    }
}
```

For a production environment, the `-prod true` flag is required at startup.

In this case, you can't start the server with the default build-in configuration *if the config file is missing*, so a config file is needed to run in production.

Yes, you still can start prod with the db password `qwerty` and `ololo` pepper, but that's not a good idea at all.
