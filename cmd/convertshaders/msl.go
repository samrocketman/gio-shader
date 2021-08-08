// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

// MSL is hlsl compiler that targets the Metal shading language
type MSL struct {
	WorkDir WorkDir
}

// MetalLibs contains compiled .metallib programs for all supported platforms.
type MetalLibs struct {
	MacOS        []byte
	IOS          []byte
	IOSSimulator []byte
}

// Compile compiles the input .metal program and converts it into .metallib libraries.
func (msl *MSL) Compile(path, variant string, src, srcIOS []byte) (MetalLibs, error) {
	base := msl.WorkDir.Path(filepath.Base(path), variant)
	pathinMacOS := base + ".macos.metal"
	pathinIOS := base + ".ios.metal"

	var libs MetalLibs
	if err := msl.WorkDir.WriteFile(pathinMacOS, []byte(src)); err != nil {
		return libs, fmt.Errorf("unable to write shader to disk: %w", err)
	}
	if err := msl.WorkDir.WriteFile(pathinIOS, []byte(srcIOS)); err != nil {
		return libs, fmt.Errorf("unable to write shader to disk: %w", err)
	}

	var err error
	libs.MacOS, err = msl.compileFor("macosx", "-mmacosx-version-min=10.11", pathinMacOS)
	if err != nil {
		return libs, err
	}
	libs.IOS, err = msl.compileFor("iphoneos", "-mios-version-min=10.0", pathinIOS)
	if err != nil {
		return libs, err
	}
	libs.IOSSimulator, err = msl.compileFor("iphonesimulator", "-miphonesimulator-version-min=8.0", pathinIOS)
	if err != nil {
		return libs, err
	}

	return libs, nil
}

// compileFor compiles the input .metal program and converts it into a
// .metallib library for a particular SDK.
func (msl *MSL) compileFor(sdk, minVer, path string) ([]byte, error) {
	metal := exec.Command("xcrun", "--sdk", sdk, "metal")
	pathout := path + ".metallib"

	metal.Args = append(metal.Args,
		minVer,
		"-o", pathout,
		path,
	)

	output, err := metal.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s\nfailed to run %v: %w", output, metal.Args, err)
	}

	compiled, err := ioutil.ReadFile(pathout)
	if err != nil {
		return nil, fmt.Errorf("unable to read output %q: %w", pathout, err)
	}

	return compiled, nil
}
