/**
 * Kiw DSL — JS/TS parser (mirrors Go packages/kiw)
 *
 * .kiw 2.0 Modern DSL (Svelte-Style Surface):
 * ---
 * title: Landing
 * layout: Base
 * ---
 * <div class="container" class:active={isActive}>
 *   <h1>{title}</h1>
 *   {#if user.role == 'admin'}
 *     <p>Welcome admin {user.name}!</p>
 *   {:else}
 *     <p>Welcome guest!</p>
 *   {/if}
 *   <ul>
 *     {#each tags as tag}
 *       <li>{tag}</li>
 *     {/each}
 *   </ul>
 *   <Card {title} variant="primary" />
 * </div>
 * <style scoped>.container{padding:1rem}</style>
 * <script>console.log(1)</script>
 */

export interface StyleBlock { lang: string; scoped: boolean; content: string }
export interface ScriptBlock { lang: string; hydrate: string; server: boolean; compute: boolean; content: string }

export interface ComponentCall {
  name: string
  hydrate?: string
  props?: Record<string, string>
  raw: string
}

export interface ImportDecl {
  localName: string
  importedName?: string
  source: string
  isDefault?: boolean
  isNamespace?: boolean
}

export interface ExportDecl {
  name: string
  source?: string
  kind: string
}

export interface PropSpec {
  name: string
  type?: string
  default?: string
  required?: boolean
}

export interface SlotSpec {
  name: string
  hasFallback?: boolean
  fallback?: string
}

export interface KiwModule {
  frontmatter: Record<string, any>
  body: string
  styles: string[]
  scripts: string[]
  styleBlocks: StyleBlock[]
  scriptBlocks: ScriptBlock[]
  markdown: string[]
  components?: ComponentCall[]
  componentNames?: string[]
  imports?: ImportDecl[]
  exports?: ExportDecl[]
  props?: Record<string, PropSpec>
  defaults?: Record<string, any>
  slots?: SlotSpec[]
  raw: string
}

const styleRe = /<style([^>]*)>([\s\S]*?)<\/style>/gi
const scriptRe = /<script([^>]*)>([\s\S]*?)<\/script>/gi
const markdownRe = /<markdown[^>]*>([\s\S]*?)<\/markdown>/gi
const attrRe = /([a-zA-Z_:][-a-zA-Z0-9_:.]*)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>`]+)))?/g

const compNamePattern = "(?:[a-zA-Z_][a-zA-Z0-9_-]*:)?[A-Z][a-zA-Z0-9_]*(?:\\.[A-Z][a-zA-Z0-9_]*)*"
const selfClosingCompRe = new RegExp(`<(${compNamePattern})([^>]*?)\\s*\\/>`, "g")
const pairedCompRe = new RegExp(`<(${compNamePattern})([^>]*)>([\\s\\S]*?)<\\/\\1>`, "g")
const mustacheCompRe = /\{\{\s*component\s+"([^"]+)"(?:\s+([^{}]+))?\s*\}\}/g
const compAttrRe = /([a-zA-Z_:][-a-zA-Z0-9_:.]*)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|\{([^}]*)\}|([^\s"'=<>`]+)))?/g

const identPathRe = /^[a-zA-Z_$][a-zA-Z0-9_$]*(?:\.[a-zA-Z0-9_$]+)*$/
const eachBlockRe = /\{#each\s+([a-zA-Z0-9_$.]+)\s+as\s+([a-zA-Z0-9_$]+)(?:\s*,\s*([a-zA-Z0-9_$]+))?\s*\}([\s\S]*?)\{\/each\}/g
const ifTagRe = /\{#if\s+([^}]+)\}/g
const elseIfTagRe = /\{:else\s+if\s+([^}]+)\}/g
const elseTagRe = /\{:else\}/g
const endIfTagRe = /\{\/if\}/g
const classDirRe = /class:([a-zA-Z0-9_-]+)(?:=\{([^}]*)\}|="([^"]*)"|='([^']*)')?/g
const classAttrRe = /\bclass=(?:"([^"]*)"|'([^']*)')/
const tagWithClassDirRe = /<([a-zA-Z0-9_-]+)([^>]*?\sclass:[^>]*?)>/g
const punnedPropRe = /(?:^|[\s/])\{([a-zA-Z_][a-zA-Z0-9_]*)\}/g

