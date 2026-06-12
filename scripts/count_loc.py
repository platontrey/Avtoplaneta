#!/usr/bin/env python3
import os
import sys
import argparse
import subprocess
import fnmatch
import json
import time
import concurrent.futures
from pathlib import Path

# Default directories to ignore when scanning
DEFAULT_IGNORE_DIRS = {
    '.git', '.venv', 'venv', 'node_modules', '__pycache__', 
    '.idea', '.vscode', '.claude', 'dist', 'build', 'out',
    'postgres_data', 'grafana', 'traefik', 'target', 'bin', 'obj',
    '.nuxt', '.next', '.svelte-kit', '.cache', 'vendor', 'coverage',
    '.nyc_output', '.expo', 'ios/Pods'
}

# Default file extensions and patterns to ignore
DEFAULT_IGNORE_FILES = {
    'package-lock.json', 'yarn.lock', 'pnpm-lock.yaml',
    'poetry.lock', 'Gemfile.lock', 'Cargo.lock', 'go.sum',
    'composer.lock', '*.png', '*.jpg', '*.jpeg', '*.gif', 
    '*.ico', '*.svg', '*.webp', '*.mp4', '*.mp3', '*.pdf', 
    '*.zip', '*.tar.gz', '*.rar', '*.xlsx', '*.xls', '*.db', 
    '*.sqlite', '*.pem', '*.key', '*.woff', '*.woff2', '*.ttf', 
    '*.eot', '*.exe', '*.dll', '*.so', '*.dylib', '*.bin', 
    '*.out', '*.map', '*.DS_Store'
}

def is_binary(file_path):
    """Check if a file is binary by reading its first chunk and checking for null bytes."""
    try:
        with open(file_path, 'rb') as f:
            chunk = f.read(1024)
            return b'\x00' in chunk
    except IOError:
        return True

def parse_gitignore(gitignore_path):
    """Parse basic rules from a .gitignore file."""
    if not gitignore_path.exists():
        return []
    rules = []
    with open(gitignore_path, 'r', encoding='utf-8', errors='ignore') as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith('#'):
                continue
            rules.append(line)
    return rules

def should_ignore(path: Path, root_dir: Path, custom_excludes=None, gitignore_rules=None):
    """Check if a file path should be ignored based on default rules and gitignore."""
    try:
        rel_path = path.relative_to(root_dir)
    except ValueError:
        # Fallback if path is not relative to root_dir
        rel_path = path
        
    # Check if any parent directory is in the default ignore list
    for part in rel_path.parts[:-1]:
        if part in DEFAULT_IGNORE_DIRS:
            return True
            
    # Check default file ignore patterns
    name = path.name
    for ignore_pat in DEFAULT_IGNORE_FILES:
        if fnmatch.fnmatch(name, ignore_pat):
            return True
            
    # Check custom excludes
    if custom_excludes:
        rel_str = str(rel_path)
        for pattern in custom_excludes:
            if fnmatch.fnmatch(rel_str, pattern) or fnmatch.fnmatch(name, pattern) or pattern in rel_path.parts:
                return True
                
    # Check gitignore rules
    if gitignore_rules:
        rel_str = str(rel_path)
        for rule in gitignore_rules:
            if rule.endswith('/'):
                rule_dir = rule.rstrip('/')
                if rel_str.startswith(rule_dir + '/') or rel_str == rule_dir:
                    return True
            elif fnmatch.fnmatch(rel_str, rule) or fnmatch.fnmatch(name, rule):
                return True
                
    return False

def get_git_files(root_dir):
    """Get tracked and untracked (but not ignored) files from git."""
    try:
        # Check if it's a git repo
        subprocess.run(
            ['git', 'rev-parse', '--is-inside-work-tree'],
            cwd=root_dir, stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=True
        )
        
        # Tracked files
        tracked = subprocess.run(
            ['git', 'ls-files'],
            cwd=root_dir, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=True
        ).stdout.splitlines()
        
        # Untracked files (excluding ignored ones)
        untracked = subprocess.run(
            ['git', 'ls-files', '--others', '--exclude-standard'],
            cwd=root_dir, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=True
        ).stdout.splitlines()
        
        return [Path(root_dir) / f for f in (tracked + untracked) if f]
    except (subprocess.SubprocessError, FileNotFoundError):
        return None

