// Package config loads and validates mailgraph runtime settings from Viper.
//
// Settings are resolved in order: CLI flags, MAILGRAPH_* environment variables,
// config.toml, then built-in defaults.
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config holds runtime settings for the collector and HTTP server.
type Config struct {
	// LogFile is the mail log file path to read or tail.
	LogFile string
	// LogType is the log format: "syslog" or "metalog".
	LogType string
	// Year is the starting calendar year when parsing logs without a year field.
	Year int
	// HostFilter is an optional regexp restricting syslog hostnames.
	HostFilter string
	// RRDDir is the directory where RRD database files are stored.
	RRDDir string
	// PIDFile is written when daemon mode is enabled.
	PIDFile string
	// DaemonLogFile receives verbose daemon output when configured.
	DaemonLogFile string
	// RRDName is the base filename for RRD files (e.g. mailgraph.rrd).
	RRDName string
	// IgnoreLocalhost skips mail to or from 127.0.0.1.
	IgnoreLocalhost bool
	// IgnoreHosts lists relay host regexes to ignore.
	IgnoreHosts []string
	// OnlyMailRRD updates only the main mail RRD.
	OnlyMailRRD bool
	// OnlyVirusRRD updates only virus and spam RRDs.
	OnlyVirusRRD bool
	// RBLIsSpam counts RBL rejections as spam events.
	RBLIsSpam bool
	// VirblIsVirus counts VIRBL rejections as virus events.
	VirblIsVirus bool
	// Daemon writes a PID file and detaches when true.
	Daemon bool
	// Cat processes the log once and exits without serving HTTP.
	Cat bool
	// Verbose enables detailed logging.
	Verbose bool
	// Serve starts the HTTP server when true.
	Serve bool
	// ListenAddr is the HTTP or HTTPS listen address (e.g. ":8080").
	ListenAddr string
	// Hostname is shown in graph titles.
	Hostname string
	// TLSEnabled serves HTTPS using TLSCertFile and TLSKeyFile.
	TLSEnabled bool
	// TLSCertFile is the PEM certificate path for TLS.
	TLSCertFile string
	// TLSKeyFile is the PEM private key path for TLS.
	TLSKeyFile string
	// AuthEnabled protects the web UI with HTTP Basic authentication.
	AuthEnabled bool
	// AuthUsername is the Basic Auth username.
	AuthUsername string
	// AuthPassword is the Basic Auth password.
	AuthPassword string
	// AuthRealm is the Basic Auth realm shown by browsers.
	AuthRealm string
}

// Default returns the built-in configuration defaults.
func Default() Config {
	hostname, _ := os.Hostname()
	return Config{
		LogFile:       "/var/log/mail/mail.log",
		LogType:       "syslog",
		Year:          time.Now().Year(),
		RRDDir:        "/var/www/mailgraph/rrd",
		PIDFile:       "/var/run/mailgraph.pid",
		DaemonLogFile: "/var/log/mailgraph.log",
		RRDName:       "mailgraph",
		ListenAddr:    ":8080",
		Hostname:      hostname,
		Serve:         true,
		AuthRealm:     "Mailgraph",
	}
}

