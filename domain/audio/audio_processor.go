package audio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const (
	tempDirPrefix = "/tmp/vtt2mp3_"
)

// AudioProcessor はTTSプロバイダーに依存しない音声処理操作を扱います
type AudioProcessor struct{}

// NewAudioProcessor は新しいAudioProcessorを作成します
func NewAudioProcessor() *AudioProcessor {
	return &AudioProcessor{}
}

// CreateTempDir は一時ディレクトリを作成します
func (p *AudioProcessor) CreateTempDir() (string, error) {
	tempDir := tempDirPrefix + uuid.New().String()
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("一時ディレクトリの作成に失敗しました: %v", err)
	}

	return tempDir, nil
}

// CleanupTempDir は一時ディレクトリを削除します
func (p *AudioProcessor) CleanupTempDir(tempDir string) {
	if err := os.RemoveAll(tempDir); err != nil {
		fmt.Printf("警告: 一時ディレクトリ %s の削除に失敗しました: %v\n", tempDir, err)
	}
}

// GetAudioDuration はffmpegを使用して音声ファイルの長さを取得します
func (p *AudioProcessor) GetAudioDuration(audioFile string) (time.Duration, error) {
	// ffmpegコマンドを実行して長さ情報を取得
	cmd := exec.Command("ffmpeg", "-i", audioFile, "-f", "null", "-")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// ここではエラーが発生することを想定しています（ffmpegは-f nullを使用すると終了コードでエラーを返す）
	// 長さ情報を含む標準エラー出力のみを使用します
	_ = cmd.Run() // 意図的にエラーを無視

	// 出力から長さを抽出
	output := stderr.String()
	durationRegex := regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2}\.\d{2})`)
	matches := durationRegex.FindStringSubmatch(output)

	if len(matches) < 4 {
		return 0, fmt.Errorf("ffmpeg出力から長さを抽出できませんでした")
	}

	// 時間、分、秒を解析
	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.ParseFloat(matches[3], 64)

	// time.Durationに変換
	duration := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds*float64(time.Second))

	return duration, nil
}

// MixAudioFilesWithTiming は正確なタイミングでffmpegを使用して全ての音声ファイルを結合します
func (p *AudioProcessor) MixAudioFilesWithTiming(audioFiles []string, startTimes []time.Duration, output io.Writer) error {
	if len(audioFiles) == 0 {
		return fmt.Errorf("結合する音声ファイルがありません")
	}

	if len(audioFiles) != len(startTimes) {
		return fmt.Errorf("音声ファイル数(%d)が開始時間の数(%d)と一致しません", len(audioFiles), len(startTimes))
	}

	// 出力に必要な合計長さを計算
	var maxEndTime time.Duration
	for i, audioFile := range audioFiles {
		// 音声ファイルの実際の長さを取得
		duration, err := p.GetAudioDuration(audioFile)
		if err != nil {
			return fmt.Errorf("音声ファイル %s の長さの取得に失敗しました: %v", audioFile, err)
		}

		// 実際の長さを使用して終了時間を計算
		endTime := startTimes[i] + duration
		if endTime > maxEndTime {
			maxEndTime = endTime
		}
	}

	// 正確なタイミングのための複合フィルターを作成
	filterComplex := ""

	// フィルター複合部分を作成
	for i, startTime := range startTimes {
		// 開始時間をミリ秒に変換（adelayフィルター用）
		delayMs := startTime.Milliseconds()

		// 正確な遅延を持つ音声を追加
		// adelay=delays:all=1 は遅延を全チャンネルに適用することを意味します
		filterComplex += fmt.Sprintf("[%d]adelay=%d:all=1[a%d]; ", i, delayMs, i)
	}

	// 全ての遅延された音声ストリームを結合
	if len(audioFiles) == 1 {
		// 音声ファイルが1つのみの場合、直接出力にマッピング
		filterComplex += "[a0]aformat=sample_fmts=fltp:sample_rates=44100:channel_layouts=stereo[aout]"
	} else {
		// 複数の音声ファイルの場合、結合チェーンを作成
		filterComplex += "[a0][a1]amix=inputs=2:dropout_transition=0:normalize=0[tmp0]; "

		for i := 2; i < len(audioFiles); i++ {
			filterComplex += fmt.Sprintf("[tmp%d][a%d]amix=inputs=2:dropout_transition=0:normalize=0[tmp%d]; ",
				i-2, i, i-1)
		}

		// 最後のtmpをaoutにマッピング
		filterComplex += fmt.Sprintf("[tmp%d]aformat=sample_fmts=fltp:sample_rates=44100:channel_layouts=stereo[aout]",
			len(audioFiles)-2)
	}

	// ffmpegコマンドを構築
	cmd := exec.Command("ffmpeg", "-y")

	// 全ての入力ファイルを追加
	for _, audioFile := range audioFiles {
		cmd.Args = append(cmd.Args, "-i", audioFile)
	}

	// フィルター複合と出力オプションを追加
	cmd.Args = append(cmd.Args,
		"-filter_complex", filterComplex,
		"-map", "[aout]",
		"-c:a", "libmp3lame", // 高品質MP3エンコーディングを使用
		"-q:a", "0", // 最高品質の設定を使用
		"-f", "mp3",
		"pipe:1",
	)

	cmd.Stdout = output
	cmd.Stderr = os.Stderr // デバッグ用

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("タイミング付きの音声ファイルの結合に失敗しました: %v", err)
	}

	return nil
}
