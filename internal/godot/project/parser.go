package config

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// Project represents a Godot project configuration.
type ProjectParser struct{}

// Unmarshal parses the Godot project configuration data.
//
//nolint:cyclop,funlen
func (p ProjectParser) Unmarshal(data []byte) (map[string]interface{}, error) {
	packedStringArrayPattern := regexp.MustCompile(`PackedStringArray\((.*)\)`)

	scanner := bufio.NewScanner(bytes.NewReader(data))

	var section string

	result := make(map[string]interface{})

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if len(line) == 0 || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]

			continue
		}

		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 { //nolint:mnd
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if section != "" {
				key = section + "." + key
			}

			// Handle PackedStringArray
			if matches := packedStringArrayPattern.FindStringSubmatch(value); len(matches) == 2 {
				elements := strings.Split(matches[1], ",")
				for i := range elements {
					elements[i] = strings.TrimSpace(strings.Trim(elements[i], `"`))
				}

				result[key] = elements
			} else if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
				result[key] = strings.Trim(value, `"`)
			} else {
				result[key] = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return result, nil
}

// Marshal converts a map to the Godot project configuration format.
func (p ProjectParser) Marshal(data map[string]interface{}) ([]byte, error) {
	var buffer bytes.Buffer

	sections := make(map[string]map[string]interface{})

	// Organize data into sections.
	for key, value := range data {
		parts := strings.SplitN(key, ".", 2) //nolint:mnd

		var section, param string

		if len(parts) == 2 { //nolint:mnd
			section, param = parts[0], parts[1]
		} else {
			param = parts[0]
		}

		if _, exists := sections[section]; !exists {
			sections[section] = make(map[string]interface{})
		}

		sections[section][param] = value
	}

	// Write sections and parameters.
	for section, params := range sections {
		if section != "" {
			buffer.WriteString(fmt.Sprintf("[%s]\n", section))
		}

		for param, value := range params {
			var valueStr string
			switch v := value.(type) {
			case []string:
				valueStr = fmt.Sprintf(`PackedStringArray("%s")`, strings.Join(v, `", "`))
			case string:
				valueStr = fmt.Sprintf(`"%s"`, v)
			default:
				valueStr = fmt.Sprintf("%v", v)
			}
			buffer.WriteString(fmt.Sprintf("%s=%s\n", param, valueStr))
		}

		buffer.WriteString("\n")
	}

	return buffer.Bytes(), nil
}