// SetDefaults registers default values with the global Viper instance.
func SetDefaults() {
	d := Default()
	viper.SetDefault("log.file", d.LogFile)
	viper.SetDefault("log.type", d.LogType)
	viper.SetDefault("log.year", d.Year)
	viper.SetDefault("log.host_filter", d.HostFilter)
	viper.SetDefault("rrd.dir", d.RRDDir)
	viper.SetDefault("rrd.name", d.RRDName)
	viper.SetDefault("rrd.only_mail", d.OnlyMailRRD)
	viper.SetDefault("rrd.only_virus", d.OnlyVirusRRD)
	viper.SetDefault("daemon.pid_file", d.PIDFile)
	viper.SetDefault("daemon.log_file", d.DaemonLogFile)
	viper.SetDefault("daemon.enabled", d.Daemon)
	viper.SetDefault("server.listen", d.ListenAddr)
	viper.SetDefault("server.hostname", d.Hostname)
	viper.SetDefault("server.serve", d.Serve)
	viper.SetDefault("server.tls_enabled", d.TLSEnabled)
	viper.SetDefault("server.tls_cert", d.TLSCertFile)
	viper.SetDefault("server.tls_key", d.TLSKeyFile)
	viper.SetDefault("auth.enabled", d.AuthEnabled)
	viper.SetDefault("auth.username", d.AuthUsername)
	viper.SetDefault("auth.password", d.AuthPassword)
	viper.SetDefault("auth.realm", d.AuthRealm)
	viper.SetDefault("filter.ignore_localhost", d.IgnoreLocalhost)
	viper.SetDefault("filter.ignore_hosts", d.IgnoreHosts)
	viper.SetDefault("filter.rbl_is_spam", d.RBLIsSpam)
	viper.SetDefault("filter.virbl_is_virus", d.VirblIsVirus)
	viper.SetDefault("app.verbose", d.Verbose)
}

// Load reads the effective configuration from Viper and validates it.
func Load() (Config, error) {
	cfg := Default()

	if v := viper.GetString("log.file"); v != "" {
		cfg.LogFile = v
	}
	if v := viper.GetString("log.type"); v != "" {
		cfg.LogType = v
	}
	if viper.IsSet("log.year") {
		cfg.Year = viper.GetInt("log.year")
	}
	cfg.HostFilter = viper.GetString("log.host_filter")

	if v := viper.GetString("rrd.dir"); v != "" {
		cfg.RRDDir = v
	}
	if v := viper.GetString("rrd.name"); v != "" {
		cfg.RRDName = v
	}
	cfg.OnlyMailRRD = viper.GetBool("rrd.only_mail")
	cfg.OnlyVirusRRD = viper.GetBool("rrd.only_virus")

	if v := viper.GetString("daemon.pid_file"); v != "" {
		cfg.PIDFile = v
	}
	if v := viper.GetString("daemon.log_file"); v != "" {
		cfg.DaemonLogFile = v
	}
	cfg.Daemon = viper.GetBool("daemon.enabled")

	if v := viper.GetString("server.listen"); v != "" {
		cfg.ListenAddr = v
	}
	if v := viper.GetString("server.hostname"); v != "" {
		cfg.Hostname = v
	}
	cfg.Serve = viper.GetBool("server.serve")
	cfg.TLSEnabled = viper.GetBool("server.tls_enabled")
	cfg.TLSCertFile = viper.GetString("server.tls_cert")
	cfg.TLSKeyFile = viper.GetString("server.tls_key")
	cfg.AuthEnabled = viper.GetBool("auth.enabled")
	cfg.AuthUsername = viper.GetString("auth.username")
	cfg.AuthPassword = viper.GetString("auth.password")
	cfg.AuthRealm = viper.GetString("auth.realm")

	cfg.IgnoreLocalhost = viper.GetBool("filter.ignore_localhost")
	cfg.IgnoreHosts = viper.GetStringSlice("filter.ignore_hosts")
	cfg.RBLIsSpam = viper.GetBool("filter.rbl_is_spam")
	cfg.VirblIsVirus = viper.GetBool("filter.virbl_is_virus")
	cfg.Verbose = viper.GetBool("app.verbose")

	if cfg.OnlyMailRRD && cfg.OnlyVirusRRD {
		return cfg, fmt.Errorf("cannot use rrd.only_mail and rrd.only_virus together")
	}
	if cfg.TLSEnabled && (cfg.TLSCertFile == "" || cfg.TLSKeyFile == "") {
		return cfg, fmt.Errorf("server.tls_cert and server.tls_key are required when TLS is enabled")
	}
	if cfg.AuthEnabled && (cfg.AuthUsername == "" || cfg.AuthPassword == "") {
		return cfg, fmt.Errorf("auth.username and auth.password are required when auth is enabled")
	}

	return cfg, nil
}