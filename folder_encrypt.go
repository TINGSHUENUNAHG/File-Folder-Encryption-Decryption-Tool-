package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	DefaultKeyFile   = "secret.key"
	AppID            = "FolderEncryptorV2"
	SettingsFileName = "settings.json"
)

/* =========================
   i18n：語言 & 字串表
   ========================= */

type Lang string

const (
	LangZH Lang = "zh"
	LangEN Lang = "en"
	LangJA Lang = "ja"
	LangDE Lang = "de"
)

var currentLang = LangZH

var i18n = map[Lang]map[string]string{
	LangZH: {
		"windowTitle":  "🔐 檔案/資料夾加解密 V1.0",
		"aboutBtn":     "ℹ 關於",
		"aboutTitle":   "關於本程式",
		"aboutContent": "作者: TingShuen Huang(Pecezen.org)\n\n版權宣告:\n本軟體採開源授權，轉載請註明出處。\n\n免責聲明:\n本程式僅供練習及測試使用，作者不對任何數據遺失或後果負責。\n請務必妥善保管您的金鑰或密碼。",

		"langLabel": "介面語言：",
		"langZH":    "繁體中文",
		"langEN":    "English",
		"langJA":    "日本語",
		"langDE":    "Deutsch",

		"selectFolder": "📁 選擇資料夾",
		"selectFile":   "📄 選擇單一檔案",
		"targetLabel":  "目標路徑：",
		"noTarget":     "尚未選擇目標",

		"keyPathLabel":  "金鑰路徑：",
		"passwordLabel": "密碼：",
		"changeKey":     "變更位置",
		"lockKey":       "鎖定位置",

		"modeLabel":     "模式:",
		"algoLabel":     "算法:",
		"modeKeyFile":   "使用金鑰檔案",
		"modePassword":  "使用自訂密碼",
		"algoSecretBox": "SecretBox (快速)",
		"algoAES":       "AES (標準)",

		"encryptBtn": "🔒 開始加密",
		"decryptBtn": "🔓 開始解密",

		"noticeTitle":         "提醒",
		"noticeSelectTarget":  "請先選擇檔案或資料夾",
		"noticeEnterPassword": "請輸入密碼",
		"noticePasswordMode":  "目前為密碼模式，無需選擇金鑰檔。",

		"logNoFiles": "ℹ 沒有可處理的檔案。",
		"logFailFmt": "❌ %s 失敗: %v",
		"logDoneFmt": "✅ %s 完成",
		"logAllDone": "✔ 全部處理完成！",

		"errNotEncrypted":   "非加密檔案",
		"errFileTooShort":   "檔案長度不足",
		"errDecryptFailed":  "解密失敗 (可能是密碼/金鑰錯誤)",
		"errTargetNotFound": "錯誤：找不到指定的目標路徑，請重新選擇。",

		"logEncryptFailedFmt": "❌ 加密失敗: %v",
		"logEncryptDoneFmt":   "✅ 加密完成: %s.enc",
		"logDecryptFailedFmt": "❌ 解密失敗: %v",
		"logDecryptDoneFmt":   "✅ 解密完成: %s",
	},
	LangEN: {
		"windowTitle":  "🔐 File/Folder Encryptor V1.0",
		"aboutBtn":     "ℹ About",
		"aboutTitle":   "About this App",
		"aboutContent": "Author: TingShuen Huang(Pecezen.org)\n\nCopyright:\nThis software is open-source. Please credit the author when sharing.\n\nDisclaimer:\nThis software is for practice/testing only. The author is not responsible for any data loss. \nPlease keep your keys and passwords safe.",

		"langLabel": "Language:",
		"langZH":    "繁體中文",
		"langEN":    "English",
		"langJA":    "日本語",
		"langDE":    "Deutsch",

		"selectFolder": "📁 Select Folder",
		"selectFile":   "📄 Select File",
		"targetLabel":  "Target Path:",
		"noTarget":     "No target selected",

		"keyPathLabel":  "Key Path:",
		"passwordLabel": "Password:",
		"changeKey":     "Change",
		"lockKey":       "Lock",

		"modeLabel":     "Mode:",
		"algoLabel":     "Algo:",
		"modeKeyFile":   "Key File",
		"modePassword":  "Password",
		"algoSecretBox": "SecretBox",
		"algoAES":       "AES",

		"encryptBtn": "🔒 Encrypt",
		"decryptBtn": "🔓 Decrypt",

		"noticeTitle":         "Notice",
		"noticeSelectTarget":  "Please select a file or folder first.",
		"noticeEnterPassword": "Please enter a password.",
		"noticePasswordMode":  "Currently in password mode; no key file is needed.",

		"logNoFiles": "ℹ No files to process.",
		"logFailFmt": "❌ %s failed: %v",
		"logDoneFmt": "✅ %s done",
		"logAllDone": "✔ All tasks completed!",

		"errNotEncrypted":   "not an encrypted file",
		"errFileTooShort":   "file length too short",
		"errDecryptFailed":  "decryption failed",
		"errTargetNotFound": "Error: Target path not found. Please select again.",

		"logEncryptFailedFmt": "❌ encryption failed: %v",
		"logEncryptDoneFmt":   "✅ encryption done: %s.enc",
		"logDecryptFailedFmt": "❌ decryption failed: %v",
		"logDecryptDoneFmt":   "✅ decryption done: %s",
	},
	LangJA: {
		"windowTitle":  "🔐 ファイル/フォルダの暗号化 V1.0",
		"aboutBtn":     "ℹ 情報",
		"aboutTitle":   "このプログラムについて",
		"aboutContent": "著者: TingShuen Huang(Pecezen.org)\n\n著作権:\nこのソフトウェアはオープンソースです。轉載時は著者を明記してください。\n\n免責事項:\n本ソフトは学習目的のみ。データの損失等について著者は一切の責任を負いません。",

		"langLabel": "表示言語：",
		"langZH":    "繁體中文",
		"langEN":    "English",
		"langJA":    "日本語",
		"langDE":    "Deutsch",

		"selectFolder": "📁 フォルダを選択",
		"selectFile":   "📄 ファイルを選択",
		"targetLabel":  "対象パス：",
		"noTarget":     "未選択",

		"keyPathLabel":  "鍵パス：",
		"passwordLabel": "パス：",
		"changeKey":     "変更",
		"lockKey":       "ロック",

		"modeLabel":     "モード:",
		"algoLabel":     "アルゴリズム:",
		"modeKeyFile":   "鍵ファイル",
		"modePassword":  "パスワード",
		"algoSecretBox": "SecretBox",
		"algoAES":       "AES",

		"encryptBtn": "🔒 暗号化",
		"decryptBtn": "🔓 復号",

		"noticeTitle":         "通知",
		"noticeSelectTarget":  "先にファイルかフォルダを選択してください。",
		"noticeEnterPassword": "パスワードを入力してください。",
		"noticePasswordMode":  "パスワードモードです。鍵ファイルは不要です。",

		"logNoFiles": "ℹ 対象ファイルがありません。",
		"logFailFmt": "❌ %s 失敗: %v",
		"logDoneFmt": "✅ %s 完了",
		"logAllDone": "✔ すべての處理が完了しました！",

		"errNotEncrypted":   "暗号化ファイルではありません",
		"errFileTooShort":   "ファイル長が不足しています",
		"errDecryptFailed":  "復號に失敗しました",
		"errTargetNotFound": "エラー：対象パスが見つかりません。再選擇してください。",

		"logEncryptFailedFmt": "❌ 暗號化失敗: %v",
		"logEncryptDoneFmt":   "✅ 暗號化完了: %s.enc",
		"logDecryptFailedFmt": "❌ 復號失敗: %v",
		"logDecryptDoneFmt":   "✅ 復號完了: %s",
	},
	LangDE: {
		"windowTitle":  "🔐 Datei/Ordner-Verschlüsselung V1.0",
		"aboutBtn":     "ℹ Über",
		"aboutTitle":   "Über diese App",
		"aboutContent": "Autor: TingShuen Huang(Pecezen.org)\n\nCopyright:\nDiese Software ist Open-Source. Bitte geben Sie den Autor an.\n\nHaftungsausschluss:\nDiese Software dient nur zu Übungs-/Testzwecken. Der Autor haftet nicht für Datenverlust.\nBitte bewahren Sie Ihre Schlüssel und Passwörter sicher auf.",

		"langLabel": "Sprache:",
		"langZH":    "繁體中文",
		"langEN":    "English",
		"langJA":    "日本語",
		"langDE":    "Deutsch",

		"selectFolder": "📁 Ordner wählen",
		"selectFile":   "📄 Datei wählen",
		"targetLabel":  "Zielpfad:",
		"noTarget":     "Kein Ziel ausgewählt",

		"keyPathLabel":  "Schlüsselpfad:",
		"passwordLabel": "Passwort:",
		"changeKey":     "Ändern",
		"lockKey":       "Sperren",

		"modeLabel":     "Modus:",
		"algoLabel":     "Algo:",
		"modeKeyFile":   "Schlüsseldatei",
		"modePassword":  "Passwort",
		"algoSecretBox": "SecretBox",
		"algoAES":       "AES",

		"encryptBtn": "🔒 Verschlüsseln",
		"decryptBtn": "🔓 Entschlüsseln",

		"noticeTitle":         "Hinweis",
		"noticeSelectTarget":  "Bitte wählen Sie zuerst eine Datei oder einen Ordner aus.",
		"noticeEnterPassword": "Bitte geben Sie ein Passwort ein.",
		"noticePasswordMode":  "Aktuell im Passwort-Modus; keine Schlüsseldatei erforderlich.",

		"logNoFiles": "ℹ Keine Dateien zu verarbeiten.",
		"logFailFmt": "❌ %s fehlgeschlagen: %v",
		"logDoneFmt": "✅ %s erledigt",
		"logAllDone": "✔ Alle Aufgaben abgeschlossen!",

		"errNotEncrypted":   "keine verschlüsselte Datei",
		"errFileTooShort":   "Dateilänge zu kurz",
		"errDecryptFailed":  "Entschlüsselung fehlgeschlagen",
		"errTargetNotFound": "Fehler: Zielpfad nicht gefunden. Bitte erneut wählen.",

		"logEncryptFailedFmt": "❌ Verschlüsselung fehlgeschlagen: %v",
		"logEncryptDoneFmt":   "✅ Verschlüsselung abgeschlossen: %s.enc",
		"logDecryptFailedFmt": "❌ Entschlüsselung fehlgeschlagen: %v",
		"logDecryptDoneFmt":   "✅ Entschlüsselung abgeschlossen: %s",
	},
}

