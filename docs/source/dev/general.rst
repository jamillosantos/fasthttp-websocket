How it works
------------

This section will explain how the connection will work.

1.  The client requests the connection
++++++++++++++++++++++++++++++++++++++

The websocket connection MUST start as a normal normal HTTP request. The browser
will call the given endpoint with a set of special headers asking for a
websocket connection.

It happens when the client instantiates a ``new WebSocket`` object passing the
endpoint of our server:

.. code-block:: javascript
   :linenos:

	   var socket = new WebSocket("/ws");
	..

2. The server upgrades the connection
+++++++++++++++++++++++++++++++++++++

Once the server receives the connection, it will respond upgrading the
connection.