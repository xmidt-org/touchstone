/*
Package touchhttp defines the HTTP-specific behavior for metrics within
an uber/fx app which uses the touchstone package.

Bootstrapping is similar to touchstone:  A Config object can be available
in the enclosing fx.App that will tailor certain aspects of the metrics http.Handler.

ServerBundle and ClientBundle are prebaked, opinionated metrics for instrumenting
HTTP servers and clients.  They each produce middleware given a label for a particular
server or client.
*/
package touchhttp