func tr(key string) string {
	if s, ok := i18n[currentLang][key]; ok {
		return s
	}
	if s, ok := i18n[LangEN][key]; ok {
		return s
	}
	return key
}

/* =========================
   設定檔與主題
   ========================= */

type Settings struct {
	Lang string `json:"lang"`
}

func settingsFilePath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, AppID, SettingsFileName), nil
}

func loadLanguageFromSettings() Lang {
	p, err := settingsFilePath()
	if err != nil {
		return LangZH
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return LangZH
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return LangZH
	}
	switch strings.ToLower(s.Lang) {
	case "en":
		return LangEN
	case "ja":
		return LangJA
	case "de":
		return LangDE
	default:
		return LangZH
	}
}

func saveLanguageToSettings(lang Lang) {
	p, err := settingsFilePath()
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(p), 0o700)
	s := Settings{Lang: string(lang)}
	b, _ := json.MarshalIndent(s, "", "  ")
	_ = os.WriteFile(p, b, 0o600)
}

type darkTextTheme struct{ fyne.Theme }

func (t darkTextTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameDisabled {
		return color.NRGBA{R: 100, G: 100, B: 100, A: 255}
	}
	if name == theme.ColorNameForeground {
		return color.NRGBA{R: 30, G: 30, B: 30, A: 255}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func ui(f func()) {
	drv := fyne.CurrentApp().Driver()
	if drv != nil {
		type caller interface{ CallOnMain(func()) }
		if c, ok := interface{}(drv).(caller); ok {
			c.CallOnMain(f)
			return
		}
	}
	f()
}

/* =========================
   Crypto & File Logic
   ========================= */

func deriveKeyFromPassword(password string) *[32]byte {
	hash := sha256.Sum256([]byte(password))
	var key [32]byte
	copy(key[:], hash[:])
	return &key
}

func loadOrCreateKey(path string) *[32]byte {
	var key [32]byte
	if _, err := os.Stat(path); err == nil {
		data, _ := os.ReadFile(path)
		if len(data) == 32 {
			copy(key[:], data)
		}
	} else {
		_, _ = rand.Read(key[:])
		_ = os.MkdirAll(filepath.Dir(path), 0o700)
		_ = os.WriteFile(path, key[:], 0o600)
	}
	return &key
}

func resolveDefaultKeyPath(target string, hasTarget bool) string {
	if !hasTarget || target == "" {
		return DefaultKeyFile
	}
	fi, err := os.Stat(target)
	dir := target
	if err != nil || !fi.IsDir() {
		dir = filepath.Dir(target)
	}
	return filepath.Join(dir, DefaultKeyFile)
}

func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, "tmp_enc_*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return err
	}
	_ = tmpFile.Sync()
	_ = tmpFile.Close()
	_ = os.Chmod(tmpPath, perm)
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func removeWithRetry(path string) error {
	_ = os.Chmod(path, 0o666)
	var lastErr error
	for i := 0; i < 3; i++ {
		if err := os.Remove(path); err == nil {
			return nil
		} else {
			lastErr = err
			time.Sleep(200 * time.Millisecond)
		}
	}
	return lastErr
}

func encryptFileSecretbox(filePath string, key *[32]byte) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	var nonce [24]byte
	_, _ = rand.Read(nonce[:])
	enc := secretbox.Seal(nonce[:], data, &nonce, key)
	if err := atomicWriteFile(filePath+".enc", enc, 0o644); err != nil {
		return err
	}
	return removeWithRetry(filePath)
}

