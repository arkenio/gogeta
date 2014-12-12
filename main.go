package main

import (
	"github.com/golang/glog"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

const (
	progname = "gogeta"
)

func getResolver(c *Config) (domainResolver, error) {
	switch c.resolverType {
	case "Dummy":
		return &DummyResolver{}, nil
	case "Env":
		return NewEnvResolver(c), nil
	default:
		r, err := NewEtcdResolver(c)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
}

func main() {

	glog.Infof("%s starting", progname)

	c := parseConfig()

	if c.cpuProfile != "" {
		f, err := os.Create(c.cpuProfile)
		if err != nil {
			glog.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	handleSignals(c)

	resolver, error := getResolver(c)
	if error != nil {
		panic(error)
	} else {

		resolver.init()

		p := NewProxy(c, resolver)
		p.start()
	}

}

func handleSignals(config *Config) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	signal.Notify(signals, os.Interrupt, syscall.SIGUSR1)
	signal.Notify(signals, os.Interrupt, syscall.SIGUSR2)

	go func() {
		isProfiling := false

		defer func() {
			if isProfiling {
				pprof.StopCPUProfile()
			}
		}()

		sig := <-signals
		switch sig {
		case syscall.SIGTERM, syscall.SIGINT:
			//Exit gracefully
			glog.Info("Shutting down...")
			os.Exit(0)
		case syscall.SIGUSR1:
			pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
		case syscall.SIGUSR2:
			if !isProfiling {
				f, err := os.Create(config.cpuProfile)
				if err != nil {
					glog.Fatal(err)
				} else {
					pprof.StartCPUProfile(f)
					isProfiling = true
				}
			} else {
				pprof.StopCPUProfile()
				isProfiling = false
			}

		}

	}()

}
