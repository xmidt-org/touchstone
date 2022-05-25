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

// Package touchbundle provides a simple way to create bundles of metrics
// where the names of the metrics can be optionally tailored but the cardinality,
// type, etc. cannot.
//
// Typical packages expect metrics to be of certain types, e.g. counter or gauge,
// and to have certain labels.  But a package might wish to allow metric names
// to be set through configuration or through application code.  This package aims
// to provide a standard approach to this using structs and struct tags.
package touchbundle
