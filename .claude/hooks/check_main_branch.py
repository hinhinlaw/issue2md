#!/usr/bin/env python3
"""
Claude Code Hook: Main Branch Protection
=========================================
This hook runs as a PreToolUse hook for Edit, Write, and MultiEdit tools.
It prevents file modifications when on the main branch.

Read more about hooks here: https://docs.anthropic.com/en/docs/claude-code/hooks

Configuration for hooks.json:
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "python3 /path/to/check-main-branch.py"
          }
        ]
      }
    ]
  }
}
"""

import json
import subprocess
import sys


def _get_current_branch() -> str:
    """Get the current git branch name."""
    try:
        result = subprocess.run(
            ["git", "rev-parse", "--abbrev-ref", "HEAD"],
            capture_output=True,
            text=True,
            check=True,
            timeout=5
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError:
        # Not in a git repository or git command failed
        return ""
    except subprocess.TimeoutExpired:
        return ""
    except FileNotFoundError:
        # git not installed
        return ""


def _is_main_branch(branch: str) -> bool:
    """Check if the branch is a main branch (main or master)."""
    return branch.lower() in ("main", "master")


def main():
    try:
        input_data = json.load(sys.stdin)
    except json.JSONDecodeError as e:
        print(f"Error: Invalid JSON input: {e}", file=sys.stderr)
        sys.exit(1)

    tool_name = input_data.get("tool_name", "")

    # Check if this is an Edit, Write, or MultiEdit tool
    if tool_name not in ("Edit", "Write", "MultiEdit"):
        sys.exit(0)

    # Get current git branch
    current_branch = _get_current_branch()

    # If not in a git repository, allow the operation
    if not current_branch:
        sys.exit(0)

    # Block operations on main/master branch
    if _is_main_branch(current_branch):
        print(
            f"🚫 Cannot modify files on '{current_branch}' branch.\n"
            f"   Please switch to a feature branch before making changes.",
            file=sys.stderr
        )
        # Exit code 2 blocks the tool call and shows stderr to Claude
        sys.exit(2)

    # Allow operation on other branches
    sys.exit(0)


if __name__ == "__main__":
    main()