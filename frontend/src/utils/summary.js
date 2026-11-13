import { marked } from 'marked'

/**
 * Returns true if the stored summary is a standalone HTML document.
 */
export function isHtmlSummary(content) {
  if (!content || typeof content !== 'string') return false
  return /^\s*(<!doctype\s+html|<html[\s>])/i.test(content)
}

/**
 * Extracts readable plain text from either an HTML summary document or Markdown.
 */
export function extractSummaryText(content) {
  if (!content || typeof content !== 'string') return ''
  if (!isHtmlSummary(content)) return content

  if (typeof DOMParser !== 'undefined') {
    const doc = new DOMParser().parseFromString(content, 'text/html')
    doc.querySelectorAll('style, script, noscript').forEach((el) => el.remove())
    return (doc.body?.textContent || '').replace(/\s+/g, ' ').trim()
  }

  return content
    .replace(/<style[\s\S]*?<\/style>/gi, ' ')
    .replace(/<script[\s\S]*?<\/script>/gi, ' ')
    .replace(/<[^>]+>/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

/**
 * Extracts a short preview snippet for recording cards and lists.
 */
export function extractSummaryPreview(content, maxLen = 95) {
  if (!content || typeof content !== 'string') return null

  if (isHtmlSummary(content)) {
    let snippet = ''
    if (typeof DOMParser !== 'undefined') {
      const doc = new DOMParser().parseFromString(content, 'text/html')
      doc.querySelectorAll('style, script, noscript').forEach((el) => el.remove())
      const lead = doc.querySelector('.lead') || doc.querySelector('article p') || doc.querySelector('p')
      snippet = (lead?.textContent || doc.body?.textContent || '').replace(/\s+/g, ' ').trim()
    } else {
      snippet = extractSummaryText(content)
    }
    if (!snippet) return null
    return snippet.length > maxLen ? `${snippet.substring(0, maxLen)}...` : snippet
  }

  const lines = content.split('\n')
  const firstTextLine = lines.find((l) => l.trim().length > 10 && !l.startsWith('#'))
  return firstTextLine ? `${firstTextLine.substring(0, maxLen)}...` : null
}

/**
 * Counts words in either an HTML summary document or Markdown.
 */
export function countSummaryWords(content) {
  const text = extractSummaryText(content).trim()
  if (!text) return 0
  return text.split(/\s+/).length
}

/**
 * Renders legacy Markdown summaries to HTML.
 */
export function renderMarkdownSummary(content) {
  return content ? marked(content) : ''
}
