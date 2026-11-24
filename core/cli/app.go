package cli

import (
	_ "github.com/krewire/hub/mysql"
	_ "github.com/krewire/hub/postgres"
	_ "github.com/krewire/hub/sqlite"
	"github.com/krewire/krew/core/commands"
	kiwver "github.com/krewire/krew/core/version"
	"github.com/krewire/tui/core"
)

// BuildApp builds the unified TUI CLI application with all developer toolchain commands.
func BuildApp(name string) *tui.App {
	app := tui.NewApp(name, kiwver.VersionString()).
		WithDescription("Autonomous AI Agent Engine & Developer Toolchain").
		// Agent commands
		Command(tui.NewCommand("chat", "start an interactive AI coding session in the terminal", registerChatFlags, runChatCmd).WithGroup("agent").WithCore().WithExample("krew chat")).
		Command(tui.NewCommand("agent", "execute an autonomous task or prompt headless to completion", registerAgentFlags, runAgentCmd).WithGroup("agent").WithCore().WithExample("krew agent \"Add unit tests for auth\"")).
		Command(tui.NewCommand("daemon", "launch HTTP/REST and Server-Sent Events (SSE) server for remote hosting", registerDaemonFlags, runDaemonCmd).WithGroup("agent").WithExample("krew daemon --port 8080")).
		Command(tui.NewCommand("sessions", "list past conversation sessions with message and token counts", registerSessionsFlags, runSessionsCmd).WithGroup("agent").WithExample("krew sessions")).
		Command(tui.NewCommand("export", "export conversation session transcript to markdown", registerExportFlags, runExportCmd).WithGroup("agent").WithExample("krew export <session-id>")).
		// Inspect commands
		Command(tui.NewCommand("version", "print CLI and framework versions", nil, commands.RunVersion).WithGroup("inspect").WithExample("krew version")).
		Command(tui.NewCommand("info", "print environment and project information", nil, commands.RunInfo).WithGroup("inspect").WithExample("krew info")).
		Command(tui.NewCommand("compat", "check version compatibility across all modules", nil, commands.RunCompat).WithGroup("inspect").WithExample("krew compat")).
		Command(tui.NewCommand("context", "inspect, validate, and lint authoritative project context for humans and AI agents", commands.RegisterContext, commands.RunContext).WithGroup("inspect").WithExample("krew context --lint")).
		Command(tui.NewCommand("bench", "run performance benchmarks and evaluate baselines", commands.RegisterBench, commands.RunBench).WithGroup("inspect").WithExample("krew bench --json")).
		Command(tui.NewCommand("tier", "inspect open-core tier posture and validate boundary rules", commands.RegisterTier, commands.RunTier).WithGroup("inspect").WithExample("krew tier --check free")).
		Command(tui.NewCommand("config", "validate krewire.yaml against its JSON Schema (line:col, did-you-mean), or print the schema", commands.RegisterConfig, commands.RunConfig).WithGroup("inspect").WithExample("krew config validate --strict")).
		Command(tui.NewCommand("doctor", "diagnose developer environment, toolchain prerequisites, and workspace", commands.RegisterDoctor, commands.RunDoctor).WithGroup("inspect").WithExample("krew doctor")).
		Command(tui.NewCommand("self-update", "update the krew toolchain to the latest stable release", commands.RegisterSelfUpdate, commands.RunSelfUpdate).WithGroup("inspect").WithExample("krew self-update --check")).
		// Project commands
		Command(tui.NewCommand("new", "create a project in a new directory (app by default; --site, --book, --cli, …); asks for name and kind when run bare", commands.RegisterNew, commands.RunNew).WithGroup("project").WithCore().WithExample("krew new my-site --site")).
		Command(tui.NewCommand("init", "set up Krewire in the current directory, optionally equipping a kind (--site, --book, …); asks for the kind when run bare", commands.RegisterInit, commands.RunInit).WithGroup("project").WithExample("krew init --site")).
		Command(tui.NewCommand("dev", "live development server: watches files, rebuilds, and hot-reloads sites/books or restarts apps and services", commands.RegisterDev, commands.RunDev).WithGroup("develop").WithCore().WithExample("krew dev")).
		Command(tui.NewCommand("build", "build the project for production (site, book, app, cli, worker, service, gateway, infra)", commands.RegisterBuild, commands.RunBuild).WithGroup("build").WithCore().WithExample("krew build")).
		Command(tui.NewCommand("bundle", "bundle static CSS and JS into assets/krewire.css and assets/krewire.js", commands.RegisterBundle, commands.RunBundle).WithGroup("build").WithExample("krew bundle --minify")).
		Command(tui.NewCommand("serve", "preview production output without watching: serves the built site/book from .krewire/build (run 'krew build' first); apps and CLIs run as with 'krew run'", commands.RegisterServe, commands.RunServe).WithGroup("build").WithExample("krew serve --port 8080")).
		Command(tui.NewCommand("run", "execute a runnable target or run an arbitrary Go file with framework runtime support", commands.RegisterRun, commands.RunRun).WithGroup("develop").WithCore().WithExample("krew run [task|path/to/file.go] [-- args]")).
		Command(tui.NewCommand("test", "run the tests of the current project", commands.RegisterTest, commands.RunTest).WithGroup("develop").WithCore().WithExample("krew test")).
		Command(tui.NewCommand("verify", "run reproducible quality gates and produce evidence-backed task verification record", commands.RegisterVerify, commands.RunVerify).WithGroup("develop").WithCore().WithExample("krew verify --json")).
		Command(tui.NewCommand("vet", "run go vet, security posture audit, and vulnerability checks", commands.RegisterVet, commands.RunVet).WithGroup("develop").WithExample("krew vet --sec --vuln")).
		Command(tui.NewCommand("fmt", "check formatting with gofmt/go fmt", commands.RegisterFmt, commands.RunFmt).WithGroup("develop").WithExample("krew fmt --write")).
		Command(tui.NewCommand("add", "add a package or plugin (package@version), or copy kui components into ./components (ui/<name>, ui:<pack>)", commands.RegisterAdd, commands.RunAdd).WithGroup("project").WithCore().WithExample("krew add ui/button ui:minimal")).
		Command(tui.NewCommand("remove", "remove a package or plugin", commands.RegisterRemove, commands.RunRemove).WithGroup("project").WithExample("krew remove twcss")).
		Command(tui.NewCommand("update", "update project dependencies to their latest versions", commands.RegisterUpdate, commands.RunUpdate).WithGroup("project").WithExample("krew update")).
		Command(tui.NewCommand("deploy", "stage deployable artifacts in .krewire/dist/", commands.RegisterDeploy, commands.RunDeploy).WithGroup("ship").WithExample("krew deploy --dry-run")).
		Command(tui.NewCommand("worker", "run background workers / job queues", commands.RegisterWorker, commands.RunWorker).WithGroup("develop").WithExample("krew worker")).
		Command(tui.NewCommand("dash", "local dev dashboard, services, logs & virtual workspace sandboxes", commands.RegisterDash, commands.RunDash).WithGroup("develop").WithCore().WithExample("krew dash")).
		Command(tui.NewCommand("dashboard", "alias for krew dash (local dev dashboard, logs, traces)", commands.RegisterDashboard, commands.RunDashboard).WithGroup("develop").WithHidden(true).WithExample("krew dashboard")).
		Command(tui.NewCommand("gateway", "run API gateway reverse proxy for microservices", commands.RegisterGateway, commands.RunGateway).WithGroup("develop").WithExample("krew gateway --port 8080")).
		Command(tui.NewCommand("mcp", "native Model Context Protocol server for AI coding agents", commands.RegisterMCP, commands.RunMCP).WithGroup("develop").WithExample("krew mcp --stdio")).
		Command(tui.NewCommand("ws", "workspace for monorepo, multi-repo, multi-project, microservices", commands.RegisterWs, commands.RunWs).WithGroup("ship").WithExample("krew ws help")).
		Command(tui.NewCommand("boost", "manage Boost AI agent template (install, status)", commands.RegisterBoost, commands.RunBoost).WithGroup("ship").WithExample("krew boost status")).
		Command(tui.NewCommand("release", "bump versions and stage a release across modules", commands.RegisterRelease, commands.RunRelease).WithGroup("ship").WithExample("krew release framework --bump minor --apply")).
		Command(tui.NewCommand("make", "generate a new component, page, handler, migration, model, seeder, factory, repository, service, job, or chapter", commands.RegisterMake, commands.RunMake).WithGroup("project").WithCore().WithExample("krew make component Card")).
		Command(tui.NewCommand("migrate", "apply, roll back, or list migrations: up [n], down [n], redo, status, create <name>", commands.RegisterMigrate, commands.RunMigrate).WithGroup("project").WithExample("krew migrate up")).
		Command(tui.NewCommand("seed", "seed the database with records via Go seeders and factories", commands.RegisterSeed, commands.RunSeed).WithGroup("project").WithExample("krew seed --class DatabaseSeeder")).
		Command(tui.NewCommand("db:seed", "alias for krew seed", commands.RegisterSeed, commands.RunSeed).WithGroup("project").WithHidden(true).WithExample("krew db:seed")).
		Command(tui.NewCommand("ui", "browse, install, and diff kui components: list, info <name>, add <name|pack:x>, diff <name>", commands.RegisterUI, commands.RunUI).WithGroup("project").WithExample("krew ui list")).
		Command(tui.NewCommand("pkg", "manage standalone .kiw component packages: init <name>, pack [dir], add <src>, list, export", commands.RegisterPkg, commands.RunPkg).WithGroup("project").WithCore().WithExample("krew pkg pack ./my-ui")).
		Command(tui.NewCommand("auth", "manage developer authentication: login, logout, status", commands.RegisterAuth, commands.RunAuth).WithGroup("project").WithCore().WithExample("krew auth login --token <token>"))

	for name, hint := range commands.Hints {
		app.WithHint(name, hint)
	}

	return app
}
