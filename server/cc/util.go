package cc

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/rs/zerolog/log"
)

func getAuthDetails(basicAuth string) (appId, apiKey string, e error) {
	src := strings.ReplaceAll(basicAuth, "Basic ", "")
	dst, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", "", err
	} else {
		split := strings.Split(string(dst), ":")
		if len(split) != 2 {
			err = errors.New("Malformed basic auth")
			log.Error().AnErr("error", err).Any("input", string(dst)).Msg("Malformed basic auth.")
			return "", "", err
		} else {
			return split[0], split[1], nil
		}
	}
}
