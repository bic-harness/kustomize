// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/cmd/config/ext"
	"sigs.k8s.io/kustomize/cmd/config/internal/generateddocs/commands"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

// FmtCmd returns a command FmtRunner.
func GetFmtRunner(name string) *FmtRunner {
	r := &FmtRunner{}
	c := &cobra.Command{
		Use:     "fmt DIR...",
		Short:   commands.FmtShort,
		Long:    commands.FmtLong,
		Example: commands.FmtExamples,
		RunE:    r.runE,
		PreRunE: r.preRunE,
	}
	fixDocs(name, c)
	c.Flags().StringVar(&r.FilenamePattern, "pattern", filters.DefaultFilenamePattern,
		`pattern to use for generating filenames for resources -- may contain the following
formatting substitution verbs {'%n': 'metadata.name', '%s': 'metadata.namespace', '%k': 'kind'}`)
	c.Flags().BoolVar(&r.SetFilenames, "set-filenames", false,
		`if true, set default filenames on Resources without them`)
	c.Flags().BoolVar(&r.KeepAnnotations, "keep-annotations", false,
		`if true, keep index and filename annotations set on Resources.`)
	c.Flags().BoolVar(&r.Override, "override", false,
		`if true, override existing filepath annotations.`)
	c.Flags().BoolVar(&r.UseSchema, "use-schema", false,
		`if true, uses openapi resource schema to format resources.`)
	c.Flags().BoolVarP(&r.RecurseSubPackages, "recurse-subpackages", "R", false,
		"formats resource files recursively in all the nested subpackages")
	r.Command = c
	return r
}

func FmtCommand(name string) *cobra.Command {
	return GetFmtRunner(name).Command
}

// FmtRunner contains the run function
type FmtRunner struct {
	Command            *cobra.Command
	FilenamePattern    string
	SetFilenames       bool
	KeepAnnotations    bool
	Override           bool
	UseSchema          bool
	RecurseSubPackages bool
}

func (r *FmtRunner) preRunE(c *cobra.Command, args []string) error {
	if r.SetFilenames {
		r.KeepAnnotations = true
	}
	return nil
}

func (r *FmtRunner) runE(c *cobra.Command, args []string) error {

	// format stdin if there are no args
	if len(args) == 0 {
		rw := &kio.ByteReadWriter{
			Reader:                c.InOrStdin(),
			Writer:                c.OutOrStdout(),
			KeepReaderAnnotations: r.KeepAnnotations,
		}
		return handleError(c, kio.Pipeline{
			Inputs: []kio.Reader{rw}, Filters: r.fmtFilters(), Outputs: []kio.Writer{rw}}.Execute())
	}

	for _, rootPkgPath := range args {
		e := executeCmdOnPkgs{
			writer:             c.OutOrStdout(),
			needOpenAPI:        false,
			recurseSubPackages: r.RecurseSubPackages,
			cmdRunner:          r,
			rootPkgPath:        rootPkgPath,
		}

		err := e.execute()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *FmtRunner) executeCmd(w io.Writer, pkgPath string) error {
	openAPIFileName, err := ext.OpenAPIFileName()
	if err != nil {
		return err
	}

	rw := &kio.LocalPackageReadWriter{
		NoDeleteFiles:         true,
		PackagePath:           pkgPath,
		KeepReaderAnnotations: r.KeepAnnotations, PackageFileName: openAPIFileName}
	err = kio.Pipeline{
		Inputs: []kio.Reader{rw}, Filters: r.fmtFilters(), Outputs: []kio.Writer{rw}}.Execute()

	if err != nil {
		// return err if RecurseSubPackages is false
		if !r.RecurseSubPackages {
			return err
		} else {
			// print error message and continue if RecurseSubPackages is true
			fmt.Fprintf(w, "%s in package %q\n", err.Error(), pkgPath)
		}
	} else {
		fmt.Fprintf(w, "formatted resource files in package %q\n", pkgPath)
	}
	return nil
}

func (r *FmtRunner) fmtFilters() []kio.Filter {
	fmtFilters := []kio.Filter{filters.FormatFilter{
		UseSchema: r.UseSchema,
	}}

	// format with file names
	if r.SetFilenames {
		fmtFilters = append(fmtFilters, &filters.FileSetter{
			FilenamePattern: r.FilenamePattern,
			Override:        r.Override,
		})
	}
	return fmtFilters
}
