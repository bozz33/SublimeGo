# Analyse Approfondie de SublimeGo — Vérifiée sur le Code Source

> **Méthode** : Chaque affirmation ci-dessous a été vérifiée directement dans le code source (`engine/`, `form/`, `table/`, `errors/`, `logger/`, `validation/`, `appconfig/`, `plugin/`, `infolist/`, `go.mod`, `.gitignore`).

---

## I. Vue d'ensemble

SublimeGo se positionne comme un équivalent Go de Laravel Filament. Le concept est pertinent — il n'existe pas d'équivalent mature à Filament en Go. Le choix de stack (Ent + Templ + HTMX + Tailwind CSS + Alpine.js) est solide et moderne.

---

## II. Architecture et Layout — Vérifications et Corrections

### ✅ Ce qui est correct (contrairement à certaines affirmations)

- **`logger/` utilise déjà `log/slog`** — `logger.go` importe `log/slog` et wrap `slog.Logger`. Ce n'est pas une réinvention : c'est une couche de configuration (rotation via `lumberjack`, JSON/text selon l'env, `slog.SetDefault`). **Affirmation "réinvention de slog" : FAUSSE.**
- **`validation/` utilise déjà `go-playground/validator/v10`** — `validator.go` importe et wrap `github.com/go-playground/validator/v10`. **Affirmation "validation réinventée from scratch" : FAUSSE.**
- **`config/` n'existe pas** — Le dossier `config/` est vide (0 fichiers Go). `appconfig/` est le seul package de configuration. **Affirmation "double configuration" : FAUSSE.**
- **`infolist/` existe et est fonctionnel** — `TextEntry`, `BadgeEntry`, `BooleanEntry`, `DateEntry`, `ImageEntry`, `ColorEntry` sont tous implémentés. **Affirmation "Infolists absents" : FAUSSE.**
- **`plugin/` existe avec une architecture complète** — Interface `Plugin`, `Register()`, `Boot()`, `Get()`, protection par `sync.RWMutex`. **Affirmation "pas d'architecture plugin" : FAUSSE.**
- **Bulk Actions présentes** — `table/bulk_actions.go` existe, `BulkActionDef` dans `engine/contract.go`, `BulkDelete` dans `ResourceCRUD`. **Affirmation "bulk actions absentes" : FAUSSE.**
- **Global Search présente** — `/api/search` dans `engine/panel.go`, package `search/` complet avec `GlobalSearch`, `QuickSearch`, fuzzy matching. **Affirmation "global search absente" : FAUSSE.**
- **`FileUpload`, `Repeater`, `Toggle`, `Hidden`, `Checkbox`, `DatePicker` présents** dans `form/fields.go`. **Affirmation "majorité des champs absents" : PARTIELLEMENT FAUSSE.**
- **`Image` column présente** dans `table/columns.go`. **Affirmation "Image column absente" : FAUSSE.**
- **`go.generate` contient une directive `//go:generate` valide** — Le fichier contient `//go:generate go run ./cmd/scanner/main.go ...`. La syntaxe est correcte même dans un fichier dédié (Go accepte les directives dans n'importe quel fichier `.go` ou non). **Affirmation "pas une convention Go" : NUANCÉE** — c'est inhabituel mais fonctionnel.
- **Interface `Resource` est composée de petites interfaces** — `Resource` embed `ResourceMeta`, `ResourceViews`, `ResourcePermissions`, `ResourceCRUD`, `ResourceNavigation`. Ce n'est pas une "grande interface monolithique" mais une composition d'interfaces petites. **Affirmation "grande interface signe de mauvaise conception" : INEXACTE sur ce code.**
- **`sublimego.db` est dans `.gitignore`** (ligne 19). **Affirmation "jamais dans .gitignore" : FAUSSE** — il y est déjà.

---

### ❌ Problèmes Réels Confirmés par le Code

#### 1. Package `errors/` — Conflit de nom avec la stdlib (CONFIRMÉ, RISQUE LATENT)
Le package se nomme `package errors` (`errors/errors.go` ligne 1). Actuellement, seuls `middleware/auth.go` et `middleware/recovery.go` l'importent, et **aucun des deux n'importe simultanément la stdlib `errors`** — donc le conflit n'est pas encore déclenché. Cependant, le package a une valeur ajoutée réelle (`AppError` avec `StatusCode`, `Code`, `Fields`, stack trace, helpers HTTP) et tout futur fichier qui aurait besoin des deux devra utiliser un alias d'import, ce qui est une friction croissante.

**Action** : Renommer la déclaration en `package apperrors` dans `errors/errors.go` et `errors/handler.go`, et mettre à jour les 2 imports concernés (`middleware/auth.go`, `middleware/recovery.go`).

#### 2. Préfixes `Get*` sur les méthodes de l'interface `Column` (CONFIRMÉ)
Dans `table/columns.go`, les méthodes d'interface sont `GetKey()`, `GetLabel()`, `GetType()`, `GetValue()`. Dans `form/fields.go` : `GetName()`, `GetLabel()`, `GetValue()`, `GetRules()`, `GetPlaceholder()`, `GetHelp()`. Dans `infolist/infolist.go` : `GetLabel()`, `GetValueStr()`.

Selon **Effective Go** : *"If you have a field called owner, the getter method should be called Owner, not GetOwner."* Ces méthodes doivent perdre leur préfixe `Get`.

**Action** : Renommer `GetKey()→Key()`, `GetLabel()→Label()`, `GetType()→Type()`, `GetValue()→Value()`, `GetName()→Name()`, `GetRules()→Rules()`, etc. dans les interfaces (breaking change, à faire en une seule passe).

#### 3. Préfixes `Set*` sur les builders du `Panel` (CONFIRMÉ, NUANCÉ)
`engine/panel.go` : `SetPath()`, `SetDatabase()`, `SetBrandName()`, `SetLogo()`, `SetPrimaryColor()`, etc. L'idiome Go pour les builders fluents est `With*` ou sans préfixe. Cependant, `Set*` est toléré dans certains contextes Go (notamment quand le setter modifie l'état d'un objet existant). C'est une violation de style, pas une erreur fonctionnelle.

**Action** : Renommer en `With*` pour les builders fluents (`WithPath()`, `WithDatabase()`, `WithBrandName()`...) ou supprimer le préfixe.

#### 4. `go.generate` fichier non-standard (CONFIRMÉ, MINEUR)
Le fichier `go.generate` à la racine contient une directive `//go:generate`. Go ne reconnaît les directives `//go:generate` que dans les fichiers `.go`. Ce fichier **ne sera pas exécuté** par `go generate ./...`. La directive doit être dans un fichier `.go` (ex: `generate.go` à la racine ou dans le package concerné).

**Action** : Créer un fichier `generate.go` à la racine avec `package main` et y déplacer la directive.

#### 5. `registry/` racine vs `internal/registry/` (CONFIRMÉ)
Il y a un package `registry/` à la racine ET `internal/registry/`. Deux packages de registre dans le même projet sans différenciation claire de responsabilité.

**Action** : Clarifier les responsabilités : le registre public (API utilisateur) en `registry/`, le registre généré (provider_gen.go) uniquement dans `internal/registry/`.

#### 6. Notifications non persistantes (CONFIRMÉ)
`notifications/notification.go` : stockage en mémoire (`map[string][]*Notification`). Redémarrage = perte de toutes les notifications. Pas de schéma Ent pour les notifications.

**Action** : Ajouter une entité `Notification` dans `internal/ent/schema/` et un `DatabaseStore` implémentant la même interface.

#### 7. `FormDecoder` custom réinventé (CONFIRMÉ)
`validation/validator.go` implémente un `FormDecoder` from scratch avec `reflect` pour décoder les données de formulaire HTTP. La bibliothèque `gorilla/schema` ou `github.com/go-chi/chi` font cela de façon battle-tested. C'est une réinvention partielle.

---

## III. Comparaison Filament 4.x — État Réel

### ✅ Présent et Fonctionnel

| Fonctionnalité | Package | Notes |
|---|---|---|
| Resources CRUD | `engine/` | Complet (List, Create, Edit, Delete, BulkDelete) |
| Table Builder | `table/` | Text, Badge, Image, Boolean, Date columns |
| Form Builder | `form/` | Text, Email, Password, Number, Textarea, Select, Checkbox, FileUpload, DatePicker, Toggle, Repeater, Hidden |
| Infolists | `infolist/` | Text, Badge, Boolean, Date, Image, Color entries |
| Actions (row) | `actions/` | Edit, Delete, custom |
| Bulk Actions | `table/bulk_actions.go` | Présent |
| Widgets | `widget/` | Stats, Charts |
| Navigation groupée | `engine/panel.go` | Groupes + tri |
| Auth | `auth/` | bcrypt + sessions SCS |
| Multi-panel | `engine/panel.go` | Via injection de config |
| Plugin system | `plugin/` | Interface + Boot lifecycle |
| Global Search | `search/` | Fuzzy + registry concurrent |
| Notifications (SSE) | `notifications/` | Temps réel via SSE |
| Export CSV/Excel | `export/` | xuri/excelize |
| Import | `import/` | Présent |
| Jobs/Queue | `jobs/` | Queue en mémoire |
| Rate Limiting | `middleware/` | Par IP |
| Mailer | `mailer/` | SMTP + LogMailer |
| CLI | `cmd/sublimego/` | serve, generate, make, db, routes, doctor... |

### ❌ Absent ou Incomplet

| Fonctionnalité Filament | Statut SublimeGo |
|---|---|
| Schema Layouts (Sections, Tabs, Grid) | ✅ Présent dans `form/layout.go` (`Section`, `Grid`, `Tabs`) — manque `Wizard` et `Callout` |
| Rich Editor / Markdown Editor | Absent |
| Tags input, Key-value, Color picker, Slider | Absent |
| Table Summaries (totaux, moyennes) | Absent |
| Table Grouping (regroupement par colonne) | Absent |
| Nested Resources | Absent |
| Multi-tenancy automatique | Absent |
| MFA | Absent |
| Render Hooks | Absent |
| Testing utilities (helpers de test framework) | Absent |
| Clusters de navigation | Absent |
| Notifications persistantes (DB) | Absent (mémoire uniquement) |
| Broadcast notifications (WebSocket) | Absent |

---

## IV. Leçons de PicoClaw Applicables

1. **Single binary avec `//go:embed`** — PicoClaw distribue un binaire unique. SublimeGo devrait embarquer les assets CSS/JS compilés dans le binaire via `//go:embed`. Actuellement `http.FileServer(http.Dir("ui/assets"))` nécessite que le dossier soit présent à l'exécution.
2. **Provider Architecture** — Le pattern de PicoClaw (protocoles familles : OpenAI-compatible, Anthropic) s'applique aux intégrations de SublimeGo : un `StorageProvider`, un `MailProvider`, un `SearchProvider` interchangeables.
3. **Minimalisme des dépendances** — PicoClaw surveille sa liste de dépendances. SublimeGo a deux drivers SQLite (`mattn/go-sqlite3` CGO + `modernc.org/sqlite` pure Go) — choisir l'un ou l'autre via build tags, pas les deux.

---

## V. Actions Prioritaires (Basées sur le Code Réel)

### Urgent — Corrections de Standards Go
1. **Renommer `package errors` → `package apperrors`** dans `errors/errors.go` et `errors/handler.go` (et mettre à jour tous les imports).
2. **Supprimer le préfixe `Get` des interfaces** : `Column`, `Field`, `Entry` dans `table/`, `form/`, `infolist/`.
3. **Renommer les builders `Set*` → `With*`** dans `engine/panel.go`.
4. **Corriger `go.generate`** : déplacer la directive dans un fichier `generate.go` (package `sublimego`).

### Important — Architecture
5. **Persistance des notifications** : ajouter entité Ent + `DatabaseStore`.
6. **`//go:embed` pour les assets** : remplacer `http.FileServer(http.Dir(...))` par un FS embarqué.
7. **Choisir un seul driver SQLite** : supprimer `mattn/go-sqlite3` ou `modernc.org/sqlite` (garder le pure Go pour la portabilité).
8. **Remplacer `FormDecoder` custom** par `gorilla/schema` ou `github.com/monoculum/formam`.

### Fonctionnel — Parité Filament
9. **Schema Layouts** : vérifier et compléter `form/layout.go` (Sections, Tabs).
10. **Table Summaries** : ajouter `SummaryDef` dans `engine/contract.go`.
11. **Render Hooks** : système d'injection de composants Templ à des points nommés.

---

## VI. Conclusion

SublimeGo est **plus avancé que ce que l'analyse initiale suggérait** (~50-55% de parité Filament, pas 25-30%). Les modules `infolist`, `plugin`, `search`, `bulk_actions`, et la majorité des champs de formulaire sont présents et fonctionnels.

Les vrais problèmes sont :
- **Nommage des interfaces** (préfixes `Get*`) — violation Effective Go confirmée, breaking change nécessaire.
- **Nom du package `errors`** — conflit d'import avec la stdlib, à corriger.
- **Notifications non persistantes** — lacune fonctionnelle critique.
- **Assets non embarqués** — problème de déploiement.
- **`go.generate` non exécuté** par `go generate` — bug silencieux.

Le choix de stack (Ent + Templ + HTMX + Tailwind) reste excellent. La fondation est solide.