def get_files_manually(root_dir, custom_excludes=None):
    """Recursively search for files in the directory manually."""
    files = []
    gitignore_rules = parse_gitignore(Path(root_dir) / '.gitignore')
    
    for root, dirs, filenames in os.walk(root_dir):
        # Modify dirs in-place to prune ignored directories
        dirs[:] = [d for d in dirs if d not in DEFAULT_IGNORE_DIRS]
        if custom_excludes:
            dirs[:] = [d for d in dirs if d not in custom_excludes]
            
        for filename in filenames:
            file_path = Path(root) / filename
            if not should_ignore(file_path, Path(root_dir), custom_excludes, gitignore_rules):
                files.append(file_path)
    return files

def is_generated(file_path):
    """Check if a file is auto-generated by checking its name or reading its first few lines."""
    name = file_path.name.lower()
    # Common generated files extensions/patterns
    if name.endswith('.pb.go') or name.endswith('.gen.go') or '_grpc.pb.go' in name or 'sql.go' in name:
        return True
    if name.endswith('.min.js') or name.endswith('.min.css'):
        return True
        
    try:
        with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
            # Check first 5 lines for common generator headers
            for _ in range(5):
                line = f.readline()
                if not line:
                    break
                line_lower = line.lower()
                if (
                    "code generated" in line_lower 
                    or "do not edit" in line_lower 
                    or "auto-generated" in line_lower 
                    or "autogenerated" in line_lower 
                    or "generated by" in line_lower
                ):
                    return True
        return False
    except IOError:
        return False

def count_lines(file_path: Path, include_generated=False):
    """Count total, blank, comment and code lines in a file, handling multi-line comments."""
    try:
        if is_binary(file_path):
            return None
            
        if not include_generated and is_generated(file_path):
            return None
            
        with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
            lines = f.readlines()
            
        total = len(lines)
        blank = 0
        comment = 0
        
        ext = file_path.suffix.lower()
        name = file_path.name
        group_key = ext if ext else name
        
        line_comment_marker = None
        block_comment_start = None
        block_comment_end = None
        
        # Identify comment syntax
        if group_key in {'.py'}:
            line_comment_marker = '#'
        elif group_key in {'.sh', '.bash', '.yaml', '.yml', '.toml', '.ini', '.rb', '.pl', '.r', 'dockerfile', 'makefile', '.conf', '.env', '.dockerignore'}:
            line_comment_marker = '#'
        elif group_key in {'.js', '.jsx', '.ts', '.tsx', '.go', '.java', '.c', '.cpp', '.h', '.hpp', '.cs', '.rs', '.swift', '.kt', '.scala', '.groovy', '.css', '.scss', '.sass', '.gradle'}:
            line_comment_marker = '//' if group_key not in {'.css'} else None
            block_comment_start = '/*'
            block_comment_end = '*/'
        elif group_key in {'.html', '.xml', '.vue', '.svelte', '.svg', '.storyboard', '.plist', '.xcscheme'}:
            block_comment_start = '<!--'
            block_comment_end = '-->'
            if group_key in {'.vue', '.svelte'}:
                line_comment_marker = '//'
        elif group_key in {'.sql', '.lua', '.hs'}:
            line_comment_marker = '--'
            if group_key == '.lua':
                block_comment_start = '--[['
                block_comment_end = ']]'
        elif group_key in {'.proto'}:
            line_comment_marker = '//'
            block_comment_start = '/*'
            block_comment_end = '*/'
            
        in_block = False
        in_py_docstring = None
        
        for line in lines:
            stripped = line.strip()
            if not stripped:
                blank += 1
                continue
                
            # Handle Python docstrings as comments
            if group_key == '.py':
                if in_py_docstring:
                    comment += 1
                    if in_py_docstring in stripped:
                        in_py_docstring = None
                    continue
                else:
                    if stripped.startswith('#'):
                        comment += 1
                        continue
                    elif stripped.startswith('"""') or stripped.startswith("'''"):
                        comment += 1
                        marker = '"""' if stripped.startswith('"""') else "'''"
                        # If the marker appears only once, it starts a multi-line docstring
                        if stripped.count(marker) == 1:
                            in_py_docstring = marker
                        continue
                        
            # Handle block comments
            if in_block:
                comment += 1
                if block_comment_end and block_comment_end in stripped:
                    in_block = False
                continue
                
            # Handle single-line comments
            if line_comment_marker and stripped.startswith(line_comment_marker):
                comment += 1
                continue
                
            # Handle start of block comments
            if block_comment_start and stripped.startswith(block_comment_start):
                comment += 1
                if block_comment_end and block_comment_end in stripped[len(block_comment_start):]:
                    pass # Closed on same line
                else:
                    in_block = True
                continue
                
        code = total - blank - comment
        return total, blank, comment, code
    except Exception:
        return None