func decryptFileSecretbox(encFilePath string, key *[32]byte) error {
	if !strings.HasSuffix(encFilePath, ".enc") {
		return fmt.Errorf(tr("errNotEncrypted"))
	}
	data, err := os.ReadFile(encFilePath)
	if err != nil {
		return err
	}
	if len(data) < 24 {
		return fmt.Errorf(tr("errFileTooShort"))
	}
	var nonce [24]byte
	copy(nonce[:], data[:24])
	decrypted, ok := secretbox.Open(nil, data[24:], &nonce, key)
	if !ok {
		return fmt.Errorf(tr("errDecryptFailed"))
	}
	if err := atomicWriteFile(strings.TrimSuffix(encFilePath, ".enc"), decrypted, 0o644); err != nil {
		return err
	}
	return removeWithRetry(encFilePath)
}

func encryptFileAES(filePath string, key *[32]byte) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	block, _ := aes.NewCipher(key[:])
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	_, _ = rand.Read(nonce)
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	if err := atomicWriteFile(filePath+".enc", ciphertext, 0o644); err != nil {
		return err
	}
	return removeWithRetry(filePath)
}

func decryptFileAES(encFilePath string, key *[32]byte) error {
	if !strings.HasSuffix(encFilePath, ".enc") {
		return fmt.Errorf(tr("errNotEncrypted"))
	}
	data, err := os.ReadFile(encFilePath)
	if err != nil {
		return err
	}
	block, _ := aes.NewCipher(key[:])
	gcm, _ := cipher.NewGCM(block)
	nz := gcm.NonceSize()
	if len(data) < nz {
		return fmt.Errorf(tr("errFileTooShort"))
	}
	nonce, ciphertext := data[:nz], data[nz:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf(tr("errDecryptFailed"))
	}
	if err := atomicWriteFile(strings.TrimSuffix(encFilePath, ".enc"), plaintext, 0o644); err != nil {
		return err
	}
	return removeWithRetry(encFilePath)
}

