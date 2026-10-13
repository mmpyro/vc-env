#!/usr/bin/env bats
# tests/integration.bats

setup() {
	# Load helper libraries
	load '/usr/local/lib/bats-support/load'
	load '/usr/local/lib/bats-assert/load'

	# Set up a clean environment for each test
	export VCENV_ROOT="${BATS_TMPDIR}/vc-env"
	export PATH="${VCENV_ROOT}/shims:${PATH}"
	mkdir -p "${VCENV_ROOT}"
}

teardown() {
	# Clean up after each test
	rm -rf "${VCENV_ROOT}"
}

# install_fake <version>: install a stub vcluster that echoes its arguments
# and exits with the code given as "exit=<n>" (default 0). Needs no network.
install_fake() {
	local ver="$1"
	local src="${BATS_TMPDIR}/fake-vcluster-${ver}"
	cat > "${src}" <<'SCRIPT'
#!/bin/sh
echo "fake-vcluster VCENV_VERSION=${VCENV_VERSION} args:$*"
case "$1" in exit=*) exit "${1#exit=}" ;; esac
exit 0
SCRIPT
	vc-env install "${ver}" --from-file "${src}" --silent
}

@test "vc-env init creates necessary directories and shims" {
	run vc-env init
	assert_success
	[ -d "${VCENV_ROOT}/versions" ]
	[ -f "${VCENV_ROOT}/shims/vcluster" ]
}

@test "vc-env list-remote returns a list of versions" {
	vc-env init
	run vc-env list-remote
	assert_success
	assert_output --regexp "[0-9]+\.[0-9]+\.[0-9]+"
}

@test "vc-env latest returns a stable version" {
	vc-env init
	run vc-env latest
	assert_success
	assert_output --regexp "[0-9]+\.[0-9]+\.[0-9]+$"
}

@test "vc-env latest --prerelease returns a version" {
	vc-env init
	run vc-env latest --prerelease
	assert_success
	assert_output --regexp "[0-9]+\.[0-9]+\.[0-9]+"
}

@test "vc-env latest --help shows help text" {
	run vc-env latest --help
	assert_success
	assert_output --partial "latest"
	assert_output --partial "prerelease"
}

@test "vc-env install/list/global/which flow" {
	vc-env init
	local test_ver="0.21.1"
	
	# Install
	run vc-env install "${test_ver}"
	assert_success
	[ -f "${VCENV_ROOT}/versions/${test_ver}/vcluster" ]
	
	# List
	run vc-env list
	assert_success
	assert_output --partial "${test_ver}"
	
	# Global
	run vc-env global "${test_ver}"
	assert_success
	
	run vc-env global
	assert_success
	assert_output --partial "${test_ver}"
	
	# Which
	run vc-env which
	assert_success
	assert_output --partial "${VCENV_ROOT}/versions/${test_ver}/vcluster"
}

@test "vc-env local sets directory-specific version" {
	vc-env init
	local test_ver="0.21.1"
	
	# Must install it first because setup() wipes VCENV_ROOT
	vc-env install "${test_ver}"
	
	local work_dir="${BATS_TMPDIR}/work"
	mkdir -p "${work_dir}"
	
	pushd "${work_dir}"
	run vc-env local "${test_ver}"
	assert_success
	[ -f ".vcluster-version" ]
	grep -q "${test_ver}" ".vcluster-version"
	
	run vc-env local
	assert_success
	assert_output --partial "${test_ver}"
	popd
	
	rm -rf "${work_dir}"
}

@test "vc-env shell outputs export command" {
	vc-env init
	local test_ver="0.21.1"
	vc-env install "${test_ver}"
	run vc-env shell "${test_ver}"
	assert_success
	assert_output --partial "export VCENV_VERSION=${test_ver}"
}

@test "vc-env uninstall removes version" {
	vc-env init
	local test_ver="0.21.1"
	# Pre-install
	run vc-env install "${test_ver}"
	assert_success
	
	# Uninstall
	run vc-env uninstall "${test_ver}"
	assert_success
	[ ! -d "${VCENV_ROOT}/versions/${test_ver}" ]
	
	run vc-env list
	assert_success
	refute_output --partial "${test_ver}"
}

