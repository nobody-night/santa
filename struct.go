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

import "sync/atomic"

// StructLogger is the structure of a structured logger instance.
//
// The structured logger is based on the standard logger. Structured Logger
// provides simple, fast and multi-log level structured log message API for
// applications. The structured logger allows adding a piece of text and
// one or more fields to the message of each log entry. The text is used to
// describe related events or behaviors, and one or more fields are used to
// provide metadata associated with related events or behaviors.
//
// For example, when an online server accepts a new connection, it outputs
// a log entry message with the description text "Connection has been
// accepted.", the value of the field "address" is "1.1.1.1" and the value
// of the field "port" is 54321.
//
// Please note that the structured logger defaults to enable the internal
// cache provided by the synchronizer to improve the output performance
// of log entries, but the side effect is that the time when some log entry
// data is actually written to a specific storage device will be delayed.
// If the application needs to write log entry data to a specific storage
// device in real time, disable the internal cache.
//
// Regardless of whether the internal cache is disabled or not, each logger
// needs to be explicitly closed after it is no longer in use, otherwise
// it may cause file handle leakage and loss of some log entry data. For
// details, please refer to the comment section of the Syncer interface.
//
// Unless explicitly stated, the API provided by the logger is
// thread-safe. It’s worth noting that APIs that allow post-build
// changes to logger instances are generally not thread-safe. If you
// need to change the logger instance (including but not limited to:
// minimum log entry level, etc.), use the Duplicate function to create
// a copy of the logger instance, and then make changes to the copy of
// the logger instance.
type StructLogger struct {
	StandardLogger
}

// Prints outputs a structured log message with a given log level,
// given description text and fields, and then returns any errors
// encountered.
func (l *StructLogger) Prints(level Level, text string, fields ...Field) error {
	message := pool.Message.Structure.New(text, fields)
	err := l.Output(2, level, message)
	pool.Message.Structure.Free(message)
	return err
}

// Debugs outputs a structured log message with a log level of DEBUG,
// given description text and fields, and then returns any errors
// encountered.
func (l *StructLogger) Debugs(text string, fields ...Field) error {
	message := pool.Message.Structure.New(text, fields)
	err := l.Output(2, LevelDebug, message)
	pool.Message.Structure.Free(message)
	return err
}

// Infos outputs a structured log message with a log level of INFO,
// given description text and fields, and then returns any errors
// encountered.
func (l *StructLogger) Infos(text string, fields ...Field) error {
	message := pool.Message.Structure.New(text, fields)
	err := l.Output(2, LevelInfo, message)
	pool.Message.Structure.Free(message)
	return err
}

// Warnings outputs a structured log message with a log level of WARNING,
// given description text and fields, and then returns any errors
// encountered.
func (l *StructLogger) Warnings(text string, fields ...Field) error {
	message := pool.Message.Structure.New(text, fields)
	err := l.Output(2, LevelWarning, message)
	pool.Message.Structure.Free(message)
	return err
}

// Errors outputs a structured log message with a log level of ERROR,
// given description text and fields, and then returns any errors
// encountered.
func (l *StructLogger) Errors(text string, fields ...Field) error {
	message := pool.Message.Structure.New(text, fields)
	err := l.Output(2, LevelError, message)
	pool.Message.Structure.Free(message)
	return err
}

// Fatals outputs a structured log message with a log level of FATAL,
// given description text and fields, and then returns any errors
// encountered.
func (l *StructLogger) Fatals(text string, fields ...Field) error {
	message := pool.Message.Structure.New(text, fields)
	err := l.Output(2, LevelFatal, message)
	pool.Message.Structure.Free(message)
	return err
}

// Duplicate creates and returns a copy of the logger. If the logger is
// closed, it returns nil.
//
// Please note that the application must explicitly close each copy of
// the logger, otherwise the logger may be leaked.
func (l *StructLogger) Duplicate() *StructLogger {
	if atomic.AddInt32(l.contextReferences, 1) == 1 {
		// The logger has been shut down, and using the created copy
		// may cause panic.
		return nil
	}
	instance := *l
	return &instance
}

// StructOption is a structure that contains options for structured
// loggers.
type StructOption struct {
	StandardOption
}

