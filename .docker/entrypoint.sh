#!/usr/bin/env sh
set -eu

cp /opt/SSUIBuildFiles/StationeersServerUI /app/.StationeersServerUI.new
chmod +x /app/.StationeersServerUI.new
mv /app/.StationeersServerUI.new /app/StationeersServerUI
cp /opt/SSUIBuildFiles/LICENSE /app/LICENSE
exec /app/StationeersServerUI "$@"