function capitalize(s: string): string {
  if (!s) return ""
  return s.charAt(0).toUpperCase() + s.slice(1)
}

function toGoField(path: string, loopVars?: Set<string>): string {
  if (!path) return ""
  const parts = path.split(".")
  const root = parts[0]
  if (root.startsWith("$")) {
    return [root, ...parts.slice(1).map(capitalize)].join(".")
  }
  if (loopVars && loopVars.has(root)) {
    return ["$" + root, ...parts.slice(1).map(capitalize)].join(".")
  }
  if (root.startsWith(".")) {
    return [root, ...parts.slice(1).map(capitalize)].join(".")
  }
  const prefix = loopVars && loopVars.size > 0 ? "$." : "."
  return [prefix + capitalize(root), ...parts.slice(1).map(capitalize)].join(".")
}

function desugarInfixExpr(expr: string, loopVars?: Set<string>): string {
  let e = expr.trim()
  e = e.replace(/'([^']*)'/g, '"$1"')

  const relOpMap: Record<string, string> = {
    "==": "eq",
    "!=": "ne",
    "<=": "le",
    ">=": "ge",
    "<": "lt",
    ">": "gt",
  }

  function parseAnd(sub: string): string {
    const orParts = sub.split("||").map((s) => s.trim())
    if (orParts.length > 1) {
      return `(or ${orParts.map(parseAnd).join(" ")})`
    }
    const andParts = sub.split("&&").map((s) => s.trim())
    if (andParts.length > 1) {
      return `(and ${andParts.map(parseRel).join(" ")})`
    }
    return parseRel(sub)
  }

  function parseRel(sub: string): string {
    let s = sub.trim()
    if (s.startsWith("(") && s.endsWith(")")) {
      s = s.slice(1, -1).trim()
      return parseAnd(s)
    }
    if (s.startsWith("!")) {
      return `(not ${parseRel(s.slice(1).trim())})`
    }
    for (const [op, fn] of Object.entries(relOpMap)) {
      const idx = s.indexOf(op)
      if (idx !== -1) {
        const left = parseOperand(s.slice(0, idx).trim())
        const right = parseOperand(s.slice(idx + op.length).trim())
        return `(${fn} ${left} ${right})`
      }
    }
    return parseOperand(s)
  }

  function parseOperand(op: string): string {
    if (op.startsWith('"') && op.endsWith('"')) return op
    if (op === "true" || op === "false" || !isNaN(Number(op))) return op
    return toGoField(op, loopVars)
  }

  let res = parseAnd(e)
  if (res.startsWith("(") && res.endsWith(")")) {
    res = res.slice(1, -1)
  }
  return res
}

function desugarInterpolation(src: string, loopVars?: Set<string>): string {
  let out = ""
  let i = 0
  const n = src.length
  while (i < n) {
    if (src[i] === "{" && src[i + 1] === "{") {
      const end = src.indexOf("}}", i + 2)
      if (end !== -1) {
        out += src.slice(i, end + 2)
        i = end + 2
        continue
      }
    }
    if (src[i] === "{" && src[i + 1] !== "{") {
      const closeIdx = src.indexOf("}", i + 1)
      if (closeIdx !== -1) {
        const content = src.slice(i + 1, closeIdx).trim()
        if (identPathRe.test(content) && content !== "true" && content !== "false" && content !== "nil" && content !== "null") {
          out += `{{${toGoField(content, loopVars)}}}`
          i = closeIdx + 1
          continue
        }
      }
    }
    out += src[i]
    i++
  }
  return out
}

