# How To

THE GOAL: just get your foot in the door with the simplest highlighting! Make
that first end-to-end connection! Figure out ANY build path that'll get
SOMETHING highlighted.

I'm trying to triangulate this through Sublime, VS Code, and targeting something
compatible for JetBrains to import.

The ocaml.tmb DOES import and does syntax highlight the helloworld.ml file.

I git cloned the ocaml.tmbundle project, but RubyMine wouldn't load it from that
exact location, cuz of the name problem, even if I tried renaming it in-situ. So
I copied the whole structure into this superdb-syntaxes project, and
successfully imported that into RubyMine's TextMate Bundles plugin.

zson-sublime.yaml: this is a COPY of a JSON file found
[here](https://github.com/sublimehq/Packages/blob/759d6eed9b4beed87e602a23303a121c3a6c2fb3/JSON/JSON.sublime-syntax),
and I'm trying to make the smallest of modifications to get it to highlight my
sample.zson file, which for the moment is pure JSON (and it should work, since
ZSON is a superset of JSON, right? Keys can be strings, right? this works, but
it still outputs zson with non-string keys - `zq -i zson 'yield this'
sample.json`)
    
zson-sublime.plist: a FAILED attempt IN Sublime to build zson-sublime.yaml
to plist, because of some error message about keys needing to be strings. BUT,
this SHOULD be a slightly edited copy of the JSON file inside the sublime
source repo anyway. Hmm :thinking-face:. Weird.

zson-scratch.yaml is using Sublime Text to generate a brand New Syntax... from
scratch from their menus, and seeing if we can perhaps get the SIMPLEST
highlighting working IN RubyMine.

I started wih VS Code, and the extension to generate the plist ... but it _seems
like_ Sublime is the more authoritative resource here. I came here through
Perplexity suggesting a Sublime YAML for the zq language, and then noticing
`bat` (while researching Markdown readers for the skdoc tool) has superior
support for syntax highlighting for code blocks inside Markdown. How? It points
to all of the built-in Sublime syntax packages, and then a host of other repos
throughout GitHub-land.
