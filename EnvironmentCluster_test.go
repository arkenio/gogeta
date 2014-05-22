package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func Test_cluster(t *testing.T) {
	var cluster *EnvironmentCluster

	Convey("Given an environment cluster", t, func() {
		cluster = &EnvironmentCluster{}

		Convey("When the cluster is initialized", func() {
			Convey("Then it should be empty", func() {
				So(len(cluster.instances), ShouldEqual, 0)

			})
		})

		Convey("When the cluster contains an inactive environment", func() {
			cluster.Add(getEnvironment("1", "nxio-0001", false))
			Convey("Then it can't get a next environment", func() {
				env, err := cluster.Next()

				So(len(cluster.instances), ShouldEqual, 1)
				So(env, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When the cluster contains active environment", func() {
			cluster.Add(getEnvironment("2", "nxio-0001", true))
			Convey("Then it can get a next environment", func() {
				env, err := cluster.Next()

				So(len(cluster.instances), ShouldEqual, 1)
				So(env, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("Then returned environment should always be the same", func() {
				env, _ := cluster.Next()
				firstKey := env.key
				env, _ = cluster.Next()
				So(env.key, ShouldEqual, firstKey)

			})

		})

		Convey("When the cluster contains several environments", func() {
			cluster.Add(getEnvironment("1", "nxio-0001", true))
			cluster.Add(getEnvironment("2", "nxio-0001", false))
			cluster.Add(getEnvironment("3", "nxio-0001", true))

			Convey("Then it should loadbalance between environments", func() {
				env, err := cluster.Next()
				So(env, ShouldNotBeNil)
				So(err, ShouldBeNil)

				firstKey := env.key

				env, err = cluster.Next()
				So(env, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(env.key, ShouldNotEqual, firstKey)
			})

			Convey("Then it should never loadbalance on an inactive environment", func() {
				for i := 0; i < len(cluster.instances); i++ {
					env, err := cluster.Next()
					So(env, ShouldNotBeNil)
					So(err, ShouldBeNil)
					So(env.key, ShouldNotEqual, "2")
				}
			})

			Convey("Then it can get each environment by its key", func() {

				env := cluster.Get("1")
				So(env.key, ShouldEqual, "1")
				So(env.status.current, ShouldEqual, "started")

				env = cluster.Get("2")
				So(env.key, ShouldEqual, "2")
				So(env.status.current, ShouldEqual, "stopped")
			})

		})

		Convey("When removing a key to a cluster", func() {
			cluster.Add(getEnvironment("1", "nxio-0001", true))
			cluster.Add(getEnvironment("2", "nxio-0001", false))
			cluster.Add(getEnvironment("3", "nxio-0001", true))

			initSize := len(cluster.instances)

			cluster.Remove("2")

			Convey("Then it should containe one less instance", func() {
				So(len(cluster.instances), ShouldEqual, initSize-1)

			})
		})

	})

}

func Test_Environment(t *testing.T) {
	var env1, env2 *Environment

	Convey("Given two environment with same values", t, func() {
		env1 = getEnvironment("1", "nxio-0001", true)
		env2 = getEnvironment("1", "nxio-0001", true)
		Convey("When i dont change anything", func() {
			Convey("Then they are equal", func() {

				So(env1.equals(env2), ShouldEqual, true)

			})

		})

		Convey("When host is not the same", func() {
			env2.location.Host = "otherhost"
			Convey("Then they are equal", func() {

				So(env1.equals(env2), ShouldEqual, false)

			})

		})

		Convey("When port is not the same", func() {
			env2.location.Port = 9090
			Convey("Then they are equal", func() {

				So(env1.equals(env2), ShouldEqual, false)

			})

		})

		Convey("When current status is not the same", func() {
			env2.status.current = "other"
			Convey("Then they are equal", func() {

				So(env1.equals(env2), ShouldEqual, false)

			})

		})

		Convey("When expected status is not the same", func() {
			env2.status.expected = "other"
			Convey("Then they are equal", func() {

				So(env1.equals(env2), ShouldEqual, false)

			})

		})
		Convey("When alive status is not the same", func() {
			env2.status.alive = "other"
			Convey("Then they are equal", func() {

				So(env1.equals(env2), ShouldEqual, false)

			})

		})
	})
}

func getEnvironment(key string, name string, active bool) *Environment {
	var s *Status

	if active {
		s = &Status{"1", "started", "started"}
	} else {
		s = &Status{"", "stopped", "started"}
	}

	return &Environment{
		key:      key,
		location: service{"127.0.0.1", 8080},
		domain:   "dummydomain.com",
		name:     name,
		status:   s}

}
