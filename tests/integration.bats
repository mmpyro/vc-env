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
	assert_output --regexp "^[0-9]+\.[0-9]+\.[0-9]+$"
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
