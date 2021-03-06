// MIT License
//
// Copyright (c) 2020 Nobody Night
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package santa

import (
	"context"
	"errors"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrInvalidType represents the type is invalid. This is usually
	// because the given type is invalid or unsupported.
	ErrInvalidType = errors.New("invalid type")

	// ErrClosed represents the instance has been closed. This is usually
	// because the application attempts to close an instance multiple times.
	ErrClosed = errors.New("instance has been closed")
)

// Logger is the structure of the logger instance.
//
// The logger is the foundation of all logger types. It provides simple
// and fast log entry message output API for applications, and supports
// all log entry messages that implement the Message interface. Any
// logger type should be implemented based on the logger.
//
// Under normal circumstances, applications should not directly use the
// logger. If you need to output any log entry messages that implement
// the Message interface, please use a standard logger.
//
// The API provided by the logger is thread-safe.
type Logger struct {
	name string
	level Level
	sampler Sampler
	hooks []Hook
	exporters []Exporter
	labels SerializedLabels

	addSource bool
}

// Output checks whether the log level is lower than the minimum log
// level of the logger. If it is higher than or equal to, a log entry
// of the given log level and message is generated. The generated log
// entries are then passed to the log entry sampler and one or more log
// entry hooks for processing, and finally passed to one or more log
// entry exporters for processing, and any errors encountered are
// returned.
//
// Please note that this is a low-level API, and the high-level API
// usually provided by the logger is used internally. Unless necessary,
// applications should not use this API directly.
func (l *Logger) Output(stacks int, level Level, message Message) error {
	if !l.level.Enabled(level) {
		return nil
	}
	if len(l.exporters) == 0 {
		return nil
	}

	entry := pool.Entry.New()
	entry.Name = l.name
	entry.Level = level
	entry.Time = time.Now()
	entry.Message = message
	entry.Labels = l.labels

	if l.sampler != nil && !l.sampler.Sample(entry) {
		pool.Entry.Free(entry)
		return nil
	}
	if l.addSource {
		entry.SourceLocation = newEntrySourceLocation(
			runtime.Caller(stacks))
	}

	for index := 0; index < len(l.hooks); index++ {
		err := l.hooks[index].Print(entry)

		if err != nil {
			pool.Entry.Free(entry)
			return err
		}
	}
	for index := 0; index < len(l.exporters); index++ {
		err := l.exporters[index].Export(entry)

		if err != nil {
			pool.Entry.Free(entry)
			return err
		}
	}

	pool.Entry.Free(entry)
	return nil
}

// Print outputs log entries for a given log level and message, and then
// returns any errors encountered.
func (l *Logger) Print(level Level, message Message) error {
	return l.Output(2, level, message)
}

// Option is a structure that contains options for the logger.
//
// Normally, all the logger option types of all logger types rely on the
// logger option to construct a basic logger instance. Unless necessary,
// applications should not directly use this option to build a logger
// instance.
type Option struct {
	// Name represents the name of each log entry output, usually used to
	// identify a component or resource. If not provided, the default
	// value is empty.
	Name string

	// Level represents the lowest level of log entries, and log entries
	// below the lowest level will be discarded. If not provided, the
	// default lowest level is DEBUG.
	Level Level

	// Sampler represents a log sampler. Each log entry to be output will
	// be passed to the log sampler, and the log sampler determines whether
	// a log entry should be output. If not provided, no log sampler is
	// used by default.
	//
	// For details, please refer to the comment section of the Sampler
	// interface.
	Sampler Sampler

	// Hooks represent a set of log entry hooks, and each log entry to be
	// output will be passed to each log entry hook so that the log entry
	// has the opportunity to process it before output. For example, one or
	// more log entry hooks can match each log entry and intercept the
	// output or perform other processing. If not provided, no log entry
	// hooks are used by default.
	//
	// For details, see the comment section of the Hook interface.
	//
	// Please note that this option slice will be reused during the build
	// process, and any side effects of external modifications are undefined.
	Hooks []Hook

	// Exporters represent a group of log entry exporters. Normally, each
	// exporter uses a specific encoder (such as a JSON encoder) to encode
	// log entries into specific data, and then uses a specific synchronizer
	// (such as a file synchronizer) to write the log entry data to a
	// specific storage device.
	//
	// Not only that, different exporter types may also use different
	// strategies to match each log entry to determine whether it needs to
	// be processed. For example, the standard exporter checks whether the
	// level of each log entry is included in a specific level span, and
	// only processes the included log entries.
	//
	// The advantage of using multiple exporters is that it allows the use
	// of different encoders and synchronizers for different log level
	// ranges. For details, see the comment section of the Exporter
	// interface.
	//
	// Please note that this option slice will be reused during the build
	// process, and any side effects of external modifications are undefined.
	Exporters []Exporter

	// Labels represents one or more labels related to the logger. Each label
	// is a pair of custom string keys, used to identify the attributes
	// associated with a log entry. These labels will be added to each log
	// entry to allow one or more labels to be matched when searching for a
	// set of log entries in the future.
	//
	// If not provided, no label will be added to any log entry by default.
	// For details, please refer to the annotation section of the Label
	// structure.
	Labels Labels

	// DisableSourceLocation represents whether it is necessary to obtain
	// and set the output API call source location for each log entry, so
	// that the application can track the source of each log entry. It is
	// worth noting that obtaining the source of log entries requires more
	// expensive performance overhead. If not provided, the default value
	// is false.
	DisableSourceLocation bool
}

