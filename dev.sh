#!/bin/bash

# hot reload development server

echo "starting hot reload development server..."
echo "watching: *.go, templates/*, scss/*"
echo ""

# initial build and start
go build -o taskbox ./cmd/server
./taskbox &
PID=$!

# watch for changes
while true; do
	# wait for file changes
	find . -name "*.go" -o -path "./templates/*" -o -path "./scss/*" | entr -d -r sh -c '
		echo ""
		echo "changes detected, rebuilding..."
		pkill -P '$PID' taskbox 2>/dev/null
		pkill taskbox 2>/dev/null
		go build -o taskbox ./cmd/server && ./taskbox &
		echo "server restarted"
	' 2>/dev/null
	
	# if entr not installed, fall back to simple loop
	if [ $? -eq 127 ]; then
		echo "entr not found, using basic file watching..."
		echo "install entr for better performance: sudo apt install entr"
		
		# basic file watching fallback
		LAST_CHANGE=$(find . -name "*.go" -o -path "./templates/*" | xargs stat -c %Y 2>/dev/null | sort -n | tail -1)
		
		while true; do
			sleep 2
			CURRENT_CHANGE=$(find . -name "*.go" -o -path "./templates/*" | xargs stat -c %Y 2>/dev/null | sort -n | tail -1)
			
			if [ "$CURRENT_CHANGE" != "$LAST_CHANGE" ]; then
				echo ""
				echo "changes detected, rebuilding..."
				pkill taskbox 2>/dev/null
				go build -o taskbox ./cmd/server && ./taskbox &
				PID=$!
				echo "server restarted"
				LAST_CHANGE=$CURRENT_CHANGE
			fi
		done
		break
	fi
done
