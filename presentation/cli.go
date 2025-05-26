package presentation

import (
	"flag"
	"fmt"
	"path/filepath"

	"vtt2mp3/application"
)

// CLI はアプリケーションのコマンドラインインターフェースを表します
type CLI struct {
	service *application.VTT2MP3Service
}

// NewCLI は新しいCLIを作成します
func NewCLI(service *application.VTT2MP3Service) *CLI {
	return &CLI{
		service: service,
	}
}

// Run はCLIアプリケーションを実行します
func (c *CLI) Run(args []string) error {
	// コマンドラインフラグを定義
	flagSet := flag.NewFlagSet("vtt2mp3", flag.ExitOnError)
	inputFile := flagSet.String("i", "input.vtt", "入力VTTファイル")
	outputFile := flagSet.String("o", "out.mp3", "出力MP3ファイル")
	languageCode := flagSet.String("l", "ja", "言語コード")

	// コマンドラインフラグを解析
	if err := flagSet.Parse(args); err != nil {
		return fmt.Errorf("コマンドラインフラグの解析に失敗しました: %v", err)
	}

	// オプションを表示
	fmt.Printf("%sを%sに言語%sで変換しています\n", *inputFile, *outputFile, *languageCode)

	// 出力ファイルがMP4（動画出力）かどうかを確認
	isVideoOutput := false
	if filepath.Ext(*outputFile) == ".mp4" {
		isVideoOutput = true
		fmt.Println(".mp4拡張子を検出しました、動画出力を生成します")
	}

	// VTTをMP3またはMP4に変換
	options := application.ConvertOptions{
		InputFile:     *inputFile,
		OutputFile:    *outputFile,
		LanguageCode:  *languageCode,
		IsVideoOutput: isVideoOutput,
	}
	if err := c.service.Convert(options); err != nil {
		if isVideoOutput {
			return fmt.Errorf("VTTをMP4に変換できませんでした: %v", err)
		}
		return fmt.Errorf("VTTをMP3に変換できませんでした: %v", err)
	}

	fmt.Printf("%sを%sに変換しました\n", *inputFile, *outputFile)
	return nil
}
