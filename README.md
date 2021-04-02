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

The pictures are stored in a sqlite or mysql DB, the schemas is simple see the Item
structure in [photos/data.go](./photos/data.go) for the structure.

It takes it and serves the pages via a custom golang server.

Probably could be a pure static site but I want to do an uploader that does the
resizing and such and being more dynamic while on the road/phone. Maybe in the
future, it's fun to experiment.

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

Things should be pretty quick, if it isn't I probably would be adding static
html caching...

## Upload

There is a simple upload page available in `/upload`, it's up to you to protect it
via nginx or other means.

It uses [imagemagick](https://imagemagick.org/) to resize the images so you
would need to install this. It needs to have `/usr/share/dict/words` to generate
random words for uniqueness. For example install the package `wamerican` on
debianies distros for the american word list

## Bugs/Ideas

- Doensn't start up if you don't have the fill up the DB with 6 items already
- Connect with Google Photos API, grab favourites or some other forms and
  generate from there?
- Full on static ? Upload on CI from a GIT project?
- When uploading add the status of the resizing which may take up to a minute on
  small rpi.
- Add disabled field to disable an entry without removing it.
- Use webp instead of jpeg?
