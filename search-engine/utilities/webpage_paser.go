package utilities

import (
	"encoding/json"
	"search-engine/internal/types"
)

func ParseWebpages(data []byte) (*[]types.WebpageTFIDF, error) {

	var webpages []types.WebpageTFIDF
	err := json.Unmarshal(data, &webpages)
	if err != nil {
		return nil, err
	}
	return &webpages, nil
}
