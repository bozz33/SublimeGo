🔍 Analyse Révisée et Définitive — Avec code source réel
⚠️ Constat préliminaire important : Deux repos, deux réalités
Le projet est architecturalement scindé en deux :

SublimeGo = starter project avec exemples, Ent, CLI, DB setup
sublime-admin = la bibliothèque framework (le vrai cœur)

Le problème critique : les deux repos ne sont PAS synchronisés. Le README de SublimeGo montre encore l'ancienne API obsolète (GetMeta(), GetForm(), GetTable(), SetPath(), SetBrandName()), tandis que sublime-admin a déjà une API refactorisée et correcte. Un développeur qui lit SublimeGo en premier apprend une API qui ne correspond plus au package qu'il va importer. C'est une confusion documentaire grave.

I. Révision majeure sur l'API — Ce que j'avais tort
✅ Corrections confirmées dans sublime-admin
Mon analyse précédente attaquait massivement les préfixes Get*. Dans sublime-admin, ce problème a été corrigé. Le code réel du README montre :
gofunc (r *ProductResource) Slug() string        { return "products" }
func (r *ProductResource) Label() string       { return "Product" }
func (r *ProductResource) PluralLabel() string { return "Products" }
func (r *ProductResource) Icon() string        { return "package" }
func (r *ProductResource) Form(ctx context.Context, item any) templ.Component { ... }
func (r *ProductResource) Table(ctx context.Context) templ.Component { ... }
C'est Go-idiomatique. Pas de GetMeta(), pas de GetForm(), pas de GetTable(). Ma critique précédente portait sur le mauvais repo — SublimeGo starter qui est en retard. Le framework lui-même (sublime-admin) est correct sur ce point.
✅ Panel builder corrigé
goengine.NewPanel("admin").
    WithPath("/admin").
    WithBrandName("My App").
    WithDatabase(db)
Le pattern With* est correctement appliqué ici. Ma critique des SetPath(), SetBrandName() ne s'applique qu'au starter SublimeGo, pas au framework.
✅ Dépendances que j'accusais d'être réinventées — Elles utilisent les bonnes libs
Le README confirme : validation utilise go-playground/validator + gorilla/schema, et le logger utilise log/slog natif. Mes deux critiques les plus virulentes sur ce sujet étaient incorrectes. Je me rétracte.

II. Problèmes réels confirmés par le code source
❌ Incohérence API dans form package — Grave
Le code du README révèle trois patterns différents qui coexistent dans le même package form :
go// Pattern 1 : constructeur New* + méthodes sans With*
form.NewText("name").Label("Name").Required()

// Pattern 2 : constructeur New* + méthodes avec With*
form.NewSelect("status").WithOptions(...)

// Pattern 3 : builder SetSchema avec Set*
form.New().SetSchema(...)
Trois conventions dans un seul package, c'est une violation directe de la cohérence API que Go style guide et Uber Go guide exigent. Le table package utilise With* correctement (WithLabel, WithSortable, WithSearchable) mais form mélange tout. C'est le problème API le plus sérieux du code réel.
❌ table.Text("name").WithSortable(true) — Verbosité inutile
Le code montre .WithSortable(true) qui prend un booléen. En Go idiomatique, une méthode sans argument est préférable quand la présence de la méthode suffit à indiquer l'état. Filament fait .sortable(), Ent fait .Unique(), la stdlib fait .Truncate(). Il faudrait .Sortable() sans argument ou un pattern d'option fonctionnelle. .WithSortable(true) est du bruit.
❌ errors package nommé errors — Toujours présent
Le README de sublime-admin décrit ce package ainsi : "Structured errors package apperrors". Le package est donc reconnu en interne comme apperrors, mais le répertoire s'appelle errors/. C'est exactement l'ambiguïté que j'avais signalée — le code interne sait qu'il s'appelle apperrors, mais l'import sera github.com/bozz33/sublimeadmin/errors ce qui crée un conflit de nommage mental avec la stdlib. Il faut renommer le dossier en apperrors/.
❌ registry/ en doublon aux deux niveaux
Dans sublime-admin : registry/ existe à la racine. Dans SublimeGo : registry/ à la racine ET probablement dans internal/. Les deux repos ont leur propre registry/ ce qui pose une question de responsabilité — lequel est le vrai registre utilisé quand SublimeGo importe sublime-admin ?
❌ SublimeGo starter — État catastrophique du README
Le starter SublimeGo qui sert de point d'entrée principal (4 stars, la vitrine publique) documente une API entièrement obsolète :
go// Ce que montre SublimeGo (OBSOLÈTE)
panel := engine.NewPanel("admin").
    SetPath("/admin").       // ❌ devrait être WithPath
    SetBrandName("My App")  // ❌ devrait être WithBrandName