func processFolder(folderPath string, key *[32]byte, action string, useAES bool, progress *widget.ProgressBar, log *widget.Entry) {
	var files []string
	_ = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			if action == "encrypt" && !strings.HasSuffix(path, ".enc") && filepath.Base(path) != DefaultKeyFile {
				files = append(files, path)
			} else if action == "decrypt" && strings.HasSuffix(path, ".enc") {
				files = append(files, path)
			}
		}
		return nil
	})

	total := float64(len(files))
	if total == 0 {
		ui(func() { log.SetText(log.Text + tr("logNoFiles") + "\n"); progress.SetValue(1) })
		return
	}

	var wg sync.WaitGroup
	for i, f := range files {
		wg.Add(1)
		go func(idx int, path string) {
			defer wg.Done()
			var err error
			if action == "encrypt" {
				if useAES {
					err = encryptFileAES(path, key)
				} else {
					err = encryptFileSecretbox(path, key)
				}
			} else {
				if useAES {
					err = decryptFileAES(path, key)
				} else {
					err = decryptFileSecretbox(path, key)
				}
			}
			ui(func() {
				if err != nil {
					log.SetText(log.Text + fmt.Sprintf(tr("logFailFmt")+"\n", path, err))
				} else {
					log.SetText(log.Text + fmt.Sprintf(tr("logDoneFmt")+"\n", path))
				}
				progress.SetValue(float64(idx+1) / total)
			})
		}(i, f)
		time.Sleep(20 * time.Millisecond)
	}
	wg.Wait()
	ui(func() { log.SetText(log.Text + tr("logAllDone") + "\n"); progress.SetValue(1) })
}

