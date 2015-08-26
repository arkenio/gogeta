#!/bin/bash

etcdctl set /domains/bla.dev/type service
etcdctl set /domains/bla.dev/value nxio_000001
etcdctl set /services/nxio_000001/1/config/gogeta "{\"robots\":\"User-Agent: *\\nDisallow:truc\"}"
etcdctl set /services/nxio_000001/1/location "{\"host\":\"localhost\",\"port\":80}"
etcdctl set /services/nxio_000001/1/status/expected started
etcdctl set /services/nxio_000001/1/status/current started
etcdctl set /services/nxio_000001/1/status/alive 1