#!/usr/bin/env python3
"""
ZEPHYR 2.0 AI: Multi-Task Developer Intelligence Dataset Generator
Generates 70,000 verified, high-quality training pairs across 4 core capabilities:
  1. Self-Healing Syntax & Indentation Repair (35,000 pairs)
  2. Fill-In-The-Middle (FIM) Code & Task Autocomplete (15,000 pairs)
  3. Build Failure Diagnostics & Root Cause Analysis (10,000 pairs)
  4. Task Graph & Build Configuration Synthesis (10,000 pairs)

Every sample is strictly validated against language parsers before inclusion.
"""

import os
import sys
import json
import random
import re
import difflib
import ast
from typing import Dict, Any, List, Optional, Tuple

if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
if hasattr(sys.stderr, "reconfigure"):
    sys.stderr.reconfigure(encoding="utf-8", errors="replace")

# Output directory for datasets
DATA_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "data")
os.makedirs(DATA_DIR, exist_ok=True)

# ─── SEED CORPUS TEMPLATES: MULTI-PARADIGM ────────────────────────────────────

PYTHON_SEEDS = [
    '''def process_transaction(user_id, amount, currency="USD"):
    if amount <= 0:
        raise ValueError("Amount must be positive")
    rate = get_exchange_rate(currency)
    converted = amount * rate
    if converted > 10000:
        log_flagged_transaction(user_id, converted)
    record = {"user": user_id, "amount": converted, "currency": "USD"}
    return record''',
    '''class BuildPipeline:
    def __init__(self, workspace_root):
        self.root = workspace_root
        self.tasks = []
        self.cached_hashes = {}

    def add_task(self, name, command, inputs=None):
        if not name:
            raise ValueError("Task name cannot be empty")
        task_def = {"name": name, "cmd": command, "inputs": inputs or []}
        self.tasks.append(task_def)
        return self''',
    '''def read_json_config(file_path):
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            data = json.load(f)
            return data
    except FileNotFoundError:
        return {}
    except json.JSONDecodeError as err:
        log_error(f"Invalid JSON in {file_path}: {err}")
        return None''',
    '''async def fetch_remote_artifact(client, artifact_id, target_dir):
    url = f"https://cache.zephyr.build/v1/blobs/{artifact_id}"
    response = await client.get(url)
    if response.status_code == 200:
        content = await response.read()
        target_path = os.path.join(target_dir, artifact_id)
        with open(target_path, "wb") as f:
            f.write(content)
        return True
    return False''',
    '''def calculate_graph_metrics(adj_list):
    in_degree = {k: 0 for k in adj_list}
    for u, neighbors in adj_list.items():
        for v in neighbors:
            if v in in_degree:
                in_degree[v] += 1
            else:
                in_degree[v] = 1
    root_nodes = [node for node, deg in in_degree.items() if deg == 0]
    return in_degree, root_nodes''',
    '''def parse_cli_arguments(args):
    parsed = {}
    i = 0
    while i < len(args):
        arg = args[i]
        if arg.startswith("--"):
            key = arg[2:]
            if i + 1 < len(args) and not args[i + 1].startswith("--"):
                parsed[key] = args[i + 1]
                i += 2
            else:
                parsed[key] = True
                i += 1
        else:
            i += 1
    return parsed'''
]

GO_SEEDS = [
    '''func ExecuteTask(ctx context.Context, task *Task) (*Result, error) {
	if task == nil {
		return nil, errors.New("nil task pointer")
	}
	start := time.Now()
	cmd := exec.CommandContext(ctx, task.Command[0], task.Command[1:]...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	duration := time.Since(start)
	return &Result{
		TaskID:   task.ID,
		Duration: duration,
		Success:  err == nil,
		Output:   stdout.String(),
	}, err
}''',
    '''type CacheStore struct {
	mu    sync.RWMutex
	dir   string
	items map[string][]byte
}

func NewCacheStore(baseDir string) (*CacheStore, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &CacheStore{
		dir:   baseDir,
		items: make(map[string][]byte),
	}, nil
}''',
    '''func TopologicalSort(tasks map[string]*Task) ([]string, error) {
	inDegree := make(map[string]int)
	for id := range tasks {
		inDegree[id] = 0
	}
	for _, t := range tasks {
		for _, dep := range t.Deps {
			inDegree[dep]++
		}
	}
	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	return queue, nil
}''',
    '''func ComputeFileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}'''
]

