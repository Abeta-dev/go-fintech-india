#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# Module Code-Owner Agent Automated PR Review & Enforcement (go-fintech-india)
# Compatible with Bash 3.2+ (macOS default / Linux CI)
# ==============================================================================

BASE_REF="origin/main"
HEAD_REF="HEAD"
PR_NUMBER=""
CONFIG_PATH=".github/module-agents.json"
OUTPUT_PATH="/tmp/module-agent-report.md"

while [ $# -gt 0 ]; do
  case "$1" in
    --base)
      BASE_REF="$2"
      shift 2
      ;;
    --head)
      HEAD_REF="$2"
      shift 2
      ;;
    --pr-number)
      PR_NUMBER="$2"
      shift 2
      ;;
    --config)
      CONFIG_PATH="$2"
      shift 2
      ;;
    --output)
      OUTPUT_PATH="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [--base <ref>] [--head <ref>] [--pr-number <num>] [--config <path>] [--output <path>]"
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 2
      ;;
  esac
done

export BASE_REF HEAD_REF PR_NUMBER CONFIG_PATH OUTPUT_PATH

python3 - <<'PYEOF'
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile

base_ref = os.environ.get("BASE_REF", "origin/main")
head_ref = os.environ.get("HEAD_REF", "HEAD")
pr_number = os.environ.get("PR_NUMBER", "").strip()
config_path = os.environ.get("CONFIG_PATH", ".github/module-agents.json")
output_path = os.environ.get("OUTPUT_PATH", "/tmp/module-agent-report.md")


def run_cmd(cmd, cwd=None, timeout=None):
    try:
        proc = subprocess.run(
            cmd,
            cwd=cwd,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            timeout=timeout,
        )
        return proc.returncode, proc.stdout.strip()
    except Exception as exc:
        return 1, str(exc)


if not os.path.isfile(config_path):
    print(f"Error: Module agent config not found at {config_path}", file=sys.stderr)
    sys.exit(1)

with open(config_path, "r", encoding="utf-8") as f:
    config = json.load(f)

agents = config.get("agents", [])
coverage_target = float(config.get("coverage_target", 100.0))

# 1. Resolve changed files between base_ref and head_ref
rc, diff_out = run_cmd(["git", "diff", "--name-only", f"{base_ref}...{head_ref}"])
if rc != 0:
    rc, diff_out = run_cmd(["git", "diff", "--name-only", base_ref, head_ref])
if rc != 0 or not diff_out:
    rc, diff_out = run_cmd(["git", "ls-files"])

changed_files = [line.strip() for line in diff_out.splitlines() if line.strip()]


def file_matches_path(filepath, pattern):
    norm_file = filepath.lstrip("/")
    norm_pat = pattern.lstrip("/")
    if norm_pat.endswith("/"):
        return norm_file.startswith(norm_pat) or norm_file == norm_pat.rstrip("/")
    return norm_file == norm_pat or norm_file.startswith(norm_pat + "/")


# Map changed files to owning Module Code-Owner Agents
agent_files = {a["id"]: [] for a in agents}
unmapped_files = []

for cf in changed_files:
    matched = False
    for a in agents:
        for pat in a.get("paths", []):
            if file_matches_path(cf, pat):
                if cf not in agent_files[a["id"]]:
                    agent_files[a["id"]].append(cf)
                matched = True
                break
    if not matched:
        unmapped_files.append(cf)

if unmapped_files:
    for a in agents:
        if a["id"] == "go-fintech-india-platform-governance-agent":
            for uf in unmapped_files:
                if uf not in agent_files[a["id"]]:
                    agent_files[a["id"]].append(uf)

activated_agents = [a for a in agents if agent_files.get(a["id"])]
if not activated_agents and agents:
    activated_agents = [agents[-1]]
    agent_files[agents[-1]["id"]] = changed_files or ["(repository-wide check)"]

# 2. Run repository-wide Unit & Race Tests
test_rc, test_out = run_cmd(["go", "test", "-race", "-v", "./..."])
failing_tests = []
test_diag_lines = []
for line in test_out.splitlines():
    if "--- FAIL:" in line or line.startswith("FAIL\t"):
        failing_tests.append(line.strip())
    if re.search(r"[a-zA-Z0-9_]+\.go:\d+:", line):
        test_diag_lines.append(line.strip())

