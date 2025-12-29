package handlers

import (
	"fmt"
	"regexp"
	"strings"
)

// ProcessSQL processes the SQL query by handling conditional blocks and parameter replacement.
func ProcessSQL(query string, params map[string]interface{}) string {
	// 1. Handle {{ ... }} blocks
	// Regex matches {{ followed by content, then }}
	// We use a lazy match .*? to handle multiple blocks correctly
	re := regexp.MustCompile(`\{\{(.*?)\}\}`)
	
	query = re.ReplaceAllStringFunc(query, func(match string) string {
		// Remove {{ and }}
		content := match[2 : len(match)-2]
		
		// Split by "else"
		parts := strings.Split(content, "else")
		if len(parts) == 0 {
			return "" 
		}

		mainBlock := strings.TrimSpace(parts[0])
		elseBlock := ""
		if len(parts) > 1 {
			elseBlock = strings.TrimSpace(parts[1])
			// Handle 'quote' removal if present in else block (based on user example '1 = 1')
			// The user example had '1 = 1' (with quotes). We should probably strip outer quotes if they exist,
			// or just leave it as is if it's a valid SQL fragment.
			// User example: {{/*status = :parametro_1*/ else '1 = 1'}}
			// The else block tends to be a string literal in the template syntax, so let's strip single quotes if they surround the content.
			elseBlock = stripQuotes(elseBlock)
		}

		// Check if the main block contains a parameter usage that matches our provided params
		// A parameter in the block looks like :paramName
		// We need to find all :params in the mainBlock
		paramRe := regexp.MustCompile(`:([a-zA-Z0-9_]+)`)
		paramMatches := paramRe.FindAllStringSubmatch(mainBlock, -1)

		hasAllParams := true
		if len(paramMatches) == 0 {
			// No parameters in main block? Then it's just static text? 
			// If it's a comment block like /* ... */, we should strip the comment markers if we are "using" it.
			// The user example: /*status = :parametro_1*/
			// So we need to remove /* and */ if we decide to use the main block.
			hasAllParams = true // If no params, we can use it? Or maybe valid only if it has params?
			// Let's assume if no params are PRESENT in the block, it's valid to use (maybe just uncommenting sql).
		}

		for _, pm := range paramMatches {
			paramName := pm[1]
			if _, exists := params[paramName]; !exists {
				hasAllParams = false
				break
			}
		}

		if hasAllParams && len(paramMatches) > 0 {
			// We use the main block
			// First, remove comment markers /* */ if they exist wrapping the content
			cleaned := strings.TrimSpace(mainBlock)
			if strings.HasPrefix(cleaned, "/*") && strings.HasSuffix(cleaned, "*/") {
				cleaned = cleaned[2 : len(cleaned)-2]
			}
			return strings.TrimSpace(cleaned)
		} else if hasAllParams && len(paramMatches) == 0 {
             // Case with no params, just return cleaned block
			cleaned := strings.TrimSpace(mainBlock)
			if strings.HasPrefix(cleaned, "/*") && strings.HasSuffix(cleaned, "*/") {
				cleaned = cleaned[2 : len(cleaned)-2]
			}
			return strings.TrimSpace(cleaned)
        }

		return elseBlock
	})

	// 2. Replace parameters in the entire string (including those revealed from blocks)
	// We iterate over the params provided
	for key, val := range params {
		valStr := fmt.Sprintf("%v", val)
		// Check if it's a string, if so, wrap in quotes?
		// User said: ped_situacao = "pendente"
		// So if the value is a string, we should wrap it.
		// NOTE: logic needs to be careful not to double quote if the user puts quotes in SQL.
		// But usually :param is a placeholder for a value.
		// For simplicity/safety relative to the prompt "replace da string", we'll wrap strings in double quotes.
		
		switch v := val.(type) {
		case string:
			valStr = fmt.Sprintf("\"%s\"", v)
		}

		query = strings.ReplaceAll(query, ":"+key, valStr)
	}

	return query
}

func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}