def format_num(val):
    if isinstance(val, int):
        return f"{val:,}"
    return str(val)

def format_cell(val, width, align_right=True, color_code=None):
    val_str = format_num(val)
    padding = width - len(val_str)
    if align_right:
        cell = " " * padding + val_str
    else:
        cell = val_str + " " * padding
    if color_code:
        return f"{color_code}{cell}\033[0m"
    return cell

def supports_color():
    """Check if the terminal supports color output."""
    plat = sys.platform
    supported_platform = plat != 'Pocket PC' and (plat != 'win32' or 'ANSICON' in os.environ)
    is_a_tty = hasattr(sys.stdout, 'isatty') and sys.stdout.isatty()
    return supported_platform and is_a_tty

def generate_table(headers, rows, total_row=None, use_color=True):
    col_widths = [len(h) for h in headers]
    for row in rows:
        for idx, val in enumerate(row):
            col_widths[idx] = max(col_widths[idx], len(format_num(val)))
    if total_row:
        for idx, val in enumerate(total_row):
            col_widths[idx] = max(col_widths[idx], len(format_num(val)))
            
    # ANSI escape sequences
    C_HEADER = '\033[95;1m' if use_color else ''
    C_LANG = '\033[96;1m' if use_color else ''
    C_GRAY = '\033[90m' if use_color else ''
    C_GREEN = '\033[92m' if use_color else ''
    C_YELLOW = '\033[93m' if use_color else ''
    C_CODE = '\033[92;1m' if use_color else ''
    C_BOLD = '\033[1m' if use_color else ''
    C_END = '\033[0m' if use_color else ''
    
    lines = []
    
    # Format headers
    header_cells = []
    for i, h in enumerate(headers):
        header_cells.append(format_cell(h, col_widths[i], align_right=(i > 0), color_code=C_HEADER))
    lines.append("  ".join(header_cells))
    
    # Divider
    divider_len = sum(col_widths) + 2 * (len(headers) - 1)
    lines.append(f"{C_GRAY}" + "-" * divider_len + f"{C_END}" if use_color else "-" * divider_len)
    
    # Format rows
    for row in rows:
        row_cells = []
        for i, val in enumerate(row):
            align_right = (i > 0)
            if i == 0:
                color = C_LANG
            elif i == len(row) - 1: # Code column
                color = C_CODE
            elif headers[i].lower() == "comment":
                color = C_GREEN
            elif headers[i].lower() == "blank":
                color = C_GRAY
            else:
                color = None
            row_cells.append(format_cell(val, col_widths[i], align_right, color))
        lines.append("  ".join(row_cells))
        
    # Format total
    if total_row:
        lines.append(f"{C_GRAY}" + "-" * divider_len + f"{C_END}" if use_color else "-" * divider_len)
        total_cells = []
        for i, val in enumerate(total_row):
            align_right = (i > 0)
            color = C_BOLD if i == 0 else C_CODE
            total_cells.append(format_cell(val, col_widths[i], align_right, color))
        lines.append("  ".join(total_cells))
        
    return "\n".join(lines)

def generate_markdown(headers, rows, total_row=None):
    lines = []
    lines.append("| " + " | ".join(headers) + " |")
    aligns = [":---" if i == 0 else "---:" for i in range(len(headers))]
    lines.append("| " + " | ".join(aligns) + " |")
    for row in rows:
        lines.append("| " + " | ".join(format_num(val) for val in row) + " |")
    if total_row:
        lines.append("| " + " | ".join(format_num(val) for val in total_row) + " |")
    return "\n".join(lines)

