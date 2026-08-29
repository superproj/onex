#!/bin/bash
CHROME="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
for f in 01-layout 02-navigation 03-form 04-data 05-feedback 06-mobile; do
  "$CHROME" --headless --disable-gpu --no-sandbox --disable-dev-shm-usage --hide-scrollbars \
    --force-device-scale-factor=2 --window-size=1440,900 \
    --screenshot="/tmp/ui-layout-diagrams/$f.png" \
    "file:///tmp/ui-layout-diagrams/$f.html" >/dev/null 2>&1
  echo "rendered $f.png"
done
ls -la /tmp/ui-layout-diagrams/*.png
