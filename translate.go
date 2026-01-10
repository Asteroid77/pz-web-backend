package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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

// 单个语言的翻译包
type TranslationMap map[string]string

var (
	// 缓存池: langCode -> TranslationMap
	// 例如: "CN" -> {...}, "EN" -> {...}
	translationCache = make(map[string]TranslationMap)
	cacheMutex       sync.RWMutex
)

// GetTranslationMap 获取指定语言的翻译字典
// 如果缓存里没有，就去磁盘加载
func GetTranslationMap(lang string) TranslationMap {
	cacheMutex.RLock()
	if t, ok := translationCache[lang]; ok {
		cacheMutex.RUnlock()
		return t
	}
	cacheMutex.RUnlock()

	// 加载并缓存
	return loadLanguage(lang)
}

func GetSandboxOptions(dict TranslationMap, key string) []ConfigOption {
	var options []ConfigOption
	prefix := "Sandbox_"

	// 僵毁的 option 通常是从 1 开始的整数
	for i := 1; ; i++ {
		// 构造 Key: Sandbox_DayLength_option1
		optionKey := fmt.Sprintf("%s%s_option%d", prefix, key, i)

		if label, ok := dict[optionKey]; ok {
			options = append(options, ConfigOption{
				// 注意：僵毁的 Lua 值通常就是这个 index (1, 2, 3...)
				// 但前端传回来的要是字符串
				Value: strconv.Itoa(i),
				Label: label,
			})
		} else {
			// 如果找不到 optionN，说明选项列表结束了
			break
		}
	}

	return options
}

func loadLanguage(lang string) TranslationMap {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// 双重检查
	if t, ok := translationCache[lang]; ok {
		return t
	}

	t := make(TranslationMap)
	fmt.Printf("Loading translations for language: %s\n", lang)

	// 需要加载的文件前缀
	filePrefixes := []string{"Sandbox", "UI", "Tooltip"}

	for _, prefix := range filePrefixes {
		// 构造文件名: Sandbox_CN.txt
		filename := fmt.Sprintf("%s_%s.txt", prefix, lang)
		path := filepath.Join(BaseGameDir, "lua/shared/Translate", lang, filename)

		// 如果指定语言文件不存在（比如找不到 CN），尝试回退到 EN
		if _, err := os.Stat(path); os.IsNotExist(err) && lang != "EN" {
			// fallback path
			path = filepath.Join(BaseGameDir, "lua/shared/Translate", "EN", fmt.Sprintf("%s_EN.txt", prefix))
		}

		// 依然调用原签名的 loadFile
		loadFile(path, t)
	}

	translationCache[lang] = t
	return t
}

// 根据 Key 从指定字典里查翻译
func TranslateKey(t TranslationMap, key string, contextPrefix string) (label, tooltip string) {
	fullKey := contextPrefix + key

	// 1. 找 Label
	if v, ok := t[fullKey]; ok {
		label = v
	} else if v, ok := t[key]; ok {
		label = v
	} else {
		label = key // 没翻译就用 key
	}

	// 2. 找 Tooltip
	if v, ok := t[fullKey+"_tooltip"]; ok {
		tooltip = v
	}

	return label, tooltip
}

// loadFile 读取文件并将内容存入 map
func loadFile(path string, targetMap TranslationMap) error {
	// 读取全部内容
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	finalStr := ""

	// 检测编码：如果是 UTF-8，直接用
	if isUTF8(content) {
		finalStr = string(content)
	} else {
		// 如果不是 UTF-8，根据文件名推断语言代码，然后转码
		// 例如: ".../Sandbox_CN.txt" -> "CN"
		fileName := filepath.Base(path)                                       // Sandbox_CN.txt
		fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName)) // Sandbox_CN
		parts := strings.Split(fileNameNoExt, "_")

		langCode := "EN" // 默认 fallback
		if len(parts) > 1 {
			langCode = parts[len(parts)-1] // 取最后一段作为语言代码
		}

		// 获取对应的解码器 (GBK, Windows-1250, etc)
		decoder := getEncoding(langCode).NewDecoder()
		decodedBytes, _, err := transform.Bytes(decoder, content)
		if err != nil {
			finalStr = string(content) // 转码失败，使用原文
		} else {
			finalStr = string(decodedBytes)
		}
	}

	// 解析内容
	// 正则匹配: Some_Key = "Some Value",
	re := regexp.MustCompile(`^\s*([A-Za-z0-9_]+)\s*=\s*"(.*)",?.*`)

	scanner := bufio.NewScanner(strings.NewReader(finalStr))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") || line == "{" || line == "}" {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			val := matches[2]
			// 处理转义字符
			val = strings.ReplaceAll(val, `\"`, `"`)
			val = strings.ReplaceAll(val, `<br>`, "\n")
			targetMap[key] = val
		}
	}
	return nil
}

// isUTF8 检测字节流是否为有效的 UTF-8
func isUTF8(content []byte) bool {
	// 跳过 BOM (如果存在)
	bytesToCheck := content
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		bytesToCheck = content[3:]
	}
	return utf8.Valid(bytesToCheck)
}

// getEncoding 根据 PZ 的语言代码返回对应的字符集编码
func getEncoding(lang string) encoding.Encoding {
	switch strings.ToUpper(lang) {
	case "CN":
		// 使用 GB18030 以支持更多汉字
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
