#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/.." && pwd)"

backend_dir="${BACKEND_DIR:-backend}"
coverage_min="${BACKEND_COVERAGE_MIN:-}"
coverage_profile="${BACKEND_COVERAGE_PROFILE:-}"
cleanup_profile=0

if [[ "${backend_dir}" != /* ]]; then
	backend_dir="${repo_root}/${backend_dir}"
fi

if [[ -n "${coverage_profile}" && "${coverage_profile}" != /* ]]; then
	coverage_profile="${repo_root}/${coverage_profile}"
fi

if [[ -z "${coverage_profile}" ]]; then
	coverage_profile="$(mktemp "${TMPDIR:-/tmp}/backend-coverage.XXXXXX.out")"
	cleanup_profile=1
else
	mkdir -p "$(dirname -- "${coverage_profile}")"
fi

cd "${backend_dir}"

go test -covermode=atomic -coverpkg=./... -coverprofile="${coverage_profile}" ./...

total_line="$(go tool cover -func="${coverage_profile}" | awk '/^total:/ { print; exit }')"
total_coverage="$(printf '%s\n' "${total_line}" | awk '{ gsub("%", "", $NF); print $NF }')"

if [[ -z "${total_line}" || -z "${total_coverage}" ]]; then
	printf 'failed to parse total coverage from %s\n' "${coverage_profile}" >&2
	if [[ "${cleanup_profile}" -eq 1 ]]; then
		rm -f "${coverage_profile}"
	fi
	exit 1
fi

printf '%s\n' "${total_line}"

if [[ -n "${coverage_min}" ]]; then
	if awk -v total="${total_coverage}" -v min="${coverage_min}" 'BEGIN { exit (total + 0 >= min + 0 ? 0 : 1) }'; then
		printf 'backend coverage ok: %s%% >= %s%%\n' "${total_coverage}" "${coverage_min}"
	else
		printf 'backend coverage check failed: %s%% < %s%%\n' "${total_coverage}" "${coverage_min}" >&2
		if [[ "${cleanup_profile}" -eq 0 ]]; then
			printf 'coverage profile: %s\n' "${coverage_profile}" >&2
		else
			rm -f "${coverage_profile}"
		fi
		exit 1
	fi
fi

if [[ "${cleanup_profile}" -eq 1 ]]; then
	rm -f "${coverage_profile}"
else
	printf 'coverage profile: %s\n' "${coverage_profile}"
fi