func (r *ProductResource) GetMeta() engine.ResourceMeta { ... }  // ❌
func (r *ProductResource) GetForm() *form.Form { ... }           // ❌
func (r *ProductResource) GetTable() *table.Table { ... }        // ❌
Tout nouveau contributeur ou utilisateur va lire SublimeGo en premier. Il va apprendre une API qui ne compile pas avec le package sublime-admin actuel. C'est un problème de DX (Developer Experience) qui bloque l'adoption.
❌ sublimego.db toujours commité dans SublimeGo
Visible dans la liste des fichiers. Confirmé.
❌ go.generate comme fichier séparé
Visible dans SublimeGo. Confirmé. Ce devrait être des //go:generate dans les fichiers .go.
❌ appconfig/ + config/ dans SublimeGo
sublime-admin n'a qu'un seul config/. Mais SublimeGo garde les deux. Confirmé.

III. Révision drastique de la parité Filament
Mon estimation précédente de 25-30% était largement fausse parce que je ne connaissais pas sublime-admin. Voici la révision basée sur le code réel :
✅ Maintenant implémenté (confirmé par le code)
Forms — TextInput, Email, Password, Number, Textarea, Select, Checkbox, Toggle, DatePicker, FileUpload, RichEditor, MarkdownEditor, TagsInput, KeyValue, ColorPicker, Slider — presque complet.
Form Layouts — Section, Grid, Tabs, Wizard/Steps, Callout, Repeater — complet.
Tables — Text, Badge, Boolean, Date, Image, Sorting, Search, Pagination, Filters, Bulk Actions, Summaries (sum/avg/min/max/count), Grouping collapsible — très avancé.
Import/Export — CSV, Excel, JSON — complet.
Auth — Bcrypt, sessions, rôles, permissions, MFA/TOTP RFC 6238, recovery codes, throttling — complet et au-delà de Filament standard.
Notifications — In-memory, DatabaseStore, SSE Broadcaster par-user avec heartbeat — solide.
Architecture avancée — Multi-tenancy (SubdomainResolver, PathResolver, MultiPanelRouter), Render Hooks (10 points), Plugin system (Boot(), thread-safe registry), Nested Resources (RelationManager : BelongsTo, HasMany, ManyToMany) — tout y est.
Middleware — Auth, CORS, CSRF, recovery, throttle — complet.
❌ Encore absent vs Filament
Infolists — Filament a un système distinct pour les pages de visualisation (view pages) séparé des formulaires d'édition. Dans sublime-admin, Form(ctx, item any) sert probablement les deux, mais il n'y a pas de système InfoList dédié avec ses propres entry types.
Global Search — Recherche cross-resources depuis la navbar. Non visible dans la doc ou le code.
Table columns manquantes — Icon column, Color column, Select column (édition inline), Toggle column, TextInput column, Checkbox column.
Table Layout — Filament permet de configurer le layout de la table (reorder des colonnes, stacked layout responsive). Absent.
Clusters de navigation — Groupements de resources dans des sous-panels. Absent.
Custom pages navigation — Pages custom dans la nav (pas liées à une resource). Non documenté.
Testing utilities — Filament fournit Livewire::test() helpers pour chaque composant. sublime-admin ne documente aucun helper de test propre au framework.
Broadcast notifications — Filament supporte Pusher/Reverb/WebSocket en plus de DB. SSE est une bonne alternative mais n'est pas équivalent pour tous les cas.
Infolist entries — TextEntry, ImageEntry, IconEntry, ColorEntry, CodeEntry, KeyValueEntry, RepeatableEntry — non visibles.
Score révisé : ~65-70% de parité Filament