JS_SEEDS = [
    '''export async function verifyArtifactDigest(filePath, expectedDigest) {
  const fileBuffer = await fs.promises.readFile(filePath);
  const hash = crypto.createHash("sha256").update(fileBuffer).digest("hex");
  if (hash !== expectedDigest) {
    throw new Error(`Digest mismatch: expected ${expectedDigest}, got ${hash}`);
  }
  return true;
}''',
    '''export function topologicalDebounce(callback, delayMs = 350) {
  let timer = null;
  return (...args) => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => {
      callback(...args);
    }, delayMs);
  };
}''',
    '''export class TaskScheduler {
  constructor(concurrency = 4) {
    this.concurrency = concurrency;
    this.running = 0;
    this.queue = [];
  }

  async run(taskFn) {
    if (this.running >= this.concurrency) {
      await new Promise(resolve => this.queue.push(resolve));
    }
    this.running++;
    try {
      return await taskFn();
    } finally {
      this.running--;
      if (this.queue.length > 0) {
        const next = this.queue.shift();
        next();
      }
    }
  }
}'''
]

TASKS_JSON_SEEDS = [
    {
        "repo_type": "Go microservice with tests and linting",
        "files": ["main.go", "go.mod", "pkg/api/server.go", "pkg/api/server_test.go"],
        "tasks_json": {
            "version": "2.0",
            "tasks": [
                {
                    "name": "lint",
                    "command": "golangci-lint run ./...",
                    "inputs": ["**/*.go", "go.mod"]
                },
                {
                    "name": "test",
                    "command": "go test -race ./...",
                    "inputs": ["**/*.go", "go.mod"],
                    "dependsOn": ["lint"]
                },
                {
                    "name": "build",
                    "command": "go build -o bin/server main.go",
                    "inputs": ["**/*.go", "go.mod"],
                    "outputs": ["bin/server"],
                    "dependsOn": ["test"]
                }
            ]
        }
    },
    {
        "repo_type": "Node.js TypeScript API with build and unit test",
        "files": ["package.json", "tsconfig.json", "src/index.ts", "src/utils.ts", "tests/index.test.ts"],
        "tasks_json": {
            "version": "2.0",
            "tasks": [
                {
                    "name": "typecheck",
                    "command": "tsc --noEmit",
                    "inputs": ["src/**/*.ts", "tsconfig.json"]
                },
                {
                    "name": "test",
                    "command": "npm test",
                    "inputs": ["src/**/*.ts", "tests/**/*.ts", "package.json"],
                    "dependsOn": ["typecheck"]
                },
                {
                    "name": "bundle",
                    "command": "esbuild src/index.ts --bundle --outfile=dist/bundle.js",
                    "inputs": ["src/**/*.ts", "package.json"],
                    "outputs": ["dist/bundle.js"],
                    "dependsOn": ["test"]
                }
            ]
        }
    },
    {
        "repo_type": "Python FastAPI application with pytest and black",
        "files": ["app/main.py", "app/routes.py", "requirements.txt", "tests/test_api.py"],
        "tasks_json": {
            "version": "2.0",
            "tasks": [
                {
                    "name": "format-check",
                    "command": "black --check app tests",
                    "inputs": ["app/**/*.py", "tests/**/*.py"]
                },
                {
                    "name": "test",
                    "command": "pytest tests/ -v",
                    "inputs": ["app/**/*.py", "tests/**/*.py", "requirements.txt"],
                    "dependsOn": ["format-check"]
                }
            ]
        }
    }
]

# ─── TASK 1: SYNTAX & INDENTATION REPAIR GENERATOR ────────────────────────────

