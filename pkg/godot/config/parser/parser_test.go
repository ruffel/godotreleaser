package parser_test

import (
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/ruffel/godotreleaser/pkg/godot/config/parser"
	"github.com/stretchr/testify/assert"
)

func TestParser_Unmarshal_Simple(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  any
	}{
		{
			name: "good - minimal",
			input: heredoc.Doc(`
				; Engine configuration data
				config_version=5
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"config_version": "5",
				},
			},
		},
		{
			name: "good - multi-line string",
			input: heredoc.Doc(`
				[preset.0.options]
				ssh_remote_deploy/cleanup_script="line1
				line2
				line3"
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{},
				"preset.0.options": map[string]interface{}{
					"ssh_remote_deploy/cleanup_script": "line1__NEWLINE__line2__NEWLINE__line3",
				},
			},
		},
		{
			name: "good - list",
			input: heredoc.Doc(`
				[DEFAULT]
				alist = foo
				alist = bar
				alist = baz
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"alist": []string{"foo", "bar", "baz"},
				},
			},
		},
		{
			name: "good - PackedStringArray(...)",
			input: heredoc.Doc(`
				[DEFAULT]
				array = PackedStringArray("foo", "bar", "baz")
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"array": []string{"foo", "bar", "baz"},
				},
			},
		},
		{
			name: "good - PackedStringArray(1)",
			input: heredoc.Doc(`
				[DEFAULT]
				array = PackedStringArray("foo")
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"array": []string{"foo"},
				},
			},
		},
		{
			name: "good - PackedStringArray(empty)",
			input: heredoc.Doc(`
				[DEFAULT]
				array = PackedStringArray()
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"array": []string{},
				},
			},
		},
		{
			name: "good - inline JSON",
			input: heredoc.Doc(`
				[DEFAULT]
				json = {"foo": "bar", "baz": 42 }
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"json": `{"foo": "bar", "baz": 42 }`,
				},
			},
		},
		{
			name: "good - inline multiline JSON",
			input: heredoc.Doc(`
				[DEFAULT]
				json = {
					"foo": "bar",
					"baz": 42
				}`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"json": `{ "foo": "bar", "baz": 42 }`,
				},
			},
		},
		{
			name: "good - nested sections",
			input: heredoc.Doc(`
				[preset.0.options]
				key1 = value1
				
				[preset.0.options.nested]
				key2 = value2
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{},
				"preset.0.options": map[string]interface{}{
					"key1": "value1",
				},
				"preset.0.options.nested": map[string]interface{}{
					"key2": "value2",
				},
			},
		},
		{
			name: "good - mixed list types",
			input: heredoc.Doc(`
				[DEFAULT]
				mixed_list = foo
				mixed_list = 42
				mixed_list = true
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"mixed_list": []string{"foo", "42", "true"},
				},
			},
		},
		{
			name:  "good - empty input",
			input: "",
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{},
			},
		},
		{
			name: "bad - invalid syntax",
			input: heredoc.Doc(`
				[DEFAULT
				key = value
			`),
			want: nil,
		},
		{
			name: "good - complex structure",
			input: heredoc.Doc(`
				[preset.0.options]
				option1 = "value1"
				
				[preset.1.options]
				option2 = "value2"
				
				[preset.1.options.subsection]
				option3 = "value3"
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{},
				"preset.0.options": map[string]interface{}{
					"option1": "value1",
				},
				"preset.1.options": map[string]interface{}{
					"option2": "value2",
				},
				"preset.1.options.subsection": map[string]interface{}{
					"option3": "value3",
				},
			},
		},
		{
			name: "good - multiple sections with the same key",
			input: heredoc.Doc(`
				[preset.0]
				option = "value1"

				[preset.1]
				option = "value2"
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{},
				"preset.0": map[string]interface{}{
					"option": "value1",
				},
				"preset.1": map[string]interface{}{
					"option": "value2",
				},
			},
		},
		{
			name: "good - duplicate list entries",
			input: heredoc.Doc(`
				[DEFAULT]
				alist = foo
				alist = foo
				alist = foo
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"alist": []string{"foo", "foo", "foo"},
				},
			},
		},
		{
			name: "good - duplicate list entries (packed)",
			input: heredoc.Doc(`
				[DEFAULT]
				alist = PackedStringArray("foo", "foo", "foo")
			`),
			want: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"alist": []string{"foo", "foo", "foo"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parser.Godot{}.Unmarshal(([]byte(tt.input)))

			if tt.want == nil {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParser_Marshal_Simple(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]interface{}
		want  string
	}{
		{
			name: "single default entry",
			input: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"config_version": "5",
				},
			},
			want: heredoc.Doc(`
				config_version = 5
			`),
		},
		{
			name: "multiple defaults entries",
			input: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"foo": "5",
					"bar": 5,
				},
			},
			want: heredoc.Doc(`
				bar = 5
				foo = 5
			`),
		},
		{
			name: "multiple sections",
			input: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"foo": "5",
				},
				"section": map[string]interface{}{
					"bar": "5",
				},
			},
			want: heredoc.Doc(`
				foo = 5

				[section]
				bar = 5
			`),
		},
		{
			name: "multiple sections (2)",
			input: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"foo": "5",
				},
				"section": map[string]interface{}{
					"bar": "5",
				},
				"section2": map[string]interface{}{
					"baz": "5",
				},
			},
			want: heredoc.Doc(`
				foo = 5

				[section]
				bar = 5

				[section2]
				baz = 5
			`),
		},
		{
			name: "packed string array",
			input: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"array": []string{"foo", "bar", "baz"},
				},
			},
			want: heredoc.Doc(`
				array = PackedStringArray("foo", "bar", "baz")
			`),
		},
		{
			name: "empty packed string array",
			input: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"array": []string{},
				},
			},
			want: heredoc.Doc(`
				array = PackedStringArray()
			`),
		},
		{
			name: "inline JSON",
			input: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"json": `{"foo": "bar", "baz": 42 }`,
				},
			},
			want: heredoc.Doc(`
				json = {"foo": "bar", "baz": 42 }
			`),
		},
		{
			name: "multiline strings",
			input: map[string]interface{}{
				"DEFAULT": map[string]interface{}{
					"multiline": "line1__NEWLINE__line2__NEWLINE__line3",
				},
			},
			want: heredoc.Doc(`
				multiline = """line1
				line2
				line3"""
			`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parser.Godot{}.Marshal(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}