// Build builds and returns an instance of the logger.
func (o *Option) Build() (*Logger, error) {
	return &Logger {
		name: o.Name,
		level: o.Level,
		sampler: o.Sampler,
		hooks: o.Hooks,
		exporters: o.Exporters,
		labels: NewSerializedLabels(o.Labels...),
		addSource: !o.DisableSourceLocation,
	}, nil
}

// NewOption creates and returns a logger option instance with default
// optional values.
func NewOption() *Option {
	return &Option {
		Level: LevelDebug,
		DisableSourceLocation: false,
	}
}

// New creates and returns a logger instance using the default optional
// values.
func New() (*Logger, error) {
	return NewOption().Build()
}

// StandardLogger is the structure of the standard logger instance.
//
// The standard logger is based on the logger. The standard logger
// provides a simple and fast multi-log level log output API for
// applications, and supports the output of any log messages that have
// implemented the Message interface.
//
// If the application does not need to output custom log message types,
// or does not need to trade convenience in exchange for higher log
// entry output performance, it is a good choice to use including but
// not limited to template loggers, structured loggers, etc., because
// they Provide an easier-to-use API.
//
// Please note that the standard logger defaults to enable the internal
// cache provided by the synchronizer to improve the output performance
// of log entries, but the side effect is that the time when some log
// entry data is actually written to a specific storage device will be
// delayed. If the application needs to write log entry data to a
// specific storage device in real time, disable the internal cache.
//
// Regardless of whether the internal cache is disabled or not, each
// logger needs to be explicitly closed after it is no longer in use,
// otherwise it may cause file handle leakage and loss of some log entry
// data. For details, please refer to the comment section of the Syncer
// interface.
//
// Unless explicitly stated, the API provided by the logger is
// thread-safe. It’s worth noting that APIs that allow post-build
// changes to logger instances are generally not thread-safe. If you
// need to change the logger instance (including but not limited to:
// minimum log entry level, etc.), use the Duplicate function to create
// a copy of the logger instance, and then make changes to the copy of
// the logger instance.
type StandardLogger struct {
	Logger

	context context.Context
	contextCancel context.CancelFunc
	contextWaitGroup *sync.WaitGroup
	contextReferences *int32

	closed int32
}

// Debug outputs a given log message with a log level of DEBUG, and then
// returns any errors encountered.
func (l *StandardLogger) Debug(message Message) error {
	return l.Output(2, LevelDebug, message)
}

// Info outputs a given log message with a log level of INFO, and then
// returns any errors encountered.
func (l *StandardLogger) Info(message Message) error {
	return l.Output(2, LevelInfo, message)
}

// Warning outputs a given log message with a log level of WARNING, and
// then returns any errors encountered.
func (l *StandardLogger) Warning(message Message) error {
	return l.Output(2, LevelWarning, message)
}

// Error outputs a given log message with a log level of ERROR, and then
// returns any errors encountered.
func (l *StandardLogger) Error(message Message) error {
	return l.Output(2, LevelError, message)
}

// Fatal outputs a given log message with a log level of FATAL, and then
// returns any errors encountered.
func (l *StandardLogger) Fatal(message Message) error {
	return l.Output(2, LevelFatal, message)
}