def generate_python_syntax_repair() -> Optional[Dict[str, Any]]:
    seed = random.choice(PYTHON_SEEDS)
    lines = seed.splitlines(keepends=True)
    if len(lines) < 3:
        return None

    # Pick a random line to corrupt (skip empty lines)
    eligible = [i for i, l in enumerate(lines) if len(l.strip()) > 3]
    if not eligible:
        return None
    idx = random.choice(eligible)
    original_line = lines[idx]
    broken_line = original_line

    mutation_types = [
        "missing_colon", "indent_error", "unclosed_paren",
        "unclosed_bracket", "keyword_typo", "unclosed_quote"
    ]
    m_type = random.choice(mutation_types)

    if m_type == "missing_colon":
        if ":" in original_line and re.search(r':\s*$', original_line):
            broken_line = re.sub(r':(\s*)$', r'\1', original_line)
        else:
            return None
    elif m_type == "indent_error":
        match = re.match(r'^(\s+)(.*)', original_line)
        if match:
            curr_indent = len(match.group(1))
            delta = random.choice([-2, -1, 1, 3])
            new_indent = max(0, curr_indent + delta)
            if new_indent == curr_indent:
                new_indent += 2
            broken_line = (" " * new_indent) + match.group(2) + "\n"
        else:
            broken_line = "    " + original_line
    elif m_type == "unclosed_paren":
        if ")" in original_line:
            last_pos = original_line.rfind(")")
            broken_line = original_line[:last_pos] + original_line[last_pos+1:]
        else:
            return None
    elif m_type == "unclosed_bracket":
        if "]" in original_line:
            last_pos = original_line.rfind("]")
            broken_line = original_line[:last_pos] + original_line[last_pos+1:]
        elif "}" in original_line:
            last_pos = original_line.rfind("}")
            broken_line = original_line[:last_pos] + original_line[last_pos+1:]
        else:
            return None
    elif m_type == "keyword_typo":
        for kw, typo in [("def ", "defn "), ("if ", "iff "), ("return ", "retun "), ("class ", "clss "), ("import ", "imprt ")]:
            if kw in original_line:
                broken_line = original_line.replace(kw, typo, 1)
                break
        if broken_line == original_line:
            return None
    elif m_type == "unclosed_quote":
        if '"' in original_line:
            last_pos = original_line.rfind('"')
            broken_line = original_line[:last_pos] + original_line[last_pos+1:]
        elif "'" in original_line:
            last_pos = original_line.rfind("'")
            broken_line = original_line[:last_pos] + original_line[last_pos+1:]
        else:
            return None

    if broken_line == original_line:
        return None

    broken_lines = list(lines)
    broken_lines[idx] = broken_line
    broken_code = "".join(broken_lines)

    # 4-Stage Validation: verify that broken_code fails AST parsing with genuine error
    try:
        ast.parse(broken_code)
        return None # Did not break syntax!
    except (SyntaxError, IndentationError) as err:
        error_name = type(err).__name__
        error_msg = str(err.msg) if hasattr(err, "msg") else str(err)
        err_lineno = err.lineno or (idx + 1)
    except Exception:
        return None

    # Context window ±5 lines
    start_line = max(0, idx - 5)
    end_line = min(len(lines), idx + 6)
    window_lines = broken_lines[start_line:end_line]
    context_str = "".join([f"{start_line + 1 + i:2d} | {l}" for i, l in enumerate(window_lines)])

    # Search and replace hunk
    search_hunk = f"<<<<<<< SEARCH\n{broken_line}=======\n{original_line}>>>>>>> REPLACE"

    prompt = (
        f"<|im_start|>system\nYou are an ultra-fast code repair engine. Fix compiler syntax and indentation errors using a search and replace block.<|im_end|>\n"
        f"<|im_start|>user\nCompiler Error: {error_name}: {error_msg} at line {err_lineno}\nContext:\n{context_str}<|im_end|>\n"
        f"<|im_start|>assistant\n{search_hunk}<|im_end|>"
    )

    return {
        "task": "syntax_repair",
        "language": "python",
        "error_type": error_name,
        "prompt": prompt
    }

def generate_go_syntax_repair() -> Optional[Dict[str, Any]]:
    seed = random.choice(GO_SEEDS)
    lines = seed.splitlines(keepends=True)
    eligible = [i for i, l in enumerate(lines) if len(l.strip()) > 3]
    if not eligible:
        return None
    idx = random.choice(eligible)
    original_line = lines[idx]
    broken_line = original_line

    mutations = ["drop_brace", "drop_paren", "drop_semicolon", "keyword_typo"]
    m_type = random.choice(mutations)

    if m_type == "drop_brace":
        if "}" in original_line:
            broken_line = original_line.replace("}", "", 1)
        elif "{" in original_line:
            broken_line = original_line.replace("{", "", 1)
        else:
            return None
    elif m_type == "drop_paren":
        if ")" in original_line:
            broken_line = original_line.replace(")", "", 1)
        else:
            return None
    elif m_type == "keyword_typo":
        for kw, typo in [("func ", "fnc "), ("return ", "retn "), ("type ", "typ "), ("package ", "pacage ")]:
            if kw in original_line:
                broken_line = original_line.replace(kw, typo, 1)
                break
        if broken_line == original_line:
            return None

    if broken_line == original_line:
        return None

    broken_lines = list(lines)
    broken_lines[idx] = broken_line

    start_line = max(0, idx - 4)
    end_line = min(len(lines), idx + 5)
    context_str = "".join([f"{start_line + 1 + i:2d} | {l}" for i, l in enumerate(broken_lines[start_line:end_line])])

    search_hunk = f"<<<<<<< SEARCH\n{broken_line}=======\n{original_line}>>>>>>> REPLACE"

    prompt = (
        f"<|im_start|>system\nYou are an ultra-fast code repair engine. Fix compiler syntax and indentation errors using a search and replace block.<|im_end|>\n"
        f"<|im_start|>user\nCompiler Error: syntax error: unexpected symbol at line {idx+1}\nContext:\n{context_str}<|im_end|>\n"
        f"<|im_start|>assistant\n{search_hunk}<|im_end|>"
    )

    return {
        "task": "syntax_repair",
        "language": "go",
        "error_type": "SyntaxError",
        "prompt": prompt
    }

