// This file has no purpose other than to mark this directory as its own
// Go module boundary, so `go build/vet/test ./...` run from the repo
// root does not descend into frontend/node_modules — some npm packages
// ship stray .go files (e.g. language-binding examples) that would
// otherwise be swept into the main Go module's build.
module rate-limiter-frontend-placeholder

go 1.26.6
