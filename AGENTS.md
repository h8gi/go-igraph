# Project development policy

- Prefer the cleanest current API, correctness, safety, and maintainability over
  backward compatibility.
- Breaking changes are acceptable when they materially improve the design.
- Do not add compatibility shims or preserve legacy APIs unless a task
  explicitly requires it.
- Propagate C/igraph errors through idiomatic Go errors and make ownership of C
  resources explicit.