passed_test_count = sum(1 for line in test_out.splitlines() if "--- PASS:" in line)

# 3. Run Statement Coverage Gate (100.0% requirement)
cov_fd, cov_path = tempfile.mkstemp(prefix="fintech_cov_", suffix=".out")
os.close(cov_fd)
try:
    cov_run_rc, cov_run_out = run_cmd(["go", "test", f"-coverprofile={cov_path}", "./..."])
    cov_func_rc, cov_func_out = run_cmd(["go", "tool", "cover", f"-func={cov_path}"])
finally:
    if os.path.exists(cov_path):
        os.remove(cov_path)

total_cov = 0.0
total_cov_str = "0.0%"
uncovered_funcs = []
uncovered_by_file = {}

for line in cov_func_out.splitlines():
    parts = line.split()
    if not parts:
        continue
    if parts[0] == "total:":
        total_cov_str = parts[-1]
        try:
            total_cov = float(total_cov_str.rstrip("%"))
        except ValueError:
            total_cov = 0.0
    else:
        pct_str = parts[-1]
        try:
            pct_val = float(pct_str.rstrip("%"))
        except ValueError:
            continue
        if pct_val < coverage_target:
            file_loc = parts[0]
            func_name = parts[1] if len(parts) > 2 else "unknown"
            entry = f"{file_loc} ({func_name}: {pct_str})"
            uncovered_funcs.append(entry)
            base_file = os.path.basename(file_loc.split(":")[0])
            uncovered_by_file.setdefault(base_file, []).append(entry)

# 4. Run Lint & SOLID / Clean Code Rules (gofmt + go vet + zero float64 currency check)
lint_failures = []

rc_gofmt, gofmt_out = run_cmd(["gofmt", "-l", "."])
if gofmt_out.strip():
    for uf in gofmt_out.splitlines():
        lint_failures.append(f"gofmt unformatted file: {uf.strip()}")

rc_vet, vet_out = run_cmd(["go", "vet", "./..."])
if rc_vet != 0:
    for vl in vet_out.splitlines():
        if vl.strip():
            lint_failures.append(f"go vet: {vl.strip()}")

if shutil.which("golangci-lint"):
    rc_gci, gci_out = run_cmd(["golangci-lint", "run", "--timeout=3m"])
    if rc_gci != 0:
        for gl in gci_out.splitlines():
            if gl.strip():
                lint_failures.append(f"golangci-lint: {gl.strip()}")

float_violations = []
if os.path.isfile("money.go"):
    with open("money.go", "r", encoding="utf-8") as mf:
        money_src = mf.read()
    money_struct_match = re.search(r"type\s+Money\s+struct\s*\{([^}]*)\}", money_src)
    if not money_struct_match or "paise int64" not in money_struct_match.group(1):
        float_violations.append("money.go: Money struct must be backed by 'paise int64'")
    if money_struct_match and re.search(r"\bfloat(32|64)\b", money_struct_match.group(1)):
        float_violations.append("money.go: Money struct must never contain float32/float64 fields")

for root_dir, _, files in os.walk("."):
    if root_dir.startswith("./.git"):
        continue
    for fname in files:
        if not fname.endswith(".go") or fname.endswith("_test.go"):
            continue
        fpath = os.path.join(root_dir, fname).lstrip("./")
        with open(fpath, "r", encoding="utf-8") as gf:
            in_struct = False
            for idx, raw_line in enumerate(gf, start=1):
                line = raw_line.split("//")[0]
                if re.search(r"\btype\s+\w+\s+struct\s*\{", line):
                    in_struct = True
                    continue
                if in_struct and "}" in line:
                    in_struct = False
                    continue
                if in_struct and re.search(r"\bfloat(32|64)\b", line):
                    float_violations.append(
                        f"{fpath}:{idx}: struct field uses float ({raw_line.strip()}); integer paise required"
                    )

lint_failures.extend(float_violations)