def get_distribution_bar(code, comment, blank, use_color=True):
    total = code + comment + blank
    if total == 0:
        return ""
        
    code_pct = code / total
    comment_pct = comment / total
    blank_pct = blank / total
    
    bar_width = 40
    code_chars = int(round(code_pct * bar_width))
    comment_chars = int(round(comment_pct * bar_width))
    blank_chars = bar_width - code_chars - comment_chars
    if blank_chars < 0:
        blank_chars = 0
        comment_chars = bar_width - code_chars
        
    C_GREEN = '\033[92m' if use_color else ''
    C_CYAN = '\033[96m' if use_color else ''
    C_GRAY = '\033[90m' if use_color else ''
    C_BOLD = '\033[1m' if use_color else ''
    C_END = '\033[0m' if use_color else ''
    
    bar = ""
    if code_chars > 0:
        bar += f"{C_GREEN}" + "█" * code_chars + f"{C_END}"
    if comment_chars > 0:
        bar += f"{C_CYAN}" + "░" * comment_chars + f"{C_END}"
    if blank_chars > 0:
        bar += f"{C_GRAY}" + "▒" * blank_chars + f"{C_END}"
        
    stats_str = f"Code: {code_pct:.1%}, Comments: {comment_pct:.1%}, Blank: {blank_pct:.1%}"
    return f"[{bar}] {C_BOLD}{stats_str}{C_END}"

