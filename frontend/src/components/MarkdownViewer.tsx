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

  // Парсим инлайн стили (bold, italic, code, links)
  const renderInline = (text: string) => {
    // Регулярки для токенов
    const parts: React.ReactNode[] = [];
    let remaining = text;
    let keyIdx = 0;

    while (remaining.length > 0) {
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
      const italicMatch = remaining.match(/^\*([^*]+)\*/);
      if (italicMatch) {
        parts.push(
          <em key={keyIdx++} className="italic text-muted-foreground">
            {italicMatch[1]}
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
      const nextSpecial = remaining.search(/[`*[]/);
      if (nextSpecial === -1) {
        parts.push(remaining);
        break;
      } else if (nextSpecial === 0) {
        // Если спецсимвол не совпал с правилом, берем один символ
        parts.push(remaining[0]);
        remaining = remaining.slice(1);
      } else {
        parts.push(remaining.slice(0, nextSpecial));
        remaining = remaining.slice(nextSpecial);
      }
    }

    return parts;
  };

  // Построчный парсинг блоков
  const lines = content.split('\n');
  const elements: React.ReactNode[] = [];
  let i = 0;
  let codeBlockCounter = 0;

  while (i < lines.length) {
    const line = lines[i];

    // Пустые строки
    if (!line.trim()) {
      i++;
      continue;
    }

    // Блоки кода ```lang
    if (line.trim().startsWith('```')) {
      const lang = line.trim().slice(3).trim();
      const codeLines: string[] = [];
      i++;
      while (i < lines.length && !lines[i].trim().startsWith('```')) {
        codeLines.push(lines[i]);
        i++;
      }
      i++; // пропускаем закрывающий ```
      const codeText = codeLines.join('\n');
      const currentIndex = codeBlockCounter++;

      elements.push(
        <div key={`code-${i}`} className="my-4 rounded-xl border border-border bg-muted/60 overflow-hidden shadow-sm">
          <div className="flex items-center justify-between px-4 py-2 border-b border-border bg-muted text-xs font-mono text-muted-foreground">
            <span>{lang || 'text'}</span>
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
          <pre className="p-4 text-xs font-mono overflow-x-auto text-foreground leading-relaxed">
            {codeText}
          </pre>
        </div>
      );
      continue;
    }

    // Заголовки
    if (line.startsWith('# ')) {
      elements.push(
        <h1 key={`h1-${i}`} className="text-2xl sm:text-3xl font-extrabold tracking-tight text-foreground mt-8 mb-4 border-b border-border pb-3">
          {renderInline(line.slice(2))}
        </h1>
      );
      i++;
      continue;
    }
    if (line.startsWith('## ')) {
      elements.push(
        <h2 key={`h2-${i}`} className="text-xl sm:text-2xl font-bold tracking-tight text-foreground mt-7 mb-3">
          {renderInline(line.slice(3))}
        </h2>
      );
      i++;
      continue;
    }
    if (line.startsWith('### ')) {
      elements.push(
        <h3 key={`h3-${i}`} className="text-lg font-semibold text-foreground mt-5 mb-2">
          {renderInline(line.slice(4))}
        </h3>
      );
      i++;
      continue;
    }
    if (line.startsWith('#### ')) {
      elements.push(
        <h4 key={`h4-${i}`} className="text-base font-semibold text-foreground mt-4 mb-1.5">
          {renderInline(line.slice(5))}
        </h4>
      );
      i++;
      continue;
    }

    // Цитаты / alerts >
    if (line.startsWith('>')) {
      const quoteLines: string[] = [];
      while (i < lines.length && lines[i].startsWith('>')) {
        quoteLines.push(lines[i].replace(/^>\s?/, ''));
        i++;
      }
      elements.push(
        <div key={`quote-${i}`} className="border-l-4 border-primary bg-primary/5 p-4 rounded-r-xl my-4 text-sm text-foreground">
          {quoteLines.map((ql, qIdx) => (
            <p key={qIdx} className="mb-1 last:mb-0">
              {renderInline(ql)}
            </p>
          ))}
        </div>
      );
      continue;
    }

    // Таблицы Markdown | col | col |
    if (line.includes('|') && lines[i + 1] && lines[i + 1].includes('|') && lines[i + 1].includes('-')) {
      const headerCols = line
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
                <tr key={rIdx} className="hover:bg-muted/30">
                  {row.map((cell, cIdx) => (
                    <td key={cIdx} className="py-2 px-3">
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

    // Маркированные списки - или *
    if (line.trim().startsWith('- ') || line.trim().startsWith('* ')) {
      const listItems: string[] = [];
      while (i < lines.length && (lines[i].trim().startsWith('- ') || lines[i].trim().startsWith('* '))) {
        listItems.push(lines[i].trim().slice(2));
        i++;
      }
      elements.push(
        <ul key={`ul-${i}`} className="list-disc list-inside space-y-1.5 my-3 text-sm text-muted-foreground">
          {listItems.map((item, idx) => (
            <li key={idx} className="leading-relaxed">
              {renderInline(item)}
            </li>
          ))}
        </ul>
      );
      continue;
    }

    // Нумерованные списки 1. 2.
    if (/^\d+\.\s/.test(line.trim())) {
      const listItems: string[] = [];
      while (i < lines.length && /^\d+\.\s/.test(lines[i].trim())) {
        listItems.push(lines[i].trim().replace(/^\d+\.\s/, ''));
        i++;
      }
      elements.push(
        <ol key={`ol-${i}`} className="list-decimal list-inside space-y-1.5 my-3 text-sm text-muted-foreground">
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
    if (line.trim() === '---' || line.trim() === '***') {
      elements.push(<hr key={`hr-${i}`} className="my-6 border-border" />);
      i++;
      continue;
    }

    // Обычный абзац
    elements.push(
      <p key={`p-${i}`} className="text-sm text-foreground/90 leading-relaxed my-2.5">
        {renderInline(line)}
      </p>
    );
    i++;
  }

  return <div className="space-y-1">{elements}</div>;
}
