// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package yamlprocessor ...
package yamlprocessor

import (
	"fmt"
	"io"
	"strconv"

	"github.com/k14s/ytt/pkg/cmd/template"
	yttui "github.com/k14s/ytt/pkg/cmd/ui"
	"github.com/k14s/ytt/pkg/files"
	"github.com/k14s/ytt/pkg/workspace"
	"github.com/k14s/ytt/pkg/yamlmeta"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/apis/providers/v1alpha1"
)

// DefinitionParser provides behavior to process template definition
type DefinitionParser interface {
	// ParsePath returns the path specified within the template definition.
	ParsePath([]byte) ([]v1alpha1.PathInfo, error)
}

// YttProcessorOption is a type that mutates ytt based on options defined in the option
type YttProcessorOption func(*YTTProcessor)

// InjectDefinitionParser is a YttProcessorOption that allows overriding of
// the definition parser.
func InjectDefinitionParser(dp DefinitionParser) YttProcessorOption {
	return func(p *YTTProcessor) {
		p.parser = dp
	}
}

// YTTProcessor a type for processing and parsing ytt files.
type YTTProcessor struct {
	parser   DefinitionParser
	srcPaths []v1alpha1.PathInfo
}

// TODO: Add logs

// NewYttProcessor returns an instance of the YTTProcessor.
func NewYttProcessor(opts ...YttProcessorOption) *YTTProcessor {
	p := &YTTProcessor{
		parser: NewYttDefinitionParser(),
	}

	for _, o := range opts {
		o(p)
	}
	return p
}

// NewYttProcessorWithConfigDir returns an instance of the YTTProcessor
// configured with tkg config directory
func NewYttProcessorWithConfigDir(configDir string) *YTTProcessor {
	definitionParser := InjectDefinitionParser(NewYttDefinitionParser(InjectTKGDir(configDir)))
	return NewYttProcessor(definitionParser)
}

// GetTemplateName returns the name of the template definition file for the
// specified version and plan.
func (p *YTTProcessor) GetTemplateName(version, plan string) string {
	name := "cluster-template-definition"
	if plan != "" {
		name = fmt.Sprintf("%s-%s", name, plan)
	}
	return fmt.Sprintf("%s.yaml", name)
}

func (p *YTTProcessor) getLoader(rawArtifact []byte) (*workspace.LibraryLoader, error) {
	srcPaths, err := p.getYttSrcDir(rawArtifact)
	if err != nil {
		return nil, err
	}

	yttFiles, err := p.getYttFiles(srcPaths)
	if err != nil {
		return nil, err
	}

	lib := workspace.NewRootLibrary(yttFiles)
	libCtx := workspace.LibraryExecutionContext{Current: lib, Root: lib}
	libExecFact := workspace.NewLibraryExecutionFactory(&noopUI{}, workspace.TemplateLoaderOpts{})
	return libExecFact.New(libCtx), nil
}

// GetVariables returns a list of the variables specified from the ytt data
// values.
func (p *YTTProcessor) GetVariables(rawArtifact []byte) ([]string, error) {
	libLoader, err := p.getLoader(rawArtifact)
	if err != nil {
		return nil, err
	}

	var variables []string
	values, _, err := libLoader.Values([]*workspace.DataValues{})
	if err != nil || values == nil || values.Doc == nil {
		return nil, errors.Wrap(err, "unable to load yaml document")
	}

	for _, v := range values.Doc.GetValues() {
		if t, ok := v.(*yamlmeta.Map); ok {
			for _, mapItem := range t.Items {
				k, ok := mapItem.Key.(string)
				if ok {
					variables = append(variables, k)
				}
			}
		}
	}
	return variables, nil
}

// Process returns the final yaml of the ytt templates.
func (p *YTTProcessor) Process(rawArtifact []byte, variablesClient func(string) (string, error)) ([]byte, error) {
	variables, err := p.GetVariables(rawArtifact)
	if err != nil {
		return nil, err
	}

	// build out the data values for ytt
	dataValues := make([]string, 0, len(variables))
	for _, vName := range variables {
		vValue, err := variablesClient(vName)
		if err != nil {
			// skip the variables that don't have user specified values
			continue
		}

		convs := []yamlScalarConvertable{
			nullConvertable,
			booleanConvertable,
			integerConvertable,
			floatConvertable,
			structuredConvertable,
		}

		convertable := false
		for _, conv := range convs {
			convertable = conv(vValue)
			if convertable {
				break
			}
		}

		if convertable {
			dataValues = append(dataValues, fmt.Sprintf("%s=%s", vName, vValue))
		} else {
			dataValues = append(dataValues, fmt.Sprintf("%s=\"%s\"", vName, vValue))
		}
	}
	dvf := template.DataValuesFlags{
		KVsFromYAML: dataValues,
	}

	// add the data values as overlays to the ytt templates
	overlayValuesDoc, _, err := dvf.AsOverlays(false)
	if err != nil {
		return nil, err
	}

	libLoader, err := p.getLoader(rawArtifact)
	if err != nil {
		return nil, err
	}

	valuesDoc, libraryValueDoc, err := libLoader.Values(overlayValuesDoc)
	if err != nil {
		return nil, err
	}

	result, err := libLoader.Eval(valuesDoc, libraryValueDoc)
	if err != nil {
		return nil, err
	}
	return result.DocSet.AsBytes()
}

