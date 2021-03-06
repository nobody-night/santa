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

// TemplateLogger is the structure of the template logger instance.
//
// The template logger is based on the standard logger. Template Logger
// provides simple, easy-to-use and multi-log level template log message
// API for applications. The template logger allows to specify a formatting
// template string and one or more parameters for each log entry message,
// just like using fmt.Sprintf to format a log message string and output,
// it is very easy to use.
//
// If the application needs better log entry output performance, using a
// structured logger or a standard logger is a good choice.
//
// Please note that the template logger defaults to enable the internal
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
type TemplateLogger struct {
	StandardLogger
}

// Printf outputs a template log message with a given log level, a given
// template string and one or more parameters, and then returns any errors
// encountered.
func (l *TemplateLogger) Printf(level Level, template string, args ...interface { }) error {
	message := pool.Message.Template.New(template, args)
	err := l.Output(2, level, message)
	pool.Message.Template.Free(message)
	return err
}

// Debugf outputs a template log message with a log level of DEBUG, a given
// template string and one or more parameters, and then returns any errors
// encountered.
func (l *TemplateLogger) Debugf(template string, args ...interface { }) error {
	message := pool.Message.Template.New(template, args)
	err := l.Output(2, LevelDebug, message)
	pool.Message.Template.Free(message)
	return err
}

// Infof outputs a template log message with a log level of INFO, a given
// template string and one or more parameters, and then returns any errors
// encountered.
func (l *TemplateLogger) Infof(template string, args ...interface { }) error {
	message := pool.Message.Template.New(template, args)
	err := l.Output(2, LevelInfo, message)
	pool.Message.Template.Free(message)
	return err
}

// Warningf outputs a template log message with a log level of WARNING, a
// given template string and one or more parameters, and then returns any
// errors encountered.
func (l *TemplateLogger) Warningf(template string, args ...interface { }) error {
	message := pool.Message.Template.New(template, args)
	err := l.Output(2, LevelWarning, message)
	pool.Message.Template.Free(message)
	return err
}

// Errorf outputs a template log message with a log level of ERROR, a given
// template string and one or more parameters, and then returns any errors
// encountered.
func (l *TemplateLogger) Errorf(template string, args ...interface { }) error {
	message := pool.Message.Template.New(template, args)
	err := l.Output(2, LevelError, message)
	pool.Message.Template.Free(message)
	return err
}

// Fatalf outputs a template log message with a log level of FATAL, a given
// template string and one or more parameters, and then returns any errors
// encountered.
func (l *TemplateLogger) Fatalf(template string, args ...interface { }) error {
	message := pool.Message.Template.New(template, args)
	err := l.Output(2, LevelFatal, message)
	pool.Message.Template.Free(message)
	return err
}

// Duplicate creates and returns a copy of the logger. If the logger is
// closed, it returns nil.
//
// Please note that the application must explicitly close each copy of
// the logger, otherwise the logger may be leaked.
func (l *TemplateLogger) Duplicate() *TemplateLogger {
	if atomic.AddInt32(l.contextReferences, 1) == 1 {
		// The logger has been shut down, and using the created copy
		// may cause panic.
		return nil
	}
	instance := *l
	return &instance
}

// TemplateOption is a structure that contains options for the template
// logger.
type TemplateOption struct {
	StandardOption
}

// UseName uses the given name as the value of the option Name. For details,
// please refer to the comment section of the Name option. Then return to
// the option instance itself.
func (o *TemplateOption) UseName(name string) *TemplateOption {
	o.Name = name
	return o
}

// UseLevel uses the given log level as the value of the option Level. For
// details, please refer to the comment section of the Level option. Then
// return to the option instance itself.
func (o *TemplateOption) UseLevel(level Level) *TemplateOption {
	o.Level = level
	return o
}

// UseHook appends the given Hook value to the Hook option slice. For details,
// see the comment section of the Hook option. Then return to the option
// instance itself.
func (o *TemplateOption) UseHook(hook Hook) *TemplateOption {
	o.Hooks = append(o.Hooks, hook)
	return o
}

// UseHooks appends the given one or more hooks to the o.Hooks option slice,
// and then returns the option instance itself. For details, please refer to
// the comment section of the o.Hooks option.
func (o *TemplateOption) UseHooks(hooks ...Hook) *TemplateOption {
	o.Hooks = append(o.Hooks, hooks...)
	return o
}

// UseLabels appends the given one or more labels to the o.Labels option
// slice, and then returns the option instance itself. For details, please
// refer to the comment section of the o.Labels option.
func (o *TemplateOption) UseLabels(labels ...Label) *TemplateOption {
	o.Labels = append(o.Labels, labels...)
	return o
}

// UseEncoding uses the given encoding option as the value of the option
// Encoding, please refer to the comment section of the Encoding option for
// details. Then return to the option instance itself.
func (o *TemplateOption) UseEncoding(option *EncodingOption) *TemplateOption {
	o.Encoding = *option
	return o
}

// UseOutputting uses the given output option as the value of option
// Outputting. For details, please refer to the comment section of Outputting
// option. Then return to the option instance itself.
func (o *TemplateOption) UseOutputting(option *OutputtingOption) *TemplateOption {
	o.Outputting = *option
	return o
}

// UseErrorOutputting uses the given output option as the value of option
// ErrorOutputting. For details, please refer to the comment section of
// ErrorOutputting option. Then return to the option instance itself.
func (o *TemplateOption) UseErrorOutputting(option *OutputtingOption) *TemplateOption {
	o.ErrorOutputting = *option
	return o
}

// UseFlushing Use the given flushing option as the value of the Flushing
// option. For details, see the comment section of the Flushing option. Then
// return to the option instance itself.
func (o *TemplateOption) UseFlushing(option *FlushingOption) *TemplateOption {
	o.Flushing = *option
	return o
}

// DisableCache disable the internal cache of output and error output. For
// details, please refer to the DisableCache option of the OutputtingOption
// structure. Then return to the option instance itself.
func (o *TemplateOption) DisableCache() *TemplateOption {
	o.Outputting.DisableCache = true
	o.ErrorOutputting.DisableCache = true
	return o
}

// DisableSampling disable sampling of log entries. For details, see the
// comment section of the Type option of the SamplingOption structure.
// Then return to the option instance itself.
func (o *TemplateOption) DisableSampling() *TemplateOption {
	o.Sampling = SamplingOption { }
	return o
}

// DisableFlushing Disables automatic flushing of cached log entry data.
// For details, see Flushing option. Then return to the option instance
// itself.
func (o *TemplateOption) DisableFlushing() *TemplateOption {
	o.Flushing.Interval = 0
	return o
}

// Build builds and returns a template logger instance.
func (o *TemplateOption) Build() (*TemplateLogger, error) {
	logger, err := o.StandardOption.Build()
	if err != nil {
		return nil, err
	}
	return &TemplateLogger {
		StandardLogger: *logger,
	}, nil
}

// NewTemplateOption creates an instance of a template logger option with
// default optional values.
func NewTemplateOption() *TemplateOption {
	return &TemplateOption {
		StandardOption: *NewStandardOption(),
	}
}

// NewTemplate creates and returns a template logger instance using default
// optional values.
func NewTemplate() (*TemplateLogger, error) {
	return NewTemplateOption().Build()
}

// NewTemplateBenchmark creates and returns an instance of a template logger
// suitable for benchmark performance testing and any errors encountered.
func NewTemplateBenchmark(sampling bool, encoder string) (*TemplateLogger, error) {
	option := NewTemplateOption()
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
