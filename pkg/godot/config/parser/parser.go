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

//nolint:cyclop,funlen
func sanitizeData(data []byte) []byte {
	var sane bytes.Buffer

	var buffer strings.Builder

	var insideJSON, insideMultilineString bool

	// Placeholder for newlines in multi-line strings
	const newlinePlaceholder = "__NEWLINE__"

	// Create a scanner to read the data line by line
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Handle PackedStringArray
		pattern := regexp.MustCompile(`PackedStringArray\((.*?)\)`)
		line = pattern.ReplaceAllString(line, "[$1]")

		// Handle multi-line strings
		if insideMultilineString {
			buffer.WriteString(line)
			buffer.WriteString(newlinePlaceholder)

			// If this line ends the multi-line string, flush the buffer
			//
			// HACK: Double check that it's not an escaped string. This is lazy,
			// find a better way.
			if strings.HasSuffix(line, "\"") && !strings.HasSuffix(line, "\\\"") {
				sane.WriteString(buffer.String())
				sane.WriteString("\n")
				buffer.Reset()

				insideMultilineString = false
			}

			continue
		}

		// Detect the start of a multi-line string (e.g., key="multi-line-start...)
		if idx := strings.Index(line, "=\""); idx != -1 && !strings.HasSuffix(line, "\"") {
			insideMultilineString = true

			buffer.WriteString(line)
			buffer.WriteString(newlinePlaceholder)

			continue
		}

		// Detect the start of JSON-like structures (e.g., `{`)
		if strings.Contains(line, "{") && !strings.HasSuffix(line, "}") {
			insideJSON = true
		}

		// If inside JSON, compress into a single line
		if insideJSON {
			buffer.WriteString(strings.TrimSpace(line))

			// If this line ends the JSON, flush the buffer
			if strings.Contains(line, "}") {
				sane.WriteString(buffer.String())
				sane.WriteString("\n")
				buffer.Reset()

				insideJSON = false
			}

			continue
		}

		// Write non-buffered lines directly to sane
		sane.WriteString(line + "\n")
	}

	return sane.Bytes()
}
