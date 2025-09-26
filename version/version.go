package version

// Version information variables - these will be overridden by build flags
var (
	// Version is the current version of the application
	Version = "local-dev"
	// GitCommit is the git commit hash
	GitCommit = "unknown"
	// GitBranch is the git branch name
	GitBranch = "unknown"
	// BuildTime is when the binary was built
	BuildTime = "unknown"
)
