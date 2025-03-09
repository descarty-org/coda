package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	// Loading Configuration Tests
	t.Run("Load_Success", func(t *testing.T) {
		t.Parallel()

		validConfig := `
server:
  host: 127.0.0.1
  port: 8080
  allowedOrigins:
    - http://localhost:3000
logging:
  format: json
llm:
  openai:
    apiKey: test-api-key
  langfuse:
    privateKey: test-private-key
    publicKey: test-public-key
`
		tempDir := setupConfigDir(t, validConfig, "local.yaml")

		cfg, err := Load(ENVLocal, tempDir)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Validate the loaded config
		if cfg.Server.Host != "127.0.0.1" {
			t.Errorf("Expected host to be 127.0.0.1, got %s", cfg.Server.Host)
		}
		if cfg.Server.Port != 8080 {
			t.Errorf("Expected port to be 8080, got %d", cfg.Server.Port)
		}
		if len(cfg.Server.AllowedOrigins) != 1 || cfg.Server.AllowedOrigins[0] != "http://localhost:3000" {
			t.Errorf("Expected allowed origins to be [http://localhost:3000], got %v", cfg.Server.AllowedOrigins)
		}
		if cfg.Logging.Format != "json" {
			t.Errorf("Expected logging format to be json, got %s", cfg.Logging.Format)
		}
		if cfg.Global.Env != ENVLocal {
			t.Errorf("Expected env to be local, got %s", cfg.Global.Env)
		}
		if cfg.LLM.OpenAI.APIKey != "test-api-key" {
			t.Errorf("Expected OpenAI API key to be test-api-key, got %s", cfg.LLM.OpenAI.APIKey)
		}
		if cfg.LLM.Langfuse.PrivateKey != "test-private-key" {
			t.Errorf("Expected Langfuse private key to be test-private-key, got %s", cfg.LLM.Langfuse.PrivateKey)
		}
		if cfg.LLM.Langfuse.PublicKey != "test-public-key" {
			t.Errorf("Expected Langfuse public key to be test-public-key, got %s", cfg.LLM.Langfuse.PublicKey)
		}
	})

	// Error Handling Tests
	t.Run("Load_Errors", func(t *testing.T) {
		t.Parallel()

		// Table-driven test for error cases
		tests := []struct {
			name      string
			content   string
			setupFunc func(t *testing.T) string
			expectErr bool
		}{
			{
				name:    "FileNotFound",
				content: "",
				setupFunc: func(t *testing.T) string {
					return setupConfigDir(t, "", "nonexistent.yaml")
				},
				expectErr: true,
			},
			{
				name: "InvalidYAML",
				content: `
server:
  host: 127.0.0.1
  port: 8080
  this is not valid yaml
`,
				setupFunc: func(t *testing.T) string {
					return setupConfigDir(t, `
server:
  host: 127.0.0.1
  port: 8080
  this is not valid yaml
`, "local.yaml")
				},
				expectErr: true,
			},
			{
				name: "ValidationFailure",
				content: `
logging:
  format: json
# Missing required server configuration
`,
				setupFunc: func(t *testing.T) string {
					return setupConfigDir(t, `
logging:
  format: json
# Missing required server configuration
`, "local.yaml")
				},
				expectErr: true,
			},
		}

		for _, tc := range tests {
			tc := tc // Capture range variable
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				tempDir := tc.setupFunc(t)

				_, err := Load(ENVLocal, tempDir)
				if tc.expectErr && err == nil {
					t.Errorf("Expected error, got nil")
				} else if !tc.expectErr && err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			})
		}
	})

	// Environment Variable Override Tests
	t.Run("Load_WithEnvOverrides", func(t *testing.T) {
		validConfig := `
server:
  host: 127.0.0.1
  port: 8080
  allowedOrigins:
    - http://localhost:3000
logging:
  format: json
llm:
  openai:
    apiKey: test-api-key
  langfuse:
    privateKey: test-private-key
    publicKey: test-public-key
`
		tempDir := setupConfigDir(t, validConfig, "local.yaml")

		// Test with valid environment variable overrides
		t.Run("ValidOverrides", func(t *testing.T) {
			// Set environment variables to override config
			os.Setenv("HOST", "0.0.0.0")
			os.Setenv("PORT", "9090")
			os.Setenv("ALLOWED_ORIGINS", "https://example.com,https://test.com")
			os.Setenv("OPENAI_API_KEY", "env-api-key")
			t.Cleanup(func() {
				os.Unsetenv("HOST")
				os.Unsetenv("PORT")
				os.Unsetenv("ALLOWED_ORIGINS")
				os.Unsetenv("OPENAI_API_KEY")
			})

			cfg, err := Load(ENVLocal, tempDir)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Validate the loaded config with environment overrides
			if cfg.Server.Host != "0.0.0.0" {
				t.Errorf("Expected host to be 0.0.0.0 (from env), got %s", cfg.Server.Host)
			}
			if cfg.Server.Port != 9090 {
				t.Errorf("Expected port to be 9090 (from env), got %d", cfg.Server.Port)
			}
			expectedOrigins := []string{"https://example.com", "https://test.com"}
			if len(cfg.Server.AllowedOrigins) != len(expectedOrigins) {
				t.Errorf("Expected %d allowed origins, got %d", len(expectedOrigins), len(cfg.Server.AllowedOrigins))
			} else {
				for i, origin := range cfg.Server.AllowedOrigins {
					if origin != expectedOrigins[i] {
						t.Errorf("Expected allowed origin %d to be %s, got %s", i, expectedOrigins[i], origin)
					}
				}
			}
			if cfg.LLM.OpenAI.APIKey != "env-api-key" {
				t.Errorf("Expected OpenAI API key to be env-api-key, got %s", cfg.LLM.OpenAI.APIKey)
			}
		})

		// Test with invalid environment variable
		t.Run("InvalidOverride", func(t *testing.T) {
			os.Setenv("PORT", "not-a-number")
			t.Cleanup(func() {
				os.Unsetenv("PORT")
			})

			_, err := Load(ENVLocal, tempDir)
			if err == nil {
				t.Error("Expected error when environment variable is invalid, got nil")
			}
		})
	})

	// Environment Variable Loading Tests
	t.Run("LoadEnvVariables", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name         string
			envVars      map[string]string
			initialCfg   Config
			expectedCfg  Config
			expectErr    bool
			errorMessage string
		}{
			{
				name:    "NoEnvVars",
				envVars: map[string]string{},
				initialCfg: Config{
					Server: Server{
						Host:           "127.0.0.1",
						Port:           8080,
						AllowedOrigins: []string{"http://localhost:3000"},
					},
				},
				expectedCfg: Config{
					Server: Server{
						Host:           "127.0.0.1",
						Port:           8080,
						AllowedOrigins: []string{"http://localhost:3000"},
					},
				},
				expectErr: false,
			},
			{
				name: "AllOverrides",
				envVars: map[string]string{
					"HOST":            "0.0.0.0",
					"PORT":            "9090",
					"ALLOWED_ORIGINS": "https://example.com,https://test.com",
					"OPENAI_API_KEY":  "env-api-key",
				},
				initialCfg: Config{
					Server: Server{
						Host:           "127.0.0.1",
						Port:           8080,
						AllowedOrigins: []string{"http://localhost:3000"},
					},
					LLM: LLM{
						OpenAI: OpenAI{
							APIKey: "test-api-key",
						},
					},
				},
				expectedCfg: Config{
					Server: Server{
						Host:           "0.0.0.0",
						Port:           9090,
						AllowedOrigins: []string{"https://example.com", "https://test.com"},
					},
					LLM: LLM{
						OpenAI: OpenAI{
							APIKey: "env-api-key",
						},
					},
				},
				expectErr: false,
			},
			{
				name: "InvalidPort",
				envVars: map[string]string{
					"PORT": "not-a-number",
				},
				initialCfg: Config{
					Server: Server{
						Host:           "127.0.0.1",
						Port:           8080,
						AllowedOrigins: []string{"http://localhost:3000"},
					},
				},
				expectedCfg: Config{}, // Not used when expectErr is true
				expectErr:   true,
			},
		}

		for _, tc := range tests {
			tc := tc // Capture range variable
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				// Set environment variables
				for k, v := range tc.envVars {
					os.Setenv(k, v)
					t.Cleanup(func() {
						os.Unsetenv(k)
					})
				}

				// Create a copy of the initial config
				config := tc.initialCfg

				// Call the function being tested
				err := LoadEnvVariables(&config)

				// Check error expectations
				if tc.expectErr {
					if err == nil {
						t.Errorf("Expected error, got nil")
					}
					return
				}

				if err != nil {
					t.Errorf("LoadEnvVariables() error = %v, wantErr false", err)
					return
				}

				// Validate the config
				if config.Server.Host != tc.expectedCfg.Server.Host {
					t.Errorf("Expected Host to be %s, got %s", tc.expectedCfg.Server.Host, config.Server.Host)
				}
				if config.Server.Port != tc.expectedCfg.Server.Port {
					t.Errorf("Expected Port to be %d, got %d", tc.expectedCfg.Server.Port, config.Server.Port)
				}
				if len(config.Server.AllowedOrigins) != len(tc.expectedCfg.Server.AllowedOrigins) {
					t.Errorf("Expected %d AllowedOrigins, got %d", len(tc.expectedCfg.Server.AllowedOrigins), len(config.Server.AllowedOrigins))
					return
				}
				for i, origin := range config.Server.AllowedOrigins {
					if origin != tc.expectedCfg.Server.AllowedOrigins[i] {
						t.Errorf("Expected AllowedOrigins[%d] to be %s, got %s", i, tc.expectedCfg.Server.AllowedOrigins[i], origin)
					}
				}

				// Check LLM config if present in expected
				if tc.expectedCfg.LLM.OpenAI.APIKey != "" && config.LLM.OpenAI.APIKey != tc.expectedCfg.LLM.OpenAI.APIKey {
					t.Errorf("Expected OpenAI API key to be %s, got %s", tc.expectedCfg.LLM.OpenAI.APIKey, config.LLM.OpenAI.APIKey)
				}
			})
		}
	})

	// Options Tests
	t.Run("Options_SetDefaults", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			options  Options
			expected Options
		}{
			{
				name:     "EmptyOptions",
				options:  Options{},
				expected: Options{Env: ENVLocal, ConfigDir: "config"},
			},
			{
				name:     "CustomEnv",
				options:  Options{Env: ENVProduction},
				expected: Options{Env: ENVProduction, ConfigDir: "config"},
			},
			{
				name:     "CustomConfigDir",
				options:  Options{ConfigDir: "/etc/myapp"},
				expected: Options{Env: ENVLocal, ConfigDir: "/etc/myapp"},
			},
			{
				name:     "CustomEnvAndConfigDir",
				options:  Options{Env: ENVDevelopment, ConfigDir: "/etc/myapp"},
				expected: Options{Env: ENVDevelopment, ConfigDir: "/etc/myapp"},
			},
		}

		for _, tc := range tests {
			tc := tc // Capture range variable
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				options := tc.options
				options.setDefaults()
				if options.Env != tc.expected.Env {
					t.Errorf("Expected Env to be %s, got %s", tc.expected.Env, options.Env)
				}
				if options.ConfigDir != tc.expected.ConfigDir {
					t.Errorf("Expected ConfigDir to be %s, got %s", tc.expected.ConfigDir, options.ConfigDir)
				}
			})
		}
	})

	// Candidates Tests
	t.Run("Options_Candidates", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			env      ENV
			expected []string
		}{
			{
				name:     "LocalEnv",
				env:      ENVLocal,
				expected: []string{"local.yaml"},
			},
			{
				name:     "DevelopmentEnv",
				env:      ENVDevelopment,
				expected: []string{"development.yaml"},
			},
			{
				name:     "ProductionEnv",
				env:      ENVProduction,
				expected: []string{"production.yaml"},
			},
			{
				name:     "TestEnv",
				env:      ENVTest,
				expected: []string{"test.yaml"},
			},
			{
				name:     "CustomEnv",
				env:      ENV("staging"),
				expected: []string{"staging.yaml"},
			},
		}

		for _, tc := range tests {
			tc := tc // Capture range variable
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				options := Options{Env: tc.env}
				candidates := options.candidates()
				if len(candidates) != len(tc.expected) {
					t.Errorf("Expected %d candidates, got %d", len(tc.expected), len(candidates))
					return
				}
				for i, candidate := range candidates {
					if candidate != tc.expected[i] {
						t.Errorf("Expected candidate %d to be %s, got %s", i, tc.expected[i], candidate)
					}
				}
			})
		}
	})

	// Config Files Tests
	t.Run("FindConfigFiles", func(t *testing.T) {
		// Create a temporary directory for test config files
		tempDir, err := os.MkdirTemp("", "config-files-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		t.Cleanup(func() {
			_ = os.RemoveAll(tempDir)
		})

		// Create test config files
		for _, env := range []string{"local", "development", "production"} {
			if err := os.WriteFile(filepath.Join(tempDir, env+".yaml"), []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}
		}

		tests := []struct {
			name      string
			options   Options
			expectErr bool
		}{
			{
				name:      "LocalEnv",
				options:   Options{Env: ENVLocal, ConfigDir: tempDir},
				expectErr: false,
			},
			{
				name:      "DevelopmentEnv",
				options:   Options{Env: ENVDevelopment, ConfigDir: tempDir},
				expectErr: false,
			},
			{
				name:      "ProductionEnv",
				options:   Options{Env: ENVProduction, ConfigDir: tempDir},
				expectErr: false,
			},
			{
				name:      "NonExistentEnv",
				options:   Options{Env: ENV("nonexistent"), ConfigDir: tempDir},
				expectErr: true,
			},
			{
				name:      "NonExistentConfigDir",
				options:   Options{Env: ENVLocal, ConfigDir: filepath.Join(tempDir, "nonexistent")},
				expectErr: true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				files, err := findConfigFiles(tc.options)
				if tc.expectErr {
					if err == nil {
						t.Error("Expected error, got nil")
					}
					return
				}
				if err != nil {
					t.Errorf("findConfigFiles() error = %v, wantErr false", err)
					return
				}
				if len(files) != 1 {
					t.Errorf("Expected 1 file, got %d", len(files))
				}
			})
		}
	})
}

// Helper function to create a temporary directory with a config file
func setupConfigDir(t *testing.T, content string, filename string) string {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	if content != "" {
		if err := os.WriteFile(filepath.Join(tempDir, filename), []byte(content), 0644); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to write test config file: %v", err)
		}
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	return tempDir
}
