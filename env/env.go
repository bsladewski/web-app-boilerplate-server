// Package env provides convenience functions for reading environment variables.
package env

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// GetString retrieves the specified environment variable as a string.
func GetString(key string) string {
	return os.Getenv(key)
}

// GetStringSafe retrieves the specified environment variable as a string
// returning the supplied default value if the environment variable is not set.
func GetStringSafe(key, defaultVal string) string {
	if val := GetString(key); val != "" {
		return val
	}
	return defaultVal
}

// MustGetString retrieves the specified environment variable as a string
// logging a fatal error if the environment variable is not set.
func MustGetString(key string) string {
	if val := GetString(key); val != "" {
		return val
	}
	logrus.Fatalf("environment variable '%s' not set", key)
	return ""
}

// GetInt retrieves the specified environment variable as an int returning the
// zero value if the environment variable is not set.
func GetInt(key string) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return 0, nil
	}
	return strconv.Atoi(val)
}

// GetIntSafe retrieves the specified environment variable as an int returning
// the supplied default value if the environment variable is not set or is not
// valid.
func GetIntSafe(key string, defaultVal int) int {
	val, err := GetInt(key)
	if err != nil {
		logrus.Error(err)
		return defaultVal
	} else if val == 0 {
		return defaultVal
	}
	return val
}

// MustGetInt retrieves the specified environment variable as an int logging a
// fatal error if the environment variable is not set or is invalid.
func MustGetInt(key string) int {
	val, err := GetInt(key)
	if err != nil {
		logrus.Fatal(err)
		return 0
	} else if val == 0 {
		logrus.Fatalf("environment variable '%s' not set", key)
		return 0
	}
	return val
}

// GetFloat64 retrieves the specified environment variable as a float64
// returning the zero value if the environment variable is not set.
func GetFloat64(key string) (float64, error) {
	val := os.Getenv(key)
	if val == "" {
		return 0.0, nil
	}
	return strconv.ParseFloat(val, 64)
}

// GetFloat64Safe retrieves the specified environment variable as a float64
// returning the supplied default value if the environment variable is not set
// or is not valid.
func GetFloat64Safe(key string, defaultVal float64) float64 {
	val, err := GetFloat64(key)
	if err != nil {
		logrus.Error(err)
		return defaultVal
	} else if val == 0.0 {
		return defaultVal
	}
	return val
}

// MustGetFloat64 retrieves the specified environment variable as a float64
// logging a fatal error if the environment variable not set or is invalid.
func MustGetFloat64(key string, defaultVal float64) float64 {
	val, err := GetFloat64(key)
	if err != nil {
		logrus.Error(err)
		return 0.0
	} else if val == 0.0 {
		logrus.Fatalf("environment variable '%s' not set", key)
		return 0.0
	}
	return val
}

// GetBool retrieves the specified environment variable as a bool returning the
// zero value if the environment variable is not set.
func GetBool(key string) (bool, error) {
	val := os.Getenv(key)
	if val == "" {
		return false, nil
	}
	return strconv.ParseBool(val)
}

// GetBoolSafe retrieves the specified environment variable as a bool returning
// the supplied default value if the environment variable is not set or is not
// valid.
func GetBoolSafe(key string, defaultVal bool) bool {
	val, err := GetBool(key)
	if err != nil {
		logrus.Error(err)
		return defaultVal
	} else if !val {
		return defaultVal
	}
	return val
}
