package main

import "path"

// UpDowngradeRelease is used to up/downgrade a release.
func UpDowngradeRelease(VersionHash string) {
	// TODO: Actually roll back!

	// Print that we are running a new release.
	println("[IRONFIST] Now running " + VersionHash + ".")

	// Run the application.
	LaunchApplication(path.Join(FolderPath, VersionHash))
}