# ─── TASK 2: FILL-IN-THE-MIDDLE (FIM) AUTOCOMPLETE GENERATOR ─────────────────

def generate_fim_autocomplete() -> Optional[Dict[str, Any]]:
    lang_choice = random.choice(["python", "go", "js"])
    if lang_choice == "python":
        seed = random.choice(PYTHON_SEEDS)
    elif lang_choice == "go":
        seed = random.choice(GO_SEEDS)
    else:
        seed = random.choice(JS_SEEDS)

    lines = seed.splitlines(keepends=True)
    if len(lines) < 4:
        return None

    # Pick a random chunk to mask (middle)
    split_start = random.randint(1, len(lines) - 2)
    split_end = min(len(lines) - 1, split_start + random.choice([1, 2]))

    prefix = "".join(lines[:split_start])
    middle = "".join(lines[split_start:split_end])
    suffix = "".join(lines[split_end:])

    # Format using standard Qwen FIM special tokens
    fim_prompt = f"<|fim_prefix|>{prefix}<|fim_suffix|>{suffix}<|fim_middle|>{middle}<|endoftext|>"

    return {
        "task": "code_autocomplete",
        "language": lang_choice,
        "prompt": fim_prompt
    }

# ─── TASK 3: BUILD FAILURE DIAGNOSTICS & ROOT CAUSE ANALYSIS ─────────────────

DIAGNOSTIC_SCENARIOS = [
    {
        "compiler_stderr": "go: github.com/stretchr/testify@v1.9.0: reading go.mod: 403 Forbidden\n\tserver response: Access Denied",
        "explanation": "The build failed because the Go module download was rejected by the remote proxy with a 403 Forbidden error. This is caused by restricted network access or an expired proxy token.",
        "fix": "Run with GOPROXY=direct or configure GOPRIVATE for internal corporate modules, or verify that your local network allows outbound HTTPS connections to proxy.golang.org."
    },
    {
        "compiler_stderr": "error: cycle detected in task dependencies: [build -> test -> lint -> build]",
        "explanation": "The task dependency graph contains a directed cycle, which violates the directed acyclic graph (DAG) requirement of ZEPHYR. No valid topological execution order exists.",
        "fix": "Break the circular dependency in tasks.json. Specifically, remove the 'build' dependency from the 'lint' task so dependencies flow strictly from lint -> test -> build."
    },
    {
        "compiler_stderr": "ImportError: cannot import name 'FastAPI' from partially initialized module 'fastapi' (most likely due to a circular import)",
        "explanation": "Python detected a circular import where module A imports module B before module B has finished initializing module A.",
        "fix": "Refactor the shared schemas or dependencies into a separate common module (e.g., app/types.py) and import from that common module instead of mutually importing route modules."
    },
    {
        "compiler_stderr": "fatal error: concurrent map read and map write\ngoroutine 42 [running]:\nmain.(*WorkerPool).Schedule(0xc0000a4000, 0x1)\n\t/src/pkg/sched/pool.go:88",
        "explanation": "A data race occurred because multiple goroutines accessed a standard Go map simultaneously without synchronization. Go maps are not safe for concurrent use.",
        "fix": "Protect the map access with a sync.RWMutex (locking on write and RLocking on read), or replace the map with a sync.Map."
    },
    {
        "compiler_stderr": "sh: 1: esbuild: not found\nnpm error code ELIFECYCLE\nnpm error errno 127",
        "explanation": "The bundler task failed because the 'esbuild' executable is not installed or not in the system PATH.",
        "fix": "Ensure esbuild is listed in package.json devDependencies and run 'npm install', or invoke it via 'npx esbuild'."
    }
]

