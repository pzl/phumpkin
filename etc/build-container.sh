#!/bin/bash

set -e
curdir="$(dirname $(readlink -f "$0"))"

cleanup() {
	set +e
	status=$?
	[ ! -z "$bcmt" ] && echo "[i] unmounting compile container" && buildah umount $bc > /dev/null
	[ ! -z "$bc" ] &&   echo "[i] removing compile container $bc" && buildah rm $bc > /dev/null
	#[ ! -z "$img" ] && echo "[i] removing build image $img" && buildah rmi $img
	[ ! -z "$cmt" ] && echo "[i] unmounting build container" && buildah umount $c > /dev/null
	[ ! -z "$c" ] &&   echo "[i] removing build container $c" && buildah rm $c > /dev/null
	set -e
	exit "$status"
}
trap cleanup INT TERM EXIT

#echo "[i] making sure assets are up to date"
#make cmd/phumpkin/assets.go


bc=$(buildah from golang:alpine)
echo "[i] created compile container $bc"
bcmt=$(buildah mount $bc)
echo "[i] mounted compile container $bc at $bcmt"

echo "[i] installing vips"
buildah run -- $bc apk add --no-cache vips-dev

echo "[i] installing build tools"
buildah run -- $bc apk add --no-cache gcc musl-dev

echo "[i] copying code"
mkdir -p $bcmt/app
cp -r cmd $bcmt/app
cp -r pkg $bcmt/app
cp go.mod go.sum $bcmt/app

echo "[i] compiling code"
mkdir -p $bcmt/app/bin
buildah run -- $bc sh -c 'cd /app && go build -o bin ./cmd/...'



c=$(buildah from alpine:edge)
echo "[i] created container $c"
cmt=$(buildah mount $c)
echo "[i] mounted container $c at $cmt"


echo "[i] copying app to container"
cp $bcmt/app/bin/phumpkin $cmt/phumpkin

echo "[i] installing darktable"
buildah run -- $c apk add --no-cache darktable
cp etc/darktablerc $cmt/usr/share/darktable/darktablerc

echo "[i] installing exiftool"
buildah run -- $c apk add --no-cache exiftool

echo "[i] installing vips"
buildah run -- $c apk add --no-cache vips-dev

echo "[i] getting most recent lensdb lenses"
rm -rf $cmt/usr/share/lensfun/version_1/*
curl -s https://wilson.bronger.org/lensfun-db/version_1.tar.bz2 | tar -xj -C $cmt/usr/share/lensfun/version_1


echo "[i] setting app container configs"
buildah config --label name=phumpkin $c
buildah config --port 80 $c
buildah config --author 'Dan Panzarella <dan@panzarel.la>' $c
buildah config --cmd "/phumpkin" $c

echo "[i] committing app container"
img=$(buildah commit $c phumpkin)

#echo "[i] copying committed container to tar file"
#skopeo copy containers-storage:$img oci-archive:phumpkin.tar