// getYttSrcDir returns the cached srcDir if it called multiple times else it
// parses it from the template definition.
func (p *YTTProcessor) getYttSrcDir(rawArtifact []byte) ([]v1alpha1.PathInfo, error) {
	if len(p.srcPaths) != 0 {
		return p.srcPaths, nil
	}

	srcPaths, err := p.parser.ParsePath(rawArtifact)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse raw artifact bytes")
	}

	p.srcPaths = srcPaths
	return p.srcPaths, nil
}

// getYttFiles returns list of ytt files object in a sorted order
func (p *YTTProcessor) getYttFiles(srcPaths []v1alpha1.PathInfo) ([]*files.File, error) {
	allPaths := []string{}
	fileMarkMap := make(map[string][]*files.File)
	allowSymlinks := files.SymlinkAllowOpts{}

	// Store all files to fileMarkMap if the FileMark is present
	// This will be used later for applying FileMark to the file object
	// After all the files are sorted
	for _, path := range srcPaths {
		allPaths = append(allPaths, path.Path)
		if path.FileMark != "" {
			fileMarkPaths, err := files.NewSortedFilesFromPaths([]string{path.Path}, allowSymlinks)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to get all relative path of files for path: %v", path.Path)
			}
			existingFileMarkPaths, exists := fileMarkMap[path.FileMark]
			if !exists {
				fileMarkMap[path.FileMark] = fileMarkPaths
			} else {
				fileMarkMap[path.FileMark] = append(existingFileMarkPaths, fileMarkPaths...)
			}
		}
	}

	// sort all file path in a single function call NewSortedFilesFromPaths
	// as each file returned has order property set within the object which
	// is used to determine the file processing order with ytt
	sortedFiles, err := files.NewSortedFilesFromPaths(allPaths, allowSymlinks)

	// Apply FileMark to the returned sortedFiles from the fileMarkMap
	for fileMark, allfiles := range fileMarkMap {
		for _, pathFileMark := range allfiles {
			for _, f := range sortedFiles {
				if f.RelativePath() == pathFileMark.RelativePath() {
					if err := p.updateFilesMetadata(f, fileMark); err != nil {
						return nil, errors.Wrap(err, "unable to update file mark")
					}
				}
			}
		}
	}

	return sortedFiles, err
}

func (p *YTTProcessor) updateFilesMetadata(srcFile *files.File, fileMark string) error {
	updateFileType := func(fileType files.Type, markTemplate bool) {
		srcFile.MarkType(fileType)
		srcFile.MarkTemplate(markTemplate)
	}

	switch fileMark {
	case "yaml-template": // yaml template processing
		updateFileType(files.TypeYAML, true)
	case "yaml-plain": // no template processing
		updateFileType(files.TypeYAML, false)
	case "text-template":
		updateFileType(files.TypeText, true)
	case "text-plain":
		updateFileType(files.TypeText, false)
	case "starlark":
		updateFileType(files.TypeStarlark, false)
	case "data":
		updateFileType(files.TypeUnknown, false)
	case "":
		// If filemark is not provided, treat each file based on the file header
		// example: files with #@data/values header will be treated as data files,
		// by default ytt determines file type based on file header
	default:
		return errors.Errorf("unknown filemark type: '%s'", fileMark)
	}
	return nil
}

type noopUI struct{}

var _ yttui.UI = noopUI{}

func (ui noopUI) Printf(str string, args ...interface{}) {
	// noop
}

func (ui noopUI) Debugf(str string, args ...interface{}) {
	// noop
}

func (ui noopUI) Warnf(str string, args ...interface{}) {
	// noop
}

func (ui noopUI) DebugWriter() io.Writer {
	return noopWriter{}
}

type noopWriter struct{}

func (n noopWriter) Write(p []byte) (int, error) {
	return 0, nil
}

var _ io.Writer = noopWriter{}

type yamlScalarConvertable func(in string) bool

func structuredConvertable(in string) bool {
	var result interface{}
	if err := yaml.Unmarshal([]byte(in), &result); err == nil && result != nil {
		return true
	}
	return false
}

func nullConvertable(in string) bool {
	return in == "~" || in == "null"
}

func booleanConvertable(in string) bool {
	if _, err := strconv.ParseBool(in); err == nil {
		return true
	}
	return false
}

func integerConvertable(in string) bool {
	if _, err := strconv.ParseUint(in, 0, 0); err == nil {
		return true
	}
	return false
}

func floatConvertable(in string) bool {
	if _, err := strconv.ParseFloat(in, 64); err == nil {
		return true
	}
	return false
}
