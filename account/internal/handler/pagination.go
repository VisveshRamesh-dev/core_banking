package handler

import (
	"fmt"
	"strconv"
)

func parseOffset(token string, out *int) (string, error) {
	n, err := strconv.Atoi(token)
	if err != nil {
		return token, err
	}
	*out = n
	return token, nil
}

func formatOffset(offset int) string {
	return fmt.Sprintf("%d", offset)
}