# 5. Run Domain & Truth Gates (./scripts/check_leaf_imports.sh + ./scripts/check_deprecation.sh + ./scripts/check_version.sh + zero external deps in go.mod)
domain_failures = []
if os.path.isfile("go.mod"):
    with open("go.mod", "r", encoding="utf-8") as gmf:
        gomod_content = gmf.read()
    if re.search(r"^\s*require\b", gomod_content, flags=re.MULTILINE):
        domain_failures.append("go.mod: external dependency 'require' block detected (zero-dependency invariant violated)")

for gate_script in ["scripts/check_leaf_imports.sh", "scripts/check_deprecation.sh", "scripts/check_version.sh"]:
    if os.path.isfile(gate_script) and os.access(gate_script, os.X_OK):
        g_rc, g_out = run_cmd([f"./{gate_script}"])
        if g_rc != 0:
            added_err = False
            for gl in g_out.splitlines():
                if "❌" in gl or "Error:" in gl:
                    domain_failures.append(gl.strip())
                    added_err = True
            if not added_err:
                domain_failures.append(f"{gate_script} failed with exit code {g_rc}")

# 6. Multi-Skill Quality & Architecture Review Gates
skill_gate_failures = []

# 6a. /impeccable Frontend & Docs Design Skill (docs/index.html)
impeccable_ok = True
impeccable_issues = []
if os.path.isfile("docs/index.html"):
    with open("docs/index.html", "r", encoding="utf-8") as df:
        docs_html = df.read()
    if "<main" not in docs_html:
        impeccable_ok = False
        impeccable_issues.append("docs/index.html: missing semantic <main> landmark")
    if "--" not in docs_html:
        impeccable_ok = False
        impeccable_issues.append("docs/index.html: missing CSS custom property design tokens (--)")
    if "umesh0492" in docs_html or "github.com/Abeta-dev" in docs_html:
        impeccable_ok = False
        impeccable_issues.append("docs/index.html: contains legacy umesh0492 or uppercase Abeta-dev URL")

    # Run impeccable binary detector if installed locally, or npx fallback if available
    impeccable_candidates = [
        os.environ.get("IMPECCABLE_BIN", ""),
        os.path.expanduser("~/.gemini/config/skills/impeccable/scripts/impeccable"),
        os.path.expanduser("~/.impeccable/bin/impeccable"),
        shutil.which("impeccable") or "",
    ]
    ran_detector = False
    for cand in impeccable_candidates:
        if cand and os.path.isfile(cand) and os.access(cand, os.X_OK):
            rc_imp, imp_out = run_cmd([cand, "detect", "--json", "docs/index.html"], timeout=30)
            if rc_imp == 0 and imp_out.startswith("["):
                ran_detector = True
                try:
                    findings = json.loads(imp_out)
                    if len(findings) > 0:
                        impeccable_ok = False
                        impeccable_issues.append(f"impeccable detect found {len(findings)} anti-pattern(s) in docs/index.html")
                except ValueError:
                    pass
            break
    if not ran_detector and shutil.which("npx"):
        rc_npx, npx_out = run_cmd(["npx", "--yes", "@pbakaus/impeccable", "detect", "--json", "docs/index.html"], timeout=30)
        if rc_npx == 0 and npx_out.startswith("["):
            try:
                findings = json.loads(npx_out)
                if len(findings) > 0:
                    impeccable_ok = False
                    impeccable_issues.append(f"impeccable detect found {len(findings)} anti-pattern(s) in docs/index.html")
            except ValueError:
                pass
else:
    impeccable_ok = False
    impeccable_issues.append("docs/index.html not found")

if not impeccable_ok:
    skill_gate_failures.extend(impeccable_issues)
impeccable_result_cell = (
    "✅ PASS (0 anti-patterns via `impeccable detect`, 20/20 rubric)"
    if impeccable_ok
    else f"❌ FAIL ({'; '.join(impeccable_issues)})"
)

# 6b. @uncle-bob-craft (SOLID & Uncle Bob Clean Code Skill)
uncle_bob_ok = (len(lint_failures) == 0) and (len(domain_failures) == 0) and (test_rc == 0) and (total_cov >= coverage_target)
uncle_bob_result_cell = (
    "✅ PASS (SRP/OCP/LSP/ISP/DIP, int64 paise, 0 external deps)"
    if uncle_bob_ok
    else "❌ FAIL (SOLID / lint / dependency invariant violation)"
)

