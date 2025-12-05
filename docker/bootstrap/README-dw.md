## Testing Image Modifications

In order to fix running tests, `goimports` has been added over the existing images.

These images have already been manually built and pushed, the details below is to document what has been done.

```
$ docker build -t vitess/bootstrap:22.11-percona80 -f docker/bootstrap/Dockerfile.percona80.dw .
$ docker tag docker.io/vitess/bootstrap:22.11-percona80 docker.ihs.demonware.net/platform-dbtech/bootstrap:22.11-percona80
$ docker push docker.ihs.demonware.net/platform-dbtech/bootstrap:22.11-percona80

```
This has been done for each just percona80

