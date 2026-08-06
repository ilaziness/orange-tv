/**
 * injectAdScripts parses an ad code string (e.g. AdSense snippet), injects
 * non-script HTML into the container via innerHTML, and re-creates script
 * elements dynamically (since innerHTML won't execute scripts).
 *
 * Returns the injected script elements so the caller can remove them on cleanup.
 *
 * Used by AdCodeRenderer (React component) and Player (raw DOM in Artplayer layer).
 */
export function injectAdScripts(container: HTMLElement, code: string): HTMLScriptElement[] {
  const parser = new DOMParser()
  const doc = parser.parseFromString(code, 'text/html')
  const scripts = Array.from(doc.querySelectorAll('script'))

  // Remove script tags from the parsed doc, set remaining HTML as innerHTML.
  scripts.forEach((s) => s.remove())
  container.innerHTML = doc.body.innerHTML

  // Re-create each script element (innerHTML won't execute scripts).
  const injected: HTMLScriptElement[] = []
  for (const original of scripts) {
    const script = document.createElement('script')
    if (original.src) {
      script.src = original.src
      script.async = original.async
      script.defer = original.defer
    } else {
      script.textContent = original.textContent
    }
    // Copy other attributes (type, crossorigin, etc.)
    for (const attr of Array.from(original.attributes)) {
      if (attr.name !== 'src' && attr.name !== 'async' && attr.name !== 'defer') {
        script.setAttribute(attr.name, attr.value)
      }
    }
    container.appendChild(script)
    injected.push(script)
  }
  return injected
}
