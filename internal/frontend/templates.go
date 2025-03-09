package frontend

import (
	"coda/internal/config"
	"coda/internal/logger"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

// Embedded file systems for templates and static assets
//
//go:embed assets/templates
var templatesFS embed.FS

//go:embed assets/static
var staticFS embed.FS

// Default template configuration
const (
	defaultBaseTemplate  = "assets/templates/base.gohtml"
	defaultTemplateExt   = ".gohtml"
	defaultTemplatesPath = "internal/frontend/assets/templates"
)

// TemplateManager handles the loading, rendering, and hot-reloading of templates.
// It provides methods for rendering full pages and components, and manages
// template caching and synchronization.
type TemplateManager struct {
	templates     map[string]*template.Template // Cached page templates
	components    *template.Template            // Component templates
	mu            sync.RWMutex                  // Mutex for thread-safe template access
	funcMap       template.FuncMap              // Template functions
	baseTemplate  string                        // Path to base template
	templateExt   string                        // Template file extension
	templatesPath string                        // Path to templates directory
	cancelWatcher context.CancelFunc            // Function to cancel file watcher
}

// renderMarkdown converts markdown text to HTML with appropriate extensions and settings.
// This is used as a template function to render markdown content within templates.
// The HTML output is sanitized using bluemonday to prevent script injection.
func renderMarkdown(md string) template.HTML {
	// Create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)

	// Parse the markdown text
	doc := p.Parse([]byte(md))

	// Create HTML renderer with options
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{
		Flags: htmlFlags,
	}
	renderer := html.NewRenderer(opts)

	// Convert to HTML
	unsafeHTML := markdown.Render(doc, renderer)

	// Create a bluemonday policy for sanitizing HTML
	// UGCPolicy is designed for user-generated content and allows a reasonable set of HTML elements and attributes
	// while blocking potentially dangerous ones like <script> tags and javascript: URLs
	policy := bluemonday.UGCPolicy()

	// Add additional allowed elements and attributes for code blocks and syntax highlighting
	policy.AllowAttrs("class").OnElements("code", "pre")
	policy.AllowAttrs("data-language").OnElements("pre")

	// Sanitize the HTML
	safeHTML := policy.SanitizeBytes(unsafeHTML)

	// Return as template.HTML to avoid escaping
	return template.HTML(safeHTML) //nolint:gosec
}

// newTemplateManager creates a new TemplateManager with the given configuration.
// It initializes the template cache, sets up template functions, and loads templates.
func newTemplateManager(cfg *config.Config) (*TemplateManager, error) {
	// Create template functions
	funcMap := template.FuncMap{
		"appEnv": func() string {
			return string(cfg.Global.Env)
		},
		"markdown": renderMarkdown,
	}

	// Create template manager with default settings
	tm := &TemplateManager{
		templates:     make(map[string]*template.Template),
		funcMap:       funcMap,
		baseTemplate:  defaultBaseTemplate,
		templateExt:   defaultTemplateExt,
		templatesPath: defaultTemplatesPath,
	}

	// Load templates from embedded filesystem
	if loadErr := tm.Load(); loadErr != nil {
		return nil, fmt.Errorf("loading templates: %w", loadErr)
	}

	return tm, nil
}

// watchFiles sets up a file watcher for hot-reloading templates in development.
// This is only enabled in the local environment.
func (tm *TemplateManager) watchFiles(cfg *config.Config) error {
	// Only watch files in local development environment
	if cfg.Global.Env != config.ENVLocal {
		return nil
	}

	// Create a new file watcher
	fw, err := NewFileWatcher(tm)
	if err != nil {
		return fmt.Errorf("creating file watcher: %w", err)
	}

	// Start watching in a background goroutine
	ctx := context.Background()
	ctx, tm.cancelWatcher = context.WithCancel(ctx)
	go func() {
		if watchErr := fw.Watch(ctx); watchErr != nil {
			log.Printf("Error watching files: %v", watchErr)
		}
	}()

	return nil
}

// Close stops the file watcher if it's running.
func (tm *TemplateManager) Close() {
	if tm.cancelWatcher != nil {
		tm.cancelWatcher()
		tm.cancelWatcher = nil
	}
}

// Load loads all templates from the embedded filesystem.
// This is used during initialization and when templates are reloaded.
func (tm *TemplateManager) Load() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Load page templates
	if err := tm.loadTemplatesFromFS(templatesFS, "assets/templates/partials/*", "assets/templates/pages"); err != nil {
		return fmt.Errorf("loading templates from FS: %w", err)
	}

	// Load component templates
	components, err := template.New("").Funcs(tm.funcMap).ParseFS(templatesFS, "assets/templates/components/*")
	if err != nil {
		return fmt.Errorf("parsing components: %w", err)
	}
	tm.components = components

	return nil
}

