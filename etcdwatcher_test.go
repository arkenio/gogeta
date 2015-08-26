package main

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
	"time"
)

func Test_EtcdWatcher(t *testing.T) {
	if os.Getenv("IT_Test") != "" {
		IT_EtcdWatcher(t)
	}
}

func IT_EtcdWatcher(t *testing.T) {

	c := parseConfig()
	client, _ := c.getEtcdClient()

	client.Delete("/domains", true)
	client.Delete("/services", true)

	var w *watcher

	Convey("Given a watcher", t, func() {
		domains := make(map[string]*Domain)
		services := make(map[string]*ServiceCluster)

		w, _ = NewEtcdWatcher(c, domains, services)
		w.init()

		Convey("When it is started", func() {

			Convey("It doesn't contains any domain", func() {
				So(len(w.domains), ShouldEqual, 0)
			})

			Convey("It doesn't contains any service", func() {
				So(len(w.services), ShouldEqual, 0)
			})
		})

		Convey("When I add a domain", func() {

			_, err := client.Set("/domains/mydomain.com/type", "service", 0)
			So(len(w.domains), ShouldEqual, 0)
			if err != nil {
				panic(err)
			}
			_, err = client.Set("/domains/mydomain.com/value", "my_service", 0)

			WaitEtcd()
			if err != nil {
				panic(err)
			}
			So(len(w.domains), ShouldEqual, 1)
			domain := w.domains["mydomain.com"]
			So(domain.typ, ShouldEqual, "service")
			So(domain.value, ShouldEqual, "my_service")

		})

		Convey("When I remove the domain in etcd", func() {
			_, err := client.Delete("/domains/mydomain.com", true)
			if err != nil {
				panic(err)
			}
			Convey("Then the domain is removed from the list of domains", func() {
				So(len(w.domains), ShouldEqual, 0)

			})

		})

		Convey("When I add a service", func() {
			_, err := client.Set("/services/my_service/1/domain", "mydomain.com", 0)
			WaitEtcd()
			if err != nil {
				panic(err)
			}

			Convey("Then there should be one service", func() {
				So(len(w.services), ShouldEqual, 1)
			})

			Convey("Then the status should be nil", func() {
				_, err := w.services["my_service"].Next()
				So(err, ShouldNotBeNil)
				So(err.(StatusError).status, ShouldBeNil)

			})

		})

		// Creates a service that has not status, meaning started by default
		Convey("When I add a location to the service", func() {

			b, _ := json.Marshal(&location{Host: "127.0.0.1", Port: 8080})
			client.Set("/services/my_service/1/location", string(b[:]), 0)
			WaitEtcd()

			Convey("Then it should be started", func() {
				service, err := w.services["my_service"].Next()
				So(err, ShouldBeNil)
				So(service.status.compute(), ShouldEqual, STARTED_STATUS)
			})
		})

		// When we create a service, it should be stopped
		Convey("When i add a stopped status", func() {
			client.Set("/services/my_service/1/status/expected", "stopped", 0)
			client.Set("/services/my_service/1/status/current", "stopped", 0)
			WaitEtcd()
			Convey("Then it should be stopped", func() {
				_, err := w.services["my_service"].Next()
				So(err, ShouldNotBeNil)
				So(err.(StatusError).computedStatus, ShouldEqual, STOPPED_STATUS)
			})

		})

		Convey("When I add an expected started status", func() {
			client.Set("/services/my_service/1/status/expected", "started", 0)
			WaitEtcd()
			Convey("Then the service should be in error (meaning unit has not set starting as current status)", func() {
				_, err := w.services["my_service"].Next()
				So(err, ShouldNotBeNil)
				So(err.(StatusError).computedStatus, ShouldEqual, ERROR_STATUS)
			})
		})

		Convey("When I add a current starting status", func() {
			client.Set("/services/my_service/1/status/current", "starting", 0)
			WaitEtcd()
			Convey("Then the service should be in starting", func() {
				_, err := w.services["my_service"].Next()
				So(err, ShouldNotBeNil)
				So(err.(StatusError).computedStatus, ShouldEqual, STARTING_STATUS)
			})
		})

		Convey("When I add a current started status", func() {
			client.Set("/services/my_service/1/status/current", "started", 0)
			WaitEtcd()
			Convey("Then the service should be in error (if not alive it should be in error)", func() {
				_, err := w.services["my_service"].Next()
				So(err, ShouldNotBeNil)
				So(err.(StatusError).computedStatus, ShouldEqual, ERROR_STATUS)
			})
		})

		Convey("When I add an alive key", func() {
			client.Set("/services/my_service/1/status/current", "started", 0)
			client.Set("/services/my_service/1/status/alive", "1", 0)
			WaitEtcd()
			Convey("Then the service should be starting", func() {
				service, err := w.services["my_service"].Next()
				So(err, ShouldBeNil)
				So(service.status.compute(), ShouldEqual, STARTED_STATUS)
			})
		})

		Convey("When I passivate the service", func() {
			client.Set("/services/my_service/1/status/current", STOPPED_STATUS, 0)
			client.Set("/services/my_service/1/status/expected", PASSIVATED_STATUS, 0)
			WaitEtcd()
			Convey("Then the service should be starting", func() {
				_, err := w.services["my_service"].Next()
				So(err, ShouldNotBeNil)
				So(err.(StatusError).computedStatus, ShouldEqual, PASSIVATED_STATUS)
			})
		})

		Convey("When I add a gogeta config value", func() {
			client.Set("/services/my_service/1/status/current", "started", 0)
			client.Set("/services/my_service/1/status/alive", "1", 0)
			WaitEtcd()
			service, err := w.services["my_service"].Next()
			So(err, ShouldBeNil)
			So(service.config, ShouldNotBeNil)
			So(service.config.Robots, ShouldEqual, "")
			b, _ := json.Marshal(&ServiceConfig{Robots: "UserAgent: *"})
			client.Set("/services/my_service/1/config/gogeta", string(b[:]), 0)
			WaitEtcd()
			Convey("Then the robots config should be set", func() {
				service, _ = w.services["my_service"].Next()
				So(service.config.Robots, ShouldEqual, "UserAgent: *")
			})
		})

		Convey("When I add a gogeta config value", func() {
			client.Delete("/services/my_service/1/config/gogeta", true)
			WaitEtcd()
			Convey("Then the robots config should be set", func() {
				service, _ := w.services["my_service"].Next()
				So(service.config.Robots, ShouldEqual, "")
			})
		})

	})

}

func WaitEtcd() {
	time.Sleep(500 * time.Millisecond)
}
