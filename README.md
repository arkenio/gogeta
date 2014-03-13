etcd-reverse-proxy
==================

Reverse proxy based on etcd hierarchy


How it works
------------

It is basically an HTTP cluster router that holds its configuration in etcd. Default behavior
is to use the IoEtcdResolver :

  * client asks for mycustomdomain.com
  * proxy looks at `/nuxeo.io/domains/mycustomdomain.com/[type,value]`
  * if type is io container we look for `/nuxeo.io/envs/{value}/[ip,port]`
  * the request is proxied to `http://{ip}:{port}/`
  * if type is uri
  * the requestion is proxies to the value `/nuxeo.io/domains/mycustomdomain.com/value`

It also provides to custom resolvers :

  * EnvResolver : it serves `http://{envid}.local/ to the host referenced at `/nuxeo.io/envs/{envid}/[ip,port]`
  * DummyResolver : it always proxies to `http://localhost:8080/`


Configuration
-------------

Several parameters allow to configure the way the proxy behave :

 * `domainDir` allows to select the prefix of the key where it watches for domain
 * `envDir` allows to select the prefix of the key where it watches for environments
 * `etcdAddress` specify the address of the `etcd` server
 * `port` port to listen
 * `resolverType` : choose the resolver to use
    * `Env` : EnvResolver
    * `Dummy` : DummyResolver
    * by default : IoEtcd

How to contribute
-----------------

See this page for practical information:
<http://doc.nuxeo.com/display/NXDOC/Nuxeo+contributors+welcome+page>

This presentation will give you more insight about "the Nuxeo way":
<http://www.slideshare.net/nuxeo/nuxeo-world-session-becoming-a-contributor-how-to-get-started>


About Nuxeo
-----------

Nuxeo provides a modular, extensible Java-based
[open source software platform for enterprise content management](http://www.nuxeo.com/en/products/ep),
and packaged applications for [document management](http://www.nuxeo.com/en/products/document-management),
[digital asset management](http://www.nuxeo.com/en/products/dam) and
[case management](http://www.nuxeo.com/en/products/case-management).

Designed by developers for developers, the Nuxeo platform offers a modern
architecture, a powerful plug-in model and extensive packaging
capabilities for building content applications.

More information on: <http://www.nuxeo.com/>