// Duplicate creates and returns a copy of the logger. If the logger is
// closed, it returns nil.
//
// Please note that the application must explicitly close each copy of
// the logger, otherwise the logger may be leaked.
func (l *StandardLogger) Duplicate() *StandardLogger {
	if atomic.AddInt32(l.contextReferences, 1) == 1 {
		// The logger has been shut down, and using the created copy
		// may cause panic.
		return nil
	}
	instance := *l
	return &instance
}

// SetName sets the log entry name to the given name. For details, please
// refer to the comment section of the Name field of the StandardOption
// structure.
//
// Please note that this API is not thread-safe.
func (l *StandardLogger) SetName(name string) {
	l.name = name
}

// SetLevel sets the lowest level of the log entry to the given level.
// For details, please refer to the comment section of the Level field of
// the StandardOption structure.
//
// Please note that this API is not thread-safe.
func (l *StandardLogger) SetLevel(level Level) {
	l.level = level
}

// SetSampler sets the sampler to the given sampler. For details, please
// refer to the comment section of the Sampler field of the Option
// structure.
//
// Please note that this API is not thread-safe.
func (l *StandardLogger) SetSampler(sampler Sampler) {
	l.sampler = sampler
}

// SetLabels sets the label to one or more given labels. For details,
// please refer to the comment section of the Labels field of the Option
// structure.
//
// It is worth noting that one or more labels previously set by the
// logger will be discarded because labels need to be pre-serialized.
//
// Please note that this API is not thread-safe.
func (l *StandardLogger) SetLabels(labels ...Label) {
	l.labels = NewSerializedLabels(labels...)
}

// AddHooks adds one or more hooks to the hook chain. For details,
// please refer to the comment section of the Hooks field of the Option
// option.
//
// Please note that this API is not thread-safe.
func (l *StandardLogger) AddHooks(hooks ...Hook) {
	l.hooks = append(l.hooks, hooks...)
}

// ResetHooks resets the hook chain, and the hooks that have been added
// will be removed. For details, please refer to the comment section of
// the Hooks field of the Option option.
//
// Please note that this API is not thread-safe.
func (l *StandardLogger) ResetHooks() {
	l.hooks = []Hook { }
}

// Sync writes the internal cache data of a specific synchronizer to a
// specific storage device. If the specific storage device is based on
// the file system, write the data cached by the file system to the
// persistent storage device. For details, please refer to the Sync
// function of the Syncer interface.
//
// Finally, any errors encountered are returned.
func (l *StandardLogger) Sync() error {
	for index := 0; index < len(l.exporters); index++ {
		err := l.exporters[index].Sync()

		if err != nil {
			return err
		}
	}
	return nil
}

