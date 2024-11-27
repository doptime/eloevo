package dynamicdev

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/doptime/doptime/api"
)

func keepFunctionDefinitionAndRemoveDetail_SourceCodeToArchitecture(content string) string {
	lines := strings.Split(content, "\n")
	var contentBuilder strings.Builder

	var curlyBrackets []string

	for i, lineString := range lines {
		//remove comment on function definition line
		line := strings.Split(lineString, "//")[0]
		//remove leading and trailing spaces
		line = strings.TrimSpace(line)
		//skip empty line
		l := len(line)
		if l == 0 {
			continue
		}
		funcDefinitionStarting, funcDefinitionEnding := false, false

		//capture the function definition line
		//golang hash strict ast tree, so this evidence is enough
		//seek position of func start
		exceptionCaseOfTypeDefinition := len(curlyBrackets) == 0 && len(line) > 5 && line[:4] == "type"
		if line[l-1] == '{' && !exceptionCaseOfTypeDefinition {
			curlyBrackets = append(curlyBrackets, "{")
			funcDefinitionStarting = len(curlyBrackets) == 1
		}
		//exception case of {}, such as string{} or map[string]string{} ... , return mytype{ a: 1, b: 2}
		exceptionCaseOfVarInitiation, exceptionCaseOfVarInitiationB := len(line) > 2 && line[l-1] == '}', false
		for li := l - 2; li >= 1 && exceptionCaseOfVarInitiation && !exceptionCaseOfVarInitiationB; li-- {
			exceptionCaseOfVarInitiationB = line[li] == '{'
		}
		exceptionCaseOfVarInitiation = exceptionCaseOfVarInitiation && exceptionCaseOfVarInitiationB

		//capture the function definition end
		if L2plusFuncEnd, L1FunEnd := line[l-1] == '}', line[0] == '}'; (L2plusFuncEnd || L1FunEnd) && len(curlyBrackets) > 0 && !exceptionCaseOfVarInitiation {
			//pop the last element
			if curlyBrackets[len(curlyBrackets)-1] == "{" {
				curlyBrackets = curlyBrackets[:len(curlyBrackets)-1]
			}
			funcDefinitionEnding = len(curlyBrackets) == 0
		}

		if inFunctionDefinition := len(curlyBrackets) > 0; funcDefinitionEnding || funcDefinitionStarting || !inFunctionDefinition {
			contentBuilder.WriteString(fmt.Sprintf("%d:%s\n", i+1, lineString))

		}

	}

	return strings.Trim(contentBuilder.String(), "\n")
}

func removeStandardLibraryPackages_SourceCodeToArchitecture(fileContentWithName string) (content string, err error) {
	var contentBuilder strings.Builder
	var importStatements []string
	var importStarted bool = false
	for _, line := range strings.Split(fileContentWithName, "\n") {
		seqWithLinestring := strings.SplitN(line, ":", 2)
		_, lineString := seqWithLinestring[0], seqWithLinestring[1]

		if strings.Contains(lineString, "import (") {
			importStarted = true
		} else if importStarted && strings.Contains(lineString, ")") {
			importStatements = append(importStatements, line)
			importStarted = false
			for i := len(importStatements) - 2; i >= 1; i-- {
				//remove the import statements if not contains "."
				if !strings.Contains(importStatements[i], ".") {
					importStatements = append(importStatements[:i], importStatements[i+1:]...)
				}
			}
			//if lefts no import statements, skip the import block
			if len(importStatements) <= 2 {
				importStatements = []string{}
			}
			//append left import statement to contentBuilder
			for _, importStatement := range importStatements {
				contentBuilder.WriteString(importStatement + "\n")
			}
			continue
		}

		if importStarted {
			importStatements = append(importStatements, line)
		} else {
			contentBuilder.WriteString(line + "\n")
		}
	}

	return contentBuilder.String(), nil

}

func SourceCodeToArchitecture(sourceCode string) (architecture string, err error) {
	processedPage := keepFunctionDefinitionAndRemoveDetail_SourceCodeToArchitecture(sourceCode)
	if processedPage, err = removeStandardLibraryPackages_SourceCodeToArchitecture(processedPage); err != nil {
		return "", err
	}
	return processedPage, nil
}

