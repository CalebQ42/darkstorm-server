# Darkstorm Server

Experimenting with a Go server for personal uses. Combines a simple website server with a tcp forwarder.

Configure which ports go to which addresses via /etc/darkstorm-server.conf in the form `type port address`. If type is not given, tcp is assumed.
