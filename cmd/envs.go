package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

func getStringFromEnv(envName string) (string, error) {
	env := os.Getenv(envName)
	if len(env) == 0 {
		return "", fmt.Errorf("%s not set", envName)
	}

	return env, nil
}

func getStringFromEnvOrDefault(envName string, defaultValue string) string {
	env, err := getStringFromEnv(envName)
	if err != nil {
		return defaultValue
	}

	return env
}

func getIntValueFromEnv(envName string, defaultValue int64) (int64, error) {
	sVal := os.Getenv(envName)

	if sVal != "" {
		val, err := strconv.ParseInt(sVal, 0, 32)
		if err != nil {
			return defaultValue, errors.Join(err, fmt.Errorf("cant parse %s", envName))
		}

		return val, nil
	}

	return defaultValue, nil
}
