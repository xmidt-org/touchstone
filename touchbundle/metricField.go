package touchbundle

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/multierr"
)

const (
	// TagTouchstone is the struct field tag that controls whether touchstone
	// ignores the field.  Setting this tag to "-" will cause the field to be
	// ignored and not populated.
	//
	// Use this tag to ignore struct fields that would otherwise be populated,
	// e.g. if a prometheus.Counter field should be ignored.
	TagTouchstone = "touchstone"

	// TagNamespace is the struct field tag that specifies the metric namespace.
	// If absent, the default namespace from the Factory is used.
	TagNamespace = "namespace"

	// TagSubsystem is the struct field tag that specifies the metric subsystem.
	// If absent, the default subsystem from the Factory is used.
	TagSubsystem = "subsystem"

	// TagName is the struct field tag that specifies the metric name.  If absent,
	// the struct field name is snakecased and used as the metric name, e.g.
	// a field such as "MyAppCounter *prometheus.CounterVec" has a default name
	// of "my_app_counter".
	TagName = "name"

	// TagHelp is the struct field tag that specifies the metric help.  There is
	// no default for this tag.
	TagHelp = "help"

	// TagBuckets is the struct field tag specifying the set of histogram buckets.
	// The format of this tag is a comma-delimited string containing float64 values.
	// Internal whitespace is allowed.
	TagBuckets = "buckets"

	// TagObjectives is the struct field tag specifying the set of summary objectives.
	// The format of this tag is a comma-delimited string containing float64 pairs
	// separated by semi-colons, e.g. "1.0:2.5, 3.5:6.7".  Internal whitespace
	// is allowed, e.g. "1.0: 2.5,  3.5: 6.7".
	TagObjectives = "objectives"

	// TagMaxAge is the struct field tag specifying the summary MaxAge.  This tag's
	// value must parse as a uint32.
	TagMaxAge = "maxAge"

	// TagAgeBuckets is the struct field tag specifying the summary AgeBuckets.  This
	// tag's value must parse as a uint32.
	TagAgeBuckets = "ageBuckets"

	// TagBufCap is the struct field tag specifying the summary BufCap.  This tag's
	// value must parse as a uint32.
	TagBufCap = "bufCap"

	// TagLabelNames specifies the set of label names for the metric.  Only permitted
	// for vector metrics, e.g. *prometheus.CounterVec.  This type is only valid for
	// vector types.
	TagLabelNames = "labelNames"

	// TagType is the struct field tag indicating the type of metric, e.g. histogram
	// or summary.  This tag is only valid when the struct field type doesn't
	// uniquely specify a metric, e.g. prometheus.Observer.  If the struct field type
	// does specify a metric, this tag cannot be supplied or an error is raised.
	TagType = "type"

	// TypeHistogram is the TagType value indicating that the metric is a histogram
	// or histogram vector.
	TypeHistogram = "histogram"

	// TypeSummary is the TagType value indicating that the metric is a summary
	// or summary vector.
	TypeSummary = "summary"
)

// FieldError represents an error while processing a metric field.
type FieldError struct {
	// Field is the struct field corresponding to the metric.
	Field reflect.StructField

	// Cause is the wrapped error that caused this field error.
	Cause error

	// Message is the error message associated with the field.
	Message string
}

func (fe *FieldError) Unwrap() error {
	return fe.Cause
}

func (fe *FieldError) Error() string {
	return fmt.Sprintf(
		"'%s %s': %s",
		fe.Field.Name,
		fe.Field.Type,
		fe.Message,
	)
}

