// Copyright 2016-2017 Apcera Inc. All rights reserved.

package server

import (
	"flag"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/gnatsd/conf"
	natsd "github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats-streaming-server/stores"
	"github.com/nats-io/nats-streaming-server/util"
)

// ProcessConfigFile parses the configuration file `configFile` and updates
// the given Streaming options `opts`.
func ProcessConfigFile(configFile string, opts *Options) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	m, err := conf.Parse(string(data))
	if err != nil {
		return err
	}
	// Look for a "streaming" key. If so, use only the content of this
	// map, otherwise, use all keys.
	for k, v := range m {
		name := strings.ToLower(k)
		if name == "streaming" {
			content, ok := v.(map[string]interface{})
			if !ok {
				return fmt.Errorf("expected streaming section to be a map/struct, got %v", v)
			}
			// Override `m` with the content of the streaming map.
			m = content
		}
	}
	for k, v := range m {
		name := strings.ToLower(k)
		switch name {
		case "id", "cid", "cluster_id":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			opts.ID = v.(string)
		case "discover_prefix":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			opts.DiscoverPrefix = v.(string)
		case "st", "store_type", "store", "storetype":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			switch strings.ToUpper(v.(string)) {
			case stores.TypeFile:
				opts.StoreType = stores.TypeFile
			case stores.TypeMemory:
				opts.StoreType = stores.TypeMemory
			default:
				return fmt.Errorf("unknown store type: %v", v.(string))
			}
		case "dir", "datastore":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			opts.FilestoreDir = v.(string)
		case "sd", "stan_debug":
			if err := checkType(k, reflect.Bool, v); err != nil {
				return err
			}
			opts.Debug = v.(bool)
		case "sv", "stan_trace":
			if err := checkType(k, reflect.Bool, v); err != nil {
				return err
			}
			opts.Trace = v.(bool)
		case "ns", "nats_server", "nats_server_url":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			opts.NATSServerURL = v.(string)
		case "secure":
			if err := checkType(k, reflect.Bool, v); err != nil {
				return err
			}
			opts.Secure = v.(bool)
		case "tls":
			if err := parseTLS(v, opts); err != nil {
				return err
			}
		case "limits", "store_limits", "storelimits":
			if err := parseStoreLimits(v, opts); err != nil {
				return err
			}
		case "file", "file_options":
			if err := parseFileOptions(v, opts); err != nil {
				return err
			}
		case "hbi", "hb_interval", "server_to_client_hb_interval":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			dur, err := time.ParseDuration(v.(string))
			if err != nil {
				return err
			}
			opts.ClientHBInterval = dur
		case "hbt", "hb_timeout", "server_to_client_hb_timeout":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			dur, err := time.ParseDuration(v.(string))
			if err != nil {
				return err
			}
			opts.ClientHBTimeout = dur
		case "hbf", "hb_fail_count", "server_to_client_hb_fail_count":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.ClientHBFailCount = int(v.(int64))
		case "ack_subs_pool_size", "ack_subscriptions_pool_size":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.AckSubsPoolSize = int(v.(int64))
		case "ft_group", "ft_group_name":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			opts.FTGroupName = v.(string)
		case "partitioning":
			if err := checkType(k, reflect.Bool, v); err != nil {
				return err
			}
			opts.Partitioning = v.(bool)
		}
	}
	return nil
}

// checkType returns a formatted error if `v` is not of the expected kind.
func checkType(name string, kind reflect.Kind, v interface{}) error {
	actualKind := reflect.TypeOf(v).Kind()
	if actualKind != kind {
		return fmt.Errorf("parameter %q value is expected to be %v, got %v",
			name, kind.String(), actualKind.String())
	}
	return nil
}

// parseTLS updates `opts` with TLS config
func parseTLS(itf interface{}, opts *Options) error {
	m, ok := itf.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected TLS to be a map/struct, got %v", itf)
	}
	for k, v := range m {
		name := strings.ToLower(k)
		switch name {
		case "client_cert":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			opts.ClientCert = v.(string)
		case "client_key":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			opts.ClientKey = v.(string)
		case "client_ca", "client_cacert":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			opts.ClientCA = v.(string)
		}
	}
	return nil
}