@test "vc-env version" {
	run vc-env version
	assert_success
	assert_output --regexp "[0-9]+\.[0-9]+\.[0-9]+"
}

@test "vc-env help" {
	run vc-env help
	assert_success
	assert_output --partial "Commands:"
}

@test "vc-env status shows environment information" {
	vc-env init
	local test_ver="0.21.1"
	vc-env install "${test_ver}"
	vc-env global "${test_ver}"
	
	run vc-env status
	assert_success
	assert_output --partial "VCENV_ROOT"
	assert_output --regexp "Active version:[[:space:]]+${test_ver}"
	assert_output --partial "set by global version file"
	assert_output --partial "* ${test_ver}"
}

@test "vc-env exec runs specific version" {
	vc-env init
	local test_ver="0.21.1"
	vc-env install "${test_ver}"
	
	# Run a command via exec
	run vc-env exec "${test_ver}" version
	assert_success
	assert_output --regexp "([vV]ersion|[0-9]+\.[0-9]+\.[0-9])"
}

@test "vc-env exec strips one leading -- and sets VCENV_VERSION" {
	vc-env init
	install_fake 9.9.1

	run vc-env exec 9.9.1 -- --flag value -- tail
	assert_success
	assert_output "fake-vcluster VCENV_VERSION=9.9.1 args:--flag value -- tail"
}

@test "vc-env exec propagates the child exit code" {
	vc-env init
	install_fake 9.9.1

	run vc-env exec 9.9.1 exit=3
	[ "$status" -eq 3 ]
	assert_output --partial "args:exit=3"
	refute_output --partial "exit status"

	run vc-env exec 9.9.1 -- exit=42
	[ "$status" -eq 42 ]
}

@test "vc-env exec propagates a real vcluster failure" {
	vc-env init
	vc-env install 0.21.1

	vc-env exec 0.21.1 version
	run vc-env exec 0.21.1 -- definitely-not-a-subcommand
	assert_failure
	refute_output --partial "exit status"
}

@test "vc-env exec resolves a constraint spec" {
	vc-env init
	install_fake 9.9.1
	install_fake 9.9.4

	run vc-env exec 9.9 -- version
	assert_success
	assert_output --partial "VCENV_VERSION=9.9.4"
}

@test "vc-env exec fails for a missing version without auto-install" {
	vc-env init
	run vc-env exec --no-auto 9.9.9 version
	assert_failure
	assert_output --partial "not installed"
}

@test "vc-env exec -h shows help" {
	run vc-env exec -h
	assert_success
	assert_output --partial "Usage: vc-env exec [--auto|--no-auto]"
}

@test "vc-env upgrade --help prints help and does not upgrade" {
	local before
	before=$(sha256sum "$(command -v vc-env)" | awk '{print $1}')
	run vc-env upgrade --help
	assert_success
	assert_output --partial "Usage: vc-env upgrade"
	refute_output --partial "is already up to date"
	refute_output --partial "Running a dev build"
	[ "$(sha256sum "$(command -v vc-env)" | awk '{print $1}')" = "${before}" ]

	run vc-env upgrade -h
	assert_success
	assert_output --partial "Usage: vc-env upgrade"
}

@test "vc-env -h is honoured by uninstall, shell, local and global" {
	vc-env init
	local cmd
	for cmd in uninstall shell local global which status; do
		run vc-env "${cmd}" -h
		assert_success
		assert_output --partial "Usage: vc-env ${cmd}"
	done
	# No side effects: nothing was written.
	[ ! -f "${VCENV_ROOT}/version" ]
	[ ! -f ".vcluster-version" ]
}

@test "vc-env prune is a dry run by default" {
	vc-env init
	install_fake 9.9.1
	install_fake 9.9.2

	VCENV_PRUNE_SCAN_ROOTS="" run vc-env prune
	assert_success
	assert_output --partial "would remove 9.9.1"
	assert_output --partial "pass --yes to apply"
	[ -d "${VCENV_ROOT}/versions/9.9.1" ]
	[ -d "${VCENV_ROOT}/versions/9.9.2" ]
}

