package config_test

import (
	"os"
	"testing"
	"time"

	"go-expert-rater-limit/config"
)

func TestConfig(t *testing.T) {
	t.Run("should load default values when env vars are not set", func(t *testing.T) {
		// Limpa variáveis de ambiente antes do teste
		os.Clearenv()

		cfg := config.Load()

		// Verifica valores padrão
		if cfg.RedisAddr != "localhost:6379" {
			t.Errorf("Expected RedisAddr to be localhost:6379, got %s", cfg.RedisAddr)
		}
		if cfg.IPLimit != 5 {
			t.Errorf("Expected IPLimit to be 5, got %d", cfg.IPLimit)
		}
		if cfg.TokenLimit != 10 {
			t.Errorf("Expected TokenLimit to be 10, got %d", cfg.TokenLimit)
		}
		if cfg.IPDuration != time.Second {
			t.Errorf("Expected IPDuration to be 1s, got %v", cfg.IPDuration)
		}
		if cfg.IPBlockTime != 5*time.Minute {
			t.Errorf("Expected BlockTime to be 5m, got %v", cfg.IPBlockTime)
		}
		if cfg.TokenBlockTime != 6*time.Minute {
			t.Errorf("Expected TokenBlockTime to be 6m, got %v", cfg.TokenBlockTime)
		}
		if cfg.ServerPort != "8080" {
			t.Errorf("Expected ServerPort to be 8080, got %s", cfg.ServerPort)
		}
	})

	t.Run("should load values from environment variables", func(t *testing.T) {
		// Configura variáveis de ambiente para o teste
		envVars := map[string]string{
			"REDIS_ADDR":       "redis:7000",
			"IP_LIMIT":         "100",
			"TOKEN_LIMIT":      "200",
			"IP_DURATION":      "2s",
			"IP_BLOCK_TIME":    "10m",
			"TOKEN_BLOCK_TIME": "15m",
			"SERVER_PORT":      "3000",
		}

		for k, v := range envVars {
			os.Setenv(k, v)
		}
		defer os.Clearenv()

		cfg := config.Load()

		// Verifica se os valores foram carregados corretamente
		if cfg.RedisAddr != "redis:7000" {
			t.Errorf("Expected RedisAddr to be redis:7000, got %s", cfg.RedisAddr)
		}
		if cfg.IPLimit != 100 {
			t.Errorf("Expected IPLimit to be 100, got %d", cfg.IPLimit)
		}
		if cfg.TokenLimit != 200 {
			t.Errorf("Expected TokenLimit to be 200, got %d", cfg.TokenLimit)
		}
		if cfg.IPDuration != 2*time.Second {
			t.Errorf("Expected IPDuration to be 2s, got %v", cfg.IPDuration)
		}
		if cfg.IPBlockTime != 10*time.Minute {
			t.Errorf("Expected IPBlockTime to be 10m, got %v", cfg.IPBlockTime)
		}
		if cfg.TokenBlockTime != 15*time.Minute {
			t.Errorf("Expected TokenBlockTime to be 15m, got %v", cfg.TokenBlockTime)
		}
		if cfg.ServerPort != "3000" {
			t.Errorf("Expected ServerPort to be 3000, got %s", cfg.ServerPort)
		}
	})

	t.Run("should handle invalid duration format", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("IP_DURATION", "invalid")
		os.Setenv("IP_BLOCK_TIME", "invalid")
		os.Setenv("TOKEN_BLOCK_TIME", "invalid")

		cfg := config.Load()

		// Deve usar valores padrão quando o formato é inválido
		if cfg.IPDuration != time.Second {
			t.Errorf("Expected IPDuration to fallback to 1s, got %v", cfg.IPDuration)
		}
		if cfg.IPBlockTime != 5*time.Minute {
			t.Errorf("Expected IPBlockTime to fallback to 5m, got %v", cfg.IPBlockTime)
		}
		if cfg.TokenBlockTime != 6*time.Minute {
			t.Errorf("Expected TokenBlockTime to fallback to 6m, got %v", cfg.TokenBlockTime)
		}
	})

	t.Run("should handle invalid integer format", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("IP_LIMIT", "invalid")
		os.Setenv("TOKEN_LIMIT", "invalid")

		cfg := config.Load()

		// Deve usar valores padrão quando o formato é inválido
		if cfg.IPLimit != 5 {
			t.Errorf("Expected IPLimit to fallback to 5, got %d", cfg.IPLimit)
		}
		if cfg.TokenLimit != 10 {
			t.Errorf("Expected TokenLimit to fallback to 10, got %d", cfg.TokenLimit)
		}
	})
}