def main():
    parser = argparse.ArgumentParser(description="Beautiful, fast & multithreaded lines of code counter.")
    parser.add_argument("path", nargs="?", default=".", help="Root directory to scan (default: current directory)")
    parser.add_argument("--exclude", nargs="*", default=[], help="Additional paths/patterns to exclude")
    parser.add_argument("--format", choices=["table", "markdown", "json"], default="table", help="Output format")
    parser.add_argument("--by-file", action="store_true", help="Show line counts for each individual file")
    parser.add_argument("--output", help="Write output to a file")
    parser.add_argument("--no-git", action="store_true", help="Do not use git to find files, scan manually")
    parser.add_argument("--no-color", action="store_true", help="Disable colored output")
    parser.add_argument("--include-generated", action="store_true", help="Include auto-generated code files in the scan")
    
    args = parser.parse_args()
    
    use_color = supports_color() and not args.no_color
    if args.output:
        use_color = False
        
    root_dir = Path(args.path).resolve()
    if not root_dir.is_dir():
        print(f"Error: Path '{root_dir}' is not a directory.", file=sys.stderr)
        sys.exit(1)
        
    start_time = time.time()
    
    # Get files list
    files = None
    used_git = False
    if not args.no_git:
        files = get_git_files(root_dir)
        if files is not None:
            used_git = True
            
    if files is None:
        files = get_files_manually(root_dir, custom_excludes=args.exclude)
    else:
        # Always filter files using should_ignore to respect DEFAULT_IGNORE_FILES & DEFAULT_IGNORE_DIRS
        files = [f for f in files if not should_ignore(f, root_dir, custom_excludes=args.exclude)]
        
    # Filter out directories
    files = [f for f in files if f.is_file()]
    
    # Collect statistics
    stats = {}
    total_files = 0
    total_lines = 0
    total_blank = 0
    total_comment = 0
    total_code = 0
    
    max_workers = min(32, (os.cpu_count() or 1) * 4)
    num_files = len(files)
    
    if num_files == 0:
        print("No source code files found to scan.")
        return
        
    # Process files concurrently
    with concurrent.futures.ThreadPoolExecutor(max_workers=max_workers) as executor:
        future_to_file = {executor.submit(count_lines, f, args.include_generated): f for f in files}
        
        for idx, future in enumerate(concurrent.futures.as_completed(future_to_file)):
            file_path = future_to_file[future]
            
            # Print progress to stderr
            if sys.stderr.isatty() and idx % 10 == 0:
                sys.stderr.write(f"\rScanning files: {idx}/{num_files}...")
                sys.stderr.flush()
                
            try:
                res = future.result()
                if res is None:
                    continue
                    
                total, blank, comment, code = res
                total_files += 1
                total_lines += total
                total_blank += blank
                total_comment += comment
                total_code += code
                
                if args.by_file:
                    rel_path = str(file_path.relative_to(root_dir))
                    stats[rel_path] = {
                        'total': total,
                        'blank': blank,
                        'comment': comment,
                        'code': code
                    }
                else:
                    ext = file_path.suffix.lower()
                    group_name = ext if ext else file_path.name
                    if not group_name:
                        group_name = "unknown"
                        
                    if group_name not in stats:
                        stats[group_name] = {'count': 0, 'total': 0, 'blank': 0, 'comment': 0, 'code': 0}
                    stats[group_name]['count'] += 1
                    stats[group_name]['total'] += total
                    stats[group_name]['blank'] += blank
                    stats[group_name]['comment'] += comment
                    stats[group_name]['code'] += code
            except Exception:
                pass
                
    if sys.stderr.isatty():
        sys.stderr.write("\r" + " " * 40 + "\r")
        sys.stderr.flush()
        
    elapsed_time = time.time() - start_time
    
    # Prepare data for output formatting
    if args.by_file:
        headers = ["File", "Lines", "Blank", "Comment", "Code"]
        sorted_stats = sorted(stats.items(), key=lambda x: x[1]['code'], reverse=True)
        rows = [[name, data['total'], data['blank'], data['comment'], data['code']] for name, data in sorted_stats]
        total_row = ["Total", total_lines, total_blank, total_comment, total_code]
    else:
        headers = ["Language/Extension", "Files", "Lines", "Blank", "Comment", "Code"]
        sorted_stats = sorted(stats.items(), key=lambda x: x[1]['code'], reverse=True)
        rows = [[ext, data['count'], data['total'], data['blank'], data['comment'], data['code']] for ext, data in sorted_stats]
        total_row = ["Total", total_files, total_lines, total_blank, total_comment, total_code]
        
    # Generate output string
    output_str = ""
    if args.format == "table":
        method_str = "Git (filtered)" if used_git else "Directory Traversal"
        C_BOLD = '\033[1m' if use_color else ''
        C_CYAN = '\033[96m' if use_color else ''
        C_END = '\033[0m' if use_color else ''
        
        output_str += f"{C_BOLD}🚀 CLOC-Py - Count Lines of Code{C_END}\n"
        output_str += f"{C_CYAN}---------------------------------{C_END}\n"
        output_str += f"📂 Scan Path:   {root_dir}\n"
        output_str += f"🔍 Scan Method: {method_str}\n"
        output_str += f"⏱️  Scan Time:   {elapsed_time:.3f}s\n"
        output_str += f"📦 Total Files: {total_files}\n\n"
        
        output_str += generate_table(headers, rows, total_row, use_color=use_color)
        
        # Add visual distribution bar
        bar_str = get_distribution_bar(total_code, total_comment, total_blank, use_color=use_color)
        if bar_str:
            output_str += f"\n\n📊 Distribution:\n{bar_str}\n"
            
    elif args.format == "markdown":
        output_str += f"### CLOC-Py Results\n"
        output_str += f"- **Scan Path**: `{root_dir}`\n"
        output_str += f"- **Scan Method**: `{'Git (filtered)' if used_git else 'Manual'}`\n"
        output_str += f"- **Scan Time**: `{elapsed_time:.3f}s`\n\n"
        output_str += generate_markdown(headers, rows, total_row)
        
    elif args.format == "json":
        output_str += json.dumps({
            "summary": {
                "scan_path": str(root_dir),
                "scan_method": "git" if used_git else "manual",
                "scan_time_seconds": elapsed_time,
                "total_files": total_files,
                "total_lines": total_lines,
                "blank": total_blank,
                "comment": total_comment,
                "code": total_code
            },
            "details": stats
        }, indent=2)
        
    # Write to file or print
    if args.output:
        try:
            with open(args.output, 'w', encoding='utf-8') as f:
                f.write(output_str + "\n")
            print(f"Results successfully written to '{args.output}'")
        except IOError as e:
            print(f"Error writing to output file: {e}", file=sys.stderr)
    else:
        print(output_str)

if __name__ == "__main__":
    main()
