# Audited rejected-validator terminal disposition
Objective: provide an explicit non-approval release for a canonical rejected-validator terminal. Branch: fix/5428-terminal-disposition; delivery: ask-on-risk.
Baseline: 1f9d5e6423e37f7d2316859045f379ba9b5d8c3a. Runner: focused Go tests, full affected packages and compiled CLI fixture.

## Specs
- S1: User request: "Vamos a ver porque ese bloqueo y arreglarlo"; target selection: "2", "Na, hazlo tu desde acá", "Work/gentle-ai".
- S2: Approved proposal: "liberación explícita y auditada de la autoridad activa, conservando el rechazo histórico y su evidencia. No aprobaría C3, no devolvería intentos de corrección y no iniciaría automáticamente otra revisión." Human answer: "Si".
- S3: Approval question: "¿Autorizas esa disposición no aprobatoria y añadir status:approved al issue Gentleman-Programming/gentle-ai#5428 para implementarla?" Human answer: "Si".
- Scope: implement only an exact-selector, evidence-bound terminal disposition for canonical rejected-validator authority without graph dependency ambiguity. Do not remove stop/recovery/abandon safeguards. Check rejected commands leave state unchanged; successful disposition retains historical verdict and audit and does not change source.
- Non-goals: no historical verdict replacement, correction replenishment, automatic review restart, RDD change or operation on the consumer C3 lineage. Same-target future review policy is not changed deliberately by this feature.

## Tasks
- T1 | S1-S3 | inline | in_progress | Reproduce current-main refusal, add regression and minimal disposition; normalize, focused/package/compiled-CLI checks and independent verification; commit pending.

## Log
- L1: "Vamos a ver porque ese bloqueo y arreglarlo"
- L2: User chose provider "2", requested "Na, hazlo tu desde acá" and identified "Work/gentle-ai". Absolute root verified.
- L3: Read-only mapper mv4bc3v7-i-sh6e and parent spotcheck establish deliberately enforced unchanged escalated stop, terminal abandon exclusion and missing disposition class; canonical failed regression is valid authority, not corruption.
- L4: Explicit "Si" approved S2-S3. Target gh actor ADMIN freshly verified; issue5428 status:approved added and readback confirmed. Root untracked ODD/telemetry work preserved; isolated worktree based clean committed main.
- L5: Route inline, no writer trigger. Risk items1,2,3,4: stored authority release, exact authorization, consumed contracts, exclusive lock/CAS. Forecast320-380 authored diff lines with existing executor/fixtures reused; hard400 gate before exceeding. Existing main guards unchanged. Independent verifier required. No source fix or test result claimed yet.

- L6: RED job28 initially failed to compile (wrong test API, corrected); RED job29 reproduced unavailable disposition for valid canonical authority. GREEN job30 passed the focused Go API tests. Compiled-CLI RED job31 reproduced missing selector discovery; it also reported an isolated Go module-cache cleanup permission failure. Test build now requests writable module-cache files. Compiled-CLI GREEN job32 passed. Final focused job33 covers normalized source; outcome pending when this log was written.
- L7: Exact-selector plan admits only canonical rejected-validator escalation with consumed correction, no recovery edge/dependency or inspection diagnostics. Existing executor retains verdict bytes and audit, authenticates digest/inventory and revision under lock. Source guards for ordinary stop/recovery/abandon unchanged. No new model transport/schema or real consumer authority operations.
- L8: Current forecast approximately345 authored additions/deletions including new tests and this log, below400. Full affected packages and independent read-only verification pending. Rollback: revert this work unit's selector eligibility/discovery/help/docs/tests; existing executor and historical repair remain unchanged. No publication, release, installation or completed-task claim.

- L9: Final focused job33 PASS for both named package selections and whitespace check. Independent verifier mv4c9yy9-j-4hef bounded S1-S3 PASS, no reproduced candidate-caused blocker. Independent focused tests (including compiled CLI) exited0: reviewtransaction1.275s/cli3.994s. Before/after integrity2136files unchanged. Diff341lines (+334/-7), six authorized paths, below400.
- L10: Full affected-package command ran once in isolated offline scratch and was interrupted at240s with emptylog; no packagePASS/FAIL observed, no remaining processes, not retried. Thus broad regression proof remains unavailable. Existing historical/content-mismatch protections inspected, not fully regression-tested. Canonical-invalid/incomplete negative behavior inspected without separate new runtime fixture. WriterRED not independently repeated. No source changes from verifier.
- L11: Needs human decision on proceeding with explicitly limited proof. No work-unit commit, native review closure, provider publication or application to real consumer C3 yet; RDD remainsON and historical C3 escalated. No further suite automatically launched.

- L12: Human answered "si" to continuing with local commit and native review while explicitly accepting the unavailable full affected-package proof. Scoped verification remains PASS only in its stated scope; no broader green claim, publication, installation or consumer disposition is authorized.