@test "vc-env prune --yes removes unreferenced versions" {
	vc-env init
	install_fake 9.9.1
	install_fake 9.9.2
	vc-env global 9.9.2

	VCENV_PRUNE_SCAN_ROOTS="" run vc-env prune --yes
	assert_success
	assert_output --partial "removed 9.9.1"
	[ ! -d "${VCENV_ROOT}/versions/9.9.1" ]
	[ -d "${VCENV_ROOT}/versions/9.9.2" ]
}

@test "vc-env prune keeps versions referenced by .vcluster-version files" {
	vc-env init
	install_fake 9.9.1
	install_fake 9.9.2
	install_fake 9.10.0
	local scan="${BATS_TMPDIR}/prune-scan"
	mkdir -p "${scan}/a" "${scan}/b"
	echo "9.9.1" > "${scan}/a/.vcluster-version"
	echo "9.10" > "${scan}/b/.vcluster-version"

	VCENV_PRUNE_SCAN_ROOTS="${scan}" run vc-env prune --yes
	assert_success
	[ -d "${VCENV_ROOT}/versions/9.9.1" ]
	[ -d "${VCENV_ROOT}/versions/9.10.0" ]
	[ ! -d "${VCENV_ROOT}/versions/9.9.2" ]
	rm -rf "${scan}"
}

@test "vc-env prune --older-than rejects invalid durations" {
	vc-env init
	run vc-env prune --older-than bogus
	assert_failure
	assert_output --partial "invalid --older-than value"
}

@test "vc-env doctor succeeds after init with shims on PATH" {
	vc-env init
	install_fake 9.9.1
	vc-env global 9.9.1

	run vc-env doctor
	assert_success
	assert_output --partial "[OK]"
	assert_output --partial "Active version"
	refute_output --partial "[FAIL]"
}

@test "vc-env doctor fails when shims are not on PATH" {
	vc-env init
	install_fake 9.9.1
	vc-env global 9.9.1

	local bindir
	bindir="$(dirname "$(command -v vc-env)")"
	run env PATH="${bindir}:/usr/bin:/bin" vc-env doctor
	assert_failure
	assert_output --partial "[FAIL]"
	assert_output --partial "not on PATH"
}

@test "vc-env doctor --fix regenerates a tampered shim" {
	vc-env init
	install_fake 9.9.1
	vc-env global 9.9.1
	echo "#!/bin/sh" > "${VCENV_ROOT}/shims/vcluster"

	run vc-env doctor
	assert_output --partial "Shim drift"

	run vc-env doctor --fix
	assert_success
	refute_output --partial "Shim drift"
	grep -q "vc-env resolve" "${VCENV_ROOT}/shims/vcluster"
}

@test "vcluster shim end-to-end" {
	vc-env init
	local test_ver="0.21.1"
	
	vc-env install "${test_ver}"
	vc-env global "${test_ver}"
	
	# Ensure shim is in PATH
	run vcluster version
	assert_output --regexp "([vV]ersion|[0-9]+\.[0-9]+\.[0-9])"
}

@test "vc-env completion bash outputs native bash script" {
	run vc-env completion bash
	assert_success
	assert_output --partial "_vc_env_completions()"
	assert_output --partial "complete -F _vc_env_completions vc-env"
	assert_output --partial "vc-env __complete-versions installed"
	assert_output --partial "vc-env __complete-versions remote"
}

@test "vc-env completion zsh outputs native zsh compdef script" {
	run vc-env completion zsh
	assert_success
	assert_output --partial "#compdef vc-env"
	assert_output --partial "_vc_env()"
	assert_output --partial "compdef _vc_env vc-env"
}

@test "vc-env completion fish outputs native fish complete script" {
	run vc-env completion fish
	assert_success
	assert_output --partial "complete -c vc-env"
	assert_output --partial "__fish_use_subcommand"
	assert_output --partial "__fish_seen_subcommand_from install"
}

@test "vc-env completion powershell outputs ArgumentCompleter script" {
	run vc-env completion powershell
	assert_success
	assert_output --partial "Register-ArgumentCompleter"
	assert_output --partial "-CommandName vc-env"
}