var dirOfDefaultProject, _ = os.Getwd()

type GetProjectArchitectureInfoIn struct {
	//default is current dir
	ProjectDir       string
	SkipDirs         []string
	SkipFiles        []string
	IncludedFileExts []string
	UseRawText       bool
}
type GetProjectArchitectureInfoOut struct {
	AbsPath      string
	RelFile2Arch map[string]string
}

var APIGetProjectArchitectureInfo = api.Api(func(packInfo *GetProjectArchitectureInfoIn) (architectures GetProjectArchitectureInfoOut, err error) {

	var surffixType = map[string]string{".go": "go", ".js": "js", ".ts": "js", ".vue": "js", ".jsx": "js", ".tsx": "js", ".html": "text", ".md": "text", ".json": "text", ".mdx": "text", ".toml": "text", ".txt": "text", "yaml": "text"}
	for _, surffix := range packInfo.IncludedFileExts {
		if len(surffix) == 0 {
			continue
		} else if surffix[0] != '.' {
			surffix = "." + surffix
		}
		surffixType[surffix] = "text"
	}

	architectures = GetProjectArchitectureInfoOut{
		AbsPath:      dirOfDefaultProject,
		RelFile2Arch: make(map[string]string),
	}
	if len(packInfo.ProjectDir) > 0 {
		architectures.AbsPath = packInfo.ProjectDir
	}
	var skipDirs = map[string]bool{".vscode": true, "node_modules": true}
	for _, skippedDir := range packInfo.SkipDirs {
		skipDirs[skippedDir] = true
	}
	var skipFiles = map[string]bool{}
	var skipFilesWithLeadingWildCard = []string{}
	var skipFilesWithTrailingWildCard = []string{}
	for _, skipFile := range packInfo.SkipFiles {
		if len(skipFile) == 0 {
			continue
		}
		skipFiles[skipFile] = true
		if skipFile[0] == '*' {
			skipFilesWithLeadingWildCard = append(skipFilesWithLeadingWildCard, strings.Trim(skipFile, "*"))
		}
		if skipFile[len(skipFile)-1] == '*' {
			skipFilesWithTrailingWildCard = append(skipFilesWithTrailingWildCard, strings.Trim(skipFile, "*"))
		}
	}
	// walkDir recursively walks through a directory and processes all .go files
	filepath.WalkDir(architectures.AbsPath+"/.", func(path string, info os.DirEntry, err error) error {
		if err == filepath.SkipDir {
			return nil
		}
		if info.IsDir() {
			if skiodir, ok := skipDirs[info.Name()]; ok && skiodir {
				return filepath.SkipDir
			}
			return nil
		} else if skipfile, ok := skipFiles[info.Name()]; ok && skipfile {
			return nil
		}
		for _, skipFileWithLeadingWildCard := range skipFilesWithLeadingWildCard {
			if strings.HasSuffix(info.Name(), skipFileWithLeadingWildCard) {
				return nil
			}
		}
		for _, skipFileWithTrailingWildCard := range skipFilesWithTrailingWildCard {
			if strings.HasPrefix(info.Name(), skipFileWithTrailingWildCard) {
				return nil
			}
		}
		doctype, typeExisted := surffixType[filepath.Ext(path)]
		if !typeExisted {
			return nil
		}

		page, _ := ReadInFile(path)
		corrupted := false
		if doctype == "go" {
			// 先确认语法树是否正确，如果正确再进行替换
			_, err = parser.ParseFile(token.NewFileSet(), "", page, parser.ParseComments)
			corrupted = err != nil
		}
		RelName := path[len(architectures.AbsPath):]
		architectures.RelFile2Arch[RelName] = page
		if (doctype == "go" || doctype == "js") && !corrupted && !packInfo.UseRawText {
			architectures.RelFile2Arch[RelName], _ = SourceCodeToArchitecture(page)
		}
		return nil
	})

	return architectures, nil
}).Func