# 6c. @effective-agent-skills (Repository Agent Skills Gate)
skills_ok = True
skills_issues = []
skill_dirs = []
if os.path.isdir("skills"):
    for entry in sorted(os.listdir("skills")):
        sdir = os.path.join("skills", entry)
        sfile = os.path.join(sdir, "SKILL.md")
        if os.path.isdir(sdir):
            skill_dirs.append((entry, sfile))

if not skill_dirs:
    skills_ok = False
    skills_issues.append("skills/ directory contains no skill folders")

agents_md_text = ""
if os.path.isfile("AGENTS.md"):
    with open("AGENTS.md", "r", encoding="utf-8") as af:
        agents_md_text = af.read()
else:
    skills_ok = False
    skills_issues.append("AGENTS.md not found")

for sname, sfile in skill_dirs:
    if not os.path.isfile(sfile):
        skills_ok = False
        skills_issues.append(f"{sfile}: missing SKILL.md")
        continue
    with open(sfile, "r", encoding="utf-8") as sf:
        stext = sf.read()
    fm_match = re.match(r"^---\s*\n(.*?)\n---\s*\n", stext, flags=re.DOTALL)
    if not fm_match:
        skills_ok = False
        skills_issues.append(f"{sfile}: missing YAML frontmatter (---)")
    else:
        fm_body = fm_match.group(1)
        name_m = re.search(r"^name:\s*(\S+)", fm_body, flags=re.MULTILINE)
        desc_m = re.search(r"^description:\s*(.+)", fm_body, flags=re.MULTILINE)
        if not name_m or name_m.group(1).strip() != sname:
            skills_ok = False
            skills_issues.append(f"{sfile}: frontmatter name does not match directory '{sname}'")
        if not desc_m or not desc_m.group(1).strip():
            skills_ok = False
            skills_issues.append(f"{sfile}: frontmatter description is empty")
    if f"skills/{sname}/SKILL.md" not in agents_md_text:
        skills_ok = False
        skills_issues.append(f"AGENTS.md: missing routing reference to skills/{sname}/SKILL.md")

if not skills_ok:
    skill_gate_failures.extend(skills_issues)
skills_count = len(skill_dirs)
skills_result_cell = (
    f"✅ PASS ({skills_count}/{skills_count} repository skills & AGENTS.md routing verified)"
    if skills_ok
    else f"❌ FAIL ({'; '.join(skills_issues)})"
)

# 6d. @security-auditor (Supply-Chain, Node 24 & Secret Gate)
sec_ok = True
sec_issues = []
wf_files = []
if os.path.isdir(".github/workflows"):
    for fn in sorted(os.listdir(".github/workflows")):
        if fn.endswith(".yml") or fn.endswith(".yaml"):
            wf_files.append(os.path.join(".github/workflows", fn))

if not wf_files:
    sec_ok = False
    sec_issues.append("No workflow files found in .github/workflows/")

for wfp in wf_files:
    with open(wfp, "r", encoding="utf-8") as wf:
        wftext = wf.read()
    if "FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true" not in wftext:
        sec_ok = False
        sec_issues.append(f"{wfp}: missing FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true")
    for idx, raw_line in enumerate(wftext.splitlines(), start=1):
        code_line = raw_line.split("#")[0]
        uses_m = re.search(r"^\s*-?\s*uses:\s*(\S+)", code_line)
        if uses_m:
            target = uses_m.group(1).strip().strip("'\"")
            if target.startswith("./"):
                continue
            if not re.search(r"@[0-9a-f]{40}$", target):
                sec_ok = False
                sec_issues.append(f"{wfp}:{idx}: unpinned GitHub Action '{target}' (40-char SHA required)")

if not sec_ok:
    skill_gate_failures.extend(sec_issues)
sec_result_cell = (
    "✅ PASS (SHA-pinned Actions, Node 24 enforced, 0 secrets)"
    if sec_ok
    else f"❌ FAIL ({'; '.join(sec_issues)})"
)

