# ChmouPhotos

Website for <https://photos.chmouel.com>

## ScreenShot

<img width="1351" alt="image" src="https://user-images.githubusercontent.com/98980/113452108-a345b780-9403-11eb-9f6d-50f96d7aa24b.png">

## Description

This used to be a Ghost website but it was becoming too heavy for my rpi. My
issues was with the nodejs files was killing my SD card when updating, consumes
more than 1GB of mem to just install the damn thing and pretty heavy on RAM
while running (~300mb) for small servers like RPI.

I still wanted to use the excellent theme from
[GodoFreddo](https://godofredo.ninja) and didn't need a lot of the fancy
editing features from Ghost, since I just need only a few metadatas.

The pictures are stored in a DB as [supported by Gorm
](https://gorm.io/docs/connecting_to_the_database.html) with simple schemas.

It takes the information from the DB and serves the pages via a custom golang server.

Probably could be a pure static site but I wanted to have an uploader that does the
resizing and such and being more dynamic while on the road/phone. Maybe in the
future if this website (which currently only receives bots and my own hitview) gets popular.

## Setup

The service is served under [systemd](./systemd/chmouphoto.service) and only
consumes a few MBS.

Nginx serves the assets and images and proxy thru for the html stuff, here is
the snippet from my config :

```conf
    location ~ /(content|assets) {
        root /home/www/photos/;
    }

    location / {
        max_ranges 0;
        proxy_set_header   X-Forwarded-For      $proxy_add_x_forwarded_for;
        proxy_set_header   Host              $http_host;
        proxy_set_header   X-Forwarded-Proto    $scheme;
        proxy_cache my_cache;
        proxy_cache_revalidate on;
        proxy_cache_min_uses 3;
        proxy_cache_use_stale error timeout updating http_500 http_502
                              http_503 http_504;
        proxy_cache_background_update on;
        proxy_cache_lock on;
        proxy_pass         http://127.0.0.1:8483;
    }
```

It's probably not reusable as is yet, but you can inspire yourself by it if you
move a Ghost website to a static config.

Things should be pretty quick, if it isn't I probably could add some simple HIT/MISS static
html caching...

## Config

Uses environment variable to configure the service (which makes it easy to plug
in cloud native environement). 

Environement variables are : 

* **PHOTOS_HTML_DIRECTORY**: Html directory of content, asset and html (required)
* **PHOTOS_DB**: The database DSN to connect see [gorm
  documentation](https://gorm.io/docs/connecting_to_the_database.html)
  (required)
* **PHOTOS_HOST**: The host where the service will bind (default: **127.0.0.1**)
* **PHOTOS_PORT**: The port where the service will bind (default: **8483**)


## Upload

There is a simple upload page available in `/upload`, it's up to you to protect it
via nginx or other means.

It uses [imagemagick](https://imagemagick.org/) to resize the images so you
would need to install this. It needs to have `/usr/share/dict/words` to generate
random words for uniqueness. For example install the package `wamerican` on
debianies distros for the american word list

## Bugs/Ideas

- Currently does not start up if you don't have your DB filed-up with 6 items already
- Connect with Google Photos API, grab favourites or some other forms and
  generate from there?
- Full on static ? Upload on CI from a GIT project?
- When uploading add the status of the resizing which may take up to a minute on
  small rpi.
- Add disabled field to disable an entry without removing it.
- Use webp instead of jpeg?