var (
	counterType      = reflect.TypeOf((*prometheus.Counter)(nil)).Elem()
	counterVecType   = reflect.TypeOf((*prometheus.CounterVec)(nil))
	gaugeType        = reflect.TypeOf((*prometheus.Gauge)(nil)).Elem()
	gaugeVecType     = reflect.TypeOf((*prometheus.GaugeVec)(nil))
	histogramType    = reflect.TypeOf((*prometheus.Histogram)(nil)).Elem()
	histogramVecType = reflect.TypeOf((*prometheus.HistogramVec)(nil))
	summaryType      = reflect.TypeOf((*prometheus.Summary)(nil)).Elem()
	summaryVecType   = reflect.TypeOf((*prometheus.SummaryVec)(nil))
	observerType     = reflect.TypeOf((*prometheus.Observer)(nil)).Elem()
	observerVecType  = reflect.TypeOf((*prometheus.ObserverVec)(nil)).Elem()

	histogramTagNames = []string{TagBuckets}
	summaryTagNames   = []string{TagObjectives, TagMaxAge, TagAgeBuckets, TagBufCap}
	observerTagNames  = append(
		append([]string{}, histogramTagNames...),
		summaryTagNames...,
	)
)

// metricField is a type alias for reflect.StructField with functionality
// around extracting metrics information from code.
type metricField reflect.StructField

// skip runs standard tests against a struct field to see if touchstone
// should ignore it.
func (mf metricField) skip() bool {
	return len(mf.PkgPath) > 0 ||
		mf.Anonymous ||
		mf.Tag.Get(TagTouchstone) == "-"
}

// name returns the metric name for this field.
func (mf metricField) name() string {
	return MetricName(reflect.StructField(mf))
}

// newCounterOpts constructs a prometheus.CounterOpts from this struct field.
// The tag names that would never apply to any counter are also checked and,
// if any are present, this method returns an error.
func (mf metricField) newCounterOpts() (opts prometheus.CounterOpts, err error) {
	err = mf.checkTagNotAllowed(err, observerTagNames...)
	opts = prometheus.CounterOpts{
		Name:      mf.name(),
		Help:      mf.help(),
		Namespace: mf.namespace(),
		Subsystem: mf.subsystem(),
	}

	return
}

// newGaugeOpts constructs a prometheus.GaugeOpts from this struct field.
// The tag names that would never apply to any gauge are also checked and,
// if any are present, this method returns an error.
func (mf metricField) newGaugeOpts() (opts prometheus.GaugeOpts, err error) {
	err = mf.checkTagNotAllowed(err, observerTagNames...)
	opts = prometheus.GaugeOpts{
		Name:      mf.name(),
		Help:      mf.help(),
		Namespace: mf.namespace(),
		Subsystem: mf.subsystem(),
	}

	return
}

func (mf metricField) newHistogramOpts() (opts prometheus.HistogramOpts, err error) {
	opts = prometheus.HistogramOpts{
		Name:      mf.name(),
		Help:      mf.help(),
		Namespace: mf.namespace(),
		Subsystem: mf.subsystem(),
	}

	var parseErr error
	opts.Buckets, parseErr = mf.buckets()
	err = multierr.Append(err, parseErr)

	return
}

func (mf metricField) newSummaryOpts() (opts prometheus.SummaryOpts, err error) {
	opts = prometheus.SummaryOpts{
		Name:      mf.name(),
		Help:      mf.help(),
		Namespace: mf.namespace(),
		Subsystem: mf.subsystem(),
	}

	opts.Objectives, err = mf.objectives(err)
	opts.MaxAge, err = mf.maxAge(err)
	opts.AgeBuckets, err = mf.ageBuckets(err)
	opts.BufCap, err = mf.bufCap(err)

	return
}

func (mf metricField) newObserverOpts() (opts interface{}, err error) {
	var (
		metricType        = mf.Tag.Get(TagType)
		ambiguousTagNames []string
	)

	switch {
	case metricType == TypeHistogram:
		opts, err = mf.newHistogramOpts()
		ambiguousTagNames = summaryTagNames

	case metricType == TypeSummary:
		opts, err = mf.newSummaryOpts()
		ambiguousTagNames = histogramTagNames

	case len(metricType) > 0:
		err = &FieldError{
			Field:   reflect.StructField(mf),
			Message: fmt.Sprintf("'%s' is not a valid observer metric type", metricType),
		}

	// failing an explicit type, autodetect the type, falling back to a histogram

	case mf.hasAnyTagNames(summaryTagNames...):
		opts, err = mf.newSummaryOpts()
		ambiguousTagNames = histogramTagNames

	default:
		opts, err = mf.newHistogramOpts()
		ambiguousTagNames = summaryTagNames
	}

	err = mf.checkTagAmbiguous(err, ambiguousTagNames...)
	return
}

