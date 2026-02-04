# Templating Guide

This guide explains how to work with Templ templates in SublimeGo.

## What is Templ?

Templ is a type-safe templating language for Go that compiles to Go code.

## Quick Start

### Install Templ CLI

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

### Create a Template

Create `views/hello.templ`:

```templ
package views

templ Hello(name string) {
    <div class="p-4">
        <h1>Hello, { name }!</h1>
    </div>
}
```

### Generate Go Code

```bash
templ generate
```

### Use in Handler

```go
views.Hello("World").Render(r.Context(), w)
```

## Syntax Reference

See the full Templ documentation at https://templ.guide
