# Project Status

## Current State

The `djot-fmt` project is now set up as a standalone CLI tool for formatting djot files, following Go best practices.

### ✅ Completed

**Core Functionality**
- CLI argument parsing (`-w`, `-c`, `-o` flags)
- File I/O handling (stdin/stdout, file input/output)
- AST-based formatting using godjot parser
- List formatting (simple and nested lists)
- Inline formatting (emphasis, strong, links)
- Paragraph spacing
- Heading formatting

**Project Structure**
- Go module setup with local godjot dependency
- Internal packages following best practices
- Comprehensive test coverage
- Development tooling configuration (hk, mise, golangci-lint)

**Documentation**
- README with usage examples
- CONTRIBUTING guide
- AGENTS.md for AI assistant guidance
- Project structure documentation

**Testing**
- Unit tests for formatter logic
- Unit tests for CLI argument parsing
- Test fixtures in testdata/

### 🚧 Currently Supported Node Types

- DocumentNode
- ParagraphNode
- UnorderedListNode
- ListItemNode
- TextNode
- EmphasisNode (`_text_`)
- StrongNode (`*text*`)
- LinkNode
- HeadingNode

### 📋 Known Limitations

**Unsupported Node Types** (will pass through but may not format correctly):
- OrderedListNode (numbered lists)
- TaskListNode (checkboxes)
- DefinitionListNode
- CodeNode (code blocks)
- QuoteNode (blockquotes)
- TableNode variants
- ThematicBreakNode (horizontal rules)
- SectionNode
- LineBreakNode
- Inline nodes: subscript, superscript, math, etc.

**Formatting Options**
- No configuration options (uses hardcoded 2-space indentation)
- No line wrapping
- No code block formatting

## Build and Test

```bash
# Build
GOWORK=off go build -o djot-fmt

# Run tests
GOWORK=off go test ./...

# Install locally
GOWORK=off go install
```

## Next Steps for Distribution

### Before Publishing

1. **Complete Testing**
   - [ ] Add integration tests with real-world djot files
   - [ ] Test with malformed input
   - [ ] Add benchmark tests

2. **Documentation**
   - [ ] Add more examples to README
   - [ ] Document supported vs. unsupported features
   - [ ] Create changelog

3. **CI/CD**
   - [ ] Set up GitHub Actions workflow
   - [ ] Add goreleaser configuration
   - [ ] Set up automated releases

4. **Repository Setup**
   - [ ] Initialize git repository
   - [ ] Create GitHub repo
   - [ ] Add issue templates
   - [ ] Set up branch protection

5. **Release Preparation**
   - [ ] Tag v0.1.0
   - [ ] Publish to GitHub releases
   - [ ] Test installation from published version

### Optional Enhancements

- Support for additional node types
- Configuration file support
- Plugin system for custom formatters
- Integration with editors (VSCode extension, etc.)
- Diff mode to show what would change
- Batch processing mode

## Dependencies

- github.com/sivukhin/godjot/v2 - Djot parser and AST
- github.com/stretchr/testify - Testing assertions

## Development Tools

- mise - Tool version management
- hk - Git hooks and linting
- golangci-lint - Go linter
- commitizen - Commit message formatting
