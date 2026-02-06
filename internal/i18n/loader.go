package i18n

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// TranslationMap 单个语言的翻译包。
type TranslationMap map[string]string

type Loader struct {
	baseGameDir string

	mu    sync.RWMutex
	cache map[string]TranslationMap
}

func NewLoader(baseGameDir string) *Loader {
	return &Loader{
		baseGameDir: baseGameDir,
		cache:       make(map[string]TranslationMap),
	}
}

// GetTranslationMap 获取指定语言的翻译字典（带缓存）。
func (l *Loader) GetTranslationMap(lang string) TranslationMap {
	l.mu.RLock()
	if t, ok := l.cache[lang]; ok {
		l.mu.RUnlock()
		return t
	}
	l.mu.RUnlock()

	return l.loadLanguage(lang)
}

func (l *Loader) loadLanguage(lang string) TranslationMap {
	l.mu.Lock()
	defer l.mu.Unlock()

	if t, ok := l.cache[lang]; ok {
		return t
	}

	t := make(TranslationMap)

	filePrefixes := []string{"Sandbox", "UI", "Tooltip"}
	for _, prefix := range filePrefixes {
		filename := fmt.Sprintf("%s_%s.txt", prefix, lang)
		path := filepath.Join(l.baseGameDir, "lua/shared/Translate", lang, filename)

		if _, err := os.Stat(path); os.IsNotExist(err) && lang != "EN" {
			path = filepath.Join(l.baseGameDir, "lua/shared/Translate", "EN", fmt.Sprintf("%s_EN.txt", prefix))
		}

		_ = loadFile(path, t)
	}

	l.cache[lang] = t
	return t
}

// TranslateKey 根据 Key 从指定字典里查翻译。
func TranslateKey(t TranslationMap, key string, contextPrefix string) (label, tooltip string) {
	fullKey := contextPrefix + key

	if v, ok := t[fullKey]; ok {
		label = v
	} else if v, ok := t[key]; ok {
		label = v
	} else {
		label = key
	}

	if v, ok := t[fullKey+"_tooltip"]; ok {
		tooltip = v
	}

	return label, tooltip
}

func loadFile(path string, targetMap TranslationMap) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	finalStr := ""
	if isUTF8(content) {
		finalStr = string(content)
	} else {
		fileName := filepath.Base(path)
		fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		parts := strings.Split(fileNameNoExt, "_")

		langCode := "EN"
		if len(parts) > 1 {
			langCode = parts[len(parts)-1]
		}

		decoder := getEncoding(langCode).NewDecoder()
		decodedBytes, _, decodeErr := transform.Bytes(decoder, content)
		if decodeErr != nil {
			finalStr = string(content)
		} else {
			finalStr = string(decodedBytes)
		}
	}

	re := regexp.MustCompile(`^\s*([A-Za-z0-9_]+)\s*=\s*"(.*)",?.*`)
	scanner := bufio.NewScanner(strings.NewReader(finalStr))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") || line == "{" || line == "}" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			val := matches[2]
			val = strings.ReplaceAll(val, `\"`, `"`)
			val = strings.ReplaceAll(val, `<br>`, "\n")
			targetMap[key] = val
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func isUTF8(content []byte) bool {
	bytesToCheck := content
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		bytesToCheck = content[3:]
	}
	return utf8.Valid(bytesToCheck)
}

func getEncoding(lang string) encoding.Encoding {
	switch strings.ToUpper(lang) {
	case "CN":
		return simplifiedchinese.GB18030
	case "TW":
		return traditionalchinese.Big5
	case "JP":
		return japanese.ShiftJIS
	case "KO":
		return korean.EUCKR
	case "RU", "UA":
		return charmap.Windows1251
	case "CS", "PL", "HU", "RO", "CH":
		return charmap.Windows1250
	case "TR":
		return charmap.Windows1254
	case "AR":
		return charmap.Windows1256
	case "TH":
		return charmap.Windows874
	default:
		return charmap.Windows1252
	}
}
