#!/usr/bin/env python3

import subprocess
import re
import sys
import argparse

def run_command(command):
    result = subprocess.run(command, check=False, shell=True, capture_output=True, text=True)
    if result.returncode != 0:
        raise Exception(f"Command failed: {command}\n{result.stderr}")  # pylint: disable=broad-exception-raised

    return result.stdout.strip()


def get_latest_version():
    try:
        run_command("git fetch --tags")
        tags = run_command("git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname").split()
        if not tags:
            raise ValueError("No valid tags found")

        return tags[0]
    except Exception as exception:
        raise Exception(f"Failed to get the latest version: {exception}") from exception  # pylint: disable=broad-exception-raised


def parse_version(version):
    match = re.match(r'v?(\d+)\.(\d+)\.(\d+)', version)
    if not match:
        raise ValueError(f"Invalid version format: {version}")
    return tuple(map(int, match.groups()))


def get_next_version(current_version, version_type) -> str:
    major, minor, patch = parse_version(current_version)

    if version_type == 'major':
        major += 1
        minor = 0
        patch = 0
    elif version_type == 'minor':
        minor += 1
        patch = 0
    elif version_type == 'patch':
        patch += 1
    else:
        raise ValueError("Invalid version type")

    return f"v{major}.{minor}.{patch}"


def validate_version(version: str) -> bool:
    if not validate_version_format(version):
        raise ValueError(f"Invalid version format: {version}")

    if not validate_version_unique(version):
        raise ValueError(f"Version already tagged: {version}")

    return True


def validate_version_format(version: str) -> bool:
    if re.match(r'^v[0-9]+\.[0-9]+\.[0-9]+$', version) is None:
        return False

    return True


def validate_version_unique(version: str) -> bool:
    return version not in run_command("git tag --list 'v[0-9]*.[0-9]*.[0-9]*'").split()


def main(version_input):
    try:
        if re.match(r'^v[0-9]+\.[0-9]+\.[0-9]+$', version_input):
            next_version = version_input
        else:
            next_version = get_next_version(get_latest_version(), version_input)

        validate_version(next_version)

        print(next_version)
    except Exception as exception:  # pylint: disable=broad-exception-caught
        print(f"An error occurred: {exception}", file=sys.stderr)

        sys.exit(1)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(formatter_class=argparse.RawTextHelpFormatter)
    parser.add_argument('version_input', type=str, help="Version input or type (major, minor, patch)")

    args = parser.parse_args()
    main(args.version_input)
