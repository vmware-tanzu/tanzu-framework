# What is going on here?

* Yes, these files are empty.
* We check them into the repo so that the `go:embed` directives in
  `cmd/cli/plugin/pinniped-auth/login.go` don't cause compilation to fail.
* The `hack/embed-pinniped-binary.sh` script will create "real" versions of these files that are
  actual binaries that are embedded into the `pinniped-auth` plugin.