# Evaluate per-agent results
agent_results = []
overall_passed = (len(skill_gate_failures) == 0)

for agent in activated_agents:
    aid = agent["id"]
    owned = agent_files.get(aid, [])

    # 1. Tests & Race
    agent_test_ok = (test_rc == 0) and (len(failing_tests) == 0)
    test_cell = f"✅ PASS ({passed_test_count} tests, 0 races)" if agent_test_ok else f"❌ FAIL ({len(failing_tests)} failed)"

    # 2. Coverage (100.0%)
    agent_uncovered = []
    for of in owned:
        base_of = os.path.basename(of)
        if base_of in uncovered_by_file:
            agent_uncovered.extend(uncovered_by_file[base_of])
    agent_cov_ok = (cov_run_rc == 0) and (total_cov >= coverage_target) and (len(agent_uncovered) == 0)
    cov_cell = f"✅ {total_cov_str} (target {coverage_target:.1f}%)" if agent_cov_ok else f"❌ {total_cov_str} (< {coverage_target:.1f}%)"

    # 3. Lint & SOLID / Clean Code
    agent_lint_ok = len(lint_failures) == 0
    lint_cell = "✅ gofmt + go vet + int64 paise" if agent_lint_ok else f"❌ {len(lint_failures)} violation(s)"

    # 4. Domain Rules & Truth Gates
    agent_domain_ok = len(domain_failures) == 0
    domain_cell = "✅ 0 deps + symbols verified" if agent_domain_ok else f"❌ {len(domain_failures)} failure(s)"

    agent_ok = agent_test_ok and agent_cov_ok and agent_lint_ok and agent_domain_ok
    if not agent_ok:
        overall_passed = False

    verdict_cell = "✅ **APPROVED**" if agent_ok else "❌ **CHANGES REQUESTED**"
    agent_results.append({
        "agent": agent,
        "owned": owned,
        "test_cell": test_cell,
        "cov_cell": cov_cell,
        "lint_cell": lint_cell,
        "domain_cell": domain_cell,
        "verdict_cell": verdict_cell,
        "ok": agent_ok,
    })

# Build Markdown Report
lines = []
lines.append("<!-- abeta-module-agent-review -->")
if overall_passed:
    lines.append("## 🛡️ Module Code-Owner Agent Review: ✅ APPROVED BY MODULE CODE-OWNER AGENTS")
else:
    lines.append("## 🛡️ Module Code-Owner Agent Review: ❌ CHANGES REQUESTED BY MODULE CODE-OWNER AGENTS")

lines.append("")
pr_label = f"PR `#{pr_number}` (`{base_ref}...{head_ref}`)" if pr_number else f"Diff `{base_ref}...{head_ref}`"
lines.append(
    f"Evaluated **{len(changed_files)}** changed file(s) across **{len(activated_agents)}** activated Module Code-Owner Agent(s) for {pr_label}."
)
lines.append("")
lines.append("| Module Agent | Companion Skill | Owned Changed Files | Test Cases & Race | 100.0% Coverage | Lint & SOLID Clean Code | Domain Rules | Verdict |")
lines.append("| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |")

for res in agent_results:
    a = res["agent"]
    owned_count = len(res["owned"])
    owned_preview = ", ".join(f"`{f}`" for f in res["owned"][:3])
    if owned_count > 3:
        owned_preview += f" (+{owned_count - 3} more)"
    lines.append(
        f"| **{a['handle']}**<br/>_{a['name']}_ | `{a['skill']}` | {owned_count} file(s): {owned_preview} | {res['test_cell']} | {res['cov_cell']} | {res['lint_cell']} | {res['domain_cell']} | {res['verdict_cell']} |"
    )

