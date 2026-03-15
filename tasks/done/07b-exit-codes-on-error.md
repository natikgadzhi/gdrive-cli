# Task 07b: Define exit code behavior for command errors

**Phase**: Follow-up (from PR #7 /simplify review)
**Depends on**: 07 (Auth commands)
**Blocks**: none — should be addressed in error handling audit (task 15)
**Priority**: medium

## Description

Auth commands (and future commands) currently return exit code 0 even when the operation fails (e.g., not authenticated, login failed). The error is reported as JSON `{"status": "error", "message": "..."}` to stdout, but `RunE` returns `nil` since the JSON write succeeded.

This means callers cannot use the exit code to detect failures — they must parse the JSON `status` field.

## Decision needed

Option A: Keep exit code 0 for all commands, require JSON parsing (current Python behavior).
Option B: Return non-zero exit code on logical errors while still printing JSON error to stdout.

If Option B: use Cobra's `SilenceErrors: true` on root command so Cobra doesn't print its own error message, then return a sentinel error from RunE after printing the JSON.

## Origin

/simplify review feedback on PR #7 (auth commands).

## Acceptance Criteria

- [ ] Decision documented
- [ ] All commands follow the chosen convention consistently
- [ ] Error handling audit (task 15) incorporates this decision

## Status
backlog — fold into task 15 (error handling audit)
