# Contributing

Thanks for your interest in task_prioritizer. Issues and pull requests are welcome.

## Before you start

- For bugs, open an issue with your OS, terminal emulator, terminal size and the steps to reproduce.
- For new features, open an issue first so we can agree on the scope before you write code.

## Development

Requirements: Go 1.23 or newer.

```sh
make demo           # run the app with sample data
make check          # gofmt -l, go vet, go test
make update-golden  # regenerate the TUI golden files after an intended UI change
```

- Keep the domain packages (`internal/tasks`, `internal/pomodoro`) free of UI imports.
- Add or update tests with every behaviour change. UI changes need regenerated goldens;
  review the diff of `internal/tui/testdata` before committing.
- Run `make check` before opening a pull request; CI runs the same checks.
- Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages
  (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).

## Releasing

Tag the commit on `main` and push the tag:

```sh
git tag v1.2.3
git push origin v1.2.3
```

The release workflow runs the tests, builds the Linux amd64/arm64 tarballs with `make dist`,
and publishes them with `checksums.txt` as a GitHub Release that `install.sh` downloads.
Use `make dist` locally to check the tarball for your architecture.

By contributing you agree that your contributions are licensed under the MIT License.
