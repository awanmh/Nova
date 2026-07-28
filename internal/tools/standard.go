package tools

// RegisterStandardTools adds all 8 core built-in NOVA tools to the registry.
func RegisterStandardTools(rootDir string, reg *Registry) error {
	tools := []Tool{
		NewReadFileTool(rootDir),
		NewWriteFileTool(rootDir),
		NewListDirTool(rootDir),
		NewSearchFilesTool(rootDir),
		NewRunCmdTool(rootDir),
		NewGitStatusTool(rootDir),
		NewGitDiffTool(rootDir),
		NewApplyPatchTool(rootDir),
	}

	for _, t := range tools {
		if err := reg.Register(t); err != nil {
			return err
		}
	}
	return nil
}
