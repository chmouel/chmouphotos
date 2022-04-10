# ChmouPhotos - A simple photos website

Website for <https://photos.chmouel.com>

## ScreenShot

![chmouphotos](https://user-images.githubusercontent.com/98980/113452108-a345b780-9403-11eb-9f6d-50f96d7aa24b.png)

## Description

This used to be a Ghost website, but it was becoming too heavy for my rpi. My
issues was with the nodejs files was killing my SD card when updating, consumes
more than 1GB of mem to just install the damn thing and pretty heavy on RAM
while running (~300mb) for small servers like RPI.

I still wanted to use the excellent theme from
[GodoFreddo](https://godofredo.ninja) and didn't need a lot of the fancy
editing features from Ghost, since I just need only a few metadatas.

I then moved it to a golang based site, which was lean but needed somewhere to
run.... when cloudfare adn github pages are just plain free. I converted my
golang stuff to generate html sites but it was awkward with regard to the
database and stuff.

So I converted everything to a hugo site and theme. may be I'll make it
independent theme one day if some people want it.