function desugarClassDirectives(src: string): string {
  return src.replace(tagWithClassDirRe, (_m, tagName, attrContent) => {
    let cleanAttr = attrContent.replace(classDirRe, "").trimEnd()
    const hasClassAttr = classAttrRe.test(cleanAttr)
    const condClasses: string[] = []
    classDirRe.lastIndex = 0
    let cm: RegExpExecArray | null
    let idx = 0
    while ((cm = classDirRe.exec(attrContent)) !== null) {
      const clsName = cm[1]
      let expr = (cm[2] ?? cm[3] ?? cm[4] ?? "").trim()
      if (!expr) expr = clsName
      const goCond = desugarInfixExpr(expr)
      const space = idx === 0 && !hasClassAttr ? "" : " "
      condClasses.push(`{{if ${goCond}}}${space}${clsName}{{end}}`)
      idx++
    }
    const joined = condClasses.join("")
    if (hasClassAttr) {
      cleanAttr = cleanAttr.replace(classAttrRe, (_full, dQ, sQ) => {
        const q = dQ !== undefined ? '"' : "'"
        const orig = dQ ?? sQ ?? ""
        return `class=${q}${orig}${joined}${q}`
      })
    } else {
      cleanAttr += ` class="${joined}"`
    }
    return `<${tagName}${cleanAttr}>`
  })
}

function desugarLoops(src: string): string {
  return src.replace(eachBlockRe, (_match, itemsExpr, itemVar, indexVar, body) => {
    const loopVars = new Set<string>([itemVar])
    if (indexVar) loopVars.add(indexVar)
    const goItems = toGoField(itemsExpr.trim())
    const rangeHeader = indexVar
      ? `{{range $${indexVar}, $${itemVar} := ${goItems}}}`
      : `{{range $${itemVar} := ${goItems}}}`
    const innerBody = desugarTemplate(body, loopVars)
    return `${rangeHeader}${innerBody}{{end}}`
  })
}

function desugarConditionals(src: string, loopVars?: Set<string>): string {
  let res = src.replace(ifTagRe, (_m, expr) => `{{if ${desugarInfixExpr(expr.trim(), loopVars)}}}`)
  res = res.replace(elseIfTagRe, (_m, expr) => `{{else if ${desugarInfixExpr(expr.trim(), loopVars)}}}`)
  res = res.replace(elseTagRe, "{{else}}")
  res = res.replace(endIfTagRe, "{{end}}")
  return res
}

function parseAttrs(tag: string): Record<string, string> {
  const out: Record<string, string> = {}
  let m: RegExpExecArray | null
  attrRe.lastIndex = 0
  while ((m = attrRe.exec(tag)) !== null) {
    const key = m[1].toLowerCase()
    const val = m[2] ?? m[3] ?? m[4] ?? ""
    out[key] = val
  }
  return out
}