// newOpts creates a metric *Opts struct, along with label names, if this
// field is of a type supported by touchstone.
func (mf metricField) newOpts() (opts interface{}, labelNames []string, err error) {
	switch mf.Type {
	case counterType:
		opts, err = mf.newCounterOpts()
		err = mf.checkTagNotAllowed(err, TagType, TagLabelNames)

	case counterVecType:
		opts, err = mf.newCounterOpts()
		err = mf.checkTagNotAllowed(err, TagType)
		labelNames, err = mf.labelNames(err)

	case gaugeType:
		opts, err = mf.newGaugeOpts()
		err = mf.checkTagNotAllowed(err, TagType, TagLabelNames)

	case gaugeVecType:
		opts, err = mf.newGaugeOpts()
		err = mf.checkTagNotAllowed(err, TagType)
		labelNames, err = mf.labelNames(err)

	case histogramType:
		opts, err = mf.newHistogramOpts()
		err = mf.checkTagNotAllowed(err, TagType, TagLabelNames)
		err = mf.checkTagNotAllowed(err, summaryTagNames...)

	case histogramVecType:
		opts, err = mf.newHistogramOpts()
		err = mf.checkTagNotAllowed(err, TagType)
		err = mf.checkTagNotAllowed(err, summaryTagNames...)
		labelNames, err = mf.labelNames(err)

	case summaryType:
		opts, err = mf.newSummaryOpts()
		err = mf.checkTagNotAllowed(err, TagType, TagLabelNames)
		err = mf.checkTagNotAllowed(err, histogramTagNames...)

	case summaryVecType:
		opts, err = mf.newSummaryOpts()
		err = mf.checkTagNotAllowed(err, TagType)
		err = mf.checkTagNotAllowed(err, histogramTagNames...)
		labelNames, err = mf.labelNames(err)

	case observerType:
		opts, err = mf.newObserverOpts()
		err = mf.checkTagNotAllowed(err, TagLabelNames)

	case observerVecType:
		opts, err = mf.newObserverOpts()
		labelNames, err = mf.labelNames(err)
	}

	return
}

func (mf metricField) help() string {
	return mf.Tag.Get(TagHelp)
}

func (mf metricField) namespace() string {
	return mf.Tag.Get(TagNamespace)
}

func (mf metricField) subsystem() string {
	return mf.Tag.Get(TagSubsystem)
}

// buckets parses any TagBuckets field tag and returns the result.
func (mf metricField) buckets() (buckets []float64, err error) {
	tagValue := mf.Tag.Get(TagBuckets)
	if len(tagValue) == 0 {
		return
	}

	tagValues := strings.Split(tagValue, ",")
	buckets = make([]float64, len(tagValues))
	for i := 0; i < len(tagValues); i++ {
		var parseErr error
		buckets[i], parseErr = strconv.ParseFloat(strings.TrimSpace(tagValues[i]), 64)
		err = multierr.Append(err, parseErr)
	}

	return
}

// objectives parses any TagObjectives field tag and returns the result.
func (mf metricField) objectives(appendErr error) (objectives map[float64]float64, err error) {
	err = appendErr
	tagValue := mf.Tag.Get(TagObjectives)
	if len(tagValue) == 0 {
		return
	}

	tagValues := strings.Split(tagValue, ",")
	objectives = make(map[float64]float64, len(tagValues))
	for i := 0; i < len(tagValues); i++ {
		pair := strings.Split(tagValues[i], ":")
		if len(pair) != 2 {
			err = multierr.Append(err,
				mf.fieldErrorf("Invalid objective entry '%s'", tagValues[i]),
			)

			continue
		}

		key, parseErr := strconv.ParseFloat(strings.TrimSpace(pair[0]), 64)
		err = mf.appendError(err, parseErr)

		value, parseErr := strconv.ParseFloat(strings.TrimSpace(pair[1]), 64)
		err = mf.appendError(err, parseErr)

		if parseErr == nil {
			objectives[key] = value
		}
	}

	return
}

