import os

# We compute project_root dynamically relative to this script's location
script_dir = os.path.dirname(os.path.abspath(__file__))
project_root = os.path.abspath(os.path.join(script_dir, ".."))

# Directories to ignore entirely
ignored_dirs = {
    "node_modules", ".git", ".venv", ".idea", ".vscode", "build", "ios", "android", 
    ".claude", "grafana", ".dart_tool", "gen", "sqlc"
}

# Individual files to ignore
ignored_files = {
    "common_parts.dart", "defect_report_screen.dart", "DefectReport.tsx", 
    "package-lock.json", "go.sum", "drom_session.json", "yarn.lock"
}

# Suffixes of files to ignore (generated files)
ignored_suffixes = [
    ".g.dart", ".freezed.dart", ".pb.go", ".pb.gw.go", ".sql.go", ".mocks.dart"
]

# Map file extensions/names to languages
lang_map = {
    ".go": "Go (Backend)",
    ".dart": "Dart (Mobile)",
    ".tsx": "TypeScript React (Frontend)",
    ".ts": "TypeScript (Frontend/Common)",
    ".sql": "SQL (Database Migrations/Queries)",
    ".py": "Python (Scripts/Utilities)",
    ".js": "JavaScript",
    ".css": "CSS",
    ".yml": "YAML Configs",
    ".yaml": "YAML Configs",
    ".sh": "Shell Scripts",
    "Dockerfile": "Dockerfiles",
    "Makefile": "Makefiles",
    ".env": "Environment Configurations"
}

stats = {}

for root, dirs, files in os.walk(project_root):
    # Prune directories in-place to ignore specified folders
    dirs[:] = [d for d in dirs if d not in ignored_dirs]
    
    for file in files:
        # Check explicit file exclusions
        if file in ignored_files:
            continue
            
        # Check generated file suffixes
        is_generated = any(file.endswith(suffix) for suffix in ignored_suffixes)
        if is_generated:
            continue
            
        ext = os.path.splitext(file)[1]
        lang = None
        
        if file in lang_map:
            lang = lang_map[file]
        elif ext in lang_map:
            lang = lang_map[ext]
            
        if lang:
            path = os.path.join(root, file)
            try:
                with open(path, "r", encoding="utf-8", errors="ignore") as f:
                    lines = f.readlines()
                    file_lines = len(lines)
                    blank_lines = sum(1 for line in lines if not line.strip())
                    
                    if lang not in stats:
                        stats[lang] = {"files": 0, "total": 0, "blank": 0}
                        
                    stats[lang]["files"] += 1
                    stats[lang]["total"] += file_lines
                    stats[lang]["blank"] += blank_lines
            except Exception:
                pass

print("### Общая статистика строк кода всего проекта (Без автогенерации и дефектных ведомостей)\n")
print("| Язык / Тип | Файлов | Всего строк | Код + Комментарии | Пустые строки |")
print("| :--- | :---: | :---: | :---: | :---: |")

grand_total_files = 0
grand_total_lines = 0
grand_total_code = 0
grand_total_blank = 0

for lang, data in sorted(stats.items(), key=lambda x: x[1]["total"], reverse=True):
    code_comments = data["total"] - data["blank"]
    print(f"| {lang} | {data['files']} | {data['total']} | {code_comments} | {data['blank']} |")
    grand_total_files += data["files"]
    grand_total_lines += data["total"]
    grand_total_code += code_comments
    grand_total_blank += data["blank"]

print("| **ИТОГО** | **{}** | **{}** | **{}** | **{}** |".format(
    grand_total_files, grand_total_lines, grand_total_code, grand_total_blank
))
