# vtt2mp3

Google Cloud Text-to-Speech APIを使用してWebVTT字幕ファイルをMP3音声ファイルまたはMP4動画ファイルに変換するコマンドラインツールです。

## 特徴

- WebVTT字幕ファイルをMP3音声ファイルに変換
- WebVTT字幕ファイルをMP4動画ファイル（黒背景に字幕付き）に変換
- VTTファイルからタイミング情報を保持
- Google Cloud Text-to-Speech APIによる複数言語のサポート
- 入力および出力ファイルパスのカスタマイズ可能

## 前提条件

- Go 1.24以降
- システムにffmpegがインストールされていること
- Text-to-Speech API用のGoogle Cloud認証情報が設定されていること

## Google Cloud認証の設定

このアプリケーションはGoogle Cloud Text-to-Speech APIを使用するため、Google Cloudの認証設定が必要です。以下の手順に従って設定してください：

1. [Google Cloud Console](https://console.cloud.google.com/)にアクセスし、プロジェクトを作成または選択します。
2. [APIとサービス] > [ライブラリ]から「Cloud Text-to-Speech API」を検索し、有効化します。
3. [APIとサービス] > [認証情報]から、サービスアカウントを作成します。
4. サービスアカウントに「Text-to-Speech ユーザー」のロールを付与します。
5. サービスアカウントのJSONキーファイルを作成し、ダウンロードします。
6. 環境変数を設定して、アプリケーションがこの認証情報を使用できるようにします：

```shell script
# Linux/macOS
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your-project-credentials.json"

# Windows PowerShell
$env:GOOGLE_APPLICATION_CREDENTIALS="C:\path\to\your-project-credentials.json"

# Windows コマンドプロンプト
set GOOGLE_APPLICATION_CREDENTIALS=C:\path\to\your-project-credentials.json
```

詳細は[Google Cloud認証ドキュメント](https://cloud.google.com/docs/authentication/getting-started)を参照してください。

## インストール

```shell script
# リポジトリのクローン
git clone https://github.com/junara/vtt2mp3.git
cd vtt2mp3

# アプリケーションのビルド
go build -o vtt2mp3 .

# またはインストール
go install
```

### Windowsバイナリのビルド

Windowsで実行可能なバイナリを作成するには、以下の方法があります：

#### 1. Makefileを使用する方法（推奨）

```shell script
# Makefileのbuild-windowsターゲットを使用
make build-windows
```

これにより、`bin/vtt2mp3.exe`というWindowsで実行可能なバイナリが作成されます。

#### 2. 直接コマンドを使用する方法

```shell script
# Windows用にクロスコンパイル
GOOS=windows GOARCH=amd64 go build -o vtt2mp3.exe ./cmd/vtt2mp3
```

#### 3. Windows環境でビルドする方法

Windowsマシン上でGo開発環境を設定している場合は、以下のコマンドでビルドできます：

```cmd
go build -o vtt2mp3.exe ./cmd/vtt2mp3
```

### Linuxバイナリのビルド

Linuxで実行可能なバイナリを作成するには、以下の方法があります：

#### 1. Makefileを使用する方法（推奨）

```shell script
# Makefileのbuild-linuxターゲットを使用
make build-linux
```

これにより、`bin/vtt2mp3-linux`というLinuxで実行可能なバイナリが作成されます。

#### 2. 直接コマンドを使用する方法

```shell script
# Linux用にクロスコンパイル
GOOS=linux GOARCH=amd64 go build -o vtt2mp3-linux ./cmd/vtt2mp3
```

#### 3. Linux環境でビルドする方法

Linuxマシン上でGo開発環境を設定している場合は、以下のコマンドでビルドできます：

```shell script
go build -o vtt2mp3 ./cmd/vtt2mp3
```


## 使用方法

```shell script
# デフォルトオプションでの基本的な使用法
vtt2mp3

# 入力VTTファイルの指定
vtt2mp3 -i path/to/subtitles.vtt

# 出力MP3ファイルの指定
vtt2mp3 -o path/to/output.mp3

# 出力MP4ファイルの指定（動画出力）
vtt2mp3 -o path/to/output.mp4

# 言語コードの指定
vtt2mp3 -l en

# すべてのオプションを組み合わせる（MP3出力）
vtt2mp3 -i path/to/subtitles.vtt -o path/to/output.mp3 -l en

# すべてのオプションを組み合わせる（MP4出力）
vtt2mp3 -i path/to/subtitles.vtt -o path/to/output.mp4 -l en
```


### コマンドラインオプション

- `-i string`: 入力VTTファイル（デフォルト "input.vtt"）
- `-o string`: 出力ファイル（デフォルト "out.mp3"）
  - 拡張子が `.mp3` の場合は音声ファイルを出力
  - 拡張子が `.mp4` の場合は動画ファイル（黒背景に字幕付き）を出力
- `-l string`: 言語コード（デフォルト "ja"）

## 例

```shell script
# サンプルVTTファイルをMP3に変換
vtt2mp3 -i examples/sample50_ja.vtt -o output.mp3 -l ja

# サンプルVTTファイルをMP4動画に変換
vtt2mp3 -i examples/sample50_ja.vtt -o output.mp4 -l ja
```

```shell script
# サンプルVTTファイルをMP3に変換（ドイツ語）
vtt2mp3 -i examples/sample50_de.vtt -o output.mp3 -l de
# サンプルVTTファイルをMP4動画に変換（ドイツ語）
vtt2mp3 -i examples/sample50_de.vtt -o output.mp4 -l de

# サンプルVTTファイルをMP3に変換（英語）
vtt2mp3 -i examples/sample50_en.vtt -o output.mp3 -l en
# サンプルVTTファイルをMP4動画に変換（英語）
vtt2mp3 -i examples/sample50_en.vtt -o output.mp4 -l en

# サンプルVTTファイルをMP3に変換（フランス語）
vtt2mp3 -i examples/sample50_fr.vtt -o output.mp3 -l fr
# サンプルVTTファイルをMP4動画に変換（フランス語）
vtt2mp3 -i examples/sample50_fr.vtt -o output.mp4 -l fr

# サンプルVTTファイルをMP3に変換（韓国語）
vtt2mp3 -i examples/sample50_ko.vtt -o output.mp3 -l ko
# サンプルVTTファイルをMP4動画に変換（韓国語）
vtt2mp3 -i examples/sample50_ko.vtt -o output.mp4 -l ko

# サンプルVTTファイルをMP3に変換（中国語）
vtt2mp3 -i examples/sample50_zh.vtt -o output.mp3 -l zh
# サンプルVTTファイルをMP4動画に変換（中国語）
vtt2mp3 -i examples/sample50_zh.vtt -o output.mp4 -l zh

```

## プロジェクト構造

このプロジェクトはドメイン駆動設計（DDD）の原則に従っています：

- `domain`: コアビジネスロジック
  - `vtt`: VTTファイルの解析と表現
  - `tts`: テキスト読み上げドメインロジック
  - `audio`: 音声ファイルの生成と管理
- `infrastructure`: 外部サービス連携
  - `google`: Google Cloud Text-to-Speech API連携
- `application`: プロセスを調整するアプリケーションサービス
- `presentation`: ユーザーインターフェース（CLI）
- `cmd/vtt2mp3`: アプリケーションのエントリーポイント

## 開発

### lint

このプロジェクトはコード品質チェックに[golangci-lint](https://golangci-lint.run/)を使用しています。

リンターを実行するには：

```shell script
# Makeを使用
make lint

# または直接
golangci-lint run --no-config --timeout=5m ./...
```


可能な場合に問題を自動修正するには：

```shell script
make lint-fix
```


リンターは以下のような一般的な問題をチェックします：
- 未確認のエラー戻り値
- コード簡略化の機会
- 潜在的なバグ
- 非効率なコードパターン

### Makefileコマンド

プロジェクトには以下のコマンドを含むMakefileが含まれています：

```shell script
make all           # lint、build、testを実行
make build         # アプリケーションをビルド
make build-windows # Windows用アプリケーションをビルド（bin/vtt2mp3.exe）
make build-linux   # Linux用アプリケーションをビルド（bin/vtt2mp3-linux）
make test          # テストを実行
make clean         # ビルド成果物をクリーン
make lint          # golangci-lintを実行
make lint-fix      # 自動修正付きでgolangci-lintを実行
make help          # ヘルプメッセージを表示
```


## ライセンス

MIT