// parseUint32 parses a struct field tag that should parse as a uint32.
func (mf metricField) parseUint32(tagName string) (v uint32, err error) {
	tagValue := mf.Tag.Get(tagName)
	if len(tagValue) > 0 {
		var u64 uint64
		u64, err = strconv.ParseUint(tagValue, 10, 32)
		v = uint32(u64)
	}

	return
}

// ageBuckets returns the AgeBuckets for this metric.
func (mf metricField) ageBuckets(appendErr error) (uint32, error) {
	v, parseErr := mf.parseUint32(TagAgeBuckets)
	return v, mf.appendError(appendErr, parseErr)
}

// bufCap returns the BufCap for this metric.
func (mf metricField) bufCap(appendErr error) (uint32, error) {
	v, parseErr := mf.parseUint32(TagBufCap)
	return v, mf.appendError(appendErr, parseErr)
}

// parseDuration parses a struct field tag that should parse as a time.Duration.
func (mf metricField) parseDuration(tagName string) (v time.Duration, err error) {
	tagValue := mf.Tag.Get(tagName)
	if len(tagValue) > 0 {
		v, err = time.ParseDuration(tagValue)
	}

	return
}

// maxAge returns the MaxAge value for this metric.  If there is no TagMaxAge,
// this method returns time.Duration(0).
func (mf metricField) maxAge(appendErr error) (time.Duration, error) {
	v, err := mf.parseDuration(TagMaxAge)
	return v, mf.appendError(appendErr, err)
}

// checkInvalidTagNames examines the field for tags that are considered to be invalid
// for the metric type in question.  The format string is used to create the error
// message for each field, and it is expected to take a single %s argument which
// is the tag name that is invalid.
func (mf metricField) checkInvalidTagNames(appendErr error, format string, tagNames ...string) (err error) {
	err = appendErr
	for _, tagName := range tagNames {
		if _, ok := mf.Tag.Lookup(tagName); ok {
			err = multierr.Append(err,
				&FieldError{
					Field:   reflect.StructField(mf),
					Message: fmt.Sprintf(format, tagName),
				},
			)
		}
	}

	return
}

func (mf metricField) checkTagNotAllowed(appendErr error, tagNames ...string) error {
	return mf.checkInvalidTagNames(appendErr, "tag '%s' is not allowed", tagNames...)
}

func (mf metricField) checkTagAmbiguous(appendErr error, tagNames ...string) error {
	return mf.checkInvalidTagNames(appendErr, "tag '%s' is ambiguous", tagNames...)
}

// hasAnyTagNames tests if any of the given tag names are actually present on this field.
func (mf metricField) hasAnyTagNames(tagNames ...string) (ok bool) {
	for i := 0; !ok && i < len(tagNames); i++ {
		_, ok = mf.Tag.Lookup(tagNames[i])
	}

	return
}

// labelNames returns the labels described by this field.  An error is raised
// if there are no label names or the tag wasn't found.
func (mf metricField) labelNames(appendErr error) (values []string, err error) {
	err = appendErr
	v := mf.Tag.Get(TagLabelNames)
	values = strings.Split(v, ",")
	for i, ln := range values {
		values[i] = strings.TrimSpace(ln)
	}

	if len(values) == 1 && len(values[0]) == 0 {
		err = multierr.Append(err,
			mf.fieldErrorf("tag '%s' is required and cannot be empty for vector metrics", TagLabelNames),
		)
	}

	return
}

func (mf metricField) fieldErrorf(format string, args ...interface{}) *FieldError {
	return &FieldError{
		Field:   reflect.StructField(mf),
		Message: fmt.Sprintf(format, args...),
	}
}

func (mf metricField) wrapError(cause error) *FieldError {
	return &FieldError{
		Field:   reflect.StructField(mf),
		Cause:   cause,
		Message: cause.Error(),
	}
}

func (mf metricField) appendError(appendErr, cause error) error {
	if cause == nil {
		return appendErr
	}

	return multierr.Append(appendErr, mf.wrapError(cause))
}