// parseStoreLimits updates `opts` with store limits
func parseStoreLimits(itf interface{}, opts *Options) error {
	m, ok := itf.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected store limits to be a map/struct, got %v", itf)
	}
	for k, v := range m {
		name := strings.ToLower(k)
		switch name {
		case "mc", "max_channels", "maxchannels":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.MaxChannels = int(v.(int64))
		case "channels", "channels_limits", "channelslimits", "per_channel", "per_channel_limits":
			if err := parsePerChannelLimits(v, opts); err != nil {
				return err
			}
		default:
			// Check for the global limits (MaxMsgs, MaxBytes, etc..)
			if err := parseChannelLimits(&opts.ChannelLimits, k, name, v, true); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseChannelLimits updates `cl` with channel limits.
func parseChannelLimits(cl *stores.ChannelLimits, k, name string, v interface{}, isGlobal bool) error {
	switch name {
	case "msu", "max_subs", "max_subscriptions", "maxsubscriptions":
		if err := checkType(k, reflect.Int64, v); err != nil {
			return err
		}
		cl.MaxSubscriptions = int(v.(int64))
		if !isGlobal && cl.MaxSubscriptions == 0 {
			cl.MaxSubscriptions = -1
		}
	case "mm", "max_msgs", "maxmsgs", "max_count", "maxcount":
		if err := checkType(k, reflect.Int64, v); err != nil {
			return err
		}
		cl.MaxMsgs = int(v.(int64))
		if !isGlobal && cl.MaxMsgs == 0 {
			cl.MaxMsgs = -1
		}
	case "mb", "max_bytes", "maxbytes":
		if err := checkType(k, reflect.Int64, v); err != nil {
			return err
		}
		cl.MaxBytes = v.(int64)
		if !isGlobal && cl.MaxBytes == 0 {
			cl.MaxBytes = -1
		}
	case "ma", "max_age", "maxage":
		if err := checkType(k, reflect.String, v); err != nil {
			return err
		}
		dur, err := time.ParseDuration(v.(string))
		if err != nil {
			return err
		}
		cl.MaxAge = dur
		if !isGlobal && cl.MaxAge == 0 {
			cl.MaxAge = -1
		}
	}
	return nil
}

// parsePerChannelLimits updates `opts` with per channel limits.
func parsePerChannelLimits(itf interface{}, opts *Options) error {
	m, ok := itf.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected per channel limits to be a map/struct, got %v", itf)
	}
	for channelName, limits := range m {
		limitsMap, ok := limits.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected channel limits to be a map/struct, got %v", limits)
		}
		if !util.IsSubjectValid(channelName, true) {
			return fmt.Errorf("invalid channel name %q", channelName)
		}
		cl := &stores.ChannelLimits{}
		for k, v := range limitsMap {
			name := strings.ToLower(k)
			if err := parseChannelLimits(cl, k, name, v, false); err != nil {
				return err
			}
		}
		sl := &opts.StoreLimits
		sl.AddPerChannel(channelName, cl)
	}
	return nil
}

