# 🔮 Grimoire

> *Your personal book of spells for command-line incantations*

**Grimoire** is a simple shell command snippet storage and retrieval tool. Save your most powerful ~~CLI snippets~~ spells, give them memorable names, document their ~~use cases~~ arcane functionality, and invoke them with custom ~~parameters~~ sigils whenever you need.

**📦 Depends on [fzf](https://github.com/junegunn/fzf) being in your PATH**

## ✨ Features

- 📚 **Store command snippets** with descriptive names and documentation
- 🎯 **Parameterize snippets** for flexible reuse
- 🔍 **Quick search and retrieval** of your saved spells
- ⚡ **Execute commands directly** from your grimoire
- 🪄 **Simple, magic-themed interface** that makes CLI work feel like wizardry

## 🎓 Inspiration

Grimoire was inspired by [pet](https://github.com/knqyf263/pet) but reimagined with:
- A magical, wizard-themed aesthetic
- A simpler, more focused codebase
- Easier customization and understanding of internals for my own personal use

## 🚀 Quick Start

```sh
# Add a new spell to your grimoire
grimoire add

# Cast a spell from your grimoire
grimoire cast

# Edit an existing spell by opening it in your $EDITOR (fallback editor is vi if $EDITOR is empty or undefined)
grimoire edit

# Echo just the spell details to your terminal
grimoire echo

# View all spell details including the name and description
grimoire view <spell-name>
```

## 📖 Example Spells

```txt
Spell: openssl x509 -inform DER -outform PEM -in <path>
Name: commit-all
Description: Convert a DER-encoded x.509 certificate to PEM
```

```txt
Spell: echo "${<var>//\//-}"
Name: slash-to-dash
Description: Convert all forward slashes in a variable to dashes.
```

## 🛠️ Installation

```sh
# Clone the grimoire
git clone https://github.com/yourusername/grimoire.git
cd grimoire

# Build from source
go build .

# Install the binary
go install .
```

---

*‘It is our choices, Harry, that show what we truly are, far more than our abilities.’ - Albus Dumbledore 🧙‍♂️✨*
