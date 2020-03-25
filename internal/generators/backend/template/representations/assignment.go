package representations

const Assignment = `

import (
	"bufio"
	"encoding/csv"
	"io"
	"os"
	"strings"

	{{ template "import_curve" . }}
	{{if ne .Curve "GENERIC"}}
   	"github.com/consensys/gnark/backend"
	{{end}}
)


// Assignment is used to specify inputs to the Prove and Verify functions
type Assignment struct {
	Value    fr.Element
	IsPublic bool // default == false (assignemnt is private)
}

// Assignments is used to specify inputs to the Prove and Verify functions
type Assignments map[string]Assignment

// NewAssignment returns an empty Assigments object
func NewAssignment() Assignments {
	return make(Assignments)
}

// Assign assign a value to a Secret/Public input identified by its name
func (a Assignments) Assign(visibility {{if ne .Curve "GENERIC"}} backend.{{- end}}Visibility, name string, v interface{}) {
	if _, ok := a[name]; ok {
		panic(name + " already assigned")
	}
	switch visibility {
	case {{if ne .Curve "GENERIC"}} backend.{{- end}}Secret:
		a[name] = Assignment{Value: fr.FromInterface(v)}
	case {{if ne .Curve "GENERIC"}} backend.{{- end}}Public:
		a[name] = Assignment{
			Value:    fr.FromInterface(v),
			IsPublic: true,
		}
	default:
		panic("supported visibility attributes are SECRET and PUBLIC")
	}
}

// ReadFile parse r1cs.Assigments from given file
func (assignment Assignments) ReadFile(filePath string) error {
	csvFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer csvFile.Close()
	return assignment.Read(csvFile)
}

// Read parse r1cs.Assigments from given io.Reader
func (assigment Assignments) Read(r io.Reader) error {
	reader := csv.NewReader(bufio.NewReader(r))
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		} else if len(line) != 3 {
			return {{if ne .Curve "GENERIC"}} backend.{{- end}}ErrInvalidInputFormat
		}
		visibility := strings.ToLower(strings.TrimSpace(line[0]))
		name := strings.TrimSpace(line[1])
		value := strings.TrimSpace(line[2])

		assigment.Assign({{if ne .Curve "GENERIC"}} backend.{{- end}}Visibility(visibility), name, value)
	}
	return nil
}


// WriteFile serialize given assigment to disk
func (assignment Assignments) WriteFile(path string) error {
	csvFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer csvFile.Close()
	return assignment.Write(csvFile)
}

// Write serialize given assigment to io.Writer
func (assignment Assignments) Write(w io.Writer) error {
	writer := csv.NewWriter(w)
	for k, v := range assignment {
		r := v.Value
		record := []string{string({{if ne .Curve "GENERIC"}} backend.{{- end}}Secret), k, r.String()}
		if v.IsPublic {
			record[0] = string({{if ne .Curve "GENERIC"}} backend.{{- end}}Public)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
}

// DiscardSecrets returns a copy of self without Secret Assigment
func (assignments Assignments) DiscardSecrets() Assignments {
	toReturn := NewAssignment()
	for k, v := range assignments {
		if v.IsPublic {
			toReturn[k] = v
		}
	}
	return toReturn
}

`