/* =========================
   UI References & Apply
   ========================= */

type UIRefs struct {
	w                 fyne.Window
	aboutBtn          *widget.Button
	selectedPathLabel *widget.Label
	selectedPath      *widget.Label
	selectFolderBtn   *widget.Button
	selectFileBtn     *widget.Button
	keyPathLabel      *widget.Label
	passwordLabel     *widget.Label
	modeLabel         *widget.Label
	algoLabel         *widget.Label
	langLabel         *widget.Label
	btnChangeKey      *widget.Button
	modeOptions       *widget.RadioGroup
	algoOptions       *widget.RadioGroup
	langSwitch        *widget.RadioGroup
	encryptBtn        *widget.Button
	decryptBtn        *widget.Button
	keyPath           *widget.Entry
	passwordEntry     *widget.Entry
	logArea           *widget.Entry
	progress          *widget.ProgressBar
}

func applyLanguage(u *UIRefs, locked *bool, hasTarget *bool, isPasswordMode bool, isAES bool) {
	ui(func() {
		u.w.SetTitle(tr("windowTitle"))
		u.aboutBtn.SetText(tr("aboutBtn"))
		u.selectedPathLabel.SetText(tr("targetLabel"))
		if !*hasTarget {
			u.selectedPath.SetText(tr("noTarget"))
		}
		u.selectFolderBtn.SetText(tr("selectFolder"))
		u.selectFileBtn.SetText(tr("selectFile"))
		u.keyPathLabel.SetText(tr("keyPathLabel"))
		u.passwordLabel.SetText(tr("passwordLabel"))
		u.modeLabel.SetText(tr("modeLabel"))
		u.algoLabel.SetText(tr("algoLabel"))
		u.langLabel.SetText(tr("langLabel"))

		if *locked {
			u.btnChangeKey.SetText(tr("changeKey"))
		} else {
			u.btnChangeKey.SetText(tr("lockKey"))
		}

		// 更新選項並保持目前的預選狀態
		u.modeOptions.Options = []string{tr("modeKeyFile"), tr("modePassword")}
		if isPasswordMode {
			u.modeOptions.SetSelected(tr("modePassword"))
		} else {
			u.modeOptions.SetSelected(tr("modeKeyFile"))
		}
		u.modeOptions.Refresh()

		u.algoOptions.Options = []string{tr("algoSecretBox"), tr("algoAES")}
		if isAES {
			u.algoOptions.SetSelected(tr("algoAES"))
		} else {
			u.algoOptions.SetSelected(tr("algoSecretBox"))
		}
		u.algoOptions.Refresh()

		u.langSwitch.Options = []string{tr("langZH"), tr("langEN"), tr("langJA"), tr("langDE")}
		u.langSwitch.Refresh()

		u.encryptBtn.SetText(tr("encryptBtn"))
		u.decryptBtn.SetText(tr("decryptBtn"))
	})
}