@test "vc-env completion rejects unknown shells" {
	run vc-env completion tcsh
	assert_failure
	assert_output --partial "unsupported shell"
}

@test "vc-env completion --help shows help text" {
	run vc-env completion --help
	assert_success
	assert_output --partial "vc-env completion [<shell>] [--shell <shell>]"
	assert_output --partial "pwsh"
	assert_output --partial "bash"
	assert_output --partial "zsh"
	assert_output --partial "fish"
	assert_output --partial "powershell"
}

@test "vc-env autocompletion still works as deprecated alias" {
	run vc-env autocompletion bash
	assert_success
	assert_output --partial "_vc_env_completions()"
	assert_output --partial "complete -F _vc_env_completions vc-env"
	assert_output --partial "vc-env: 'autocompletion' is deprecated; use 'vc-env completion'"
}

@test "vc-env autocompletion takes the same arguments as completion" {
	run vc-env autocompletion --shell zsh
	assert_success
	assert_output --partial "#compdef vc-env"
	assert_output --partial "'autocompletion' is deprecated"
}

@test "vc-env completion works for every shell, positional and --shell" {
	local shell marker
	for shell in bash zsh fish powershell pwsh; do
		case "${shell}" in
			bash) marker="complete -F _vc_env_completions vc-env" ;;
			zsh) marker="#compdef vc-env" ;;
			fish) marker="complete -c vc-env -f" ;;
			*) marker="Register-ArgumentCompleter" ;;
		esac
		run vc-env completion "${shell}"
		assert_success
		assert_output --partial "${marker}"
		run vc-env completion --shell "${shell}"
		assert_success
		assert_output --partial "${marker}"
		run vc-env completion --shell="${shell}"
		assert_success
		assert_output --partial "${marker}"
	done
}

@test "vc-env completion detects the shell from SHELL" {
	SHELL=/usr/bin/fish run vc-env completion
	assert_success
	assert_output --partial "complete -c vc-env -f"
	refute_output --partial "defaulting to bash"
}

@test "vc-env completion falls back to bash with a note" {
	SHELL=/bin/tcsh run vc-env completion
	assert_success
	assert_output --partial "complete -F _vc_env_completions vc-env"
	assert_output --partial "defaulting to bash"
}

@test "vc-env completion --shell rejects unknown shells" {
	run vc-env completion --shell tcsh
	assert_failure
	assert_output --partial "unsupported shell"
}

@test "vc-env __complete-versions installed lists installed versions" {
	vc-env init
	mkdir -p "${VCENV_ROOT}/versions/0.22.0"
	mkdir -p "${VCENV_ROOT}/versions/0.21.1"
	run vc-env __complete-versions installed
	assert_success
	assert_output --partial "0.22.0"
	assert_output --partial "0.21.1"
}

@test "vc-env install accepts MAJOR.MINOR alias" {
	vc-env init
	# 0.21 should resolve to the highest 0.21.x stable release and install it.
	run vc-env install 0.21
	assert_success
	# Verify at least one 0.21.x directory now exists.
	ls "${VCENV_ROOT}/versions" | grep -E "^0\.21\.[0-9]+$"
}

@test "vc-env resolve prints concrete version for alias" {
	vc-env init
	local test_ver="0.21.1"
	vc-env install "${test_ver}"

	# Installed versions win: 0.21 resolves to the installed 0.21.1 even
	# though newer 0.21.x releases exist upstream.
	run vc-env resolve 0.21
	assert_success
	assert_output "0.21.1"
}

@test "vc-env resolve with no argument uses the active global version" {
	vc-env init
	install_fake 9.9.1
	install_fake 9.9.3
	vc-env global 9.9

	run vc-env resolve
	assert_success
	assert_output "9.9.3"

	VCENV_VERSION=9.9.1 run vc-env resolve
	assert_success
	assert_output "9.9.1"
}

@test "vc-env resolve matches an explicit constraint against installed versions" {
	vc-env init
	install_fake 9.9.1
	install_fake 9.10.0

	run vc-env resolve "~9.9.0"
	assert_success
	assert_output "9.9.1"

	run vc-env resolve latest
	assert_success
	assert_output "9.10.0"
}

