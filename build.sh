#!/bin/sh
# Builds the Windows uploader the one way it's always built, and checks the result is byte for byte the file
# in RELEASE. Same source, same Go version and same flags give the same file, so Chrome and
# Windows keep the download history they've seen for it instead of treating every deploy as a new, unknown file.
#
#   sh build.sh OUT.exe            build and check
#   UPDATE=1 sh build.sh OUT.exe   build and record a new release (only when Melktruck asks for a new uploader)
set -eu
cd "$(dirname "$0")"
out="${1:?usage: build.sh OUT.exe}"
want_go="go1.24.7"
have_go="$(GOTOOLCHAIN=local go env GOVERSION)"
if [ "$have_go" != "$want_go" ]; then
  echo "The uploader is built with $want_go so the file never changes; this is $have_go." >&2
  exit 1
fi
GOTOOLCHAIN=local GOFLAGS=-mod=readonly GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
  go build -trimpath -buildvcs=false -ldflags "-H=windowsgui -buildid=" -o "$out" .
sum="$(sha256sum "$out" | cut -d' ' -f1)"
version="$(sed -n 's/^[[:space:]]*version[[:space:]]*= "\(.*\)"$/\1/p' main.go)"
if [ "${UPDATE:-}" = 1 ]; then
  printf '%s %s\n' "$version" "$sum" > RELEASE
  echo "Recorded uploader $version: $sum"
  exit 0
fi
read -r rel_version rel_sum < RELEASE
if [ "$sum" != "$rel_sum" ] || [ "$version" != "$rel_version" ]; then
  echo "The uploader would change: built $version $sum, but RELEASE says $rel_version $rel_sum." >&2
  echo "Members' browsers treat a changed file as new and warn again, so it only changes on purpose:" >&2
  echo "bump version in main.go and run UPDATE=1 sh tools/uploader/build.sh /tmp/u.exe." >&2
  exit 1
fi
echo "Uploader $version: $sum (unchanged)"