func parseFileOptions(itf interface{}, opts *Options) error {
	m, ok := itf.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected file options to be a map/struct, got %v", itf)
	}
	for k, v := range m {
		name := strings.ToLower(k)
		switch name {
		case "compact", "compact_enabled":
			if err := checkType(k, reflect.Bool, v); err != nil {
				return err
			}
			opts.FileStoreOpts.CompactEnabled = v.(bool)
		case "compact_frag", "compact_fragmentation":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.FileStoreOpts.CompactFragmentation = int(v.(int64))
		case "compact_interval":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.FileStoreOpts.CompactInterval = int(v.(int64))
		case "compact_min_size":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.FileStoreOpts.CompactMinFileSize = v.(int64)
		case "buffer_size":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.FileStoreOpts.BufferSize = int(v.(int64))
		case "crc", "do_crc":
			if err := checkType(k, reflect.Bool, v); err != nil {
				return err
			}
			opts.FileStoreOpts.DoCRC = v.(bool)
		case "crc_poly":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.FileStoreOpts.CRCPolynomial = v.(int64)
		case "sync", "do_sync", "sync_on_flush":
			if err := checkType(k, reflect.Bool, v); err != nil {
				return err
			}
			opts.FileStoreOpts.DoSync = v.(bool)
		case "slice_max_msgs", "slice_max_count", "slice_msgs", "slice_count":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.FileStoreOpts.SliceMaxMsgs = int(v.(int64))
		case "slice_max_bytes", "slice_max_size", "slice_bytes", "slice_size":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.FileStoreOpts.SliceMaxBytes = v.(int64)
		case "slice_max_age", "slice_age", "slice_max_time", "slice_time_limit":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			dur, err := time.ParseDuration(v.(string))
			if err != nil {
				return err
			}
			opts.FileStoreOpts.SliceMaxAge = dur
		case "slice_archive_script", "slice_archive", "slice_script":
			if err := checkType(k, reflect.String, v); err != nil {
				return err
			}
			opts.FileStoreOpts.SliceArchiveScript = v.(string)
		case "file_descriptors_limit", "fds_limit":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.FileStoreOpts.FileDescriptorsLimit = v.(int64)
		case "parallel_recovery":
			if err := checkType(k, reflect.Int64, v); err != nil {
				return err
			}
			opts.FileStoreOpts.ParallelRecovery = int(v.(int64))
		}
	}
	return nil
}