function desugarComponent(name: string, attrStr: string, bodyContent: string, raw: string): [string, ComponentCall] {
  const call: ComponentCall = { name, props: {}, raw }
  const trimmedAttr = attrStr.trim()
  const trimmedBody = bodyContent.trim()

  if (!trimmedAttr && !trimmedBody) {
    return [`{{component "${name}"}}`, call]
  }

  if (trimmedAttr === "." && !trimmedBody) {
    call.props = { ".": "." }
    return [`{{component "${name}" .}}`, call]
  }

  if (trimmedAttr.startsWith("(dict") && !trimmedBody) {
    return [`{{component "${name}" ${trimmedAttr}}}`, call]
  }

  const dictPairs: string[] = []
  let cleanAttr = attrStr
  if (trimmedAttr) {
    punnedPropRe.lastIndex = 0
    let pm: RegExpExecArray | null
    while ((pm = punnedPropRe.exec(attrStr)) !== null) {
      const propName = pm[1]
      if (propName === "true" || propName === "false" || propName === "nil") continue
      call.props![propName] = "." + propName
      dictPairs.push(`"${propName}" .${propName}`)
    }
    cleanAttr = cleanAttr.replace(punnedPropRe, (m, p) => {
      if (p === "true" || p === "false" || p === "nil") return m
      return " "
    })
  }

  if (cleanAttr.trim()) {
    compAttrRe.lastIndex = 0
    let sm: RegExpExecArray | null
    while ((sm = compAttrRe.exec(cleanAttr)) !== null) {
      if (!sm[1] || !sm[1].trim()) continue
      const key = sm[1]
      if (!sm[0].includes("=")) {
        call.props![key] = "true"
        dictPairs.push(`"${key}" true`)
        continue
      }
      let val = ""
      let isExpr = false
      if (sm[2] !== undefined) {
        val = sm[2]
        if (val.startsWith("{{") && val.endsWith("}}")) {
          val = val.slice(2, -2).trim()
          isExpr = true
        } else if (val.startsWith("{") && val.endsWith("}")) {
          val = val.slice(1, -1).trim()
          isExpr = true
          if (identPathRe.test(val) && !val.startsWith(".") && !val.startsWith("$") && val !== "true" && val !== "false") {
            val = toGoField(val)
          }
        }
      } else if (sm[3] !== undefined) {
        val = sm[3]
      } else if (sm[4] !== undefined) {
        val = sm[4].trim()
        isExpr = true
        if (identPathRe.test(val) && !val.startsWith(".") && !val.startsWith("$") && val !== "true" && val !== "false") {
          val = toGoField(val)
        }
      } else if (sm[5] !== undefined) {
        val = sm[5]
        if (val === "true" || val === "false" || !isNaN(Number(val)) || val.startsWith(".") || val.startsWith("$")) {
          isExpr = true
        }
      }

      call.props![key] = val
      if (isExpr || val === "true" || val === "false" || !isNaN(Number(val))) {
        dictPairs.push(`"${key}" ${val}`)
      } else {
        dictPairs.push(`"${key}" ${JSON.stringify(val)}`)
      }
    }
  }

  if (trimmedBody) {
    call.props!["Body"] = trimmedBody
    dictPairs.push(`"Body" ${JSON.stringify(trimmedBody)}`)
  }

  if (dictPairs.length === 0) {
    return [`{{component "${name}"}}`, call]
  }
  return [`{{component "${name}" (dict ${dictPairs.join(" ")})}}`, call]
}

