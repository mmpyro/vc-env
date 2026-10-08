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
	assert_output --partial "vc-env completion <shell>"
	assert_output --partial "bash"
	assert_output --partial "zsh"
	assert_output --partial "fish"
	assert_output --partial "powershell"
}

@test "vc-env autocompletion still works as deprecated alias" {
	run vc-env autocompletion
	assert_success
	assert_output --partial "_vc_env_completions()"
	assert_output --partial "complete -F _vc_env_completions vc-env"
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

	run vc-env resolve 0.21
	assert_success
	assert_output --partial "0.21.1"
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