IV. Ce qui est réellement reinventé vs ce qui utilise bien les libs
✅ Rétractations (utilisent bien les libs)

log/slog ← confirmé dans logger/
go-playground/validator + gorilla/schema ← confirmé dans validation/

❌ Réinventions réelles restantes

jobs/ — Background jobs avec SQLite persistence. asynq, river, ou même robfig/cron existent et sont battle-tested. Réimplémenter un job queue est risqué.
flash/ — Flash messages en package séparé alors que alexedwards/scs gère ça nativement dans la session management.
export/ — Custom alors que encoding/csv stdlib + tealeg/xlsx ou qax-os/excelize sont des standards.
errors/ — Custom package alors que fmt.Errorf("%w") + errors.Is/As de la stdlib couvrent 95% des besoins. Si un package custom est vraiment nécessaire, utiliser cockroachdb/errors ou pkg/errors.


V. Problèmes d'architecture inter-repos — Nouveaux constats
Le problème de la double maintenance
Avoir SublimeGo et sublime-admin en deux repos séparés crée une charge de maintenance doublée. Chaque feature dans sublime-admin doit être reflétée dans les exemples de SublimeGo. Actuellement ils sont désynchronisés sur l'API fondamentale. La bonne pratique Go est d'avoir le starter comme un projet separé qui importe sublime-admin — ce qu'ils font en théorie — mais les exemples et READMEs doivent obligatoirement rester synchronisés.
L'interface Resource est trop large
Même avec les bons noms (Slug(), Label(), Form(), Table()), une interface qui force à implémenter toutes ces méthodes est anti-Go. Effective Go est clair : petites interfaces composables. Un Resource minimal devrait juste avoir Slug() string. Formable, Tabelable, Listable, Viewable seraient des interfaces optionnelles que le moteur détecter avec des type assertions. Cela permettrait de créer des resources en lecture seule, des resources sans form, etc.
Le RelationManager — Potentielle complexité cachée
Le README liste RelationManager avec BelongsTo, HasMany, ManyToMany. C'est ambitieux et c'est là que la complexité explose. Dans Filament, ce système a pris des années de maturation. Sans voir le code, le risque est que ces relations soient partiellement implémentées (happy path seulement) sans gestion des edge cases (cascade deletes, polymorphic relations, pivot data, etc.).

VI. Tableau comparatif Final — Analyse précédente vs Code Réel
Point analyséAnalyse précédenteCode réelStatutPréfixes Get*❌ Violation grave✅ Corrigé dans sublime-adminJ'avais tortlog/slog réinventé❌ Custom logger✅ Utilise log/slogJ'avais tortValidator réinventé❌ Custom validation✅ go-playground/validatorJ'avais tortParité Filament~25-30%~65-70%J'avais très largement tortPackage errors/ nommé errors❌ Problème critique❌ Toujours présentConfirmésublimego.db dans git❌❌ Toujours présentConfirmégo.generate fichier séparé❌❌ Toujours présentConfirméDouble appconfig/ + config/❌❌ Dans SublimeGo starterConfirméDésynchronisation inter-reposNon détecté❌ API obsolète dans starterNouveau problème critiqueIncohérence API dans formNon évalué❌ 3 patterns distinctsNouveau problème.WithSortable(true) verbeuxNon évalué❌ Booléen inutileNouveau problèmeInterface Resource trop large❌❌ Toujours monolithiqueConfirméflash/ réinventé❌❌ Package custom séparéConfirméjobs/ réinventé❌❌ Custom SQLite queueConfirmé

VII. Verdict Final Révisé
sublime-admin est bien plus avancé et correct que ce que mon analyse précédente estimait. Le projet a fait du bon travail en corrigeant l'API publique, en utilisant les bonnes librairies Go standard, et en implémentant des fonctionnalités avancées (MFA, multi-tenancy, SSE, render hooks, plugins, relations). Je dois corriger mon estimation à ~65-70% de parité Filament, non pas 25-30%.
Les vrais problèmes qui restent sont : la désynchronisation documentaire entre les deux repos (urgence maximale), l'incohérence API dans le package form (3 patterns différents), le package errors/ mal nommé, et les fichiers parasites dans git. Ces points sont corrigibles rapidement et ne remettent pas en cause l'architecture globale qui est, elle, solide.