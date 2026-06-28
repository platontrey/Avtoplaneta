import re
import sys

def process(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # match { watch("field") && ( <button ... </button> )}
    content = re.sub(r'\{\s*watch\([^)]+\)\s*&&\s*\(\s*<button.*?title="Очистить".*?</button>\s*\)\s*\}', '', content, flags=re.DOTALL)
    
    with open(filepath, 'w') as f:
        f.write(content)

process("frontend/src/components/AddPart.tsx")
process("frontend/src/components/DefectReport.tsx")
