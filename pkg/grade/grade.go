package grade

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

