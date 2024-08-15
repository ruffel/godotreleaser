package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/ini.v1"
)

type Godot struct{}

func (g Godot) Unmarshal(data []byte) (map[string]interface{}, error) {
	sane := sanitizeData(data)

	// Load sanitized data into INI object
	d, err := ini.Load(sane)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	result := make(map[string]interface{})

	// Traverse the INI sections and keys
	for _, section := range d.Sections() {
		sectionMap := make(map[string]interface{})

		for _, key := range section.Keys() {
			sectionMap[key.Name()] = key.Value()
		}

		result[section.Name()] = sectionMap
	}

	return result, nil
}

func (g Godot) Marshal(data map[string]interface{}) ([]byte, error) {
	cfg := ini.Empty()

	for sectionName, sectionData := range data {
		section, err := cfg.NewSection(sectionName)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		if sectionMap, ok := sectionData.(map[string]interface{}); ok {
			for key, value := range sectionMap {
				_, err := section.NewKey(key, fmt.Sprintf("%v", value))
				if err != nil {
					return nil, err //nolint:wrapcheck
				}
			}
		}
	}

	var buf bytes.Buffer
	if _, err := cfg.WriteTo(&buf); err != nil {
		return nil, err //nolint:wrapcheck
	}

	return buf.Bytes(), nil
}

func sanitizeData(data []byte) []byte {
	var sane bytes.Buffer

	var buffer strings.Builder

	pattern := regexp.MustCompile(`PackedStringArray\((.*?)\)`)

	// Placeholder for newlines in multi-line strings, not 100% sure this is the best approach.
	const newlinePlaceholder = "__NEWLINE__"

	// Create a scanner to read the data line by line
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()
		line = pattern.ReplaceAllString(line, "[$1]")

		// If buffer contains data, we are processing a multi-line string
		if buffer.Len() != 0 {
			buffer.WriteString(line)
			buffer.WriteString(newlinePlaceholder)

			// If this line ends the multi-line string, flush the buffer
			if strings.HasSuffix(line, "\"") {
				sane.WriteString(buffer.String())
				sane.WriteString("\n")
				buffer.Reset()
			}

			continue
		}

		// Detect the start of a multi-line string (e.g., key="multi-line-start...)
		if idx := strings.Index(line, "=\""); idx != -1 && !strings.HasSuffix(line, "\"") {
			buffer.WriteString(line)
			buffer.WriteString(newlinePlaceholder)

			continue
		}

		sane.WriteString(line + "\n")
	}

	return sane.Bytes()
}
