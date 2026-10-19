#!/bin/sh
# Installs task_prioritizer for the current user, with a desktop menu entry.
#
#   curl -fsSL https://raw.githubusercontent.com/xilen0x/task_prioritizer_tui/main/install.sh | sh
#   curl -fsSL .../install.sh | sh -s -- --uninstall
#
# Environment:
#   PREFIX                    install prefix (default: $HOME/.local)
#   TASK_PRIORITIZER_VERSION  release tag to install, e.g. v1.2.3 (default: latest)
#   TASK_PRIORITIZER_TARBALL  install from this local tarball instead of downloading
set -eu

REPO=xilen0x/task_prioritizer_tui
NAME=task_prioritizer
PREFIX=${PREFIX:-$HOME/.local}
BIN_DIR=$PREFIX/bin
APP_DIR=$PREFIX/share/applications
ICON_ROOT=$PREFIX/share/icons/hicolor
ICON_DIR=$ICON_ROOT/scalable/apps

say() { printf '%s\n' "$*"; }
die() {
	printf 'error: %s\n' "$*" >&2
	exit 1
}
have() { command -v "$1" >/dev/null 2>&1; }

fetch() { # URL DEST
	if have curl; then
		curl -fsSL -o "$2" "$1"
	elif have wget; then
		wget -qO "$2" "$1"
	else
		die "curl or wget is required"
	fi
}

refresh_menus() {
	if have update-desktop-database; then update-desktop-database "$APP_DIR" >/dev/null 2>&1 || true; fi
	if have gtk-update-icon-cache; then gtk-update-icon-cache -q -t "$ICON_ROOT" >/dev/null 2>&1 || true; fi
}

detect_arch() {
	case $(uname -m) in
	x86_64 | amd64) echo amd64 ;;
	aarch64 | arm64) echo arm64 ;;
	*) die "unsupported architecture: $(uname -m) (supported: x86_64, aarch64)" ;;
	esac
}

# download_release TMP: fetches and verifies the release tarball, prints its path.
download_release() {
	asset=${NAME}_linux_$(detect_arch).tar.gz
	version=${TASK_PRIORITIZER_VERSION:-latest}
	if [ "$version" = latest ]; then
		base=https://github.com/$REPO/releases/latest/download
	else
		base=https://github.com/$REPO/releases/download/$version
	fi
	say "Downloading $asset ($version)..." >&2
	fetch "$base/$asset" "$1/$asset" || die "download failed: $base/$asset"
	fetch "$base/checksums.txt" "$1/checksums.txt" || die "download failed: $base/checksums.txt"
	have sha256sum || die "sha256sum is required to verify the download (install coreutils)"
	want=$(awk -v f="$asset" '$2 == f || $2 == "*" f { print $1 }' "$1/checksums.txt")
	[ -n "$want" ] || die "$asset is not listed in checksums.txt"
	got=$(sha256sum "$1/$asset" | cut -d ' ' -f 1)
	[ "$want" = "$got" ] || die "checksum mismatch for $asset"
	say "Checksum verified." >&2
	echo "$1/$asset"
}

install_app() {
	# The launcher path goes into a quoted desktop Exec value, where these
	# characters would need escaping; reject them up front.
	case $BIN_DIR in
	*[\\\"\`\$%]* | *"
"*) die "PREFIX must not contain \\, \", \`, \$, % or a newline: $PREFIX" ;;
	esac

	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT
	trap 'exit 1' INT TERM

	if [ -n "${TASK_PRIORITIZER_TARBALL:-}" ]; then
		tarball=$TASK_PRIORITIZER_TARBALL
		[ -f "$tarball" ] || die "tarball not found: $tarball"
		say "Installing from $tarball"
	else
		tarball=$(download_release "$tmp")
	fi

	mkdir "$tmp/x"
	tar -xzf "$tarball" -C "$tmp/x" || die "cannot extract $tarball"
	src=
	for d in "$tmp/x/${NAME}_linux_"*; do src=$d; done
	[ -n "$src" ] && [ -f "$src/$NAME" ] || die "unexpected tarball layout in $tarball"

	mkdir -p "$BIN_DIR" "$APP_DIR" "$ICON_DIR"
	install -m 755 "$src/$NAME" "$BIN_DIR/$NAME"
	install -m 755 "$src/$NAME-launcher" "$BIN_DIR/$NAME-launcher"
	install -m 644 "$src/$NAME.svg" "$ICON_DIR/$NAME.svg"
	# Absolute Exec path: the menu entry works even if $BIN_DIR is not in PATH.
	EXEC_PATH="$BIN_DIR/$NAME-launcher" awk '/^Exec=/ { print "Exec=\"" ENVIRON["EXEC_PATH"] "\""; next } { print }' \
		"$src/$NAME.desktop" >"$APP_DIR/$NAME.desktop"
	chmod 644 "$APP_DIR/$NAME.desktop"
	refresh_menus

	# Best effort: a failing --version must not fail an otherwise complete install.
	label=$("$BIN_DIR/$NAME" --version 2>/dev/null) || label=
	say "Installed ${label:-$NAME} to $BIN_DIR"
	say "Menu entry: $APP_DIR/$NAME.desktop"
	case ":$PATH:" in
	*":$BIN_DIR:"*) ;;
	*) say "note: $BIN_DIR is not in your PATH; add it to run '$NAME' from a shell." ;;
	esac
}

uninstall_app() {
	for f in "$BIN_DIR/$NAME" "$BIN_DIR/$NAME-launcher" "$APP_DIR/$NAME.desktop" "$ICON_DIR/$NAME.svg"; do
		if [ -e "$f" ]; then
			rm -f "$f"
			say "Removed $f"
		fi
	done
	refresh_menus
	say "Uninstalled. Saved tasks in ${XDG_DATA_HOME:-$HOME/.local/share}/$NAME were kept."
}

main() {
	[ "$(uname -s)" = Linux ] || die "this installer supports Linux only; elsewhere use: go install github.com/$REPO/cmd/$NAME@latest"
	case ${1:-} in
	"") install_app ;;
	--uninstall) uninstall_app ;;
	-h | --help) say "usage: install.sh [--uninstall]  (see the header of this script for options)" ;;
	*) die "unknown option: $1 (use --uninstall or --help)" ;;
	esac
}

main "$@"
