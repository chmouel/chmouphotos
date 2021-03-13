# ChmouPhotos

Website for <https://photos.google.com>

This used to be a Ghost website but it was becoming too heavy for my rpi. My
issues was with the nodejs files was killing my SD card when updating, consumes
more than 1GB of mem to just install the damn thing and pretty heavy on RAM
while running (~300mb) for small servers like RPI.

I still wanted to use the excellent theme from
[GodoFreddo](https://godofredo.ninja) and didn't need a lot of the fancy
editing features from Ghost, since I just need only a few metadatas.

I made a `config.json` with a list of `images`/`href`/`desc`, which looks like
this :

```json
[
  {
    "image": "2021/03/IMG_20210311_122657-EFFECTS.jpg",
    "href": "checking-out-the-weather-from-the-balcony",
    "desc": "Checking out the weather from the balcony"
  },
  {
    "image": "2021/03/IMG_20210214_091010.jpg"
    "href": "snowy-sunday-group-ride-and-friends",
    "desc": "Snowy sunday group ride and friends"
  },
...
]
```

and serve the pages via a custom golang server to.

Probably could do pure static but I want to do an uploader that does the
resizing and such and being more dynamic while on the road/phone.

The service is served under [systemd](./systemd/chmouphoto.service) and only
consumes a few MBS,

Nginx serves the assets and images and pass thru for the other stuff, here is
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


