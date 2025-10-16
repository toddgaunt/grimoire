package main

import (
	"fmt"
	"regexp"
	"strings"
)

var paramRegex = regexp.MustCompile(`<([^\s][^<>]*[^\s])>`)

// Param is a single parameter in a spell that indicates a value to be substituted
type Param struct {
	Name          string
	DefaultValues []string
}

// SpellSegments represents a spell split into segments where parameters can be substituted
type SpellSegments struct {
	Segments     []string // All segments (text and parameter names)
	ParamIndices []int    // Indices in Segments that contain parameter names
	Params       []Param  // Unique parameters with their default values
}

// Reconstruct rebuilds the spell with the given parameter values
func (ss *SpellSegments) Reconstruct(paramValues map[string]string) (string, error) {
	result := make([]string, len(ss.Segments))
	copy(result, ss.Segments)

	// Replace parameter segments with their values
	for _, idx := range ss.ParamIndices {
		paramName := result[idx]
		if value, exists := paramValues[paramName]; exists {
			result[idx] = value
		} else {
			return "", fmt.Errorf("no value provided for parameter '%s'", paramName)
		}
	}

	return strings.Join(result, ""), nil
}

// ParseSpell splits a spell into segments and identifies parameter locations.
// Parameter segments are replaced with just the parameter name, and are marked
// by ParamIndices for later substitution.
// Format can be: <param_name> or <param_name=default;default2>
func ParseSpell(spell string) (*SpellSegments, error) {
	matches := paramRegex.FindAllStringSubmatchIndex(spell, -1)

	if len(matches) == 0 {
		// No parameters, return a single segment
		return &SpellSegments{
			Segments:     []string{spell},
			ParamIndices: []int{},
			Params:       []Param{},
		}, nil
	}

	var segments []string
	var paramIndices []int
	paramMap := make(map[string]Param)
	var paramOrder []string

	lastEnd := 0

	for _, match := range matches {
		// match[0] is start of full match, match[1] is end of full match
		start := match[0]
		end := match[1]

		// match[2] is start of captured group, match[3] is end of captured group
		paramStart := match[2]
		paramEnd := match[3]

		// Add the text segment before this parameter
		if start > lastEnd {
			segments = append(segments, spell[lastEnd:start])
		}

		if paramStart >= 0 && paramEnd >= 0 {
			name, defaultValues := extractParamNameAndDefaults(spell, paramStart, paramEnd)

			if _, exists := paramMap[name]; exists {
				if len(defaultValues) > 1 {
					return nil, fmt.Errorf("parameter '%s' appears multiple times with default values - defaults only allowed on first occurrence", name)
				}
			} else {
				paramMap[name] = Param{
					Name:          name,
					DefaultValues: defaultValues,
				}

				paramOrder = append(paramOrder, name)
			}

			// Add the parameter name alone as a segment, this
			// allows us to use it to map values later. Mark
			// its index for later substitution.
			segments = append(segments, name)
			paramIndices = append(paramIndices, len(segments)-1)
		}

		lastEnd = end
	}

	// Add any remaining text after the last parameter
	if lastEnd < len(spell) {
		segments = append(segments, spell[lastEnd:])
	}

	// Convert paramMap to slice in order of first occurrence
	var params []Param
	for _, name := range paramOrder {
		params = append(params, paramMap[name])
	}

	return &SpellSegments{
		Segments:     segments,
		ParamIndices: paramIndices,
		Params:       params,
	}, nil
}

func extractParamNameAndDefaults(spell string, paramStart, paramEnd int) (string, []string) {
	paramText := strings.TrimSpace(spell[paramStart:paramEnd])
	parts := strings.SplitN(paramText, "=", 2)
	name := strings.TrimSpace(parts[0])

	var defaultValues []string
	if len(parts) > 1 {
		defaultValues = strings.Split(parts[1], ";")
	}

	return name, defaultValues
}
