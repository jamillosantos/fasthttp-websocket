[![license](https://img.shields.io/github/license/mashape/apistatus.svg)]()
[![Documentation Status](https://readthedocs.org/projects/fasthttp-websocket/badge/?version=latest)](http://fasthttp-websocket.readthedocs.io/en/latest/?badge=latest)
[![Build Status](https://travis-ci.org/jamillosantos/websocket.svg?branch=master)](https://travis-ci.org/jamillosantos/websocket)
[![Coverage Status](https://coveralls.io/repos/github/jamillosantos/websocket/badge.svg?branch=master)](https://coveralls.io/github/jamillosantos/websocket?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jamillosantos/websocket)](https://goreportcard.com/report/github.com/jamillosantos/websocket)

**This library is under development**

# Websocket

A WebSocket implementation on top of the fasthttp.

# Implementation

The [RFC 6455](https://tools.ietf.org/html/rfc6455) describes the WebSocket
Protocol. This is the main source of information used for this implementation.

# Motivation

Yay! The Gorilla WebSocket works great, and some code of this library are
actually based on their implementation, but it does not support the Valyala
Fasthttp library.

At first, I tried to use a the Leavengood
([https://github.com/leavengood/websocket](https://github.com/leavengood/websocket))
fork of the Gorilla Websocket. However, I could not find it useful. Actually I
could not find even a reference for the fasthttp at the master branch. How
strange is that!?

Hence, I decided to come up with a websocket protocol implementation aiming
specially the fasthttp.

# Testing

In order to make sure everything is working properly, after each minor release
an Autobahn Test Suite report will be released to keep track of the supported
features.

More info at [https://github.com/crossbario/autobahn-testsuite](https://github.com/crossbario/autobahn-testsuite).
