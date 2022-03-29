package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// JSON 形式のダンプファイルからモデルの集合をロードする
func (s *Set[T]) LoadJSON(jsonFile string) error {
	// 引数に渡されたファイルを開く
	file, err := os.Open(jsonFile)
	if err != nil {
		return err
	}
	defer file.Close()

	models := []T{}

	// JSON 形式としてデコード
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&models); err != nil {
		return err
	}

	// デコードしたモデルを先頭から順に Set に追加
	for _, model := range models {
		if !s.Add(model) {
			return fmt.Errorf("Unexpected error on dump loading: %v", model)
		}
	}

	return nil
}