// Close close all specific exporters, and then return any errors
// encountered. For details, please refer to the comment section of the
// Close function of the Exporter interface.
//
// If there are multiple copies of the logger, this function only reduces
// the reference count of the logger. If the logger's reference count is 0,
// it will actually be closed.
//
// Please note that this function is not guaranteed to succeed. If any
// errors are encountered, the state of the application may change. The
// best practice is to exit the application.
func (l *StandardLogger) Close() error {
	if !atomic.CompareAndSwapInt32(&l.closed, 0, 1) {
		return ErrClosed
	}
	references := atomic.AddInt32(l.contextReferences, -1)
	switch {
	case references > 0:
		// The other logger copy is using the logger and cannot be closed
		// now, otherwise it may cause panic.
		return nil
	case references < 0:
		// This is usually because the application tries to close the
		// logger repeatedly.
		return ErrClosed
	}
	l.contextCancel()
	l.contextWaitGroup.Wait()
	for index := 0; index < len(l.exporters); index++ {
		err := l.exporters[index].Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// IsClosed checks whether the logger instance has been closed.
func (l *StandardLogger) IsClosed() bool {
	return atomic.LoadInt32(&l.closed) == 1
}

// flushHandler calls the Sync function at a given time interval to
// automatically refresh the internal cache and file system cache until
// the context has been marked as complete and returns.
//
// This function should run in an independent coroutine context.
func (l *StandardLogger) flushHandler(interval time.Duration) {
	if interval < (time.Microsecond * 100) {
		// The interval must not be less than 100 milliseconds.
		interval = (time.Microsecond * 100)
	}
	defer l.contextWaitGroup.Done()
	for {
		select {
		case <-l.context.Done():
			return
		case <-time.After(interval):
			// Discard any errors encountered.
			_ = l.Sync()
		}
	}
}

const (
	// SamplerText represents the type of sampler as text sampler. For
	// details, please refer to the comment section of the TextSampler
	// structure.
	SamplerText = "text"
)

// SamplingOption is a structure that contains options for sampling log
// entries.
type SamplingOption struct {
	// Type represents the type of sampler, and its options are defined
	// by the constants beginning with Sampler... If the log entry message
	// does not implement the parsing interface of a specific sampler type,
	// the sampler may not work. If not provided, the default value is
	// the SamplerText constant.
	//
	// If the value of this option is empty, no valid log entry sampler
	// instance will be built.
	Type string

	// Option represents the value of the optional items of the sampler.
	// The actual data type of the value is determined by the value of the
	// option Type. The optional types of different sampler types are
	// different. If not provided, the default value is the default optional
	// value for the specific sampler type.
	Option interface { }
}

// UseText uses the text sampler (SamplerText constant) as the value of the
// option Type. For details, please refer to the comment section of the
// SamplerText constant. Then return to the option instance itself.
func (o *SamplingOption) UseText() *SamplingOption {
	o.Type = SamplerText
	o.Option = NewTextSamplerOption()
	return o
}

// UseTextOption uses the text sampler (SamplerText constant) as the value
// of the option Type, and then uses the value of the given option as the
// value of the option. If the value of the given option is nil, the default
// option is used. For details, please refer to the comment section of the
// SamplerText constant. Then return to the option instance itself.
func (o *SamplingOption) UseTextOption(option *TextSamplerOption) *SamplingOption {
	o.Type = SamplerText
	if option == nil {
		option = NewTextSamplerOption()
	}
	o.Option = option
	return o
}

// Build builds and returns a sampler instance.
func (o *SamplingOption) Build() (Sampler, error) {
	if len(o.Type) == 0 {
		return nil, nil
	}
	switch o.Type {
	case SamplerText:
		return o.Option.(*TextSamplerOption).Build()
	default:
		return nil, ErrInvalidType
	}
}

// NewSamplingOption creates and returns a sampling option instance with
// default optional values.
func NewSamplingOption() *SamplingOption {
	return &SamplingOption {
		Type: SamplerText,
		Option: NewTextSamplerOption(),
	}
}

const (
	// EncoderStandard represents that the type of encoder is a standard
	// encoder. For details, please refer to the comment section of the
	// StandardEncoder structure.
	EncoderStandard = "standard"

	// EncoderJSON represents that the type of encoder is a JSON
	// encoder. For details, please refer to the comment section of the
	// JSONEncoder structure.
	EncoderJSON = "json"
)

// EncodingOption is a structure that contains options for encoding log
// entries.
type EncodingOption struct {
	// Type represents the type of encoder, and its options are defined
	// by the constants beginning with Encoder... If the log entry message
	// does not implement the formatter interface of a specific encoder
	// type, the encoder may not work. If not provided, the default value
	// depends on the logger type.
	Type string

	// Option represents the value of the optional option of the encoder.
	// The actual data type of the value depends on the value of the option
	// Type. The optional types of different encoder types are different.
	// If not provided, the default value is the default optional value for
	// the specific encoder type.
	Option interface { }

	// DisableSourceLocation represents whether it is necessary to obtain
	// and set the output API call source location for each log entry, so
	// that the application can track the source of each log entry. It is
	// worth noting that obtaining the source of log entries requires more
	// expensive performance overhead. If not provided, the default value
	// is false.
	DisableSourceLocation bool
}

// UseStandard uses the standard encoder (EncoderStandard constant) as the
// value of option Type. For details, please refer to the comment section
// of the EncoderStandard constant. Then return to the option instance
// itself.
func (o *EncodingOption) UseStandard() *EncodingOption {
	o.Type = EncoderStandard
	o.Option = NewStandardEncoderOption()
	return o
}

// UseStandardOption uses the standard encoder (EncoderStandard constant)
// as the value of the option Type, and then uses the value of the given
// option as the value of the option. If the value of the given option is
// nil, the default option is used. For details, please refer to the
// comment section of the EncoderStandard constant. Then return to the
// option instance
func (o *EncodingOption) UseStandardOption(option *StandardEncoderOption) *EncodingOption {
	o.Type = EncoderStandard
	if option == nil {
		option = NewStandardEncoderOption()
	}
	o.Option = option
	return o
}

// UseJSON uses the standard encoder (EncoderJSON constant) as the
// value of option Type. For details, please refer to the comment section
// of the EncoderJSON constant. Then return to the option instance
// itself.
func (o *EncodingOption) UseJSON() *EncodingOption {
	o.Type = EncoderJSON
	o.Option = NewJSONEncoderOption()
	return o
}

// UseJSONOption uses the standard encoder (EncoderJSON constant)
// as the value of the option Type, and then uses the value of the given
// option as the value of the option. If the value of the given option is
// nil, the default option is used. For details, please refer to the
// comment section of the EncoderJSON constant. Then return to the option
// instance itself.
func (o *EncodingOption) UseJSONOption(option *JSONEncoderOption) *EncodingOption {
	o.Type = EncoderJSON
	if option == nil {
		option = NewJSONEncoderOption()
	}
	o.Option = option
	return o
}

// Build builds and returns a encoder instance.
func (o *EncodingOption) Build() (Encoder, error) {
	switch o.Type {
	case EncoderStandard:
		option := o.Option.(*StandardEncoderOption)
		option.EncodeSourceLocation = !o.DisableSourceLocation
		return option.Build()
	case EncoderJSON:
		option := o.Option.(*JSONEncoderOption)
		option.EncodeSourceLocation = !o.DisableSourceLocation
		return option.Build()
	default:
		return nil, ErrInvalidType
	}
}

// NewEncodingOption creates and returns a encoding option instance with
// default optional values.
func NewEncodingOption() *EncodingOption {
	return &EncodingOption {
		Type: EncoderStandard,
		Option: NewStandardEncoderOption(),
	}
}

const (
	// SyncerStandard represents that the type of synchronizer is a standard
	// synchronizer. For details, please refer to the notes section of
	// StandardSyncer structure.
	SyncerStandard = "standard"

	// SyncerFile represents that the type of synchronizer is a file
	// synchronizer. For details, please refer to the notes section of
	// FileSyncer structure.
	SyncerFile = "file"

	// SyncerNetwork represents the type of synchronizer is a network
	// synchronizer. For details, please refer to the notes section of the
	// NetworkSyncer structure.
	SyncerNetwork = "network"

	// SyncerDiscard represents that the type of synchronizer is a discard
	// synchronizer. For details, please refer to the notes section of
	// DiscardSyncer structure.
	SyncerDiscard = "discard"
)

// OutputtingOption is a structure that contains options for outputting
// log entries.
type OutputtingOption struct {
	// Type represents the type of synchronizer, and its optional options
	// are constants starting with Syncer... If not provided, the default
	// value is the SyncerDiscard constant.
	Type string

	// Option represents the value of the option of the synchronizer. The
	// data type of the value depends on the value of the option Type. The
	// option type of different synchronizer types is different. If not
	// provided, the default value is the default optional value for the
	// specific synchronizer type.
	Option interface { }

	// DisableCache represents whether to disable the internal cache
	// provided by the synchronizer. Using internal cache can significantly
	// improve the output performance of log entries, but it also brings
	// some side effects. For details, please refer to the notes section of
	// the Syncer interface. If not provided, the default value is false.
	DisableCache bool
}

// UseStandard uses the standard synchronizer (SyncerFile constant) as
// the value of the option Type. For details, please refer to the comment
// section of the SyncerFile constant. Then return to the option
// instance itself.
func (o *OutputtingOption) UseStandard(writer io.Writer) *OutputtingOption {
	o.Type = SyncerStandard
	o.Option = NewStandardSyncerOption().UseWriter(writer)
	return o
}

// UseFile uses the file synchronizer (SyncerFile constant) as
// the value of the option Type. For details, please refer to the comment
// section of the SyncerFile constant. Then return to the option
// instance itself.
func (o *OutputtingOption) UseFile(name string) *OutputtingOption {
	o.Type = SyncerFile
	o.Option = NewFileSyncerOption().UseName(name)
	return o
}

// UseNetwork uses the network synchronizer (SyncerNetwork constant) as the
// value of the option Type. For details, please refer to the comment section
// of the Type option and SyncerNetwork constant. Then return to the option
// instance itself.
//
// The optional value of the parameter protocol is defined by the constants
// at the beginning of Protocol..., and the format of the value of the
// parameter address depends on the value of the parameter protocol.
func (o *OutputtingOption) UseNetwork(protocol, address string) *OutputtingOption {
	o.Type = SyncerNetwork
	o.Option = NewNetworkSyncerOption().UseProtocol(protocol).UseAddress(address)
	return o
}

// UseDiscard uses the discard synchronizer (SyncerDiscard constant) as
// the value of the option Type. For details, please refer to the comment
// section of the SyncerDiscard constant. Then return to the option
// instance itself.
func (o *OutputtingOption) UseDiscard() *OutputtingOption {
	o.Type = SyncerDiscard
	return o
}

// Build builds and returns a syncer instance.
func (o *OutputtingOption) Build() (Syncer, error) {
	switch o.Type {
	case SyncerStandard:
		if o.DisableCache {
			o.Option.(*StandardSyncerOption).UseCacheCapacity(0)
		}
		return o.Option.(*StandardSyncerOption).Build()
	case SyncerFile:
		if o.DisableCache {
			o.Option.(*FileSyncerOption).UseCacheCapacity(0)
		}
		return o.Option.(*FileSyncerOption).Build()
	case SyncerNetwork:
		if o.DisableCache {
			o.Option.(*NetworkSyncerOption).UseCacheCapacity(0)
		}
		return o.Option.(*NetworkSyncerOption).Build()
	case SyncerDiscard:
		return NewDiscardSyncer()
	default:
		return nil, ErrInvalidType
	}
}

// NewOutputtingOption creates and returns a outputting option instance with
// default optional values.
func NewOutputtingOption() *OutputtingOption {
	return &OutputtingOption {
		Type: SyncerDiscard,
	}
}

// FlushingOption is a structure that contains options for automatic flushing
// of log entry data.
//
// Automatic flushing can refresh the data in the internal cache (if enabled)
// and the file system cache (if used) to the persistent storage device
// periodically and automatically to avoid untimely synchronization of the
// output log entry data or unexpected system interruption Part of the log
// entry data is lost.
//
// Normally, the value of the default option is a best practice, and the
// value of the default option maintains a balance between performance and
// data availability. If the throughput and data availability of log entry
// data are predictable, perhaps disabling automatic flushing is a good
// choice.
type FlushingOption struct {
	// Interval represents the interval time period of each automatic
	// flushing. The value of this option cannot be less than 100
	// milliseconds. If the value of this option is 0, it means that
	// automatic flushing is disabled. If not provided, the default
	// value is 1 second.
	//
	// If the system availability is high or the application does not
	// require high data availability, the value of this option can be
	// appropriately relaxed. Frequent flushing operations will cause
	// log entry data I/O throughput performance to decline. When
	// automatic flushing is performed, all log entry output operations
	// on the same log will be blocked.
	Interval time.Duration
}

// UseInterval uses the given interval as the value of the Interval option.
// For details, please refer to the comment section of the Interval option.
// Then return to the option instance itself.
func (o *FlushingOption) UseInterval(interval time.Duration) *FlushingOption {
	o.Interval = interval
	return o
}

// NewFlushingOption creates and returns an instance of a flushing option
// with default optional values.
func NewFlushingOption() *FlushingOption {
	return &FlushingOption {
		Interval: time.Second,
	}
}

// StandardOption is a structure that contains options for the standard
// logger.
type StandardOption struct {
	// Name represents the name of each log entry output, usually used to
	// identify a component or resource. If not provided, the default
	// value is empty.
	Name string

	// Level represents the lowest level of log entries, and log entries
	// below the lowest level will be discarded. If not provided, the
	// default lowest level is DEBUG.
	Level Level

	// Sampling represents the value of the log entry sampling options,
	// which contains the options related to log entry sampling. For
	// details, please refer to the comment section of the SamplingOption
	// structure. If not provided, no log sampling strategy is used by
	// default.
	Sampling SamplingOption

	// Encoding represents the value of the log entry encoding option,
	// which contains the options related to the log entry encoding. For
	// details, please refer to the comment section of the EncodingOption
	// structure. If not provided, the default value depends on the type
	// of logger.
	Encoding EncodingOption

	// Outputting represents the value of the log entry output option,
	// which contains the log entry output related options with the log
	// level from DEBUG to WARNING. For details, please refer to the
	// comment section of the OutputtingOption structure. If not provided,
	// the default output is to the standard output device (os.Stdout).
	Outputting OutputtingOption

	// ErrorOutputting represents the value of the log entry output
	// options, which contains the log entry output options from ERROR to
	// FATAL. For details, please refer to the comment section of the
	// OutputtingOption structure. If not provided, the default output is
	// to the standard error device (os.Stderr).
	ErrorOutputting OutputtingOption

	// Flushing represents the value of an option for automatic flushing
	// of log entry data. Automatic flushing can periodically flush the
	// internal cache (if enabled) and the data in the file system cache
	// to the persistent storage device. For details, see the comment
	// section of the FlushingOption structure. If not provided, the
	// default value depends on the type of logger.
	Flushing FlushingOption

	// Hooks represent a set of log entry hooks, and each log entry to be
	// output will be passed to each log entry hook so that the log entry
	// has the opportunity to process it before output. For example, one or
	// more log entry hooks can match each log entry and intercept the
	// output or perform other processing. If not provided, no log entry
	// hooks are used by default.
	//
	// For details, see the comment section of the Hook interface.
	//
	// Please note that this option slice will be reused during the build
	// process, and any side effects of external modifications are undefined.
	Hooks []Hook

	// Labels represents one or more labels related to the logger. Each label
	// is a pair of custom string keys, used to identify the attributes
	// associated with a log entry. These labels will be added to each log
	// entry to allow one or more labels to be matched when searching for a
	// set of log entries in the future.
	//
	// If not provided, no label will be added to any log entry by default.
	// For details, please refer to the annotation section of the Label
	// structure.
	Labels Labels
}

// UseName uses the given name as the value of the option Name. For details,
// please refer to the comment section of the Name option. Then return to
// the option instance itself.
func (o *StandardOption) UseName(name string) *StandardOption {
	o.Name = name
	return o
}

// UseLevel uses the given log level as the value of the option Level. For
// details, please refer to the comment section of the Level option. Then
// return to the option instance itself.
func (o *StandardOption) UseLevel(level Level) *StandardOption {
	o.Level = level
	return o
}

// UseHooks appends the given one or more hooks to the o.Hooks option slice,
// and then returns the option instance itself. For details, please refer to
// the comment section of the o.Hooks option.
func (o *StandardOption) UseHooks(hooks ...Hook) *StandardOption {
	o.Hooks = append(o.Hooks, hooks...)
	return o
}

// UseLabels appends the given one or more labels to the o.Labels option
// slice, and then returns the option instance itself. For details, please
// refer to the comment section of the o.Labels option.
func (o *StandardOption) UseLabels(labels ...Label) *StandardOption {
	o.Labels = append(o.Labels, labels...)
	return o
}

// UseSampling uses the given sampling option as the value of option Sampling.
// For details, please refer to the comment section of the Sampling option.
// Then return to the option instance itself.
func (o *StandardOption) UseSampling(option *SamplingOption) *StandardOption {
	o.Sampling = *option
	return o
}

// UseEncoding uses the given encoding option as the value of the option
// Encoding, please refer to the comment section of the Encoding option for
// details. Then return to the option instance itself.
func (o *StandardOption) UseEncoding(option *EncodingOption) *StandardOption {
	o.Encoding = *option
	return o
}

// UseOutputting uses the given output option as the value of option
// Outputting. For details, please refer to the comment section of Outputting
// option. Then return to the option instance itself.
func (o *StandardOption) UseOutputting(option *OutputtingOption) *StandardOption {
	o.Outputting = *option
	return o
}

// UseErrorOutputting uses the given output option as the value of option
// ErrorOutputting. For details, please refer to the comment section of
// ErrorOutputting option. Then return to the option instance itself.
func (o *StandardOption) UseErrorOutputting(option *OutputtingOption) *StandardOption {
	o.ErrorOutputting = *option
	return o
}

// UseFlushing Use the given flushing option as the value of the Flushing
// option. For details, see the comment section of the Flushing option. Then
// return to the option instance itself.
func (o *StandardOption) UseFlushing(option *FlushingOption) *StandardOption {
	o.Flushing = *option
	return o
}

// DisableCache disable the internal cache of output and error output. For
// details, please refer to the DisableCache option of the OutputtingOption
// structure. Then return to the option instance itself.
func (o *StandardOption) DisableCache() *StandardOption {
	o.Outputting.DisableCache = true
	o.ErrorOutputting.DisableCache = true
	return o
}

// DisableSampling disable sampling of log entries. For details, see the
// comment section of the Type option of the SamplingOption structure.
// Then return to the option instance itself.
func (o *StandardOption) DisableSampling() *StandardOption {
	o.Sampling = SamplingOption { }
	return o
}

// DisableFlushing Disables automatic flushing of cached log entry data.
// For details, see Flushing option. Then return to the option instance
// itself.
func (o *StandardOption) DisableFlushing() *StandardOption {
	o.Flushing.Interval = 0
	return o
}

// Build builds and returns a standard logger instance.
func (o *StandardOption) Build() (*StandardLogger, error) {
	sampler, err := o.Sampling.Build()
	if err != nil {
		return nil, err
	}
	encoder, err := o.Encoding.Build()
	if err != nil {
		return nil, err
	}
	syncer, err := o.Outputting.Build()
	if err != nil {
		return nil, err
	}
	exporter, err := NewStandardExporterOption().
		UseSpan(LevelDebug, LevelWarning).
		UseEncoder(encoder).
		UseSyncer(syncer).Build()
	if err != nil {
		_ = syncer.Close()
		return nil, err
	}
	errorSyncer, err := o.ErrorOutputting.Build()
	if err != nil {
		_ = exporter.Close()
		return nil, err
	}
	errorExporter, err := NewStandardExporterOption().
		UseSpan(LevelError, LevelFatal).
		UseEncoder(encoder).
		UseSyncer(errorSyncer).Build()
	if err != nil {
		_ = exporter.Close()
		_ = errorSyncer.Close()
		return nil, err
	}

	logger, err := (&Option {
		Name: o.Name,
		Level: o.Level,
		Sampler: sampler,
		Hooks: o.Hooks,
		Exporters: []Exporter {
			exporter,
			errorExporter,
		},
		Labels: o.Labels,
		DisableSourceLocation: (!encoder.Option().
			EncodeSourceLocation),
	}).Build()

	if err != nil {
		_ = exporter.Close()
		_ = errorSyncer.Close()
		return nil, err
	}

	context, contextCancel := context.WithCancel(
		context.Background())
	instance := &StandardLogger {
		Logger: *logger,

		context: context,
		contextCancel: contextCancel,
		contextWaitGroup: &sync.WaitGroup { },
		contextReferences: new(int32),
	}

	// Initialize the logger reference count to 1 to avoid
	// repeated close logger.
	atomic.AddInt32(instance.contextReferences, 1)

	if o.Flushing.Interval > 0 {
		instance.contextWaitGroup.Add(1)
		go instance.flushHandler(o.Flushing.Interval)
	}
	return instance, nil
}

// NewStandardOption creates and returns an instance of the standard logger
// option with default optional values.
func NewStandardOption() *StandardOption {
	return &StandardOption {
		Level: LevelDebug,
		Sampling: *NewSamplingOption(),
		Encoding: *NewEncodingOption(),
		Outputting: *NewOutputtingOption().UseStandard(os.Stdout),
		ErrorOutputting: *NewOutputtingOption().UseStandard(os.Stderr),
		Flushing: *NewFlushingOption(),
	}
}

// NewStandard creates and returns a standard logger instance using the
// default optional values.
func NewStandard() (*StandardLogger, error) {
	return NewStandardOption().Build()
}

// NewStandardBenchmark creates and returns an instance of a standard
// logger suitable for benchmark performance testing and any errors
// encountered.
func NewStandardBenchmark(sampling bool, encoder string) (*StandardLogger, error) {
	option := NewStandardOption()
	switch encoder {
	case EncoderStandard:
		option.Encoding.UseStandard()
	case EncoderJSON:
		option.Encoding.UseJSON()
	default:
		return nil, ErrInvalidType
	}
	option.Encoding.DisableSourceLocation = true
	option.Flushing.Interval = 0
	option.Outputting.UseDiscard()
	option.ErrorOutputting.UseDiscard()
	option.UseLevel(LevelDebug)
	if !sampling {
		option.DisableSampling()
	}
	return option.Build()
}
