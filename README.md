
# gifserver

`gifserver` is a service written in Go that transcodes GIFs to videos on the
fly. Useful for displaying user uploaded GIFs on a website without incurring
slowdowns from excess bandwidth.

The server is a wrapper around `ffmpeg`, you must have the `ffmpeg` command
installed on your system.

The server will automatically cache transcoded GIFs to disk so subsequent
requests will avoid the initial conversion time. The server is aware of what's
currently being converted so multiple requests to the same GIF during a
conversion do not trigger multiple conversions.

## Install

```bash
go get github.com/leafo/gifserver
go install github.com/leaf/gifserver

gifserver -help
```

## Running

```bash
gifserver
```

By default the command will look for a config file named `gifserver.json` in
the current directory. You can override the config file's path with the
`-config` command line flag.

To resize a GIF just request the server at the `/transcode` path with the URL
parameter to the URL you want transcoded (The `http` is optional):

```
http://localhost:9090/transcode?url=leafo.net/dump/test.gif
```

An MP4 is returned by default. If there are any errors a 500 error is returned
and the server returns an error message.

## Config

A JSON file is used to configure the server. The default config is as follows:

```json
{
	"Address":  ":9090",
	"Secret":   "",
	"CacheDir": "gifcache",
	"MaxBytes": 1024 * 1024 * 5
}
```

Your config can replace any combination of the defaults with your own values.

* **Address** -- the address to bind the server to
* **Secret** -- secret key to use for singed URLs. If `""` is the secret key then signed URLs are disabled
* **CacheDir** -- where to cache transcoded GIFs
* **MaxBytes** -- the max size of GIF in bytes allowed to be processed

## Signed URLs

You should used signed URLs whenever possible to avoid malicious users from
triggering a [denial-of-service-attack][0] on your server by sending a large
amount of conversion requests.

To enable signed URLs provide a `Secret` in your config file. The server uses
hmac sha1 to generate signatures for URLs. Generating the signature is
relatively simple:

Given the following request to transcode a GIF:

```
http://localhost:9090/transcode?url=leafo.net/dump/test.gif
```

Take the entire path and query from the URL an perform a SHA1 HMAC digest on
it. Base64 encode the sum. Then append the URL escaped signature to the end of
the URL as the query parameter `sig`:

```
sum = hmac_sha1(theSecret, "/transcode?url=leafo.net/dump/test.gif")
signature = encode_base64(sum)

signed_url = url + "&sig=" + url_escape(signature)
```

Then perform the request with the signed URL. You should always append the
signature at the end of the URL. You must not change the order of the original
query parameters in any way when appending the signature otherwise the
signature is invalid.

## About

Author: Leaf Corcoran (leafo) ([@moonscript](http://twitter.com/moonscript))  
Email: leafot@gmail.com  
Homepage: <http://leafo.net>  
License: MIT, Copyright (C) 2014 by Leaf Corcoran


  [0]: http://en.wikipedia.org/wiki/Denial-of-service_attack

