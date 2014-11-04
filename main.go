package main

import (
	"github.com/golang/glog"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"bufio"
	"fmt"
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

	handleSignals()

	resolver, error := getResolver(c)
	if error != nil {
		panic(error)
	} else {

		resolver.init()

		p := NewProxy(c, resolver)
		p.start()
	}

}

func handleSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	signal.Notify(signals, os.Interrupt, syscall.SIGUSR1)
	go func() {
		sig := <-signals
		switch sig {
		case syscall.SIGTERM, syscall.SIGINT:
			//Exit gracefully
			glog.Info("Shutting down...")
			os.Exit(0)
		case syscall.SIGUSR1:
			pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
			profileFile := "/tmp/gogeta.profile"
			DumpCPUProfile(profileFile)
			FileToStdOut(profileFile)
		}

	}()
}

func DumpCPUProfile(filePath string)  {
	f, err := os.Create(filePath)
	if err == nil {
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
}


func FileToStdOut(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Sprintf("Unable to open file : ", filePath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Printf(scanner.Text())
	}

}
