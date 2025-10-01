//go:build mage

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/autonomouskoi/akcore/modules"
	"github.com/autonomouskoi/mageutil"
)

var (
	baseDir     string
	outDir      string
	pluginDir   string
	version     string
	webSrcDir   string
	webOutDir   string
	webOutPBDir string
	nodeDir     string

	serverDir       string
	serverWebDir    string
	serverWebOutDir string

	Default = Plugin
)

func init() {
	// set up our paths
	var err error
	baseDir, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	outDir = filepath.Join(baseDir, "out")
	pluginDir = filepath.Join(outDir, "plugin")
	webSrcDir = filepath.Join(baseDir, "web")
	webOutDir = filepath.Join(webSrcDir, "out")
	webOutPBDir = filepath.Join(webOutDir, "pb")
	nodeDir = filepath.Join(webSrcDir, "node_modules")

	serverDir = filepath.Join(baseDir, "server")
	serverWebDir = filepath.Join(serverDir, "html")
	serverWebOutDir = filepath.Join(serverWebDir, "out")
}

// clean up intermediate products
func Clean() error {
	for _, dir := range []string{
		outDir, webOutDir,
	} {
		if err := sh.Rm(dir); err != nil {
			return fmt.Errorf("deleting %s: %w", dir, err)
		}
	}
	return cleanGoProtos()
}

// What's needed for running the plugin as a dev plugin
func Plugin() {
	mg.Deps(
		Icon,
		Manifest,
		WASM,
		Web,
	)
}

// Load plugin version from VERSION file
func readVersion() error {
	b, err := os.ReadFile(filepath.Join(baseDir, "VERSION"))
	if err != nil {
		return fmt.Errorf("reading VERSION: %w", err)
	}
	version = strings.TrimSpace(string(b))
	return nil
}

// Build the plugin for release
func Release() error {
	mg.Deps(Plugin, readVersion)
	filename := fmt.Sprintf("trackstar-live-%s.akplugin", version)
	return mageutil.ZipDir(pluginDir, filepath.Join(outDir, filename))
}

func cleanGoProtos() error {
	goDir := filepath.Join(baseDir, "go")
	protoFiles, err := mageutil.DirGlob(goDir, "*.pb.go")
	if err != nil {
		return fmt.Errorf("globbing %s/*.pb.go: %w", goDir, err)
	}
	for _, protoFile := range protoFiles {
		protoFile = filepath.Join(goDir, protoFile)
		if err := sh.Rm(protoFile); err != nil {
			return fmt.Errorf("deleting %s: %w", protoFile, err)
		}
	}
	return nil
}

// Generate tinygo code for our protos
func GoProtos() error {
	protos, err := mageutil.DirGlob(baseDir, "*.proto")
	if err != nil {
		return fmt.Errorf("globbing %s: %w", baseDir)
	}

	for _, protoFile := range protos {
		srcPath := filepath.Join(baseDir, protoFile)
		protoFile := strings.TrimSuffix(protoFile, ".proto") + ".pb.go"
		dstPath := filepath.Join(baseDir, "go", protoFile)
		err := mageutil.TinyGoProto(dstPath, srcPath, filepath.Join(baseDir, ".."))
		if err != nil {
			return fmt.Errorf("generating from %s: %w", srcPath, err)
		}
		dstPath = filepath.Join(baseDir, protoFile)
		err = mageutil.GoProto(dstPath, srcPath, baseDir, baseDir, "module=github.com/autonomouskoi/trackstar-live")
		if err != nil {
			return fmt.Errorf("generating server proto from %s: %w", srcPath, err)
		}
		mageutil.ReplaceInFile(dstPath, "package trackstar_live", "package server")
		srvPath := filepath.Join(baseDir, "server", protoFile)
		if err := os.Rename(dstPath, srvPath); err != nil {
			return fmt.Errorf("moving %d -> %d: %w", dstPath, srvPath, err)
		}
	}
	return nil
}

// Create our output dir
func mkOutDir() error {
	return mageutil.Mkdir(outDir)
}

// Create our plugin dir
func mkPluginDir() error {
	mg.Deps(mkOutDir)
	return mageutil.Mkdir(pluginDir)
}

// Compile our WASM code
func WASM() error {
	mg.Deps(mkPluginDir, GoProtos)

	goDir := filepath.Join(baseDir, "go")
	srcDir := filepath.Join(goDir, "main")
	outFile := filepath.Join(pluginDir, "plugin.wasm")

	return mageutil.TinyGoWASM(outFile, srcDir, goDir)
}

// Copy our icon
func Icon() error {
	mg.Deps(mkPluginDir)
	iconPath := filepath.Join(baseDir, "icon.svg")
	outPath := filepath.Join(pluginDir, "icon.svg")
	return mageutil.CopyFiles(map[string]string{
		iconPath: outPath,
	})
}

