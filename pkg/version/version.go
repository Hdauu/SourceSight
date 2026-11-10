// Package version holds build metadata. Release builds override Version via
// -ldflags "-X sourcesight/pkg/version.Version=v2.1.0".
package version

// Version of this build. "dev" disables update checks.
var Version = "dev"

// Repo is the GitHub "owner/name" that publishes releases.
const Repo = "Hdauu/SourceSight"

// RepoURL is the public project page.
const RepoURL = "https://github.com/" + Repo
