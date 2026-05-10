# Skill Registry — judo-app

> Generated: 2026-05-10 | Project: judo-app | Stack: Go + Wails + Angular

## User Skills (C:\Users\Admin\.config\opencode\skills)

### branch-pr
**Trigger**: creating, opening, or preparing PRs for review
**Path**: `C:\Users\Admin\.config\opencode\skills\branch-pr\SKILL.md`
**Compact Rules**:
- Every PR MUST link an issue with `status:approved` label — no exceptions
- Every PR MUST have exactly one `type:*` label
- Branch naming: `type/description` — must match `^(feat|fix|chore|docs|style|refactor|perf|test|build|ci|revert)\/[a-z0-9._-]+$`
- PR body must contain: `Closes #N`, PR type checkbox, summary, changes table, test plan
- Conventional commit format: `type(scope): description`; no `Co-Authored-By` trailers
- Automated checks: issue reference, `status:approved`, `type:*` label, shellcheck

---

### chained-pr
**Trigger**: PRs over 400 lines, stacked PRs, review slices
**Path**: `C:\Users\Admin\.config\opencode\skills\chained-pr\SKILL.md`
**Compact Rules**:
- Split any PR exceeding 400 changed lines; target ≤60 min reviewer effort per PR
- One deliverable work unit per PR; tests/docs travel with the unit they verify
- Every chained PR must include a dependency diagram marking the current PR with `📍`
- Feature Branch Chain: create draft tracker PR; child #1 targets tracker, later children target their immediate parent
- Do not mix chain strategies once chosen
- Treat polluted diffs as base bugs; retarget/rebase before PR

---

### chained-pr
**Trigger**: PRs over 400 lines, stacked PRs, review slices
**Path**: `C:\Users\Admin\.config\opencode\skills\chained-pr\SKILL.md`

---

### cognitive-doc-design
**Trigger**: writing guides, READMEs, RFCs, onboarding, architecture, or review-facing docs
**Path**: `C:\Users\Admin\.config\opencode\skills\cognitive-doc-design\SKILL.md`
**Compact Rules**:
- Lead with the answer — decision/action first, context after
- Progressive disclosure: happy path → details → edge cases → references
- Chunk content into small sections; flat lists short
- Signpost with headings, labels, callouts, summaries
- Prefer tables, checklists, examples over prose that must be remembered
- For PR docs: state what to review first, what is out of scope, link prior/next PR in chain

---

### comment-writer
**Trigger**: PR feedback, issue replies, reviews, Slack messages, or GitHub comments
**Path**: `C:\Users\Admin\.config\opencode\skills\comment-writer\SKILL.md`
**Compact Rules**:
- Start with the actionable point; no PR recap first
- Sound like a thoughtful teammate; keep to 1-3 short paragraphs or a tight bullet list
- Always explain WHY when asking for a change
- Match thread language; in Spanish use Rioplatense/voseo (`podés`, `tenés`, `fijate`)
- No em dashes — use commas, periods, or parentheses instead

---

### go-testing
**Trigger**: Go tests, go test coverage, Bubbletea teatest, golden files
**Path**: `C:\Users\Admin\.config\opencode\skills\go-testing\SKILL.md`
**Compact Rules**:
- Use table-driven tests with `t.Run(tt.name, ...)` for multiple cases
- Test behavior and state transitions, not implementation details
- Use `t.TempDir()` for filesystem tests; never rely on real home directory
- Integration tests that run external commands must be skippable with `testing.Short()`
- For Bubbletea: test `Model.Update()` directly for state; `teatest` for interactive flows only
- Golden files must be deterministic; update only via `-update` flag, rerun without it after

---

### issue-creation
**Trigger**: creating GitHub issues, bug reports, or feature requests
**Path**: `C:\Users\Admin\.config\opencode\skills\issue-creation\SKILL.md`
**Compact Rules**:
- Blank issues are disabled; must use Bug Report or Feature Request template
- Every issue auto-gets `status:needs-review`; maintainer must add `status:approved` before any PR
- Questions go to Discussions, not issues
- Search for duplicates before creating
- Bug reports need: pre-flight checks, description, steps to reproduce, expected/actual behavior, OS, agent, shell

---

### judgment-day
**Trigger**: judgment day, dual review, adversarial review, juzgar
**Path**: `C:\Users\Admin\.config\opencode\skills\judgment-day\SKILL.md`
**Compact Rules**:
- Resolve project skills from registry; inject same `Project Standards` into both judge and fix prompts
- Launch two blind judges in parallel with identical target and criteria
- Classify warnings as `WARNING (real)` only if normal use can trigger them; otherwise INFO
- Ask before fixing Round 1 confirmed issues
- Re-launch both judges in parallel after any fix agent runs, before done/commit
- Terminal states only: `JUDGMENT: APPROVED` or `JUDGMENT: ESCALATED`
- After 2 fix iterations with remaining issues, ask user whether to continue

---

### skill-creator
**Trigger**: new skills, agent instructions, documenting AI usage patterns
**Path**: `C:\Users\Admin\.config\opencode\skills\skill-creator\SKILL.md`
**Compact Rules**:
- Check `docs/skill-style-guide.md` first; use inline fallback only if unavailable
- `description` must be one physical line, quoted, YAML-safe, trigger words first, ≤250 chars
- Required frontmatter: `name`, `description`, `license`, `metadata.author`, `metadata.version`
- Target 180-450 body tokens; hard max 1000; put examples/schemas in `assets/`, edge cases in `references/`
- References must point to local files
- Register project skills in `AGENTS.md`

---

### work-unit-commits
**Trigger**: implementation, commit splitting, chained PRs, keeping tests and docs with code
**Path**: `C:\Users\Admin\.config\opencode\skills\work-unit-commits\SKILL.md`
**Compact Rules**:
- Commit by deliverable behavior, not file type
- Tests belong in the same commit as the behavior they verify
- Docs belong with the user-visible change they explain
- Each commit should leave the repo meaningful on its own; rollback must be clean
- SDD >400-line forecast → group commits into chained PR slices before implementation
- Commit message explains outcome, not file list

---

## SDD Skills (user-level — excluded from registry per scan rules)

The following SDD skills are installed but excluded from this registry per scan rules (skip `sdd-*`):
`sdd-init`, `sdd-explore`, `sdd-propose`, `sdd-spec`, `sdd-design`, `sdd-tasks`, `sdd-apply`, `sdd-verify`, `sdd-archive`, `sdd-onboard`

Also excluded: `_shared`, `skill-registry`

---

## Convention Files

- `C:\Users\Admin\.config\opencode\AGENTS.md` — global Gentle AI persona, engram protocol, skill loading rules

---

## Project Skills

None — no project-level skill directory found at `Q:\Programacion\Proyectos\judo-app\.agent/skills` or similar.