const slotFallbackRe = /<slot(?:\s+name=["']([a-zA-Z0-9_-]+)["'])?\s*>([\s\S]*?)<\/slot>/gi
const slotEmptyRe = /<slot(?:\s+name=["']([a-zA-Z0-9_-]+)["'])?\s*\/>/gi
const snippetDefRe = /\{#snippet\s+([a-zA-Z0-9_$]+)\s*(?:\(([^)]*)\))?\s*\}([\s\S]*?)\{\/snippet\}/gi
const renderCallRe = /\{@render\s+([a-zA-Z0-9_$]+)\s*(?:\(([^)]*)\))?\s*\}/gi

function desugarSlots(src: string): string {
  let res = src.replace(slotFallbackRe, (_m, name, fallback) => {
    const target = name ? "." + capitalize(name) : ".Body"
    const trimmed = (fallback || "").trim()
    if (!trimmed) return `{{${target}}}`
    return `{{if ${target}}}{{${target}}}{{else}}${fallback}{{end}}`
  })
  res = res.replace(slotEmptyRe, (_m, name) => {
    const target = name ? "." + capitalize(name) : ".Body"
    return `{{${target}}}`
  })
  return res
}

function desugarSnippets(src: string, loopVars?: Set<string>): string {
  const snippetParams: Record<string, string[]> = {}
  let res = src.replace(snippetDefRe, (_m, name, paramsRaw, body) => {
    const paramList: string[] = []
    const scope = new Set<string>(loopVars)
    if (paramsRaw && paramsRaw.trim()) {
      for (const p of paramsRaw.split(",")) {
        const tr = p.trim()
        if (tr) {
          paramList.push(tr)
          scope.add(tr)
        }
      }
    }
    snippetParams[name] = paramList
    const desugaredBody = desugarTemplate(body, scope)
    return `{{define "${name}"}}${desugaredBody}{{end}}`
  })

  res = res.replace(renderCallRe, (_m, name, argsRaw) => {
    if (!argsRaw || !argsRaw.trim()) {
      return `{{template "${name}" .}}`
    }
    const params = snippetParams[name] || []
    const args = argsRaw.split(",").map((a: string) => a.trim()).filter(Boolean)
    const pairs: string[] = []
    args.forEach((arg: string, i: number) => {
      let key = ""
      let val = arg
      if (arg.includes("=")) {
        const kv = arg.split("=")
        key = capitalize(kv[0].trim())
        val = kv.slice(1).join("=").trim()
      } else if (i < params.length) {
        key = capitalize(params[i])
      } else {
        key = `Arg${i}`
      }
      if (identPathRe.test(val) && val !== "true" && val !== "false" && !val.startsWith(".") && !val.startsWith("$") && isNaN(Number(val))) {
        val = toGoField(val, loopVars)
      } else if (!val.startsWith('"') && !val.startsWith("'") && isNaN(Number(val)) && val !== "true" && val !== "false" && !val.startsWith(".") && !val.startsWith("$")) {
        val = JSON.stringify(val)
      }
      pairs.push(`"${key}" ${val}`)
    })
    if (pairs.length === 0) return `{{template "${name}" .}}`
    return `{{template "${name}" (dict ${pairs.join(" ")})}}`
  })

  return res
}

export function desugarTemplate(src: string, loopVars?: Set<string>): string {
  let res = desugarClassDirectives(src)
  res = desugarLoops(res)
  res = desugarConditionals(res, loopVars)
  res = desugarSnippets(res, loopVars)
  res = res.replace(selfClosingCompRe, (match, name, attrStr) => {
    return desugarComponent(name, attrStr || "", "", match)[0]
  })
  res = res.replace(pairedCompRe, (match, name, attrStr, body) => {
    return desugarComponent(name, attrStr || "", body || "", match)[0]
  })
  res = desugarSlots(res)
  res = desugarInterpolation(res, loopVars)
  return res
}

export function parseKiw(src: string): KiwModule {
  let frontmatter: Record<string, any> = {}
  let body = src
  const trimmed = src.trimStart()
  if (trimmed.startsWith("---")) {
    const rest = trimmed.slice(3)
    const idx = rest.indexOf("\n---")
    if (idx >= 0) {
      const fmRaw = rest.slice(0, idx)
      body = rest.slice(idx + 4).replace(/^\r?\n/, "")
      try {
        frontmatter = parseYamlMinimal(fmRaw)
      } catch {}
    }
  }

  const styles: string[] = []
  const styleBlocks: StyleBlock[] = []
  body = body.replace(styleRe, (_m, attrs, inner) => {
    const a = parseAttrs(attrs)
    const content = inner.trim()
    styles.push(content)
    const lang = (a["lang"] || "css").toLowerCase()
    styleBlocks.push({ lang, scoped: "scoped" in a, content })
    return ""
  })

  const scripts: string[] = []
  const scriptBlocks: ScriptBlock[] = []
  body = body.replace(scriptRe, (_m, attrs, inner) => {
    const a = parseAttrs(attrs)
    const content = inner.trim()
    scripts.push(content)
    let lang = (a["lang"] || "js").toLowerCase()
    let hydrate = (a["hydrate"] || "").toLowerCase()
    const server = "server" in a
    const compute = "compute" in a
    if ((lang === "js" || lang === "ts" || lang === "go") && !server && !compute && !hydrate) hydrate = "load"
    scriptBlocks.push({ lang, hydrate, server, compute, content })
    return ""
  })

  const markdown: string[] = []
  body = body.replace(markdownRe, (_m, inner) => {
    markdown.push(inner.trim())
    return "\n" + inner.trim() + "\n"
  })

  const imports: ImportDecl[] = []
  const exports: ExportDecl[] = []
  const props: Record<string, PropSpec> = {}
  const defaults: Record<string, any> = {}
  const slots: SlotSpec[] = []

  // Parse frontmatter props/imports
  if (frontmatter.imports && typeof frontmatter.imports === "object") {
    for (const [k, v] of Object.entries(frontmatter.imports)) {
      const srcStr = String(v)
      const isNS = !srcStr.endsWith(".kiw") && (srcStr.startsWith("@") || !srcStr.includes("/"))
      imports.push({ localName: k, source: srcStr, isDefault: !isNS, isNamespace: isNS })
    }
  }

  // Parse @import in body
  body = body.replace(/^\s*@import\s+([A-Z][a-zA-Z0-9_]*)\s+from\s+['"]([^'"]+)['"]\s*;?/gm, (_m, name, s) => {
    imports.push({ localName: name, source: s, isDefault: true })
    return ""
  })
  body = body.replace(/^\s*@import\s+\*\s+as\s+([a-zA-Z0-9_]+)\s+from\s+['"]([^'"]+)['"]\s*;?/gm, (_m, name, s) => {
    imports.push({ localName: name, source: s, isNamespace: true })
    return ""
  })
  body = body.replace(/^\s*@import\s*\{([^}]+)\}\s*from\s+['"]([^'"]+)['"]\s*;?/gm, (_m, parts, s) => {
    parts.split(",").forEach((p: string) => {
      const tr = p.trim()
      if (!tr) return
      let imp = tr, loc = tr
      if (tr.includes(" as ")) {
        const sub = tr.split(" as ")
        imp = sub[0].trim()
        loc = sub[1].trim()
      }
      imports.push({ localName: loc, importedName: imp, source: s })
    })
    return ""
  })

  // Parse script imports / export let
  for (const sb of scriptBlocks) {
    const sc = sb.content
    let m: RegExpExecArray | null
    const defImpRe = /^\s*import\s+([A-Z][a-zA-Z0-9_]*)\s+from\s+['"]([^'"]+)['"]\s*;?/gm
    while ((m = defImpRe.exec(sc)) !== null) {
      imports.push({ localName: m[1], source: m[2], isDefault: true })
    }
    const expLetRe = /^\s*export\s+let\s+([a-zA-Z0-9_$]+)(?:\s*:\s*([^=;\n]+))?(?:\s*=\s*([^;\n]+))?\s*;?/gm
    while ((m = expLetRe.exec(sc)) !== null) {
      const propName = m[1]
      const propType = (m[2] || "").trim()
      const propDef = (m[3] || "").trim().replace(/^["']|["']$/g, "")
      props[propName] = { name: propName, type: propType, default: propDef }
      if (propDef) defaults[propName] = propDef
      exports.push({ name: propName, kind: "prop" })
    }
  }

  body = desugarTemplate(body)

  const components: ComponentCall[] = []
  const nameSet = new Set<string>()
  mustacheCompRe.lastIndex = 0
  let mm: RegExpExecArray | null
  while ((mm = mustacheCompRe.exec(body)) !== null) {
    const name = mm[1]
    components.push({ name, raw: mm[0] })
    nameSet.add(name)
  }

  return {
    frontmatter,
    body: body.trim(),
    styles,
    scripts,
    styleBlocks,
    scriptBlocks,
    markdown,
    components,
    componentNames: Array.from(nameSet).sort(),
    imports,
    exports,
    props,
    defaults,
    slots,
    raw: src,
  }
}

function parseYamlMinimal(src: string): Record<string, any> {
  const out: Record<string, any> = {}
  for (const line of src.split("\n")) {
    const t = line.trim()
    if (!t || t.startsWith("#")) continue
    const colon = t.indexOf(":")
    if (colon === -1) continue
    const k = t.slice(0, colon).trim()
    const v = t.slice(colon + 1).trim().replace(/^["']|["']$/g, "")
    out[k] = v
  }
  return out
}

export function parseKiwFile(src: string): KiwModule {
  return parseKiw(src)
}
