import sys
import re

def process(filepath):
    with open(filepath, 'r') as f:
        content = f.read()
    
    lines = content.split('\n')
    out_lines = []
    i = 0
    changed = False
    
    while i < len(lines):
        line = lines[i]
        
        if '<div className="grid grid-cols-1 sm:grid-cols-4 items-start sm:items-center gap-2 sm:gap-4">' in line:
            if i + 1 < len(lines) and '<Label' in lines[i+1]:
                html_for = ""
                m = re.search(r'htmlFor="([^"]+)"', lines[i+1])
                if m:
                    html_for = m.group(1)
                
                label_text = lines[i+2].strip()
                
                label_end = i + 1
                while '</Label>' not in lines[label_end]:
                    label_end += 1
                
                div_count = 1
                closing_j = -1
                for k in range(label_end + 1, len(lines)):
                    div_count += len(re.findall(r'<div\b', lines[k]))
                    div_count -= len(re.findall(r'</div>', lines[k]))
                    if div_count == 0:
                        closing_j = k
                        break
                
                if closing_j != -1:
                    indent = len(line) - len(line.lstrip())
                    spaces = " " * indent
                    
                    out_lines.append(f'{spaces}<FormRow label="{label_text}" htmlFor="{html_for}">')
                    
                    for k in range(label_end + 1, closing_j):
                        inner = lines[k]
                        inner = inner.replace('className="sm:col-span-3"', '')
                        inner = inner.replace('className="relative sm:col-span-3"', 'className="relative"')
                        # if empty classname left, we can clean it, but it's fine.
                        out_lines.append(inner)
                        
                    out_lines.append(f'{spaces}</FormRow>')
                    i = closing_j + 1
                    changed = True
                    continue
        
        out_lines.append(line)
        i += 1
        
    if changed:
        final_content = '\n'.join(out_lines)
        if 'import FormRow' not in final_content:
            final_content = final_content.replace('import { Label } from "@/components/ui/label";', 'import { Label } from "@/components/ui/label";\nimport FormRow from "./FormRow";')
        with open(filepath, 'w') as f:
            f.write(final_content)
        print(f"Refactored {filepath}")
    else:
        print(f"No changes in {filepath}")

for f in sys.argv[1:]:
    process(f)
