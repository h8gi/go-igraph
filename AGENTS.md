# Project development policy

- Prefer the cleanest current API, correctness, safety, and maintainability over
  backward compatibility.
- Breaking changes are acceptable when they materially improve the design.
- Do not add compatibility shims or preserve legacy APIs unless a task
  explicitly requires it.
- Propagate C/igraph errors through idiomatic Go errors and make ownership of C
  resources explicit.

## Pull request workflow

- Branch from the current `main`, use a short-lived `codex/issue-*` branch, and
  target the protected `main` branch directly. Do not add a long-lived
  `develop` branch without a separate release-management requirement.
- Keep each pull request to one coherent issue and include `Closes #<issue>`.
- Open the pull request as a draft, review its complete diff, then mark it ready
  only after the checks below pass. Squash-merge it and delete the branch after
  required GitHub checks succeed and review conversations are resolved.
- Run `make verify` before marking a pull request ready for review.
- Document whether public inputs are borrowed or copied and whether returned
  values are Go-owned.
- Cover initialization failure, upstream errors, early returns, empty values,
  and use after `Close` when the binding can exercise those paths.