// UseName uses the given name as the value of the option Name. For details,
// please refer to the comment section of the Name option. Then return to
// the option instance itself.
func (o *StructOption) UseName(name string) *StructOption {
	o.Name = name
	return o
}

// UseLevel uses the given log level as the value of the option Level. For
// details, please refer to the comment section of the Level option. Then
// return to the option instance itself.
func (o *StructOption) UseLevel(level Level) *StructOption {
	o.Level = level
	return o
}

// UseHooks appends the given one or more hooks to the o.Hooks option slice,
// and then returns the option instance itself. For details, please refer to
// the comment section of the o.Hooks option.
func (o *StructOption) UseHooks(hooks ...Hook) *StructOption {
	o.Hooks = append(o.Hooks, hooks...)
	return o
}

// UseLabels appends the given one or more labels to the o.Labels option
// slice, and then returns the option instance itself. For details, please
// refer to the comment section of the o.Labels option.
func (o *StructOption) UseLabels(labels ...Label) *StructOption {
	o.Labels = append(o.Labels, labels...)
	return o
}

// UseSampling uses the given sampling option as the value of option Sampling.
// For details, please refer to the comment section of the Sampling option.
// Then return to the option instance itself.
func (o *StructOption) UseSampling(option *SamplingOption) *StructOption {
	o.Sampling = *option
	return o
}

// UseEncoding uses the given encoding option as the value of the option
// Encoding, please refer to the comment section of the Encoding option for
// details. Then return to the option instance itself.
func (o *StructOption) UseEncoding(option *EncodingOption) *StructOption {
	o.Encoding = *option
	return o
}

// UseOutputting uses the given output option as the value of option
// Outputting. For details, please refer to the comment section of Outputting
// option. Then return to the option instance itself.
func (o *StructOption) UseOutputting(option *OutputtingOption) *StructOption {
	o.Outputting = *option
	return o
}

// UseErrorOutputting uses the given output option as the value of option
// ErrorOutputting. For details, please refer to the comment section of
// ErrorOutputting option. Then return to the option instance itself.
func (o *StructOption) UseErrorOutputting(option *OutputtingOption) *StructOption {
	o.ErrorOutputting = *option
	return o
}

// UseFlushing Use the given flushing option as the value of the Flushing
// option. For details, see the comment section of the Flushing option. Then
// return to the option instance itself.
func (o *StructOption) UseFlushing(option *FlushingOption) *StructOption {
	o.Flushing = *option
	return o
}

// DisableCache Disable the internal cache of output and error output. For
// details, please refer to the DisableCache option of the OutputtingOption
// structure. Then return to the option instance itself.
func (o *StructOption) DisableCache() *StructOption {
	o.Outputting.DisableCache = true
	o.ErrorOutputting.DisableCache = true
	return o
}

// DisableSampling disable sampling of log entries. For details, see the
// comment section of the Type option of the SamplingOption structure.
// Then return to the option instance itself.
func (o *StructOption) DisableSampling() *StructOption {
	o.Sampling = SamplingOption { }
	return o
}

// DisableFlushing Disables automatic flushing of cached log entry data.
// For details, see Flushing option. Then return to the option instance
// itself.
func (o *StructOption) DisableFlushing() *StructOption {
	o.Flushing.Interval = 0
	return o
}

// Build builds and returns a structured logger instance.
func (o *StructOption) Build() (*StructLogger, error) {
	logger, err := o.StandardOption.Build()
	if err != nil {
		return nil, err
	}
	return &StructLogger {
		StandardLogger: *logger,
	}, nil
}

// NewStructOption creates an instance of a structured logger option with
// default optional values.
func NewStructOption() *StructOption {
	return &StructOption {
		StandardOption: *NewStandardOption().
			UseEncoding(NewEncodingOption().
				UseJSON()),
	}
}

// NewStruct creates and returns a structured logger instance using default
// optional values.
func NewStruct() (*StructLogger, error) {
	return NewStructOption().Build()
}

// NewStructBenchmark creates and returns an instance of a structured logger
// suitable for benchmark performance testing and any errors encountered.
func NewStructBenchmark(sampling bool, encoder string) (*StructLogger, error) {
	option := NewStructOption()
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
