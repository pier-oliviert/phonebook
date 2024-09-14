package utils

import (
	"fmt"
	"os"

	"k8s.io/utils/env"
)

const kPhonebookConfighPath = "/var/run/configs/provider"

// First check if the environment variable is set, if not, let's look for the
// token at `${kProviderConfigPath}/${kCloudflareAPIKeyName}` and read the content
// of that file into token
func RetrieveValueFromEnvOrFile(envNameOrFileName string) (content string, err error) {
	content = env.GetString(envNameOrFileName, "")

	if content == "" {
		path := fmt.Sprintf("%s/%s", kPhonebookConfighPath, envNameOrFileName)
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("E#4002: %s does not exist as an environment variable and a file(%s) with this name could not be found", envNameOrFileName, path)
		}
		content = string(data)
	}

	return content, nil
}
