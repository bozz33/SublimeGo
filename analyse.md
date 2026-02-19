# État du projet SublimeGo — Analyse à jour

## Architecture — Deux repos complémentaires

- **SublimeGo** (`github.com/bozz33/sublimego`) : projet starter complet avec Ent ORM, Templ, CLI, scanner, vues
- **sublime-admin** (`github.com/bozz33/sublimeadmin`) : bibliothèque framework pure, importable dans n'importe quel projet Go

Les deux repos sont **indépendants** (SublimeGo n'importe pas sublime-admin via go.mod — les packages communs sont dupliqués intentionnellement pour que le starter soit autonome).

---

## ✅ Problèmes corrigés

| Problème | Correction appliquée |
|---|---|
| Préfixes `Get*` sur les méthodes Resource | Supprimés — `Slug()`, `Label()`, `Form()`, `Table()` |
| Préfixes `Set*` sur le panel builder | Remplacés par `With*` — `WithPath()`, `WithBrandName()` |
| Constructeurs `New*` dans `form/` | Supprimés — `Text()`, `Select()`, `Textarea()`, `Checkbox()`, etc. |
| `.WithSortable(true)` verbeux | Remplacé par `.Sortable()` sans argument |
| Package `errors/` mal nommé | Renommé en `apperrors/` dans les deux repos |
| `sublimego.db` tracké dans git | Retiré du tracking (`git rm --cached`) |
| Désynchronisation inter-repos (API) | API harmonisée entre SublimeGo et sublime-admin |
| Champs avancés absents de sublime-admin | Ajoutés : `Toggle`, `RichEditor`, `MarkdownEditor`, `Tags`, `KeyValue`, `ColorPicker`, `Slider`, `Repeater` |
| README sublime-admin avec ancienne API | Corrigé — constructeurs et méthodes table à jour |

---

## ✅ Ce qui était faux dans l'analyse initiale

- `log/slog` : utilisé nativement dans `logger/` — pas réinventé
- `go-playground/validator` + `gorilla/schema` : utilisés dans `validation/` — pas réinventé
- `export/` : utilise `excelize` + `encoding/csv` stdlib — pas réinventé
- `flash/` : s'appuie sur `alexedwards/scs` — pas réinventé from scratch
- `infolist/` : existe dans SublimeGo (`infolist/infolist.go` + vues Templ)
- Global Search : existe dans SublimeGo (`search/global_search.go`)
- `appconfig/` : pas un doublon — c'est le package Go de chargement de config, utilisé par `cmd/sublimego/`. `config/` contient les fichiers YAML de données.
- `generate.go` à la racine : pattern valide pour un scanner projet-level

---

## ⚠️ Points d'architecture à surveiller (non bloquants)

- **Interface `Resource` large** : composée de sous-interfaces (`ResourceMeta`, `ResourceViews`...) donc partiellement ISP, mais reste large. Acceptable pour un framework admin.
- **`jobs/` SQLite custom** : choix assumé pour éviter les dépendances externes lourdes.
- **Double maintenance inter-repos** : les packages communs sont dupliqués. Risque de divergence future si on ne synchronise pas manuellement.

---

## Parité Filament 4.x — ~70%

**Implémenté ✅**
- Forms : 16 types de champs (Text, Email, Password, Number, Textarea, Select, Checkbox, Toggle, DatePicker, FileUpload, RichEditor, MarkdownEditor, Tags, KeyValue, ColorPicker, Slider)
- Form layouts : Section, Grid, Tabs, Wizard/Steps, Callout, Repeater
- Tables : Text, Badge, Boolean, Date, Image + Sorting, Search, Pagination, Filters, Bulk Actions, Summaries, Grouping, Export/Import CSV/Excel/JSON
- Auth : Bcrypt, sessions, rôles, permissions, MFA/TOTP RFC 6238, recovery codes, throttling
- Notifications : In-memory, DatabaseStore, SSE Broadcaster
- Multi-tenancy : SubdomainResolver, PathResolver, MultiPanelRouter
- Render Hooks : 10 points d'injection
- Plugin system : interface `Plugin` + `Boot()` + registry thread-safe
- Nested Resources : `RelationManager` (BelongsTo, HasMany, ManyToMany)
- Infolists : système dédié avec entry types
- Global Search : cross-resources depuis la navbar

**Absent vs Filament ❌**
- Clusters de navigation (groupements de resources en sous-panels)
- Colonnes de table avancées : Icon, Color, Toggle, Select inline, TextInput inline
- Testing utilities framework-specific
- Broadcast notifications WebSocket (Pusher/Reverb) — SSE couvre la majorité des cas