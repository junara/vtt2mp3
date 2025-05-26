package vtt

import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// エラー定義
var (
	ErrInvalidVTTHeader     = errors.New("invalid VTT file: missing WEBVTT header")
	ErrInvalidTimestamp     = errors.New("invalid timestamp format")
	ErrInvalidSecondsFormat = errors.New("invalid seconds format")
)

// 定数定義
const (
	vttHeader        = "WEBVTT"
	timestampPattern = `(\d{2}:\d{2}:\d{2}\.\d{3}) --> (\d{2}:\d{2}:\d{2}\.\d{3})`
)

// Subtitle はVTTファイル内の単一の字幕エントリーを表します
type Subtitle struct {
	ID        string
	StartTime time.Duration
	EndTime   time.Duration
	Text      string
}

// VTTFile は解析されたVTTファイルを表します
type VTTFile struct {
	Subtitles []Subtitle
}

// ParseVTTFile はVTTファイルを解析し、VTTFile構造体を返します
func ParseVTTFile(filePath string) (vttFile *VTTFile, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)

	// ファイルが"WEBVTT"で始まるかどうかを確認
	if !scanner.Scan() || !strings.HasPrefix(scanner.Text(), vttHeader) {
		return nil, ErrInvalidVTTHeader
	}

	subtitles, err := parseSubtitles(scanner)
	if err != nil {
		return nil, err
	}

	return &VTTFile{Subtitles: subtitles}, nil
}

// parseSubtitles はヘッダーが処理された後にスキャナーから字幕を抽出します
func parseSubtitles(scanner *bufio.Scanner) ([]Subtitle, error) {
	timestampRegex := regexp.MustCompile(timestampPattern)
	subtitles := []Subtitle{}

	var currentSubtitle *Subtitle
	var textLines []string

	for scanner.Scan() {
		line := scanner.Text()

		// 空行は字幕エントリーの区切りとして扱う
		if line == "" {
			if currentSubtitle != nil && len(textLines) > 0 {
				currentSubtitle.Text = strings.Join(textLines, "\n")
				subtitles = append(subtitles, *currentSubtitle)
				currentSubtitle = nil
				textLines = nil
			}
			continue
		}

		// タイムスタンプ行のチェック
		matches := timestampRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			// 前の字幕が処理中なら保存する
			if currentSubtitle != nil && len(textLines) > 0 {
				currentSubtitle.Text = strings.Join(textLines, "\n")
				subtitles = append(subtitles, *currentSubtitle)
			}

			startTime, err := parseTimestamp(matches[1])
			if err != nil {
				return nil, err
			}

			endTime, err := parseTimestamp(matches[2])
			if err != nil {
				return nil, err
			}

			currentSubtitle = &Subtitle{
				StartTime: startTime,
				EndTime:   endTime,
			}
			textLines = []string{}
			continue
		}

		// 現在の字幕にテキスト行を追加
		if currentSubtitle != nil {
			textLines = append(textLines, line)
		}
	}

	// 最後の字幕を追加
	if currentSubtitle != nil && len(textLines) > 0 {
		currentSubtitle.Text = strings.Join(textLines, "\n")
		subtitles = append(subtitles, *currentSubtitle)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return subtitles, nil
}

// parseTimestamp はタイムスタンプ文字列（HH:MM:SS.mmm）をtime.Durationに変換します
func parseTimestamp(timestamp string) (time.Duration, error) {
	parts := strings.Split(timestamp, ":")
	if len(parts) != 3 {
		return 0, ErrInvalidTimestamp
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	secondsParts := strings.Split(parts[2], ".")
	if len(secondsParts) != 2 {
		return 0, ErrInvalidSecondsFormat
	}

	seconds, err := strconv.Atoi(secondsParts[0])
	if err != nil {
		return 0, err
	}

	milliseconds, err := strconv.Atoi(secondsParts[1])
	if err != nil {
		return 0, err
	}

	duration := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second +
		time.Duration(milliseconds)*time.Millisecond

	return duration, nil
}