// loadTemplatesFromFS loads page templates from the embedded filesystem.
// It parses each template file in the specified directory, combining it with
// the base template and partials.
func (tm *TemplateManager) loadTemplatesFromFS(fsys fs.FS, partials, dir string) error {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		// Skip directories and non-template files
		if entry.IsDir() || filepath.Ext(entry.Name()) != tm.templateExt {
			continue
		}

		// Parse the template with base template and partials
		file := filepath.Join(dir, entry.Name())
		tmpl, err := template.New("").Funcs(tm.funcMap).ParseFS(fsys, partials, tm.baseTemplate, file)
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", file, err)
		}

		// Store the template by name (without extension)
		name := strings.TrimSuffix(filepath.Base(entry.Name()), tm.templateExt)
		tm.templates[name] = tmpl
	}

	return nil
}

// LoadFromFiles loads templates from the filesystem instead of the embedded FS.
// This is used for hot-reloading templates during development.
func (tm *TemplateManager) LoadFromFiles() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Prepare paths for template files
	baseTemplate := filepath.Join(tm.templatesPath, filepath.Base(tm.baseTemplate))
	partials := filepath.Join(tm.templatesPath, "partials")
	partialTemplates, err := filepath.Glob(filepath.Join(partials, "*"+tm.templateExt))
	if err != nil {
		return fmt.Errorf("globbing partials: %w", err)
	}

	// Load component templates
	components := filepath.Join(tm.templatesPath, "components")
	componentsTemplate, err := template.New("").Funcs(tm.funcMap).ParseGlob(filepath.Join(components, "*"+tm.templateExt))
	if err != nil {
		return fmt.Errorf("parsing components: %w", err)
	}
	tm.components = componentsTemplate

	// Walk through page templates and load each one
	return filepath.Walk(filepath.Join(tm.templatesPath, "pages"), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != tm.templateExt {
			return err
		}

		// Parse the template with base template and partials
		files := append([]string{baseTemplate, path}, partialTemplates...)
		tmpl, err := template.New("").Funcs(tm.funcMap).ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", path, err)
		}

		// Store the template by name (without extension)
		name := strings.TrimSuffix(filepath.Base(path), tm.templateExt)
		tm.templates[name] = tmpl
		return nil
	})
}

// RenderComponent renders a component template with the given data.
// Components are partial templates that can be rendered independently.
func (tm *TemplateManager) RenderComponent(w http.ResponseWriter, r *http.Request, name string, data any) {
	if err := tm.components.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		logger.Error(r.Context(), "Failed to execute template", "err", err)
	}
}

// Render renders a full page template with the given data.
// It sets appropriate headers and handles errors.
func (tm *TemplateManager) Render(w http.ResponseWriter, r *http.Request, name string, data any) {
	tmpl, err := tm.getTemplate(r.Context(), name)
	if err != nil {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		logger.Error(r.Context(), "Template not found", "err", err)
		return
	}

	// Set content type header
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the template
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		logger.Error(r.Context(), "Failed to execute template", "err", err)
		return
	}
}

// getTemplate retrieves a template by name from the cache.
// It returns a clone of the template to avoid concurrent modification issues.
func (tm *TemplateManager) getTemplate(_ context.Context, name string) (*template.Template, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tmpl, ok := tm.templates[name]
	if !ok {
		return nil, fmt.Errorf("template %s not found", name)
	}
	return tmpl.Clone()
}

// fileWatcher watches template files for changes and triggers reloading.
// It's used for hot-reloading templates during development.
type fileWatcher struct {
	watcher *fsnotify.Watcher
	tm      *TemplateManager
}

// NewFileWatcher creates a new file watcher for the given template manager.
func NewFileWatcher(tm *TemplateManager) (*fileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating watcher: %w", err)
	}

	return &fileWatcher{watcher: watcher, tm: tm}, nil
}

// Watch starts watching template directories for changes.
// It runs until the context is canceled.
func (fw *fileWatcher) Watch(ctx context.Context) error {
	defer fw.watcher.Close()

	// Start the event loop in a goroutine
	go fw.watchLoop(ctx)

	// Add template directories to the watcher
	templateDirs := []string{"", "pages", "partials", "components"}
	for _, dir := range templateDirs {
		dirPath := filepath.Join(fw.tm.templatesPath, dir)
		if err := fw.watcher.Add(dirPath); err != nil {
			return fmt.Errorf("adding watcher for %s: %w", dirPath, err)
		}
	}

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// watchLoop handles file system events and errors.
// It runs in a separate goroutine.
func (fw *fileWatcher) watchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down file watcher")
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				log.Println("File watcher event channel closed")
				return
			}
			fw.handleFileEvent(event)
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				log.Println("File watcher error channel closed")
				return
			}
			log.Printf("Error in file watcher: %v", err)
		}
	}
}

// handleFileEvent processes a file system event.
// It reloads templates when a template file is modified.
func (fw *fileWatcher) handleFileEvent(event fsnotify.Event) {
	// Only handle write events for template files
	if filepath.Ext(event.Name) != fw.tm.templateExt || event.Op&fsnotify.Write != fsnotify.Write {
		return
	}

	log.Printf("Detected change in template: %s", filepath.Base(event.Name))

	// Reload all templates
	if err := fw.tm.LoadFromFiles(); err != nil {
		log.Printf("Failed to reload templates: %v", err)
	} else {
		log.Printf("Successfully reloaded templates")
	}
}
