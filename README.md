# taskbox

task management with eisenhower matrix

## setup

### production build
```bash
go build -o taskbox ./cmd/server
./taskbox
```

### development with hot reload

**using air (recommended)**
```bash
# install air (one time)
go install github.com/air-verse/air@latest

# add to PATH if not already
export PATH=$PATH:$HOME/go/bin

# run with hot reload (watches .go, .html, .scss files)
air
# or with full path
~/go/bin/air
```

visit http://localhost:1234

## features

- multi-user authentication
- inbox for task capture
- eisenhower matrix (do/decide/delegate/delete)
- drag & drop task organization
- task details with description, due date, tags, comments
- archive for completed tasks

## tech stack

- backend: go with html templates
- frontend: htmx for interactions
- drag-drop: sortable.js
- database: sqlite
- styling: scss (auto-compiled)

## notes

scss files in `/scss` are automatically compiled to `/static/css` with hot reloading