// Write our manifest
func Manifest() error {
	mg.Deps(mkPluginDir, readVersion)
	manifestPB := &modules.Manifest{
		Title:       "TS: Live",
		Id:          "352fad0f027de97e",
		Name:        "trackstar-live",
		Description: "Upload real-time Track ID data to the Trackstar Live web service",
		WebPaths: []*modules.ManifestWebPath{
			{
				Path:        "/m/trackstar-live/embed_ctrl.js",
				Type:        modules.ManifestWebPathType_MANIFEST_WEB_PATH_TYPE_EMBED_CONTROL,
				Description: "Controls for Trackstar Live",
			},
		},
	}
	manifest, err := protojson.Marshal(manifestPB)
	if err != nil {
		return fmt.Errorf("marshalling proto: %w", err)
	}
	buf := &bytes.Buffer{}
	if err := json.Indent(buf, manifest, "", "  "); err != nil {
		return fmt.Errorf("formatting manifest JSON: %w", err)
	}
	fmt.Fprintln(buf)
	manifestPath := filepath.Join(pluginDir, "manifest.json")
	_, err = os.Stat(manifestPath)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.WriteFile(manifestPath, buf.Bytes(), 0644)
}

// install NPM modules for web content
func NPMModules() error {
	if _, err := os.Stat(nodeDir); err == nil {
		return nil
	}
	if err := os.Chdir(webSrcDir); err != nil {
		return fmt.Errorf("switching to %s: %w", webSrcDir, err)
	}
	if err := sh.Run("npm", "install"); err != nil {
		return fmt.Errorf("running npm install: %w", err)
	}
	return nil
}

// Create our web output dir
func mkWebOutDir() error {
	return mageutil.Mkdir(webOutDir)
}

func mkTSPBDir() error {
	mg.Deps(mkWebOutDir)
	return mageutil.Mkdir(webOutPBDir)
}

// Generate our TypeScript protos
func TSProtos() error {
	mg.Deps(mkTSPBDir, NPMModules)
	if err := os.Chdir(webSrcDir); err != nil {
		return fmt.Errorf("switching to %s: %w", webSrcDir, err)
	}
	err := mageutil.TSProto(
		webOutPBDir,
		filepath.Join(baseDir, "live.proto"),
		filepath.Join(baseDir, ".."),
		nodeDir,
	)
	if err != nil {
		return err
	}
	// output is relative to includes so we have to move the live files
	dumbOutPath := filepath.Join(webOutPBDir, "live")
	outFiles, err := os.ReadDir(dumbOutPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading %s: %w", dumbOutPath, err)
	}
	for _, outFile := range outFiles {
		outPath := filepath.Join(dumbOutPath, outFile.Name())
		/*
			if err := mageutil.ReplaceInFile(outPath, `"../twitch/eventsub_pb.js"`, `"/m/twitch/pb/eventsub_pb.js"`); err != nil {
				return fmt.Errorf("replacing proto import in %s: %w", outPath, err)
			}
			if err := mageutil.ReplaceInFile(outPath, `"../twitch/twitch_pb.js"`, `"/m/twitch/pb/twitch_pb.js"`); err != nil {
				return fmt.Errorf("replacing proto import in %s: %w", outPath, err)
			}
		*/
		if err := os.Rename(outPath, filepath.Join(webOutPBDir, outFile.Name())); err != nil {
			return fmt.Errorf("moving %d -> %d: %w", outPath, webOutPBDir, err)
		}
	}
	sh.Rm(dumbOutPath)
	return nil
}

// Compile our TS code
func TS() error {
	mg.Deps(TSProtos)
	return mageutil.BuildTypeScript(webSrcDir, webSrcDir, webOutDir)
}

// Copy static web content
func WebSrcCopy() error {
	mg.Deps(mkWebOutDir)
	return nil
}

// All our web targets
func Web() error {
	mg.Deps(
		WebSrcCopy,
		TS,
	)
	return mageutil.SyncDirBasic(webOutDir, filepath.Join(pluginDir, "web"))
}

func ServerWebDirs() error {
	return os.MkdirAll(serverWebOutDir, 0755)
}

func ServerWebContent() error {
	mg.Deps(ServerWebDirs)

	err := mageutil.CopyInDir(serverWebOutDir, serverWebDir,
		"favicon.svg", "index.html", "main.css",
	)
	if err != nil {
		return fmt.Errorf("copying server web files: %w", err)
	}
	err = mageutil.CopyFiles(map[string]string{
		filepath.Join(baseDir, "icon.svg"): filepath.Join(serverWebOutDir, "icon.svg"),
	})
	if err != nil {
		return fmt.Errorf("copying icon: %w", err)
	}
	err = mageutil.BuildTypeScript(serverWebDir, serverWebDir, serverWebOutDir)
	if err != nil {
		return fmt.Errorf("building server web typescript: %w", err)
	}
	return nil
}