@test "vc-env resolve fails for an unresolved spec" {
	vc-env init
	install_fake 9.9.1

	run vc-env resolve 9.8
	assert_failure
	assert_output --partial "no installed version matches"

	run vc-env resolve 9.8.0
	assert_failure
	assert_output --partial "not installed"
}

@test "vc-env resolve fails when nothing is configured" {
	vc-env init
	run vc-env resolve
	assert_failure
	assert_output --partial "no vcluster version configured"
}

@test "vc-env resolve requires init" {
	rm -rf "${VCENV_ROOT}"
	run vc-env resolve 0.21.1
	assert_failure
	assert_output --partial "not initialized"
}

@test "vc-env resolve --install installs a missing version and prints only the version" {
	vc-env init
	run bash -c 'vc-env resolve 0.21.1 --install 2>/dev/null'
	assert_success
	assert_output "0.21.1"
	[ -x "${VCENV_ROOT}/versions/0.21.1/vcluster" ]
}

@test "vc-env resolve --help shows help text" {
	run vc-env resolve --help
	assert_success
	assert_output --partial "vc-env resolve [<version-or-alias>] [--install]"
}

@test "vc-env global accepts an alias and stores it verbatim" {
	vc-env init
	run vc-env global latest
	assert_success

	run cat "${VCENV_ROOT}/version"
	assert_output --partial "latest"
}

@test "VCENV_AUTO_INSTALL=1 auto-installs on shim invocation" {
	vc-env init
	local test_ver="0.21.1"

	# Set the global version without installing it.
	echo "${test_ver}" > "${VCENV_ROOT}/version"

	# Confirm the binary is not present yet.
	[ ! -f "${VCENV_ROOT}/versions/${test_ver}/vcluster" ]

	# Shim should install on demand.
	VCENV_AUTO_INSTALL=1 run vcluster version
	assert_success
	assert_output --regexp "([vV]ersion|[0-9]+\.[0-9]+\.[0-9])"
	[ -f "${VCENV_ROOT}/versions/${test_ver}/vcluster" ]
}

@test "vc-env install --from-file installs a local binary" {
	vc-env init
	local test_ver="0.21.1"

	# First get a real vcluster binary from GitHub via a normal install.
	vc-env install "${test_ver}"
	local src="${BATS_TMPDIR}/vcluster-copy"
	cp "${VCENV_ROOT}/versions/${test_ver}/vcluster" "${src}"

	# Compute checksum and remove the installed version.
	local sum
	sum=$(sha256sum "${src}" | awk '{print $1}')
	rm -rf "${VCENV_ROOT}/versions/${test_ver}"

	# Install again, this time from the local file with checksum validation.
	run vc-env install "${test_ver}" --from-file "${src}" --sha256 "${sum}"
	assert_success
	[ -x "${VCENV_ROOT}/versions/${test_ver}/vcluster" ]
}

@test "vc-env install --from-file aborts on wrong --sha256" {
	vc-env init
	local test_ver="0.21.1"
	local src="${BATS_TMPDIR}/fake-vcluster"
	echo "not a real binary" > "${src}"

	run vc-env install "${test_ver}" --from-file "${src}" --sha256 "0000000000000000000000000000000000000000000000000000000000000000"
	assert_failure
	assert_output --partial "checksum mismatch"
	[ ! -f "${VCENV_ROOT}/versions/${test_ver}/vcluster" ]
}

@test "vc-env with no arguments prints help" {
	run vc-env
	assert_success
	assert_output --partial "Commands:"
}

@test "vc-env --version and -v print the version" {
	run vc-env --version
	assert_success
	assert_output --regexp "[0-9]+\.[0-9]+\.[0-9]+"

	run vc-env -v
	assert_success
	assert_output --regexp "[0-9]+\.[0-9]+\.[0-9]+"
}

@test "vc-env unknown command fails and prints help" {
	run vc-env not-a-command
	assert_failure
	assert_output --partial "Unknown command: not-a-command"
	assert_output --partial "Commands:"
}

