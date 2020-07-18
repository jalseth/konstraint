package rego

import (
	"fmt"
	"regexp"

	"github.com/open-policy-agent/opa/ast"
)

// LoadPoliciesWithAction loads all policies from rego with rules with a given action name
func LoadPoliciesWithAction(filesContents map[string]string, action string) ([]File, error) {
	regoFiles, err := loadRegoFiles(filesContents)
	if err != nil {
		return nil, fmt.Errorf("load rego files: %w", err)
	}

	policies := getPoliciesWithAction(regoFiles, action)
	return policies, nil
}

// LoadPolicies loads all policies from rego with rules
func LoadPolicies(filesContents map[string]string) ([]File, error) {
	regoFiles, err := loadRegoFiles(filesContents)
	if err != nil {
		return nil, fmt.Errorf("load rego files: %w", err)
	}

	policies := getPoliciesWithAction(regoFiles, "")
	return policies, nil
}

// LoadLibraries loads all libraries from rego
func LoadLibraries(filesContents map[string]string) ([]File, error) {
	regoFiles, err := loadRegoFiles(filesContents)
	if err != nil {
		return nil, fmt.Errorf("load rego files: %w", err)
	}

	return regoFiles, nil
}

func getPoliciesWithAction(regoFiles []File, action string) []File {
	var matchingPolicies []File
	for _, regoFile := range regoFiles {
		if action == "" && len(regoFile.RulesActions) > 0 {
			matchingPolicies = append(matchingPolicies, regoFile)
			continue
		}

		for _, ruleAction := range regoFile.RulesActions {
			if ruleAction == action {
				matchingPolicies = append(matchingPolicies, regoFile)
			}
		}
	}

	return matchingPolicies
}

func loadRegoFiles(filesContents map[string]string) ([]File, error) {
	var regoFiles []File
	for path, contents := range filesContents {
		regoFile, err := newRegoFile(path, contents)
		if err != nil {
			return nil, fmt.Errorf("new rego file: %w", err)
		}

		regoFiles = append(regoFiles, regoFile)
	}

	return regoFiles, nil
}

func newRegoFile(filePath string, contents string) (File, error) {
	module, err := ast.ParseModule(filePath, contents)
	if err != nil {
		return File{}, fmt.Errorf("parse module: %w", err)
	}

	var importPackages []string
	for i := range module.Imports {
		importPackages = append(importPackages, module.Imports[i].Path.String())
	}

	File := File{
		FilePath:       filePath,
		PackageName:    module.Package.Path.String(),
		ImportPackages: importPackages,
		Contents:       contents,
		RulesActions:   getModuleRulesActions(module.Rules),
		Comments:       getModuleComments(module),
	}

	return File, nil
}

func getModuleRulesActions(rules []*ast.Rule) []string {
	var rulesActions []string
	re := regexp.MustCompile("^\\s*([a-z]+)\\[msg")
	for _, rule := range rules {
		match := re.FindStringSubmatch(rule.Head.String())
		if len(match) == 0 {
			continue
		}
		rulesActions = append(rulesActions, match[1])
	}
	return rulesActions
}

func getModuleComments(module *ast.Module) []string {
	var comments []string
	for _, comment := range module.Comments {
		comments = append(comments, string(comment.Text))
	}
	return comments
}