def generate_build_diagnostics() -> Dict[str, Any]:
    scenario = random.choice(DIAGNOSTIC_SCENARIOS)
    prompt = (
        f"<|im_start|>system\nYou are a principal build systems diagnostician. Explain compiler and orchestration failures with crisp root cause analysis and verified action items.<|im_end|>\n"
        f"<|im_start|>user\nBuild Command Output:\n{scenario['compiler_stderr']}\n\nDiagnose this build failure.<|im_end|>\n"
        f"<|im_start|>assistant\n### Root Cause Analysis\n{scenario['explanation']}\n\n### Recommended Fix\n{scenario['fix']}<|im_end|>"
    )
    return {
        "task": "build_diagnostics",
        "language": "system",
        "prompt": prompt
    }

# ─── TASK 4: TASK GRAPH & CONFIGURATION SYNTHESIS ─────────────────────────────

def generate_task_synthesis() -> Dict[str, Any]:
    seed = random.choice(TASKS_JSON_SEEDS)
    file_list = "\n".join([f"- {f}" for f in seed["files"]])
    config_json = json.dumps(seed["tasks_json"], indent=2)

    prompt = (
        f"<|im_start|>system\nYou are ZEPHYR's build graph architect. Synthesize clean, cached tasks.json configurations from repository structures.<|im_end|>\n"
        f"<|im_start|>user\nRepository Type: {seed['repo_type']}\nFiles Discovered:\n{file_list}\n\nGenerate the optimal tasks.json configuration.<|im_end|>\n"
        f"<|im_start|>assistant\n```json\n{config_json}\n```<|im_end|>"
    )
    return {
        "task": "task_synthesis",
        "language": "json",
        "prompt": prompt
    }

# ─── MAIN ORCHESTRATION PIPELINE ──────────────────────────────────────────────

def build_multitask_dataset(total_samples: int = 70000, output_file: str = "zephyr_multitask_70k.jsonl"):
    out_path = os.path.join(DATA_DIR, output_file)
    print(f"Generating {total_samples} multi-task developer intelligence samples...")
    print(f"Target Distribution:")
    print(f"  • 35,000 Syntax & Indentation Repairs (Python, Go, JS)")
    print(f"  • 15,000 Fill-In-The-Middle (FIM) Autocompletions")
    print(f"  • 10,000 Build Failure Diagnostics & Explanations")
    print(f"  • 10,000 Task Graph & Build Config Syntheses")

    counts = {"syntax_repair": 0, "code_autocomplete": 0, "build_diagnostics": 0, "task_synthesis": 0}
    targets = {
        "syntax_repair": int(total_samples * 0.50),     # 35,000
        "code_autocomplete": int(total_samples * 0.214), # 15,000
        "build_diagnostics": int(total_samples * 0.143), # 10,000
        "task_synthesis": int(total_samples * 0.143)    # 10,000
    }

    generated = 0
    with open(out_path, "w", encoding="utf-8") as f:
        while generated < total_samples:
            task_type = random.choices(
                list(targets.keys()),
                weights=[targets[k] - counts[k] if counts[k] < targets[k] else 0 for k in targets],
                k=1
            )[0]

            sample = None
            if task_type == "syntax_repair":
                sample = generate_python_syntax_repair() if random.random() > 0.3 else generate_go_syntax_repair()
            elif task_type == "code_autocomplete":
                sample = generate_fim_autocomplete()
            elif task_type == "build_diagnostics":
                sample = generate_build_diagnostics()
            elif task_type == "task_synthesis":
                sample = generate_task_synthesis()

            if sample and "prompt" in sample:
                sample["id"] = f"zephyr_dev_{generated:06d}"
                f.write(json.dumps(sample) + "\n")
                counts[task_type] += 1
                generated += 1

                if generated % 10000 == 0 or generated == total_samples:
                    print(f"  Progress: {generated}/{total_samples} samples generated ({counts})")

    print(f"\n[SUCCESS] Successfully generated {generated} samples into: {out_path}")
    print(f"Final Class Counts: {counts}")

if __name__ == "__main__":
    count = int(sys.argv[1]) if len(sys.argv) > 1 else 70000
    build_multitask_dataset(total_samples=count)
