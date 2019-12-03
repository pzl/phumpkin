#!/bin/bash

set -e
curdir="$(dirname $(readlink -f "$0"))"

cleanup() {
	set +e
	status=$?
	#[ ! -z "$img" ] && echo "[i] removing build image $img" && buildah rmi $img
	[ ! -z "$cmt" ] && echo "[i] unmounting build container" && buildah umount $c > /dev/null
	[ ! -z "$c" ] &&   echo "[i] removing build container $c" && buildah rm $c > /dev/null
	set -e
	exit "$status"
}
trap cleanup INT TERM EXIT


c=$(buildah from alpine:latest)
echo "[i] created container $c"
cmt=$(buildah mount $c)
echo "[i] mounted container $c at $cmt"


echo "[i] copying app to container"
cp bin/phumpkin $cmt/phumpkin

echo "[i] installing darktable"
buildah run -- $c apk add darktable

echo "[i] setting app container configs"
buildah config --label name=phumpkin $c
buildah config --port 80 $c
buildah config --author 'Dan Panzarella <dan@panzarel.la>' $c
buildah config --cmd "/phumpkin" $c

echo "[i] committing app container"
img=$(buildah commit $c phumpkin)

#echo "[i] copying committed container to tar file"
#skopeo copy containers-storage:$img oci-archive:phumpkin.tar