/* =========================
   Main
   ========================= */

func main() {
	a := app.New()
	a.Settings().SetTheme(darkTextTheme{theme.DefaultTheme()})

	// --- Icon 修復區 ---
	iconPath := "icon.png"
	if _, err := os.Stat(iconPath); err == nil {
		if iconRes, err := fyne.LoadResourceFromPath(iconPath); err == nil {
			a.SetIcon(iconRes)
		} else {
			fmt.Printf("警告: 無法讀取圖示資源: %v\n", err)
		}
	} else {
		fmt.Println("提示: 找不到 icon.png 檔案，將使用預設圖示。")
	}

	currentLang = loadLanguageFromSettings()

	var (
		usePassword       = false
		useAES            = false
		selectedHasTarget = false
		keyLocked         = true
	)

	w := a.NewWindow(tr("windowTitle"))
	if a.Icon() != nil {
		w.SetIcon(a.Icon())
	}

	aboutBtn := widget.NewButtonWithIcon(tr("aboutBtn"), theme.InfoIcon(), func() {
		dialog.ShowInformation(tr("aboutTitle"), tr("aboutContent"), w)
	})

	logArea := widget.NewMultiLineEntry()
	logArea.SetMinRowsVisible(10)
	progress := widget.NewProgressBar()

	selectedPath := widget.NewLabelWithStyle(tr("noTarget"), fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	selectedPathLabel := widget.NewLabel(tr("targetLabel"))

	keyPath := widget.NewEntry()
	keyPath.SetText(DefaultKeyFile)
	keyPath.Disable()

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.Disable()

	btnChangeKey := widget.NewButton(tr("changeKey"), nil)
	modeLabel := widget.NewLabel(tr("modeLabel"))
	algoLabel := widget.NewLabel(tr("algoLabel"))
	langLabel := widget.NewLabel(tr("langLabel"))

	modeOptions := widget.NewRadioGroup([]string{tr("modeKeyFile"), tr("modePassword")}, func(value string) {
		if strings.Contains(value, "密碼") || strings.Contains(value, "Password") || strings.Contains(value, "パスワード") || strings.Contains(value, "Passwort") {
			usePassword = true
			passwordEntry.Enable()
			keyPath.Disable()
			keyLocked = true
			btnChangeKey.Disable()
		} else {
			usePassword = false
			passwordEntry.Disable()
			keyPath.Disable()
			keyLocked = true
			btnChangeKey.Enable()
			keyPath.SetText(resolveDefaultKeyPath(selectedPath.Text, selectedHasTarget))
		}
	})
	modeOptions.Horizontal = true

	algoOptions := widget.NewRadioGroup([]string{tr("algoSecretBox"), tr("algoAES")}, func(value string) {
		useAES = strings.Contains(value, "AES")
	})
	algoOptions.Horizontal = true

	// ==========================================
	// 設定預選項目 (要在 RadioGroup 定義之後)
	// ==========================================
	modeOptions.SetSelected(tr("modeKeyFile"))
	algoOptions.SetSelected(tr("algoSecretBox"))

	btnChangeKey.OnTapped = func() {
		if usePassword {
			dialog.ShowInformation(tr("noticeTitle"), tr("noticePasswordMode"), w)
			return
		}
		if keyLocked {
			keyPath.Enable()
			keyLocked = false
			btnChangeKey.SetText(tr("lockKey"))
			dialog.ShowFileSave(func(uc fyne.URIWriteCloser, err error) {
				if uc != nil {
					p := uc.URI().Path()
					_ = uc.Close()
					ui(func() { keyPath.SetText(p) })
				}
			}, w)
		} else {
			keyPath.Disable()
			keyLocked = true
			btnChangeKey.SetText(tr("changeKey"))
		}
	}

	selectFolderBtn := widget.NewButtonWithIcon(tr("selectFolder"), theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if uri != nil {
				p := uri.Path()
				ui(func() {
					selectedPath.SetText(p)
					selectedHasTarget = true
					keyPath.SetText(resolveDefaultKeyPath(p, true))
				})
			}
		}, w)
	})

	selectFileBtn := widget.NewButtonWithIcon(tr("selectFile"), theme.FileIcon(), func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader != nil {
				p := reader.URI().Path()
				_ = reader.Close()
				ui(func() {
					selectedPath.SetText(p)
					selectedHasTarget = true
					keyPath.SetText(resolveDefaultKeyPath(p, true))
				})
			}
		}, w)
	})

	encryptBtn := widget.NewButtonWithIcon(tr("encryptBtn"), theme.ConfirmIcon(), nil)
	decryptBtn := widget.NewButtonWithIcon(tr("decryptBtn"), theme.ViewRefreshIcon(), nil)

	u := &UIRefs{
		w: w, aboutBtn: aboutBtn,
		selectedPathLabel: selectedPathLabel, selectedPath: selectedPath,
		selectFolderBtn: selectFolderBtn, selectFileBtn: selectFileBtn,
		keyPathLabel: widget.NewLabel(tr("keyPathLabel")), passwordLabel: widget.NewLabel(tr("passwordLabel")),
		modeLabel: modeLabel, algoLabel: algoLabel, langLabel: langLabel,
		btnChangeKey: btnChangeKey, modeOptions: modeOptions, algoOptions: algoOptions,
		encryptBtn: encryptBtn, decryptBtn: decryptBtn,
		keyPath: keyPath, passwordEntry: passwordEntry, logArea: logArea, progress: progress,
	}

	langSwitch := widget.NewRadioGroup([]string{tr("langZH"), tr("langEN"), tr("langJA"), tr("langDE")}, func(v string) {
		if v == tr("langEN") || v == "English" {
			currentLang = LangEN
		} else if v == tr("langJA") || v == "日本語" {
			currentLang = LangJA
		} else if v == tr("langDE") || v == "Deutsch" {
			currentLang = LangDE
		} else {
			currentLang = LangZH
		}
		// 切換語言時，傳入目前的選擇狀態以維持預選
		applyLanguage(u, &keyLocked, &selectedHasTarget, usePassword, useAES)
		saveLanguageToSettings(currentLang)
	})
	langSwitch.Horizontal = true
	u.langSwitch = langSwitch

	if currentLang == LangEN {
		langSwitch.SetSelected("English")
	} else if currentLang == LangJA {
		langSwitch.SetSelected("日本語")
	} else if currentLang == LangDE {
		langSwitch.SetSelected("Deutsch")
	} else {
		langSwitch.SetSelected("繁體中文")
	}

	encryptBtn.OnTapped = func() {
		target := selectedPath.Text
		if !selectedHasTarget || target == "" {
			dialog.ShowInformation(tr("noticeTitle"), tr("noticeSelectTarget"), w)
			return
		}
		fi, err := os.Stat(target)
		if os.IsNotExist(err) {
			dialog.ShowError(fmt.Errorf(tr("errTargetNotFound")), w)
			return
		}
		var k *[32]byte
		if usePassword {
			if passwordEntry.Text == "" {
				dialog.ShowInformation(tr("noticeTitle"), tr("noticeEnterPassword"), w)
				return
			}
			k = deriveKeyFromPassword(passwordEntry.Text)
		} else {
			k = loadOrCreateKey(keyPath.Text)
		}
		go func() {
			ui(func() { progress.SetValue(0) })
			if fi.IsDir() {
				processFolder(target, k, "encrypt", useAES, progress, logArea)
			} else {
				var err error
				if useAES {
					err = encryptFileAES(target, k)
				} else {
					err = encryptFileSecretbox(target, k)
				}
				ui(func() {
					if err != nil {
						logArea.SetText(logArea.Text + fmt.Sprintf(tr("logEncryptFailedFmt")+"\n", err))
					} else {
						logArea.SetText(logArea.Text + fmt.Sprintf(tr("logEncryptDoneFmt")+"\n", target))
					}
					progress.SetValue(1)
				})
			}
		}()
	}

	decryptBtn.OnTapped = func() {
		target := selectedPath.Text
		if !selectedHasTarget || target == "" {
			dialog.ShowInformation(tr("noticeTitle"), tr("noticeSelectTarget"), w)
			return
		}
		fi, err := os.Stat(target)
		if os.IsNotExist(err) {
			dialog.ShowError(fmt.Errorf(tr("errTargetNotFound")), w)
			return
		}
		var k *[32]byte
		if usePassword {
			if passwordEntry.Text == "" {
				dialog.ShowInformation(tr("noticeTitle"), tr("noticeEnterPassword"), w)
				return
			}
			k = deriveKeyFromPassword(passwordEntry.Text)
		} else {
			k = loadOrCreateKey(keyPath.Text)
		}
		go func() {
			ui(func() { progress.SetValue(0) })
			if fi.IsDir() {
				processFolder(target, k, "decrypt", useAES, progress, logArea)
			} else {
				var err error
				if useAES {
					err = decryptFileAES(target, k)
				} else {
					err = decryptFileSecretbox(target, k)
				}
				ui(func() {
					if err != nil {
						logArea.SetText(logArea.Text + fmt.Sprintf(tr("logDecryptFailedFmt")+"\n", err))
					} else {
						logArea.SetText(logArea.Text + fmt.Sprintf(tr("logDecryptDoneFmt")+"\n", strings.TrimSuffix(target, ".enc")))
					}
					progress.SetValue(1)
				})
			}
		}()
	}

	topBar := container.NewBorder(nil, nil, langLabel, aboutBtn, langSwitch)
	targetBox := container.NewVBox(
		container.NewHBox(selectedPathLabel, selectedPath),
		container.NewHBox(selectFolderBtn, selectFileBtn),
	)
	cryptoBox := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(modeLabel, modeOptions),
		container.NewHBox(algoLabel, algoOptions),
		container.NewBorder(nil, nil, u.keyPathLabel, btnChangeKey, keyPath),
		container.NewBorder(nil, nil, u.passwordLabel, nil, passwordEntry),
		container.NewHBox(layout.NewSpacer(), encryptBtn, decryptBtn, layout.NewSpacer()),
	)

	content := container.NewBorder(
		container.NewVBox(topBar, widget.NewSeparator(), targetBox),
		nil, nil, nil,
		container.NewVBox(cryptoBox, progress, logArea),
	)

	w.SetContent(container.NewPadded(content))
	w.Resize(fyne.NewSize(850, 650))
	w.ShowAndRun()
}