// ConfigureOptions accepts a flag set and augment it with NATS Streaming Server
// specific flags. It then invokes the corresponding function from NATS Server.
// On success, Streaming and NATS options structures are returned configured
// based on the selected flags and/or configuration files.
// The command line options take precedence to the ones in the configuration files.
func ConfigureOptions(fs *flag.FlagSet, args []string, printVersion, printHelp, printTLSHelp func()) (*Options, *natsd.Options, error) {
	sopts := GetDefaultOptions()

	var (
		stanConfigFile string
		natsConfigFile string
	)

	fs.StringVar(&sopts.ID, "cluster_id", DefaultClusterID, "stan.ID")
	fs.StringVar(&sopts.ID, "cid", DefaultClusterID, "stan.ID")
	fs.StringVar(&sopts.StoreType, "store", stores.TypeMemory, "stan.StoreType")
	fs.StringVar(&sopts.StoreType, "st", stores.TypeMemory, "stan.StoreType")
	fs.StringVar(&sopts.FilestoreDir, "dir", "", "stan.FilestoreDir")
	fs.IntVar(&sopts.MaxChannels, "max_channels", stores.DefaultStoreLimits.MaxChannels, "stan.MaxChannels")
	fs.IntVar(&sopts.MaxChannels, "mc", stores.DefaultStoreLimits.MaxChannels, "stan.MaxChannels")
	fs.IntVar(&sopts.MaxSubscriptions, "max_subs", stores.DefaultStoreLimits.MaxSubscriptions, "stan.MaxSubscriptions")
	fs.IntVar(&sopts.MaxSubscriptions, "msu", stores.DefaultStoreLimits.MaxSubscriptions, "stan.MaxSubscriptions")
	fs.IntVar(&sopts.MaxMsgs, "max_msgs", stores.DefaultStoreLimits.MaxMsgs, "stan.MaxMsgs")
	fs.IntVar(&sopts.MaxMsgs, "mm", stores.DefaultStoreLimits.MaxMsgs, "stan.MaxMsgs")
	fs.String("max_bytes", fmt.Sprintf("%v", stores.DefaultStoreLimits.MaxBytes), "stan.MaxBytes")
	fs.String("mb", fmt.Sprintf("%v", stores.DefaultStoreLimits.MaxBytes), "stan.MaxBytes")
	fs.DurationVar(&sopts.MaxAge, "max_age", stores.DefaultStoreLimits.MaxAge, "stan.MaxAge")
	fs.DurationVar(&sopts.MaxAge, "ma", stores.DefaultStoreLimits.MaxAge, "stan.MaxAge")
	fs.DurationVar(&sopts.ClientHBInterval, "hbi", DefaultHeartBeatInterval, "stan.ClientHBInterval")
	fs.DurationVar(&sopts.ClientHBInterval, "hb_interval", DefaultHeartBeatInterval, "stan.ClientHBInterval")
	fs.DurationVar(&sopts.ClientHBTimeout, "hbt", DefaultClientHBTimeout, "stan.ClientHBTimeout")
	fs.DurationVar(&sopts.ClientHBTimeout, "hb_timeout", DefaultClientHBTimeout, "stan.ClientHBTimeout")
	fs.IntVar(&sopts.ClientHBFailCount, "hbf", DefaultMaxFailedHeartBeats, "stan.ClientHBFailCount")
	fs.IntVar(&sopts.ClientHBFailCount, "hb_fail_count", DefaultMaxFailedHeartBeats, "stan.ClientHBFailCount")
	fs.BoolVar(&sopts.Debug, "SD", false, "stan.Debug")
	fs.BoolVar(&sopts.Debug, "stan_debug", false, "stan.Debug")
	fs.BoolVar(&sopts.Trace, "SV", false, "stan.Trace")
	fs.BoolVar(&sopts.Trace, "stan_trace", false, "stan.Trace")
	fs.Bool("SDV", false, "")
	fs.BoolVar(&sopts.Secure, "secure", false, "stan.Secure")
	fs.StringVar(&sopts.ClientCert, "tls_client_cert", "", "stan.ClientCert")
	fs.StringVar(&sopts.ClientKey, "tls_client_key", "", "stan.ClientKey")
	fs.StringVar(&sopts.ClientCA, "tls_client_cacert", "", "stan.ClientCA")
	fs.StringVar(&sopts.NATSServerURL, "nats_server", "", "stan.NATSServerURL")
	fs.StringVar(&sopts.NATSServerURL, "ns", "", "stan.NATSServerURL")
	fs.StringVar(&stanConfigFile, "sc", "", "")
	fs.StringVar(&stanConfigFile, "stan_config", "", "")
	fs.IntVar(&sopts.AckSubsPoolSize, "ack_subs", 0, "stan.AckSubsPoolSize")
	fs.BoolVar(&sopts.FileStoreOpts.CompactEnabled, "file_compact_enabled", stores.DefaultFileStoreOptions.CompactEnabled, "stan.FileStoreOpts.CompactEnabled")
	fs.IntVar(&sopts.FileStoreOpts.CompactFragmentation, "file_compact_frag", stores.DefaultFileStoreOptions.CompactFragmentation, "stan.FileStoreOpts.CompactFragmentation")
	fs.IntVar(&sopts.FileStoreOpts.CompactInterval, "file_compact_interval", stores.DefaultFileStoreOptions.CompactInterval, "stan.FileStoreOpts.CompactInterval")
	fs.String("file_compact_min_size", fmt.Sprintf("%v", stores.DefaultFileStoreOptions.CompactMinFileSize), "stan.FileStoreOpts.CompactMinFileSize")
	fs.String("file_buffer_size", fmt.Sprintf("%v", stores.DefaultFileStoreOptions.BufferSize), "stan.FileStoreOpts.BufferSize")
	fs.BoolVar(&sopts.FileStoreOpts.DoCRC, "file_crc", stores.DefaultFileStoreOptions.DoCRC, "stan.FileStoreOpts.DoCRC")
	fs.Int64Var(&sopts.FileStoreOpts.CRCPolynomial, "file_crc_poly", stores.DefaultFileStoreOptions.CRCPolynomial, "stan.FileStoreOpts.CRCPolynomial")
	fs.BoolVar(&sopts.FileStoreOpts.DoSync, "file_sync", stores.DefaultFileStoreOptions.DoSync, "stan.FileStoreOpts.DoSync")
	fs.IntVar(&sopts.FileStoreOpts.SliceMaxMsgs, "file_slice_max_msgs", stores.DefaultFileStoreOptions.SliceMaxMsgs, "stan.FileStoreOpts.SliceMaxMsgs")
	fs.String("file_slice_max_bytes", fmt.Sprintf("%v", stores.DefaultFileStoreOptions.SliceMaxBytes), "stan.FileStoreOpts.SliceMaxBytes")
	fs.DurationVar(&sopts.FileStoreOpts.SliceMaxAge, "file_slice_max_age", stores.DefaultFileStoreOptions.SliceMaxAge, "stan.FileStoreOpts.SliceMaxAge")
	fs.StringVar(&sopts.FileStoreOpts.SliceArchiveScript, "file_slice_archive_script", "", "stan.FileStoreOpts.SliceArchiveScript")
	fs.Int64Var(&sopts.FileStoreOpts.FileDescriptorsLimit, "file_fds_limit", stores.DefaultFileStoreOptions.FileDescriptorsLimit, "stan.FileStoreOpts.FileDescriptorsLimit")
	fs.IntVar(&sopts.FileStoreOpts.ParallelRecovery, "file_parallel_recovery", stores.DefaultFileStoreOptions.ParallelRecovery, "stan.FileStoreOpts.ParallelRecovery")
	fs.IntVar(&sopts.IOBatchSize, "io_batch_size", DefaultIOBatchSize, "stan.IOBatchSize")
	fs.Int64Var(&sopts.IOSleepTime, "io_sleep_time", DefaultIOSleepTime, "stan.IOSleepTime")
	fs.StringVar(&sopts.FTGroupName, "ft_group", "", "stan.FTGroupName")

	// First, we need to call NATS's ConfigureOptions() with above flag set.
	// It will be augmented with NATS specific flags and call fs.Parse(args) for us.
	nopts, err := natsd.ConfigureOptions(fs, args, printVersion, printHelp, printTLSHelp)
	if err != nil {
		return nil, nil, err
	}
	// At this point, if NATS config file was specified in the command line (-c of -config)
	// nopts.ConfigFile will not be empty.
	natsConfigFile = nopts.ConfigFile

	// If both nats and streaming configuration files are used, then
	// we only use the config file for the corresponding module.
	// However, if only one command line parameter was specified,
	// we use the same config file for both modules.
	if stanConfigFile != "" || natsConfigFile != "" {
		// If NATS config file was not specified, but streaming was, use
		// streaming config file for NATS too.
		if natsConfigFile == "" {
			if err := nopts.ProcessConfigFile(stanConfigFile); err != nil {
				return nil, nil, err
			}
		}
		// If NATS config file was specified, but not the streaming one,
		// use nats config file for streaming too.
		if stanConfigFile == "" {
			stanConfigFile = natsConfigFile
		}
		if err := ProcessConfigFile(stanConfigFile, sopts); err != nil {
			return nil, nil, err
		}
		// Need to call Parse() again to override with command line params.
		// No need to check for errors since this has already been called
		// in natsd.ConfigureOptions()
		fs.Parse(args)
	}

	// Special handling for some command line params
	var flagErr error
	fs.Visit(func(f *flag.Flag) {
		if flagErr != nil {
			return
		}
		switch f.Name {
		case "SDV":
			// Check value to support -SDV=false
			boolValue, _ := strconv.ParseBool(f.Value.String())
			sopts.Trace, sopts.Debug = boolValue, boolValue
		case "max_bytes", "mb":
			sopts.MaxBytes, flagErr = getBytes(f)
		case "file_compact_min_size":
			sopts.FileStoreOpts.CompactMinFileSize, flagErr = getBytes(f)
		case "file_buffer_size":
			var i64 int64
			i64, flagErr = getBytes(f)
			sopts.FileStoreOpts.BufferSize = int(i64)
		}
	})
	if flagErr != nil {
		return nil, nil, flagErr
	}
	return sopts, nopts, nil
}

// getBytes returns the number of bytes from the flag's String size.
// For instance, 1KB would return 1024.
func getBytes(f *flag.Flag) (int64, error) {
	var res map[string]interface{}
	// Use NATS parser to do the conversion for us.
	res, err := conf.Parse(fmt.Sprintf("bytes: %v", f.Value.String()))
	if err != nil {
		return 0, err
	}
	resVal := res["bytes"]
	if resVal == nil || reflect.TypeOf(resVal).Kind() != reflect.Int64 {
		return 0, fmt.Errorf("%v should be a size, got '%v'", f.Name, resVal)
	}
	return resVal.(int64), nil
}
