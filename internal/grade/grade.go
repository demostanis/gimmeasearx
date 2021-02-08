package grade

// Returns an array of grade IDs and
// their corresponding symbol.
func Grades() []map[string]interface{} {
	return []map[string]interface{}{
		{ "Symbol": "V", "Id": "grade-v", },
		{ "Symbol": "C", "Id": "grade-c", },
		{ "Symbol": "Cjs", "Id": "grade-cjs", },
		{ "Symbol": "E", "Id": "grade-e", },
	}
}

// Finds a grade ID's symbol.
func Symbol(id string) string {
	var result string
	for _, thisGrade := range Grades() {
		if thisGrade["Id"] == id {
			result = thisGrade["Symbol"].(string)
		}
	}
	return result
}

// Default grades for people whom haven't set them.
func Defaults() []string {
	return []string{"grade-v", "grade-c"}
}

// Returns a grade's comment which will be
// shown in a tooltip.
func Comment(grade string) string {
	switch grade {
		case "V":
			return `"Vanilla instance": all static files are the original ones from the searx source code`
		case "C":
			return "Some static files have been modified, but all scripts are the original ones from the searx source code"
		case "Cjs":
			return "Some static files have been modified, including scripts"
		case "E":
			return "Some files originate from another domain!"
		default:
			return "Unknown grade"
	}
}

