[![license](https://img.shields.io/github/license/mashape/apistatus.svg)]()
[![Build Status](https://travis-ci.org/jamillosantos/fasthttp-websocket.svg?branch=master)](https://travis-ci.org/jamillosantos/fasthttp-websocket)
[![Coverage Status](https://coveralls.io/repos/github/jamillosantos/fasthttp-websocket/badge.svg?branch=master)](https://coveralls.io/github/jamillosantos/fasthttp-websocket?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jamillosantos/migration)](https://goreportcard.com/report/github.com/jamillosantos/fasthttp-websocket)

**This library is under development**

# fasthttp-websocket

A WebSocket implementation on top of the fasthttp.

# Motivation

Yay! The Gorilla WebSocket works great, and some code of this library are
actually based on their implementation, but it does not support the Valyala
Fasthttp library.

At first, I tried to use a the Leavengood
([https://github.com/leavengood/websocket]()) fork of the Gorilla Websocket. 
However, I could not find it useful. Actally I could not find even a reference
for the fasthttp at the master branch. How strange is that!?

Hence, I decided to come up with a websocket protocol implementation for
the fasthttp. Meanwhile, I will be using interfaces for pretty much anything to
try to avoid any heavy bind between the fasthttp and this websocket 
implementation.

# Testing

In order to make sure everything is working properly, after each minor release
an Autobahn Test Suite report will be released to keep track of the supported
features.

More info at [](https://github.com/crossbario/autobahn-testsuite).
