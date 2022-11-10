---
title: "Static Site Encrypted Content"
date: 2020-01-06T00:00:00-08:00
draft: false
authors:
- admin
tags:
- blog
- hugo
- crypto
---

If you've browsed this site long enough you may have noticed pages
that are password protected. These aren't particularly sensitive, just
private. I wanted to be be able to host content that was only
available to friends and family, specifically photo ablums. This is an
interesting problem when you're building a static site in Hugo.

The approach is pretty simple and came together very quickly once I
had a plan in place. One of my goals was a generalizable solution that
I could apply in a variety of circumstances and wasn't too tightly
coupled to Hugo. I did spend some time investigating shortcodes and
the other extensibility options but that quickly led nowhere. My final
solution resembles a slightly more elaborate implementation of the
StatiCrypt [^1] project.

Basically, the sensitive content is encrypted and stored inline in the
pages and then decrypted using a passphrase in the browser.  To
simplify browsing the site, the passphrase is kept in the browser's
`localStorage` and when a page loads that has encrypted content it's
automatically decrypted.

The encryption tool is a simple `golang` program [^2] and performs the
following steps:

1) Opens and parses the HTML file, searching the DOM for the top most
node with a specific CSS class. This class marks the part of the DOM
with sensitive content.
2) The HTML for that sensitive area is then rendered to a string and
encrypted, then the encrypted payload is signed.
3) A `text/template` is then processed with that encrypted information
and used to generate a new fragment of HTML.
4) This new fragment replaces the ensitive one and the resulting DOM
is rendered to a new file.

In the end, the old HTML file is replaced with the encrypted one! This
creates a very seamless browsing experience when going from
unprotected to protected areas and allows for finer control of the
protected content.

There's a few ways this could be improved, specifically around how
images or other static assets are handled. I'm tossing around the idea
of obfuscating paths and/or some very basic token verification server
side, for example. For now, things require enough hoops to jump
through.

[^1]: https://robinmoisson.github.io/staticrypt/
[^2]: https://github.com/jlewallen/jacoblewallen.com/blob/master/src/secure.go
