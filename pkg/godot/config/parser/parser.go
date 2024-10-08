package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"
	"gopkg.in/ini.v1"
)

const (
	arraySentinel   = "__ARRAY_SENTINEL__"
	newlineSentinel = "__NEWLINE__"
)

var arrayPattern = regexp.MustCompile(`^(.*)=.*PackedStringArray\((.*)\).*$`)

type Godot struct{}

func (g Godot) Unmarshal(data []byte) (map[string]interface{}, error) {
	sane := sanitizeData(data)

	// Load sanitized data into INI object
	d, err := ini.LoadSources(ini.LoadOptions{AllowShadows: true, AllowDuplicateShadowValues: true}, sane)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	result := make(map[string]interface{})

	// Traverse the INI sections and keys
	for _, section := range d.Sections() {
		sectionMap := make(map[string]interface{})

		for _, key := range section.Keys() {
			values := key.ValueWithShadows()

			// We can't differentiate between an empty/single value array with a
			// scalar value.
			//
			// To work around this, we use a sentinel value to indicate that the
			// key is an array. If the sentinel is present, we prune it and
			// return the remaining values as an array.
			switch {
			case len(values) == 1 && values[0] == arraySentinel:
				sectionMap[key.Name()] = []string{}
			case len(values) > 1:
				sectionMap[key.Name()] = lo.Ternary(values[0] == arraySentinel, values[1:], values)
			default:
				sectionMap[key.Name()] = key.Value()
			}
		}

		result[section.Name()] = sectionMap
	}

	return result, nil
}

func (g Godot) Marshal(data map[string]interface{}) ([]byte, error) {
	cfg := ini.Empty()

	// Get a sorted list of section names
	sectionNames := lo.Keys(data)
	sort.StringSlice.Sort(sectionNames) // Sort the section names to ensure consistent order

	for _, sectionName := range sectionNames {
		sectionData := data[sectionName]

		section, err := cfg.NewSection(sectionName)
		if err != nil {
			return nil, err //nolint:wrapcheck
		}

		if sectionMap, ok := sectionData.(map[string]interface{}); ok {
			// Get sorted keys for the section to ensure consistent order
			keys := lo.Keys(sectionMap)
			sort.StringSlice.Sort(keys)

			for _, key := range keys {
				value := sectionMap[key]
				switch v := value.(type) {
				case []string:
					// This is an array, we need to reconstruct the PackedStringArray format
					_, _ = section.NewKey(key, reconstructPackedStringArray(v))
				default:
					// For normal strings or other data types, handle sentinel replacements
					_, _ = section.NewKey(key, reverseSentinelReplacements(fmt.Sprintf("%v", v)))
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

// Helper function to reconstruct PackedStringArray from the string slice
func reconstructPackedStringArray(values []string) string {
	if len(values) == 0 {
		return "PackedStringArray()"
	}

	// Quote each value properly
	quotedValues := lo.Map(values, func(val string, _ int) string {
		return fmt.Sprintf("\"%s\"", val)
	})

	return fmt.Sprintf("PackedStringArray(%s)", strings.Join(quotedValues, ", "))
}

// Helper function to reverse sentinel replacements
func reverseSentinelReplacements(value string) string {
	// Replace newline sentinel with actual newlines
	value = strings.ReplaceAll(value, newlineSentinel, "\n")

	// Return the modified value
	return value
}

// sanitizeData modifies the input data buffer to attempt to make it parseable by the INI parser.
func sanitizeData(data []byte) []byte {
	var sane bytes.Buffer

	// Create a scanner to read the data line by line
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var buffer strings.Builder

	var insideJSON, insideMultilineString bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if isPackedStringArray(line) {
			handlePackedStringArray(line, &sane)

			continue
		}

		if insideMultilineString {
			insideMultilineString = handleMultilineString(line, &buffer, &sane)

			continue
		}

		if isStartOfMultilineString(line) {
			insideMultilineString = true

			buffer.WriteString(line)
			buffer.WriteString(newlineSentinel)

			continue
		}

		if insideJSON {
			insideJSON = handleJSON(line, &buffer, &sane)

			continue
		}

		if isStartOfJSON(line) {
			insideJSON = true

			buffer.WriteString(strings.TrimSpace(line) + " ")

			continue
		}

		// Write non-buffered lines directly to sane
		sane.WriteString(line + "\n")
	}

	return sane.Bytes()
}

func isPackedStringArray(line string) bool {
	return arrayPattern.MatchString(line)
}

func handlePackedStringArray(line string, sane *bytes.Buffer) {
	const lineFormat = "%s = %s\n"

	matches := arrayPattern.FindStringSubmatch(line)
	k := matches[1]
	v := matches[2]

	// Split the array into individual values
	values := strings.Split(v, ",")

	// Write the array sentinel to the buffer. This differentiates between an empty array and a single value.
	sane.WriteString(fmt.Sprintf(lineFormat, k, arraySentinel))

	// Write each value to the buffer
	for _, value := range values {
		sane.WriteString(fmt.Sprintf(lineFormat, k, strings.TrimSpace(value)))
	}
}

func isStartOfMultilineString(line string) bool {
	return strings.Contains(line, "=\"") && !strings.HasSuffix(line, "\"")
}

func handleMultilineString(line string, buffer *strings.Builder, sane *bytes.Buffer) bool {
	buffer.WriteString(line)

	// End of multi-line string, flush the buffer
	if strings.HasSuffix(line, "\"") && !strings.HasSuffix(line, "\\\"") {
		sane.WriteString(buffer.String())
		sane.WriteString("\n")
		buffer.Reset()

		return false
	}

	// Still inside multi-line string
	buffer.WriteString(newlineSentinel)

	return true
}

func isStartOfJSON(line string) bool {
	return strings.Contains(line, "{") && !strings.HasSuffix(line, "}")
}

func handleJSON(line string, buffer *strings.Builder, sane *bytes.Buffer) bool {
	buffer.WriteString(strings.TrimSpace(line) + " ")

	// End of JSON-like structure, flush the buffer
	if strings.Contains(line, "}") {
		sane.WriteString(buffer.String())
		sane.WriteString("\n")
		buffer.Reset()

		return false
	}

	// Still inside JSON
	return true
}
