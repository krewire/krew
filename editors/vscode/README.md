# `.kiw` editor support

A TextMate grammar for Krewire `.kiw` single-file components, packaged as a
VS Code extension. It highlights:

- YAML frontmatter between the leading `---` lines
- template blocks `{#if}`, `{:else if}`, `{:else}`, `{/if}`, `{#each items as item, i}`, `{/each}`
- interpolation `{prop}` and Go template actions `{{ … }}`
- component tags `<Card …>` (PascalCase) with props, `{expr}` values and
  directives (`client:load`, `class:name={cond}`, `slot`)
- scoped `<style>` (CSS), `<script>` (JavaScript, or TypeScript with `lang="ts"`)
  and `<markdown>` blocks

`language-configuration.json` adds bracket matching, auto-closing pairs and
folding for `{#if}…{/if}` and `{#each}…{/each}`.

## Install in VS Code

```bash
# from this directory
npx @vscode/vsce package          # builds kiw-0.1.0.vsix
code --install-extension kiw-0.1.0.vsix
```

Or, without packaging, symlink the folder into your extensions directory and
reload the window:

```bash
ln -s "$PWD" ~/.vscode/extensions/krewire.kiw-0.1.0
```

## Other editors

`syntaxes/kiw.tmLanguage.json` is a plain TextMate grammar (scope
`text.html.kiw`); Sublime Text, Zed, Helix (via tree-sitter-less TextMate
import), IntelliJ (TextMate bundles) and GitHub Linguist-style highlighters can
load it directly. `samples/Sample.kiw` exercises every construct.

The grammar is checked by `go test ./tools/kiw/editors/...`: the JSON must
parse, every `match`/`begin`/`end` regex must compile, and the sample must lint
clean with `kiw.Lint`.
