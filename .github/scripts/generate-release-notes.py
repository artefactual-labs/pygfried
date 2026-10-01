"""Generate release notes from commits since the preceding stable tag."""

import argparse
import os
import re
import subprocess

VERSION = re.compile(r"v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)")


def git(*args: str) -> str:
    return subprocess.run(
        ["git", *args],
        check=True,
        capture_output=True,
        text=True,
    ).stdout.strip()


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("tag", help="Release tag in vX.Y.Z format")
    parser.add_argument("ref", help="Commit or ref to release")
    parser.add_argument(
        "--repository",
        default=os.environ.get("GITHUB_REPOSITORY", "artefactual-labs/pygfried"),
        help="GitHub repository in OWNER/REPO format",
    )
    args = parser.parse_args()
    match = VERSION.fullmatch(args.tag)
    if match is None:
        parser.error("tag must use stable vX.Y.Z format")
    version = tuple(map(int, match.groups()))
    release_sha = git("rev-parse", "--verify", f"{args.ref}^{{commit}}")

    candidates = []
    for tag in git("tag", "--merged", release_sha).splitlines():
        match = VERSION.fullmatch(tag)
        if match is not None:
            tag_version = tuple(map(int, match.groups()))
            if tag_version < version:
                candidates.append((tag_version, tag))
    previous_tag = max(candidates)[1] if candidates else None
    commit_range = (
        f"{previous_tag}..{release_sha}" if previous_tag is not None else release_sha
    )
    repository_url = f"https://github.com/{args.repository}"

    print("## What's Changed\n")
    commits = git("log", "--reverse", "--format=%H%x00%h%x00%s", commit_range)
    for commit in commits.splitlines():
        sha, short_sha, subject = commit.split("\x00", 2)
        subject = re.sub(r"([\\`*_{}\[\]<>])", r"\\\1", subject)
        print(f"* {subject} ([{short_sha}]({repository_url}/commit/{sha}))")
    if not commits:
        print("No new commits since the preceding release.")

    if previous_tag is None:
        changelog_url = f"{repository_url}/commits/{args.tag}"
    else:
        changelog_url = f"{repository_url}/compare/{previous_tag}...{args.tag}"
    print(f"\n**Full Changelog**: {changelog_url}")


if __name__ == "__main__":
    main()
