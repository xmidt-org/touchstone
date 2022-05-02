/**
 * Copyright 2022 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
