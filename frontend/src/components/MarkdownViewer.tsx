/*
 * Copyright (c) 2025-2026 Avtoplaneta. All rights reserved.
 */

import React, { useState } from 'react';
import { Copy, Check } from 'lucide-react';

interface MarkdownViewerProps {
  content: string;
}

export default function MarkdownViewer({ content }: MarkdownViewerProps) {
  const [copiedCodeIndex, setCopiedCodeIndex] = useState<number | null>(null);

  const copyCode = (codeText: string, index: number) => {
    navigator.clipboard.writeText(codeText);
    setCopiedCodeIndex(index);
    setTimeout(() => setCopiedCodeIndex(null), 2000);
  };

  // Парсинг инлайн стилей (bold, italic, code, links, images)
  const renderInline = (text: string) => {
    const parts: React.ReactNode[] = [];
    let remaining = text;
    let keyIdx = 0;

    while (remaining.length > 0) {
      // Image: ![alt](url)
      const imgMatch = remaining.match(/^!\[([^\]]*)\]\(([^)]+)\)/);
      if (imgMatch) {
        parts.push(
          <img
            key={keyIdx++}
            src={imgMatch[2]}
            alt={imgMatch[1]}
            className="my-3 max-w-full rounded-lg border border-border shadow-sm inline-block"
            loading="lazy"
          />
        );
        remaining = remaining.slice(imgMatch[0].length);
        continue;
      }

      // Inline code: `code`
      const codeMatch = remaining.match(/^`([^`]+)`/);
      if (codeMatch) {
        parts.push(
          <code key={keyIdx++} className="px-1.5 py-0.5 rounded bg-muted font-mono text-xs text-primary font-semibold">
            {codeMatch[1]}
          </code>
        );
        remaining = remaining.slice(codeMatch[0].length);
        continue;
      }

      // Bold: **text**
      const boldMatch = remaining.match(/^\*\*([^*]+)\*\*/);
      if (boldMatch) {
        parts.push(
          <strong key={keyIdx++} className="font-bold text-foreground">
            {boldMatch[1]}
          </strong>
        );
        remaining = remaining.slice(boldMatch[0].length);
        continue;
      }

      // Italic: *text* or _text_
      const italicMatch = remaining.match(/^(\*|_)([^*_]+)\1/);
      if (italicMatch) {
        parts.push(
          <em key={keyIdx++} className="italic text-muted-foreground">
            {italicMatch[2]}
          </em>
        );
        remaining = remaining.slice(italicMatch[0].length);
        continue;
      }

      // Link: [text](url)
      const linkMatch = remaining.match(/^\[([^\]]+)\]\(([^)]+)\)/);
      if (linkMatch) {
        const isExternal = linkMatch[2].startsWith('http');
        parts.push(
          <a
            key={keyIdx++}
            href={linkMatch[2]}
            target={isExternal ? '_blank' : undefined}
            rel={isExternal ? 'noreferrer noopener' : undefined}
            className="text-primary hover:underline font-medium"
          >
            {linkMatch[1]}
          </a>
        );
        remaining = remaining.slice(linkMatch[0].length);
        continue;
      }

      // Обычный текст до следующего спецсимвола
      const nextSpecial = remaining.search(/[`*![_]/);
      if (nextSpecial === -1) {
        parts.push(remaining);
        break;
      } else if (nextSpecial === 0) {
        parts.push(remaining[0]);
        remaining = remaining.slice(1);
      } else {
        parts.push(remaining.slice(0, nextSpecial));
        remaining = remaining.slice(nextSpecial);
      }
    }

    return parts;
  };

  // Построчный парсинг блоков Markdown
  const lines = content.split('\n');
  const elements: React.ReactNode[] = [];
  let i = 0;
  let codeBlockCounter = 0;

  while (i < lines.length) {
    const rawLine = lines[i];
    const trimmed = rawLine.trim();

    // Пустые строки
    if (!trimmed) {
      i++;
      continue;
    }

    // Блоки кода ```lang
    if (trimmed.startsWith('```')) {
      const lang = trimmed.slice(3).trim();
      const codeLines: string[] = [];
      i++;
      while (i < lines.length && !lines[i].trim().startsWith('```')) {
        codeLines.push(lines[i]);
        i++;
      }
      i++; // закрывающий ```
      const codeText = codeLines.join('\n');
      const currentIndex = codeBlockCounter++;

      elements.push(
        <div key={`code-${i}`} className="my-4 rounded-xl border border-border bg-muted/60 overflow-hidden shadow-sm">
          <div className="flex items-center justify-between px-4 py-2 border-b border-border bg-muted text-xs font-mono text-muted-foreground">
            <span className="font-semibold">{lang || 'text'}</span>
            <button
              onClick={() => copyCode(codeText, currentIndex)}
              className="flex items-center gap-1.5 px-2.5 py-1 rounded hover:bg-card text-xs transition-colors"
            >
              {copiedCodeIndex === currentIndex ? (
                <>
                  <Check className="w-3.5 h-3.5 text-emerald-500" />
                  <span className="text-emerald-500 font-medium">Скопировано</span>
                </>
              ) : (
                <>
                  <Copy className="w-3.5 h-3.5" />
                  <span>Копировать</span>
                </>
              )}
            </button>
          </div>
          <pre className="p-4 text-xs font-mono overflow-x-auto text-foreground leading-relaxed whitespace-pre">
            {codeText}
          </pre>
        </div>
      );
      continue;
    }

    // Заголовки #
    if (trimmed.startsWith('# ')) {
      elements.push(
        <h1 key={`h1-${i}`} className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground mt-8 mb-4 border-b border-border pb-3">
          {renderInline(trimmed.slice(2))}
        </h1>
      );
      i++;
      continue;
    }
    if (trimmed.startsWith('## ')) {
      elements.push(
        <h2 key={`h2-${i}`} className="text-xl sm:text-2xl font-bold tracking-tight text-foreground mt-7 mb-3 border-b border-border/40 pb-2">
          {renderInline(trimmed.slice(3))}
        </h2>
      );
      i++;
      continue;
    }
    if (trimmed.startsWith('### ')) {
      elements.push(
        <h3 key={`h3-${i}`} className="text-lg font-semibold text-foreground mt-5 mb-2">
          {renderInline(trimmed.slice(4))}
        </h3>
      );
      i++;
      continue;
    }
    if (trimmed.startsWith('#### ')) {
      elements.push(
        <h4 key={`h4-${i}`} className="text-base font-semibold text-foreground mt-4 mb-1.5">
          {renderInline(trimmed.slice(5))}
        </h4>
      );
      i++;
      continue;
    }
    if (trimmed.startsWith('##### ')) {
      elements.push(
        <h5 key={`h5-${i}`} className="text-sm font-semibold text-foreground mt-3 mb-1">
          {renderInline(trimmed.slice(6))}
        </h5>
      );
      i++;
      continue;
    }

    // Цитаты / alerts >
    if (trimmed.startsWith('>')) {
      const quoteLines: string[] = [];
      while (i < lines.length && lines[i].trim().startsWith('>')) {
        quoteLines.push(lines[i].trim().replace(/^>\s?/, ''));
        i++;
      }
      elements.push(
        <div key={`quote-${i}`} className="border-l-4 border-primary bg-primary/5 p-4 rounded-r-xl my-4 text-sm text-foreground space-y-1">
          {quoteLines.map((ql, qIdx) => (
            <p key={qIdx} className="leading-relaxed">
              {renderInline(ql)}
            </p>
          ))}
        </div>
      );
      continue;
    }

    // Таблицы Markdown | col | col |
    if (trimmed.includes('|') && lines[i + 1] && lines[i + 1].includes('|') && lines[i + 1].includes('-')) {
      const headerCols = trimmed
        .split('|')
        .slice(1, -1)
        .map((c) => c.trim());
      i += 2; // пропускаем разделитель |--|--|

      const rows: string[][] = [];
      while (i < lines.length && lines[i].includes('|')) {
        const rowCols = lines[i]
          .split('|')
          .slice(1, -1)
          .map((c) => c.trim());
        rows.push(rowCols);
        i++;
      }

      elements.push(
        <div key={`table-${i}`} className="my-4 overflow-x-auto border border-border rounded-xl shadow-sm">
          <table className="w-full text-xs text-left">
            <thead>
              <tr className="bg-muted/70 border-b border-border text-muted-foreground font-semibold">
                {headerCols.map((col, cIdx) => (
                  <th key={cIdx} className="py-2.5 px-3">
                    {renderInline(col)}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {rows.map((row, rIdx) => (
                <tr key={rIdx} className="hover:bg-muted/30 transition-colors">
                  {row.map((cell, cIdx) => (
                    <td key={cIdx} className="py-2.5 px-3 align-top leading-relaxed">
                      {renderInline(cell)}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      );
      continue;
    }

    // Маркированные списки (- или *) и чек-листы (- [ ] или - [x])
    if (trimmed.startsWith('- ') || trimmed.startsWith('* ')) {
      const listItems: { text: string; checked?: boolean }[] = [];
      while (i < lines.length && (lines[i].trim().startsWith('- ') || lines[i].trim().startsWith('* '))) {
        let rawItem = lines[i].trim().slice(2).trim();
        let checked: boolean | undefined = undefined;
        if (rawItem.startsWith('[ ] ')) {
          checked = false;
          rawItem = rawItem.slice(4).trim();
        } else if (rawItem.startsWith('[x] ') || rawItem.startsWith('[X] ')) {
          checked = true;
          rawItem = rawItem.slice(4).trim();
        }
        listItems.push({ text: rawItem, checked });
        i++;
      }
      elements.push(
        <ul key={`ul-${i}`} className="space-y-1.5 my-3 text-sm text-foreground/90">
          {listItems.map((item, idx) => (
            <li key={idx} className="flex items-start gap-2.5 leading-relaxed">
              {item.checked !== undefined ? (
                <span
                  className={`inline-flex items-center justify-center w-4 h-4 mt-0.5 rounded border text-[10px] font-bold shrink-0 ${
                    item.checked
                      ? 'bg-primary text-primary-foreground border-primary'
                      : 'border-muted-foreground/40 bg-muted/20 text-transparent'
                  }`}
                >
                  {item.checked ? '✓' : ''}
                </span>
              ) : (
                <span className="inline-block w-1.5 h-1.5 rounded-full bg-primary/70 mt-2 shrink-0" />
              )}
              <span className="flex-1">{renderInline(item.text)}</span>
            </li>
          ))}
        </ul>
      );
      continue;
    }

    // Нумерованные списки 1. 2.
    if (/^\d+\.\s/.test(trimmed)) {
      const listItems: string[] = [];
      let listStart = 1;
      const startMatch = trimmed.match(/^(\d+)\.\s/);
      if (startMatch) {
        listStart = parseInt(startMatch[1], 10);
      }
      while (i < lines.length && /^\d+\.\s/.test(lines[i].trim())) {
        listItems.push(lines[i].trim().replace(/^\d+\.\s/, ''));
        i++;
      }
      elements.push(
        <ol key={`ol-${i}`} start={listStart} className="list-decimal list-outside pl-5 space-y-1.5 my-3 text-sm text-foreground/90">
          {listItems.map((item, idx) => (
            <li key={idx} className="leading-relaxed">
              {renderInline(item)}
            </li>
          ))}
        </ol>
      );
      continue;
    }

    // Горизонтальный разделитель ---
    if (trimmed === '---' || trimmed === '***' || trimmed === '___') {
      elements.push(<hr key={`hr-${i}`} className="my-6 border-border" />);
      i++;
      continue;
    }

    // Обычный абзац
    elements.push(
      <p key={`p-${i}`} className="text-sm text-foreground/90 leading-relaxed my-2.5">
        {renderInline(rawLine)}
      </p>
    );
    i++;
  }

  return <div className="space-y-1">{elements}</div>;
}
