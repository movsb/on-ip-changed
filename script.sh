#!/bin/bash

set -eu

build() {
	local pkg='github.com/movsb/on-ip-changed/utils/version'
	local builtOn="$USER@${HOSTNAME:-$(hostname)}"
	local builtAt="${DATE:-$(date +'%F %T %z')}"
	local goVersion=$(go version | sed 's/go version //')
	local gitAuthor=$(git show -s --format='format:%aN <%ae>' HEAD)
	local gitCommit=$(git rev-parse --short HEAD)

	local ldflags="\
	-X '$pkg.BuiltOn=$builtOn' \
	-X '$pkg.BuiltAt=$builtAt' \
	-X '$pkg.GoVersion=$goVersion' \
	-X '$pkg.GitAuthor=$gitAuthor' \
	-X '$pkg.GitCommit=$gitCommit' \
	"

	go build -ldflags "$ldflags" -v
}

systemd() {
	[[ $EUID -ne 0 ]] && echo Please run this command as root. 2>&1 && exit 1

	local serviceName=on-ip-changed
	local serviceFullName="$serviceName".service
	local binPath="$(pwd)/$serviceName"
	local binDir=$(dirname "$binPath")
	cat <<-EOF > /etc/systemd/system/"$serviceFullName"
[Unit]
Description=On IP Changed
After=network.target
After=network-online.target

[Service]
ExecStart="$binPath" daemon
WorkingDirectory=$binDir
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target
EOF
	systemctl enable "$serviceFullName"
	systemctl restart "$serviceFullName"
}

"$@"