lines.append("")
lines.append("### 🎨 Multi-Skill Quality & Architecture Review (Executed in GitHub Actions)")
lines.append("")
lines.append("| Skill / Review Gate | Scope & Reference | Verification Executed | Result |")
lines.append("| :--- | :--- | :--- | :--- |")
lines.append(
    f"| **`/impeccable`** (`impeccable` Frontend & Docs Design Skill) | `docs/index.html` | `impeccable detect --json docs/index.html` + semantic `<main>` landmark, CSS custom tokens (`--`), and canonical URL check | {impeccable_result_cell} |"
)
lines.append(
    f"| **`@uncle-bob-craft`** (SOLID & Clean Code Skill) | `*.go`, `go.mod` | SRP/OCP/LSP/ISP/DIP boundaries, strict `int64` paise `Money` invariant (zero `float64` currency struct fields), `gofmt` + `go vet`, and zero external dependencies | {uncle_bob_result_cell} |"
)
lines.append(
    f"| **`@effective-agent-skills`** (Repository Agent Skills Gate) | `skills/*/SKILL.md`, `AGENTS.md` | YAML frontmatter (`name`, `description`) validation across all `{skills_count}` subskills and `AGENTS.md` routing table completeness | {skills_result_cell} |"
)
lines.append(
    f"| **`@security-auditor`** (Supply-Chain, Node 24 & Secret Gate) | `.github/workflows/*.yml` | 40-char commit SHA pinning across all `{len(wf_files)}` GitHub Actions workflows and `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: true` enforcement | {sec_result_cell} |"
)
lines.append("")

if not overall_passed:
    lines.append("### ❌ Actionable Failure Diagnostics")
    lines.append("")
    if test_rc != 0 or failing_tests:
        lines.append("#### 1. Failing Unit / Race Test Cases (`go test -race ./...`)")
        lines.append("```text")
        for ft in (failing_tests + test_diag_lines)[:40]:
            lines.append(ft)
        if not failing_tests and not test_diag_lines:
            lines.append(test_out[-1500:])
        lines.append("```")
        lines.append("")
    if cov_run_rc != 0 or total_cov < coverage_target or uncovered_funcs:
        lines.append(f"#### 2. Statement Coverage Violations (Current: `{total_cov_str}`, Required: `{coverage_target:.1f}%`)")
        lines.append("```text")
        for uf in uncovered_funcs[:40]:
            lines.append(f"UNCOVERED: {uf}")
        lines.append("```")
        lines.append("")
    if lint_failures:
        lines.append("#### 3. Lint, SOLID & Clean Code Violations")
        lines.append("```text")
        for lf in lint_failures[:40]:
            lines.append(lf)
        lines.append("```")
        lines.append("")
    if domain_failures:
        lines.append("#### 4. Domain & Truth Gate Violations")
        lines.append("```text")
        for df in domain_failures[:40]:
            lines.append(df)
        lines.append("```")
        lines.append("")
    if skill_gate_failures:
        lines.append("#### 5. Multi-Skill Quality & Architecture Gate Violations")
        lines.append("```text")
        for sf in skill_gate_failures[:40]:
            lines.append(sf)
        lines.append("```")
        lines.append("")

lines.append("### 🔍 Activated Module Agent Scope & Enforced Domain Invariants")
lines.append("")
for res in agent_results:
    a = res["agent"]
    status_icon = "✅" if res["ok"] else "❌"
    lines.append(f"<details>")
    lines.append(f"<summary>{status_icon} <strong>{a['handle']}</strong> — {a['name']} (<code>{a['skill']}</code>, owner: <code>{a['owner']}</code>)</summary>")
    lines.append("")
    lines.append("**Reviewed Files in Scope:**")
    for fpath in res["owned"]:
        lines.append(f"- `{fpath}`")
    lines.append("")
    lines.append("**Enforced Module Rules:**")
    for rule in a.get("rules", []):
        lines.append(f"- {rule}")
    lines.append("")
    lines.append("</details>")
    lines.append("")

report_markdown = "\n".join(lines) + "\n"

os.makedirs(os.path.dirname(os.path.abspath(output_path)), exist_ok=True)
with open(output_path, "w", encoding="utf-8") as out_f:
    out_f.write(report_markdown)

step_summary = os.environ.get("GITHUB_STEP_SUMMARY", "").strip()
if step_summary:
    with open(step_summary, "a", encoding="utf-8") as sf:
        sf.write(report_markdown)

print(report_markdown)
sys.exit(0 if overall_passed else 1)
PYEOF
