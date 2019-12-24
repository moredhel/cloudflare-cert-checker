# Notes

A very simple script to scrape the Cloudflare API & grab the expiration dates of respective certificates.

This is only really designed for people only using Cloudflare for DNS (and not giving CF control of their certs).
__If you aren't sure about what you're doing, just let Cloudflare manage your certs__


# Usage

If not, this is how you use it:

``` shell
$ export CLOUDFLARE_EMAIL="me@example.com"
$ export CLOUDFLARE_API_KEY="key_here"
$ go run main.go > domains.log
```

now you can simply browse through the output e.g:

``` shell
# certs that expire in the next month
$ grep "day" domains.log
```

``` shell
# certs that expire today
$ grep "\(hour\|minute\|second\)" domains.log
```

``` shell
# certs that information could not be identified
$ grep "fail" domains.log
```

