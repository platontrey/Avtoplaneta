import re
import sys

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # Remove clear buttons. Use .*? with re.DOTALL
    content = re.sub(r'\{\s*[a-zA-Z0-9_.]+\s*&&\s*\(\s*<button.*?title="Очистить".*?</button>\s*\)\s*\}', '', content, flags=re.DOTALL)
    
    # Now replace standard Select with ClearableSelect.
    # We want to replace:
    # <Select value={VAR} onValueChange={(val) => setValue('KEY', val)}>
    #     <SelectTrigger id="ID" className="CLASS">
    #         <SelectValue placeholder="PLACEHOLDER" />
    #     </SelectTrigger>
    #     <SelectContent>
    #          ITEMS
    #     </SelectContent>
    # </Select>
    
    # Let's find Select patterns and parse them.
    # Pattern to find the whole Select block:
    # We can match from <Select to </Select>
    
    def repl(m):
        select_tag = m.group(0)
        # Extract props from <Select
        val_match = re.search(r'value=\{([^}]+)\}', select_tag)
        onchange_match = re.search(r'onValueChange=\{([^}]+)\}', select_tag)
        
        # Extract from SelectTrigger
        id_match = re.search(r'<SelectTrigger[^>]*id="([^"]+)"', select_tag)
        class_match = re.search(r'<SelectTrigger[^>]*className="([^"]+)"', select_tag)
        
        # Extract placeholder
        placeholder_match = re.search(r'<SelectValue[^>]*placeholder="([^"]+)"', select_tag)
        
        # Extract items
        items_match = re.search(r'<SelectContent>(.*?)</SelectContent>', select_tag, flags=re.DOTALL)
        
        if val_match and onchange_match and placeholder_match and items_match:
            value = val_match.group(1)
            onchange = onchange_match.group(1)
            placeholder = placeholder_match.group(1)
            items = items_match.group(1)
            
            # id and class might not exist
            id_attr = f' id="{id_match.group(1)}"' if id_match else ''
            class_attr = f' className="{class_match.group(1)}"' if class_match else ''
            
            return f'<ClearableSelect\n    value={{{value}}}\n    onValueChange={{{onchange}}}\n    placeholder="{placeholder}"{id_attr}{class_attr}\n>\n    {items.strip()}\n</ClearableSelect>'
        
        return select_tag

    # Substitute <Select ... </Select>
    # But wait, what if Select has nested Selects? Unlikely here.
    content = re.sub(r'<Select\b.*?</Select>', repl, content, flags=re.DOTALL)
    
    if '<ClearableSelect' in content and 'import { ClearableSelect }' not in content:
         content = content.replace('import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"', 'import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"\nimport { ClearableSelect } from "./ClearableSelect"')
         # Sometimes it's double quotes
         content = content.replace('import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";', 'import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";\nimport { ClearableSelect } from "./ClearableSelect";')
    
    with open(filepath, 'w') as f:
        f.write(content)

process_file("frontend/src/components/AddPart.tsx")
process_file("frontend/src/components/DefectReport.tsx")