@test "vc-env init fails when VCENV_ROOT is unset" {
	run env -u VCENV_ROOT vc-env init
	assert_failure
	assert_output --partial "VCENV_ROOT not set"
}

@test "vc-env -h short-circuits install, doctor, prune, init, list and ensure" {
	local cmd
	for cmd in install doctor prune init list list-remote ensure; do
		run vc-env "${cmd}" -h
		assert_success
		assert_output --partial "Usage: vc-env ${cmd}"
	done
	# Help must not initialize or install anything.
	[ ! -d "${VCENV_ROOT}/versions" ]
}

@test "vc-env list-remote --prerelease returns versions" {
	vc-env init
	run vc-env list-remote --prerelease
	assert_success
	assert_output --regexp "[0-9]+\.[0-9]+\.[0-9]+"
}

@test "vc-env shell without an argument prints VCENV_VERSION" {
	vc-env init

	VCENV_VERSION= run vc-env shell
	assert_failure
	assert_output --partial "no shell version configured"

	VCENV_VERSION=9.9.1 run vc-env shell
	assert_success
	assert_output "9.9.1"
}

@test "vc-env ensure prints the concrete version for an installed alias" {
	vc-env init
	install_fake 9.9.1
	install_fake 9.9.4

	run vc-env ensure 9.9
	assert_success
	assert_output "9.9.4"
}

@test "vc-env ensure installs a missing version and prints only the version" {
	vc-env init
	run bash -c 'vc-env ensure 0.21.1 2>/dev/null'
	assert_success
	assert_output "0.21.1"
	[ -x "${VCENV_ROOT}/versions/0.21.1/vcluster" ]
}

@test "vc-env ensure requires a version" {
	vc-env init
	run vc-env ensure
	assert_failure
	assert_output --partial "version or alias not specified"
}

@test "vc-env exec --auto installs a missing version and runs it" {
	vc-env init
	run vc-env exec --auto 0.21.1 version
	assert_success
	assert_output --regexp "([vV]ersion|[0-9]+\.[0-9]+\.[0-9])"
	[ -x "${VCENV_ROOT}/versions/0.21.1/vcluster" ]
}

@test "vc-env exec --no-auto wins over VCENV_AUTO_INSTALL" {
	vc-env init
	VCENV_AUTO_INSTALL=1 run vc-env exec --no-auto 9.9.9 version
	assert_failure
	assert_output --partial "not installed"
	[ ! -e "${VCENV_ROOT}/versions/9.9.9" ]
}

@test "vc-env doctor --deep verifies a real binary and fails a tampered one" {
	vc-env init
	vc-env install 0.21.1
	vc-env global 0.21.1

	run vc-env doctor --deep
	assert_success
	assert_output --partial "Deep checksums"
	assert_output --partial "1 binaries verified against GitHub"

	echo "tampered" > "${VCENV_ROOT}/versions/0.21.1/vcluster"
	chmod +x "${VCENV_ROOT}/versions/0.21.1/vcluster"
	run vc-env doctor --deep
	assert_failure
	assert_output --partial "checksum mismatch"
	assert_output --partial "0.21.1"
}

@test "vc-env prune --keep-last keeps the newest unreferenced versions" {
	vc-env init
	install_fake 9.9.1
	install_fake 9.9.2
	install_fake 9.9.3

	VCENV_VERSION= VCENV_PRUNE_SCAN_ROOTS="" run vc-env prune --yes --keep-last 2
	assert_success
	assert_output --partial "removed 9.9.1"
	[ ! -d "${VCENV_ROOT}/versions/9.9.1" ]
	[ -d "${VCENV_ROOT}/versions/9.9.2" ]
	[ -d "${VCENV_ROOT}/versions/9.9.3" ]
}

@test "vc-env prune --dry-run does not remove versions even with --yes" {
	vc-env init
	install_fake 9.9.1

	VCENV_VERSION= VCENV_PRUNE_SCAN_ROOTS="" run vc-env prune --yes --dry-run
	assert_success
	assert_output --partial "would remove 9.9.1"
	[ -d "${VCENV_ROOT}/versions/9.9.1" ]
}
